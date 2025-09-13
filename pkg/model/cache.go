package model

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// CacheBackend defines the type of cache backend
type CacheBackend string

const (
	// CacheBackendMemory uses in-memory caching (default)
	CacheBackendMemory CacheBackend = "memory"
	// CacheBackendRedis uses Redis for distributed caching
	CacheBackendRedis CacheBackend = "redis"
)

// RedisConfig holds Redis-specific configuration options for distributed caching.
// It supports both simple address-based configuration and advanced pre-configured client usage.
//
// Simple configuration example:
//
//	redisConfig := &RedisConfig{
//	    Addr:      "localhost:6379",
//	    Password:  "mypassword",
//	    DB:        1,
//	    KeyPrefix: "myapp:",
//	}
//
// Advanced configuration with pre-configured client:
//
//	client := redis.NewClient(&redis.Options{
//	    Addr:     "localhost:6379",
//	    Password: "mypassword",
//	    DB:       1,
//	})
//	redisConfig := &RedisConfig{
//	    Client:    client,
//	    KeyPrefix: "myapp:",
//	}
type RedisConfig struct {
	// Addr is the Redis server address (e.g., "localhost:6379").
	// Required if Client is not provided.
	Addr string
	// Password is the Redis authentication password.
	// Optional, leave empty if Redis has no authentication.
	Password string
	// DB is the Redis database number (0-15).
	// Default is 0. Different databases provide isolation between applications.
	DB int
	// KeyPrefix is an additional prefix for all cache keys.
	// Helps avoid key collisions when multiple applications use the same Redis instance.
	// Default is "gopantic:" when using DefaultRedisCacheConfig.
	KeyPrefix string
	// Client is a pre-configured Redis client (optional).
	// If provided, Addr, Password, and DB settings are ignored.
	// Use this for advanced Redis configurations like clustering, sentinel, etc.
	Client redis.Cmdable
}

// CacheConfig holds configuration options for the cached parser.
// Supports both in-memory and Redis distributed caching backends.
//
// Memory backend example (default):
//
//	config := &CacheConfig{
//	    TTL:        time.Hour,
//	    MaxEntries: 1000,
//	    Backend:    CacheBackendMemory, // or leave empty
//	}
//
// Redis backend example:
//
//	config := &CacheConfig{
//	    TTL:     time.Hour,
//	    Backend: CacheBackendRedis,
//	    RedisConfig: &RedisConfig{
//	        Addr: "localhost:6379",
//	        DB:   0,
//	    },
//	}
//
// Or use the convenience function:
//
//	config := DefaultRedisCacheConfig("localhost:6379")
type CacheConfig struct {
	// TTL is the time-to-live for cached entries.
	// Applies to both memory and Redis backends.
	TTL time.Duration
	// MaxEntries is the maximum number of entries to keep in cache.
	// Only used for memory backend; Redis uses its own memory management.
	MaxEntries int
	// CompressionEnabled enables compression for cached values.
	// Reduces memory usage and network transfer for Redis backend.
	CompressionEnabled bool
	// Namespace is a prefix for cache keys to avoid collisions.
	// Especially important for Redis to isolate different applications/environments.
	Namespace string
	// Backend specifies the cache backend type.
	// Use CacheBackendMemory for single-instance in-memory caching (default).
	// Use CacheBackendRedis for distributed caching across multiple instances.
	Backend CacheBackend
	// RedisConfig holds Redis-specific configuration.
	// Required when Backend is CacheBackendRedis, ignored otherwise.
	RedisConfig *RedisConfig
}

// DefaultCacheConfig returns a reasonable default configuration for in-memory caching.
// Uses 1-hour TTL, 1000 entry limit, compression enabled, and memory backend.
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:                time.Hour,          // 1 hour default TTL
		MaxEntries:         1000,               // 1000 entries max
		CompressionEnabled: true,               // Enable compression by default
		Namespace:          "gopantic:parsing", // Default namespace
		Backend:            CacheBackendMemory, // Default to memory backend
	}
}

// DefaultRedisCacheConfig returns a default Redis cache configuration for distributed caching.
// This enables high-performance distributed caching across multiple application instances.
//
// Example usage:
//
//	config := model.DefaultRedisCacheConfig("localhost:6379")
//	parser, err := model.NewCachedParser[User](config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer parser.Close()
//
//	result, err := parser.Parse(jsonData)
func DefaultRedisCacheConfig(addr string) *CacheConfig {
	return &CacheConfig{
		TTL:                time.Hour,          // 1 hour default TTL
		CompressionEnabled: true,               // Enable compression by default
		Namespace:          "gopantic:parsing", // Default namespace
		Backend:            CacheBackendRedis,  // Redis backend
		RedisConfig: &RedisConfig{
			Addr:      addr,        // Redis server address
			DB:        0,           // Default database
			KeyPrefix: "gopantic:", // Redis key prefix
		},
	}
}

