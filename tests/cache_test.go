package tests

import (
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Test structures
type CachedUser struct {
	ID   int    `json:"id" yaml:"id" validate:"required,min=1"`
	Name string `json:"name" yaml:"name" validate:"required,min=2"`
	Age  int    `json:"age" yaml:"age" validate:"min=18,max=120"`
}

func TestCachedParser_Basic(t *testing.T) {
	config := &model.CacheConfig{
		TTL:        time.Minute,
		MaxEntries: 100,
	}

	parser, err := model.NewCachedParser[CachedUser](config)
	if err != nil {
		t.Fatalf("Failed to create cached parser: %v", err)
	}
	defer parser.Close()

	jsonData := []byte(`{
		"id": 123,
		"name": "John Doe",
		"age": 30
	}`)

	// First parse - should cache the result
	result1, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("First parse failed: %v", err)
	}

	if result1.ID != 123 || result1.Name != "John Doe" || result1.Age != 30 {
		t.Errorf("First parse result incorrect: %+v", result1)
	}

	// Second parse - should use cached result
	result2, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	if result2.ID != 123 || result2.Name != "John Doe" || result2.Age != 30 {
		t.Errorf("Second parse result incorrect: %+v", result2)
	}

	// Check cache stats
	stats := parser.Stats()
	if stats.Hits() == 0 {
		t.Error("Expected cache hit, but got 0 hits")
	}
}

func TestCachedParser_YAML(t *testing.T) {
	parser, err := model.NewCachedParser[CachedUser](model.DefaultCacheConfig())
	if err != nil {
		t.Fatalf("Failed to create cached parser: %v", err)
	}
	defer parser.Close()

	yamlData := []byte(`
id: 456
name: "Jane Smith"
age: 25
`)

	result, err := parser.ParseWithFormat(yamlData, model.FormatYAML)
	if err != nil {
		t.Fatalf("YAML parse failed: %v", err)
	}

	if result.ID != 456 || result.Name != "Jane Smith" || result.Age != 25 {
		t.Errorf("YAML parse result incorrect: %+v", result)
	}
}

func TestGlobalCachedFunctions(t *testing.T) {
	// Clear any existing cache
	model.ClearAllCaches()

	jsonData := []byte(`{
		"id": 789,
		"name": "Bob Wilson",
		"age": 35
	}`)

	// First call - should cache the result
	result1, err := model.ParseIntoCached[CachedUser](jsonData)
	if err != nil {
		t.Fatalf("First cached parse failed: %v", err)
	}

	if result1.ID != 789 || result1.Name != "Bob Wilson" || result1.Age != 35 {
		t.Errorf("First cached parse result incorrect: %+v", result1)
	}

	// Second call - should use cached result
	result2, err := model.ParseIntoCached[CachedUser](jsonData)
	if err != nil {
		t.Fatalf("Second cached parse failed: %v", err)
	}

	if result2.ID != 789 || result2.Name != "Bob Wilson" || result2.Age != 35 {
		t.Errorf("Second cached parse result incorrect: %+v", result2)
	}

	// Get global cache stats
	stats := model.GetGlobalCacheStats()
	userTypeKey := "tests.CachedUser"

	if userStats, exists := stats[userTypeKey]; exists {
		if userStats.Hits() == 0 {
			t.Error("Expected cache hit in global stats, but got 0 hits")
		}
	} else {
		t.Errorf("No cache stats found for user type: %s", userTypeKey)
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	parser1, err := model.NewCachedParser[CachedUser](model.DefaultCacheConfig())
	if err != nil {
		t.Fatalf("Failed to create first cached parser: %v", err)
	}
	defer parser1.Close()

	parser2, err := model.NewCachedParser[CachedUser](model.DefaultCacheConfig())
	if err != nil {
		t.Fatalf("Failed to create second cached parser: %v", err)
	}
	defer parser2.Close()

	// Same data should produce same cache keys (though we can't directly test this)
	// But different parsers should be independent
	jsonData := []byte(`{
		"id": 999,
		"name": "Cache Test",
		"age": 40
	}`)

	result1, err := parser1.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parser1 failed: %v", err)
	}

	result2, err := parser2.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parser2 failed: %v", err)
	}

	// Results should be identical
	if result1.ID != result2.ID || result1.Name != result2.Name || result1.Age != result2.Age {
		t.Error("Results from different parsers should be identical")
	}
}

