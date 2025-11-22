package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"reflect"
	"sync"
	"sync/atomic"
	"time"
)

// CacheConfig holds basic cache configuration
type CacheConfig struct {
	TTL             time.Duration // Time to live for cached entries (default: 1 hour)
	MaxEntries      int           // Maximum number of cached entries (default: 1000)
	CleanupInterval time.Duration // How often to run cleanup (default: TTL/2, 0 to disable)
}

// DefaultCacheConfig returns sensible defaults for in-memory caching
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:             time.Hour,
		MaxEntries:      1000,
		CleanupInterval: 30 * time.Minute, // Half of default TTL
	}
}

// cacheEntry represents a single cached item
type cacheEntry struct {
	value     interface{}
	timestamp time.Time
}

// CachedParser provides simple in-memory caching for parsing results
type CachedParser[T any] struct {
	cache       map[string]cacheEntry
	mu          sync.RWMutex
	config      *CacheConfig
	keyPrefix   string
	hits        uint64
	misses      uint64
	stopCleanup chan struct{}
}

// NewCachedParser creates a new cached parser with optional configuration.
// If CleanupInterval > 0, a background goroutine will periodically clean expired entries.
// Call Close() when done to stop the cleanup goroutine.
func NewCachedParser[T any](config *CacheConfig) *CachedParser[T] {
	if config == nil {
		config = DefaultCacheConfig()
	}

	var zero T
	keyPrefix := reflect.TypeOf(zero).String()

	cp := &CachedParser[T]{
		cache:       make(map[string]cacheEntry),
		config:      config,
		keyPrefix:   keyPrefix,
		stopCleanup: make(chan struct{}),
	}

	// Start cleanup goroutine if interval is configured
	if config.CleanupInterval > 0 {
		go cp.cleanupLoop()
	}

	return cp
}

// Parse parses data with caching support
func (cp *CachedParser[T]) Parse(data []byte) (T, error) {
	return cp.ParseWithFormat(data, DetectFormat(data))
}

// ParseWithFormat parses data with format specification and caching
func (cp *CachedParser[T]) ParseWithFormat(data []byte, format Format) (T, error) {
	key := cp.generateCacheKey(data, format)

	// Try cache first
	if cached, found := cp.get(key); found {
		return cached, nil
	}

	// Parse and cache
	result, err := ParseIntoWithFormat[T](data, format)
	if err != nil {
		var zero T
		return zero, err
	}

	cp.set(key, result)
	return result, nil
}

// get retrieves a value from cache with TTL check
func (cp *CachedParser[T]) get(key string) (T, bool) {
	cp.mu.RLock()
	entry, exists := cp.cache[key]
	cp.mu.RUnlock()

	if !exists {
		atomic.AddUint64(&cp.misses, 1)
		var zero T
		return zero, false
	}

	// Check TTL
	if time.Since(entry.timestamp) > cp.config.TTL {
		// Entry expired, clean up with write lock
		cp.mu.Lock()
		delete(cp.cache, key)
		cp.mu.Unlock()
		atomic.AddUint64(&cp.misses, 1)
		var zero T
		return zero, false
	}

	if result, ok := entry.value.(T); ok {
		atomic.AddUint64(&cp.hits, 1)
		return result, true
	}

	// Invalid type, clean up with write lock
	cp.mu.Lock()
	delete(cp.cache, key)
	cp.mu.Unlock()
	atomic.AddUint64(&cp.misses, 1)
	var zero T
	return zero, false
}

// set stores a value in cache with size limit enforcement
func (cp *CachedParser[T]) set(key string, value T) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	// Evict if at capacity
	if len(cp.cache) >= cp.config.MaxEntries {
		cp.evictOldest()
	}

	cp.cache[key] = cacheEntry{
		value:     value,
		timestamp: time.Now(),
	}
}

// evictOldest removes the oldest entry from cache
func (cp *CachedParser[T]) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range cp.cache {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(cp.cache, oldestKey)
	}
}

// generateCacheKey creates a unique cache key from content and format
// Uses FNV-1a for small inputs (<1KB) and SHA256 for larger inputs for optimal performance
func (cp *CachedParser[T]) generateCacheKey(data []byte, format Format) string {
	var contentHash string

	// Use faster FNV-1a hash for small inputs (typical case)
	if len(data) < 1024 {
		h := fnv.New64a()
		_, _ = h.Write(data) // hash.Hash.Write never returns an error
		contentHash = fmt.Sprintf("%x", h.Sum64())
	} else {
		// Use SHA256 for large inputs for better distribution
		hash := sha256.Sum256(data)
		contentHash = hex.EncodeToString(hash[:8])
	}

	return fmt.Sprintf("%s:%s:%v", contentHash, cp.keyPrefix, format)
}

// ClearCache removes all cached entries
func (cp *CachedParser[T]) ClearCache() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.cache = make(map[string]cacheEntry)
}

// Close stops the background cleanup goroutine if running.
// After calling Close, the parser can still be used but expired entries
// will only be cleaned up on access rather than proactively.
func (cp *CachedParser[T]) Close() {
	if cp.stopCleanup != nil {
		close(cp.stopCleanup)
	}
}

// cleanupLoop runs periodically to remove expired entries
func (cp *CachedParser[T]) cleanupLoop() {
	ticker := time.NewTicker(cp.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cp.cleanupExpired()
		case <-cp.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes all expired entries from the cache
func (cp *CachedParser[T]) cleanupExpired() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	now := time.Now()
	for key, entry := range cp.cache {
		if now.Sub(entry.timestamp) > cp.config.TTL {
			delete(cp.cache, key)
		}
	}
}

// Stats returns cache statistics including size, max size, and hit rate
func (cp *CachedParser[T]) Stats() (size, maxSize int, hitRate float64) {
	cp.mu.RLock()
	size = len(cp.cache)
	cp.mu.RUnlock()

	hits := atomic.LoadUint64(&cp.hits)
	misses := atomic.LoadUint64(&cp.misses)
	total := hits + misses

	if total == 0 {
		return size, cp.config.MaxEntries, 0.0
	}

	hitRate = float64(hits) / float64(total)
	return size, cp.config.MaxEntries, hitRate
}

// ParseIntoCached provides convenient cached parsing (falls back to non-cached for simplicity)
func ParseIntoCached[T any](data []byte) (T, error) {
	return ParseInto[T](data)
}

// ParseIntoWithFormatCached provides cached parsing with format specification (falls back to non-cached for simplicity)
func ParseIntoWithFormatCached[T any](data []byte, format Format) (T, error) {
	return ParseIntoWithFormat[T](data, format)
}