// CachedParser provides high-performance caching functionality for parsing operations.
// It supports both in-memory and Redis distributed caching backends, automatically
// handling cache key generation, TTL management, and graceful degradation.
type CachedParser[T any] struct {
	cache  *obcache.Cache
	config *CacheConfig
}

// NewCachedParser creates a new cached parser with the given configuration.
// Supports both in-memory and Redis distributed caching backends.
//
// Memory backend usage:
//
//	parser, err := model.NewCachedParser[User](model.DefaultCacheConfig())
//
// Redis backend usage:
//
//	config := model.DefaultRedisCacheConfig("localhost:6379")
//	parser, err := model.NewCachedParser[User](config)
//
// Advanced Redis configuration:
//
//	config := &model.CacheConfig{
//	    Backend: model.CacheBackendRedis,
//	    TTL:     30 * time.Minute,
//	    RedisConfig: &model.RedisConfig{
//	        Addr:      "redis.example.com:6379",
//	        Password:  "secret",
//	        DB:        1,
//	        KeyPrefix: "myapp:",
//	    },
//	}
//	parser, err := model.NewCachedParser[User](config)
//
// The parser will gracefully handle Redis connection failures by returning an error.
// Always defer parser.Close() to properly clean up resources.
func NewCachedParser[T any](config *CacheConfig) (*CachedParser[T], error) {
	if config == nil {
		config = DefaultCacheConfig()
	}

	var cacheConfig *obcache.Config
	var err error

	// Configure cache backend based on config
	switch config.Backend {
	case CacheBackendMemory, "": // Default to memory for backward compatibility
		cacheConfig = obcache.NewDefaultConfig()
		cacheConfig = cacheConfig.WithMaxEntries(config.MaxEntries).WithDefaultTTL(config.TTL)

	case CacheBackendRedis:
		// Validate Redis configuration
		if config.RedisConfig == nil {
			return nil, errors.New("Redis backend requires RedisConfig to be set")
		}

		if config.RedisConfig.Client != nil {
			// Use pre-configured client
			cacheConfig = obcache.NewRedisConfigWithClient(config.RedisConfig.Client)
		} else {
			// Create new Redis connection
			if config.RedisConfig.Addr == "" {
				return nil, errors.New("Redis address is required when Client is not provided")
			}
			cacheConfig = obcache.NewRedisConfig(config.RedisConfig.Addr)
		}

		// Apply Redis-specific settings
		redisConfig := &obcache.RedisConfig{
			Addr:      config.RedisConfig.Addr,
			Password:  config.RedisConfig.Password,
			DB:        config.RedisConfig.DB,
			KeyPrefix: config.RedisConfig.KeyPrefix,
		}
		if config.RedisConfig.Client != nil {
			redisConfig.Client = config.RedisConfig.Client
		}

		cacheConfig = cacheConfig.WithRedis(redisConfig).WithDefaultTTL(config.TTL)

	default:
		return nil, fmt.Errorf("unsupported cache backend: %s", config.Backend)
	}

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

// ParseIntoCached provides convenient cached parsing with automatic format detection.
// It uses a global cache instance per type T for maximum convenience and performance.
// This is ideal for applications that parse the same types frequently.
//
// Example:
//
//	user, err := model.ParseIntoCached[User](jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseIntoCached[T any](raw []byte) (T, error) {
	return ParseIntoWithFormatCached[T](raw, DetectFormat(raw))
}

// ParseIntoWithFormatCached provides cached parsing with explicit format specification.
// Uses a global cache instance per type T for optimal performance in repeated parsing scenarios.
//
// Example:
//
//	config, err := model.ParseIntoWithFormatCached[Config](yamlData, model.FormatYAML)
//	if err != nil {
//	    log.Fatal(err)
//	}
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

// ClearAllCaches clears all global cached parsers.
// This is useful for testing or when you need to reset all cached parsing results.
// Note: This affects all cached parsing operations across the application.
func ClearAllCaches() {
	for _, cache := range defaultCachedParsers {
		_ = cache.Clear()
	}
}

// GetGlobalCacheStats returns performance statistics for all global caches.
// The returned map contains cache statistics keyed by type name.
// Use this to monitor cache effectiveness and optimize cache configurations.
func GetGlobalCacheStats() map[string]*obcache.Stats {
	stats := make(map[string]*obcache.Stats)
	for typeKey, cache := range defaultCachedParsers {
		stats[typeKey] = cache.Stats()
	}
	return stats
}
