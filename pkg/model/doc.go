// Package model provides high-performance JSON/YAML parsing with type coercion and validation for Go.
//
// gopantic is inspired by Python's Pydantic but designed specifically for Go's type system.
// It provides automatic format detection, comprehensive type coercion, struct tag-based validation,
// and optional high-performance caching.
//
// # Quick Start
//
// Parse JSON/YAML data into typed structs with automatic type coercion:
//
//	type User struct {
//	    ID    int    `json:"id" validate:"required,min=1"`
//	    Name  string `json:"name" validate:"required,min=2"`
//	    Email string `json:"email" validate:"required,email"`
//	}
//
//	user, err := model.ParseInto[User](jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Type Coercion
//
// Automatic conversion between compatible types:
//   - String to numeric: "42" → 42, "3.14" → 3.14
//   - String to boolean: "true", "yes", "1" → true
//   - Numeric to boolean: 0 → false, non-zero → true
//   - Unix timestamps: 1673781000 → time.Time
//   - RFC3339 strings: "2023-01-15T10:30:00Z" → time.Time
//
// # Validation
//
// Built-in validators using struct tags:
//   - required: Non-zero value required
//   - min/max: Numeric bounds or string/slice length
//   - length: Exact length requirement
//   - email: Valid email format
//   - alpha: Alphabetic characters only
//   - alphanum: Alphanumeric characters only
//
// Multiple rules can be combined:
//
//	`validate:"required,min=3,max=20,alphanum"`
//
// # Format Support
//
// Automatic detection of JSON and YAML formats:
//
//	// Works with both JSON and YAML automatically
//	config, err := model.ParseInto[Config](data)
//
//	// Or specify format explicitly for performance
//	config, err := model.ParseIntoWithFormat[Config](data, model.FormatYAML)
//
// # High-Performance Caching
//
// Optional caching provides 5-27x speedup for repeated parsing:
//
//	parser := model.NewCachedParser[User](nil)
//	defer parser.Close()
//
//	user1, _ := parser.Parse(data) // Cache miss
//	user2, _ := parser.Parse(data) // Cache hit - 27x faster
//
// # Error Handling
//
// Comprehensive error reporting with field paths and aggregation:
//
//	user, err := model.ParseInto[User](invalidData)
//	if err != nil {
//	    // Multiple errors are aggregated:
//	    // "multiple errors: validation error on field 'Email': invalid email format; ..."
//	    log.Printf("Parse failed: %v", err)
//	}
//
// Specific error types for programmatic handling:
//
//	if parseErr, ok := err.(*model.ParseError); ok {
//	    log.Printf("Parse error in field %s: %s", parseErr.Field, parseErr.Message)
//	}
//	if validationErr, ok := err.(*model.ValidationError); ok {
//	    log.Printf("Validation error in field %s: %s", validationErr.Field, validationErr.Message)
//	}
//
// # Thread Safety
//
// All parsing functions are thread-safe and can be called concurrently.
// CachedParser instances use RWMutex for safe concurrent access.
//
// # Performance
//
// gopantic is optimized for practical use cases:
//   - ~5x slower than standard JSON for parsing (adds validation + coercion)
//   - 5-27x faster than uncached parsing with caching enabled
//   - Minimal memory allocations and GC pressure
//   - Thread-safe concurrent parsing
//
// See the examples directory for complete usage patterns and the docs directory
// for detailed architecture and API documentation.
package model
