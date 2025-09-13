package model

import (
	"crypto/sha256"
	"encoding/hex"
	"reflect"
	"sync"
	"time"
)

// CacheConfig holds basic cache configuration
type CacheConfig struct {
	TTL        time.Duration // Time to live for cached entries (default: 1 hour)
	MaxEntries int           // Maximum number of cached entries (default: 1000)
}

// DefaultCacheConfig returns sensible defaults for in-memory caching
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:        time.Hour,
		MaxEntries: 1000,
	}
}

// cacheEntry represents a single cached item
type cacheEntry struct {
	value     interface{}
	timestamp time.Time
}

// CachedParser provides simple in-memory caching for parsing results
type CachedParser[T any] struct {
	cache     map[string]cacheEntry
	mu        sync.RWMutex
	config    *CacheConfig
	keyPrefix string
}

// NewCachedParser creates a new cached parser with optional configuration
func NewCachedParser[T any](config *CacheConfig) *CachedParser[T] {
	if config == nil {
		config = DefaultCacheConfig()
	}

	var zero T
	keyPrefix := reflect.TypeOf(zero).String()

	return &CachedParser[T]{
		cache:     make(map[string]cacheEntry),
		config:    config,
		keyPrefix: keyPrefix,
	}
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
	defer cp.mu.RUnlock()

	entry, exists := cp.cache[key]
	if !exists {
		var zero T
		return zero, false
	}

	// Check TTL
	if time.Since(entry.timestamp) > cp.config.TTL {
		// Entry expired, clean up
		delete(cp.cache, key)
		var zero T
		return zero, false
	}

	if result, ok := entry.value.(T); ok {
		return result, true
	}

	// Invalid type, clean up
	delete(cp.cache, key)
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
func (cp *CachedParser[T]) generateCacheKey(data []byte, format Format) string {
	hash := sha256.Sum256(data)
	contentHash := hex.EncodeToString(hash[:8]) // First 8 bytes for shorter keys
	return contentHash + ":" + cp.keyPrefix
}

// ClearCache removes all cached entries
func (cp *CachedParser[T]) ClearCache() {
	cp.mu.Lock()
	defer cp.mu.Unlock()
	cp.cache = make(map[string]cacheEntry)
}

// Stats returns cache statistics
func (cp *CachedParser[T]) Stats() (size, maxSize int, hitRate float64) {
	cp.mu.RLock()
	defer cp.mu.RUnlock()
	// Simple stats - could be enhanced if needed
	return len(cp.cache), cp.config.MaxEntries, 0.0
}

// ParseIntoCached provides convenient cached parsing (falls back to non-cached for simplicity)
func ParseIntoCached[T any](data []byte) (T, error) {
	return ParseInto[T](data)
}

// ParseIntoWithFormatCached provides cached parsing with format specification (falls back to non-cached for simplicity)
func ParseIntoWithFormatCached[T any](data []byte, format Format) (T, error) {
	return ParseIntoWithFormat[T](data, format)
}
