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
