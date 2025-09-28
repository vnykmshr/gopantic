# API Reference

Complete API documentation for gopantic.

## Core Functions

### ParseInto

```go
func ParseInto[T any](data []byte) (T, error)
```

Parses JSON or YAML data into a struct with automatic format detection.

Parameters:
- `data []byte` - Raw JSON or YAML data

Returns:
- `T` - Parsed struct of type T
- `error` - Parse, coercion, or validation errors

Example:
```go
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

user, err := model.ParseInto[User]([]byte(`{"id": 1, "name": "Alice"}`))
```

### ParseIntoWithFormat

```go
func ParseIntoWithFormat[T any](data []byte, format Format) (T, error)
```

Parses data with explicit format specification.

Parameters:
- `data []byte` - Raw data
- `format Format` - `FormatJSON` or `FormatYAML`

Returns:
- `T` - Parsed struct of type T
- `error` - Parse, coercion, or validation errors

Example:
```go
user, err := model.ParseIntoWithFormat[User](yamlData, model.FormatYAML)
```

## Format Detection

### DetectFormat

```go
func DetectFormat(data []byte) Format
```

Automatically detects data format.

Parameters:
- `data []byte` - Raw data

Returns:
- `Format` - Detected format (`FormatJSON` or `FormatYAML`)

Algorithm:
- Looks for JSON markers (`{`, `[`)
- Looks for YAML markers (`---`, `:`)
- Defaults to JSON for ambiguous cases

## Caching

### NewCachedParser

```go
func NewCachedParser[T any](config *CacheConfig) (*CachedParser[T], error)
```

Creates a new cached parser instance.

Parameters:
- `config *CacheConfig` - Cache configuration (nil for defaults)

Returns:
- `*CachedParser[T]` - Parser instance with caching
- `error` - Configuration validation errors

Example:
```go
config := &model.CacheConfig{
    TTL:        5 * time.Minute,
    MaxEntries: 1000,
    Namespace:  "users",
}
parser, err := model.NewCachedParser[User](config)
defer parser.Close()
```

### CachedParser.Parse

```go
func (cp *CachedParser[T]) Parse(data []byte) (T, error)
```

Parses data with caching support.

Parameters:
- `data []byte` - Raw data

Returns:
- `T` - Parsed struct (from cache or fresh parse)
- `error` - Parse, coercion, or validation errors

### CachedParser.Stats

```go
func (cp *CachedParser[T]) Stats() (size, maxSize int, hitRate float64)
```

Returns cache statistics.

Returns:
- `size int` - Current cache entries
- `maxSize int` - Maximum cache entries
- `hitRate float64` - Cache hit rate percentage

### CachedParser.ClearCache

```go
func (cp *CachedParser[T]) ClearCache()
```

Clears all cache entries.

### ParseIntoCached

```go
func ParseIntoCached[T any](data []byte) (T, error)
```

Convenient cached parsing (currently falls back to non-cached).

## Configuration

### CacheConfig

```go
type CacheConfig struct {
    TTL        time.Duration // Cache entry lifetime
    MaxEntries int           // Maximum cache entries  
    Namespace  string        // Cache key namespace
}
```

Default Values:
- TTL: 30 minutes
- MaxEntries: 1000
- Namespace: "gopantic"

## Validation Tags

### Required

```go
`validate:"required"`
```

Field must have a non-zero value.

Applies to: All types
Error: "field is required"

### Min/Max (Numbers)

```go
`validate:"min=10"`
`validate:"max=100"`
`validate:"min=1,max=10"`
```

Numeric value constraints.

Applies to: `int`, `float64`, and variants
Error: "value must be at least N" / "value must be at most N"

### Min/Max (Strings/Slices)

```go
`validate:"min=3"`      // Minimum length
`validate:"max=50"`     // Maximum length
```

Length constraints for strings and slices.

Applies to: `string`, `[]T`
Error: "length must be at least N" / "length must be at most N"

### Length

```go
`validate:"length=8"`
```

Exact length requirement.

Applies to: `string`, `[]T`
Error: "length must be exactly N characters"

### Email

```go
`validate:"email"`
```

Valid email format validation.

Applies to: `string`
Error: "must be a valid email address"

### Alpha

```go
`validate:"alpha"`
```

Alphabetic characters only (a-z, A-Z).

