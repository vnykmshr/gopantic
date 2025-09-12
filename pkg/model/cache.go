package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"time"

	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// CacheConfig holds configuration options for the cached parser
type CacheConfig struct {
	// TTL is the time-to-live for cached entries
	TTL time.Duration
	// MaxEntries is the maximum number of entries to keep in cache
	MaxEntries int
	// CompressionEnabled enables compression for cached values
	CompressionEnabled bool
	// Namespace is a prefix for cache keys to avoid collisions
	Namespace string
}

// DefaultCacheConfig returns a reasonable default configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:                time.Hour,          // 1 hour default TTL
		MaxEntries:         1000,               // 1000 entries max
		CompressionEnabled: true,               // Enable compression by default
		Namespace:          "gopantic:parsing", // Default namespace
	}
}

// CachedParser provides caching functionality for parsing operations
type CachedParser[T any] struct {
	cache  *obcache.Cache
	config *CacheConfig
}

// NewCachedParser creates a new cached parser with the given configuration
func NewCachedParser[T any](config *CacheConfig) (*CachedParser[T], error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	// Create cache with default config first
	cacheConfig := obcache.NewDefaultConfig()

	// Apply our custom settings
	cacheConfig = cacheConfig.WithMaxEntries(config.MaxEntries).WithDefaultTTL(config.TTL)

	cache, err := obcache.New(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	return &CachedParser[T]{
		cache:  cache,
		config: config,
	}, nil
}

// Parse parses raw data with caching support
// It first checks the cache, and if not found, parses and caches the result
func (cp *CachedParser[T]) Parse(raw []byte) (T, error) {
	return cp.ParseWithFormat(raw, DetectFormat(raw))
}

// ParseWithFormat parses raw data of a specific format with caching support
func (cp *CachedParser[T]) ParseWithFormat(raw []byte, format Format) (T, error) {
	var zero T

	// Generate cache key based on content hash and type
	cacheKey := cp.generateCacheKey(raw, format, reflect.TypeOf(zero))

	// Try to get from cache first
	if cached, found := cp.cache.Get(cacheKey); found {
		if result, ok := cached.(T); ok {
			return result, nil
		}
		// If type assertion fails, remove invalid cache entry
		_ = cp.cache.Delete(cacheKey)
	}

	// Not in cache or invalid, parse normally
	result, err := ParseIntoWithFormat[T](raw, format)
	if err != nil {
		return zero, err
	}

	// Store in cache for future use
	_ = cp.cache.Set(cacheKey, result, cp.config.TTL)

	return result, nil
}

// generateCacheKey creates a unique cache key based on content, format, and target type
func (cp *CachedParser[T]) generateCacheKey(raw []byte, format Format, targetType reflect.Type) string {
	// Create hash of the content
	hasher := sha256.New()
	hasher.Write(raw)
	contentHash := hex.EncodeToString(hasher.Sum(nil))

	// Include format and type information in the key
	typeStr := targetType.String()
	formatStr := fmt.Sprintf("%d", int(format))

	// Combine into a unique cache key
	return fmt.Sprintf("%s:%s:%s:%s", cp.config.Namespace, contentHash, formatStr, typeStr)
}

// ClearCache clears all cached entries
func (cp *CachedParser[T]) ClearCache() {
	_ = cp.cache.Clear()
}

// Stats returns cache statistics
func (cp *CachedParser[T]) Stats() *obcache.Stats {
	return cp.cache.Stats()
}

// Close closes the cache and releases resources
func (cp *CachedParser[T]) Close() error {
	return cp.cache.Close()
}

// Global cached parsers for convenience
var (
	defaultCachedParsers = make(map[string]*obcache.Cache)
	defaultCacheConfig   = DefaultCacheConfig()
)

// ParseIntoCached provides a convenient cached parsing function
// It uses a global cache instance per type T
func ParseIntoCached[T any](raw []byte) (T, error) {
	return ParseIntoWithFormatCached[T](raw, DetectFormat(raw))
}

// ParseIntoWithFormatCached provides cached parsing with explicit format
func ParseIntoWithFormatCached[T any](raw []byte, format Format) (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)
	typeKey := targetType.String()

	// Get or create cache for this type
	cache, exists := defaultCachedParsers[typeKey]
	if !exists {
		cacheConfig := obcache.NewDefaultConfig().
			WithMaxEntries(defaultCacheConfig.MaxEntries).
			WithDefaultTTL(defaultCacheConfig.TTL)

		newCache, err := obcache.New(cacheConfig)
		if err != nil {
			// Fallback to basic parsing if cache creation fails
			return ParseIntoWithFormat[T](raw, format)
		}
		cache = newCache
		defaultCachedParsers[typeKey] = cache
	}

	// Generate cache key
	cacheKey := generateGlobalCacheKey(raw, format)

	// Try cache first
	if cached, found := cache.Get(cacheKey); found {
		if result, ok := cached.(T); ok {
			return result, nil
		}
		// Invalid cache entry, remove it
		_ = cache.Delete(cacheKey)
	}

	// Parse and cache
	result, err := ParseIntoWithFormat[T](raw, format)
	if err != nil {
		return zero, err
	}

	_ = cache.Set(cacheKey, result, defaultCacheConfig.TTL)
	return result, nil
}

// generateGlobalCacheKey creates a cache key for global caching functions
func generateGlobalCacheKey(raw []byte, format Format) string {
	hasher := sha256.New()
	hasher.Write(raw)
	contentHash := hex.EncodeToString(hasher.Sum(nil))
	formatStr := fmt.Sprintf("%d", int(format))
	return fmt.Sprintf("%s:%s", contentHash, formatStr)
}

// ClearAllCaches clears all global cached parsers
func ClearAllCaches() {
	for _, cache := range defaultCachedParsers {
		_ = cache.Clear()
	}
}

// GetGlobalCacheStats returns statistics for all global caches
func GetGlobalCacheStats() map[string]*obcache.Stats {
	stats := make(map[string]*obcache.Stats)
	for typeKey, cache := range defaultCachedParsers {
		stats[typeKey] = cache.Stats()
	}
	return stats
}