func TestCacheWithValidationErrors(t *testing.T) {
	parser, err := model.NewCachedParser[CachedUser](model.DefaultCacheConfig())
	if err != nil {
		t.Fatalf("Failed to create cached parser: %v", err)
	}
	defer parser.Close()

	// Invalid data - should not be cached on errors
	invalidData := []byte(`{
		"id": 0,
		"name": "A",
		"age": 15
	}`)

	_, err1 := parser.Parse(invalidData)
	if err1 == nil {
		t.Error("Expected validation error for invalid data")
	}

	_, err2 := parser.Parse(invalidData)
	if err2 == nil {
		t.Error("Expected validation error for invalid data on second try")
	}

	// Errors should not be cached, so cache hits should be 0
	stats := parser.Stats()
	if stats.Hits() > 0 {
		t.Error("Validation errors should not be cached")
	}
}

// Redis Configuration Validation Tests

func TestCachedParser_RedisConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      *model.CacheConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid Redis config with address",
			config: &model.CacheConfig{
				Backend: model.CacheBackendRedis,
				RedisConfig: &model.RedisConfig{
					Addr: "localhost:6379",
				},
			},
			expectError: false,
		},
		// Note: Pre-configured client test would require a full Redis mock
		// which is complex. This is tested in integration tests instead.
		{
			name: "Redis backend without RedisConfig",
			config: &model.CacheConfig{
				Backend: model.CacheBackendRedis,
			},
			expectError: true,
			errorMsg:    "Redis backend requires RedisConfig to be set",
		},
		{
			name: "Redis config without address and without client",
			config: &model.CacheConfig{
				Backend: model.CacheBackendRedis,
				RedisConfig: &model.RedisConfig{
					Password: "secret",
					DB:       1,
				},
			},
			expectError: true,
			errorMsg:    "Redis address is required when Client is not provided",
		},
		{
			name: "unsupported backend type",
			config: &model.CacheConfig{
				Backend: "unsupported",
			},
			expectError: true,
			errorMsg:    "unsupported cache backend: unsupported",
		},
		{
			name: "memory backend (backward compatibility)",
			config: &model.CacheConfig{
				Backend:    model.CacheBackendMemory,
				MaxEntries: 100,
			},
			expectError: false,
		},
		{
			name: "empty backend defaults to memory",
			config: &model.CacheConfig{
				MaxEntries: 100,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := model.NewCachedParser[CachedUser](tt.config)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				} else {
					parser.Close() // Clean up
				}
			}
		})
	}
}

func TestCachedParser_BackwardCompatibility(t *testing.T) {
	// Test that existing code patterns continue to work
	tests := []struct {
		name   string
		config *model.CacheConfig
	}{
		{
			name: "old style config without Backend field",
			config: &model.CacheConfig{
				TTL:        time.Minute,
				MaxEntries: 100,
				Namespace:  "test",
			},
		},
		{
			name:   "nil config defaults to memory",
			config: nil,
		},
		{
			name:   "default cache config",
			config: model.DefaultCacheConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser, err := model.NewCachedParser[CachedUser](tt.config)
			if err != nil {
				t.Fatalf("Backward compatibility broken: %v", err)
			}
			defer parser.Close()

			// Should work with memory backend (default)
			jsonData := []byte(`{"id": 1, "name": "Test User", "age": 25}`)
			result, err := parser.Parse(jsonData)
			if err != nil {
				t.Fatalf("Backward compatibility parse failed: %v", err)
			}

			if result.ID != 1 || result.Name != "Test User" {
				t.Errorf("Unexpected result: %+v", result)
			}
		})
	}
}

func TestDefaultRedisCacheConfig(t *testing.T) {
	addr := "localhost:6379"
	config := model.DefaultRedisCacheConfig(addr)

	// Verify config structure
	if config.Backend != model.CacheBackendRedis {
		t.Errorf("Expected Redis backend, got %s", config.Backend)
	}

	if config.RedisConfig == nil {
		t.Fatal("RedisConfig should not be nil")
	}

	if config.RedisConfig.Addr != addr {
		t.Errorf("Expected addr %s, got %s", addr, config.RedisConfig.Addr)
	}

	if config.RedisConfig.DB != 0 {
		t.Errorf("Expected default DB 0, got %d", config.RedisConfig.DB)
	}

	if config.RedisConfig.KeyPrefix != "gopantic:" {
		t.Errorf("Expected key prefix 'gopantic:', got %s", config.RedisConfig.KeyPrefix)
	}

	if config.TTL != time.Hour {
		t.Errorf("Expected default TTL 1 hour, got %v", config.TTL)
	}
}