Applies to: `string`
Error: "must contain only alphabetic characters"

### Alphanum

```go
`validate:"alphanum"`
```

Alphanumeric characters only (a-z, A-Z, 0-9).

Applies to: `string`
Error: "must contain only alphanumeric characters"

### Combined Rules

```go
`validate:"required,min=3,max=20,alphanum"`
```

Multiple validation rules separated by commas.

## Type Coercion

### Automatic Coercion

gopantic automatically converts between compatible types:

| Target Type | From Types | Examples |
|-------------|------------|----------|
| `int` | `string`, `float64` | `"42"` → `42`, `42.0` → `42` |
| `float64` | `string`, `int` | `"3.14"` → `3.14`, `42` → `42.0` |
| `bool` | `string`, `int` | `"true"` → `true`, `1` → `true` |
| `string` | `int`, `float64`, `bool` | `42` → `"42"`, `true` → `"true"` |
| `time.Time` | `string`, `int` | RFC3339, Unix timestamps |

### Boolean Coercion

Truthy values: `"true"`, `"yes"`, `"1"`, `"on"`, `1`, non-zero numbers
Falsy values: `"false"`, `"no"`, `"0"`, `"off"`, `""`, `0`, zero values

### Time Parsing

Supports multiple time formats:
- **RFC3339** - `"2023-01-15T10:30:00Z"`
- **RFC3339Nano** - `"2023-01-15T10:30:00.123456789Z"`
- **Date only** - `"2023-01-15"`
- **Unix timestamp** - `1673781000` (integer)
- **Unix timestamp** - `1673781000.123` (float)

## Error Types

### ParseError

```go
type ParseError struct {
    Field   string      // Field name
    Value   interface{} // Problematic value
    Type    string      // Target type
    Message string      // Error description
}
```

Returned for type coercion failures.

### ValidationError

```go
type ValidationError struct {
    Field   string      // Field name
    Value   interface{} // Invalid value  
    Rule    string      // Validation rule
    Message string      // Error description
}
```

Returned for validation failures.

### Multiple Errors

When multiple errors occur, they're aggregated:

```
"multiple errors: validation error on field 'ID': field is required; parse error on field 'Age': cannot convert string 'invalid' to integer"
```

## Struct Tag Support

### JSON Tags

```go
type User struct {
    ID       int    `json:"id"`           // Maps to "id"
    Username string `json:"username"`     // Maps to "username"  
    Hidden   string `json:"-"`            // Ignored
    Count    int    `json:"count,omitempty"` // Standard options supported
}
```

### YAML Tags

```go
type Config struct {
    Port int `yaml:"port" json:"port"` // Supports both formats
    Host string `yaml:"host"`          // YAML-specific
}
```

**Fallback behavior:** If YAML tag is missing, falls back to JSON tag.

### Validation Tags

```go
type Product struct {
    SKU   string  `json:"sku" validate:"required,length=8,alphanum"`
    Price float64 `json:"price" validate:"required,min=0.01"`
}
```

## Best Practices

### Performance

1. **Use caching** for repeated parsing of similar data
2. **Specify format explicitly** when known for slight performance gain
3. **Minimize validation rules** - only validate what's necessary
4. **Reuse parser instances** in high-throughput scenarios

### Error Handling

```go
user, err := model.ParseInto[User](data)
if err != nil {
    // Handle specific error types
    if parseErr, ok := err.(*model.ParseError); ok {
        log.Printf("Parse error in field %s: %s", parseErr.Field, parseErr.Message)
    }
    return err
}
```

### Memory Management

```go
// For high-throughput scenarios
parser := model.NewCachedParser[User](nil)
defer parser.Close() // Important: cleanup resources

// Process many requests
for data := range dataChannel {
    user, err := parser.Parse(data)
    // ... handle result
}
```

### Struct Design

```go
// Good: Clear, validated struct
type User struct {
    ID       int       `json:"id" validate:"required,min=1"`
    Email    string    `json:"email" validate:"required,email"`
    Name     string    `json:"name" validate:"required,min=2,max=50"`
    Age      *int      `json:"age,omitempty" validate:"min=0,max=150"` // Optional field
    Created  time.Time `json:"created_at"`
}

// Avoid: Over-validation
type OverValidated struct {
    ID string `json:"id" validate:"required,min=1,max=100,alphanum,length=8"` // Conflicting rules
}
```