// Redis Integration Tests (require actual Redis instance)

func TestCachedParser_Redis_Integration(t *testing.T) {
	// Skip if in short test mode or CI environment without Redis
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	config := model.DefaultRedisCacheConfig("localhost:6379")
	config.TTL = time.Minute

	parser, err := model.NewCachedParser[CachedUser](config)
	if err != nil {
		t.Skipf("Redis not available, skipping integration test: %v", err)
	}
	defer parser.Close()

	// Test basic Redis caching functionality
	jsonData := []byte(`{
		"id": 123,
		"name": "Redis User",
		"age": 30
	}`)

	// First parse - should be a cache miss and store in Redis
	result1, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("Redis parse failed: %v", err)
	}

	if result1.ID != 123 || result1.Name != "Redis User" || result1.Age != 30 {
		t.Errorf("Unexpected parse result: %+v", result1)
	}

	// Second parse - should be a cache hit from Redis
	result2, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("Redis cached parse failed: %v", err)
	}

	if result1 != result2 {
		t.Error("Expected identical results from Redis cache")
	}

	// Verify cache statistics show activity
	stats := parser.Stats()
	if stats.KeyCount() == 0 {
		t.Error("Expected Redis cache to contain keys")
	}
	if stats.Total() < 2 {
		t.Error("Expected at least 2 cache operations (miss + hit)")
	}
	if stats.Hits() == 0 {
		t.Error("Expected at least one cache hit")
	}
}

func TestCachedParser_Redis_Advanced_Config(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Test advanced Redis configuration
	config := &model.CacheConfig{
		TTL:                30 * time.Second,
		CompressionEnabled: true,
		Namespace:          "test:advanced",
		Backend:            model.CacheBackendRedis,
		RedisConfig: &model.RedisConfig{
			Addr:      "localhost:6379",
			Password:  "", // No password for test Redis
			DB:        1,  // Use database 1 for testing
			KeyPrefix: "integration-test:",
		},
	}

	parser, err := model.NewCachedParser[CachedUser](config)
	if err != nil {
		t.Skipf("Redis not available for advanced config test: %v", err)
	}
	defer parser.Close()

	// Test that it works with custom configuration
	jsonData := []byte(`{"id": 456, "name": "Advanced User", "age": 35}`)

	result, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("Advanced Redis config parse failed: %v", err)
	}

	if result.ID != 456 {
		t.Errorf("Expected ID 456, got %d", result.ID)
	}

	// Test cache hit
	result2, err := parser.Parse(jsonData)
	if err != nil {
		t.Fatalf("Advanced Redis config cached parse failed: %v", err)
	}

	if result != result2 {
		t.Error("Cache hit should return identical result")
	}
}

func TestCachedParser_Redis_Namespace_Isolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Create two parsers with different namespaces
	config1 := model.DefaultRedisCacheConfig("localhost:6379")
	config1.Namespace = "test:namespace1"
	config1.TTL = time.Minute

	config2 := model.DefaultRedisCacheConfig("localhost:6379")
	config2.Namespace = "test:namespace2"
	config2.TTL = time.Minute

	parser1, err := model.NewCachedParser[CachedUser](config1)
	if err != nil {
		t.Skipf("Redis not available for namespace test: %v", err)
	}
	defer parser1.Close()

	parser2, err := model.NewCachedParser[CachedUser](config2)
	if err != nil {
		t.Skipf("Redis not available for namespace test: %v", err)
	}
	defer parser2.Close()

	jsonData := []byte(`{"id": 789, "name": "Namespace User", "age": 40}`)

	// Parse with first parser (cache miss)
	result1, err := parser1.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parser1 parse failed: %v", err)
	}

	// Parse with second parser - should also be cache miss due to different namespace
	result2, err := parser2.Parse(jsonData)
	if err != nil {
		t.Fatalf("Parser2 parse failed: %v", err)
	}

	// Results should be the same content but from separate cache entries
	if result1.ID != result2.ID || result1.Name != result2.Name {
		t.Error("Results should have same content")
	}

	// Both should show cache operations (separate namespaces)
	stats1 := parser1.Stats()
	stats2 := parser2.Stats()

	if stats1.KeyCount() == 0 || stats2.KeyCount() == 0 {
		t.Error("Both parsers should have keys in their respective caches")
	}
	if stats1.Total() == 0 || stats2.Total() == 0 {
		t.Error("Both parsers should have performed cache operations")
	}
}
