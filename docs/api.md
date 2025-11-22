# API Reference

Complete API documentation for gopantic.

## Core Functions

### ParseInto

```go
func ParseInto[T any](data []byte) (T, error)
```

Parses JSON or YAML with automatic format detection, type coercion, and validation.

```go
user, err := model.ParseInto[User]([]byte(`{"id": 1, "name": "Alice"}`))
```

### ParseIntoWithFormat

```go
func ParseIntoWithFormat[T any](data []byte, format Format) (T, error)
```

Parses with explicit format (`FormatJSON` or `FormatYAML`).

```go
user, err := model.ParseIntoWithFormat[User](yamlData, model.FormatYAML)
```

### Validate

```go
func Validate[T any](v *T) error
```

Validates an already-parsed struct.

```go
var user User
json.Unmarshal(data, &user)
err := model.Validate(&user)
```

## Format Detection

### DetectFormat

```go
func DetectFormat(data []byte) Format
```

Auto-detects JSON or YAML format. Looks for JSON markers (`{`, `[`), YAML markers (`---`, `:`), defaults to JSON for ambiguous cases.

## Caching

### NewCachedParser

```go
func NewCachedParser[T any](config *CacheConfig) (*CachedParser[T], error)
```

Creates a cached parser instance.

```go
config := &model.CacheConfig{
    TTL:        5 * time.Minute,
    MaxEntries: 1000,
}
parser, err := model.NewCachedParser[User](config)
defer parser.Close()
```

### CachedParser Methods

```go
func (cp *CachedParser[T]) Parse(data []byte) (T, error)
func (cp *CachedParser[T]) Stats() (size, maxSize int, hitRate float64)
func (cp *CachedParser[T]) ClearCache()
func (cp *CachedParser[T]) Close()
```

## Configuration

### CacheConfig

```go
type CacheConfig struct {
    TTL             time.Duration // Default: 30 minutes
    MaxEntries      int           // Default: 1000
    Namespace       string        // Default: "gopantic"
    CleanupInterval time.Duration // Default: 5 minutes
}
```

### Global Variables

```go
var MaxInputSize = 10 * 1024 * 1024  // Max 10MB input (0 = unlimited)
var MaxCacheSize = 1000               // Max validation metadata cache (0 = unlimited)
var MaxValidationDepth = 32           // Max nested struct depth (0 = unlimited)
```

## Validation Tags

### Built-in Validators

| Tag | Applies To | Description | Example |
|-----|------------|-------------|---------|
| `required` | All types | Non-zero value required | `validate:"required"` |
| `min=N` | Numbers | Minimum value | `validate:"min=1"` |
| `max=N` | Numbers | Maximum value | `validate:"max=100"` |
| `min=N` | String, Slice | Minimum length | `validate:"min=3"` |
| `max=N` | String, Slice | Maximum length | `validate:"max=50"` |
| `length=N` | String, Slice | Exact length | `validate:"length=8"` |
| `email` | String | Valid email format | `validate:"email"` |
| `alpha` | String | Alphabetic only (a-z, A-Z) | `validate:"alpha"` |
| `alphanum` | String | Alphanumeric only | `validate:"alphanum"` |

### Combined Rules

```go
type User struct {
    Name string `json:"name" validate:"required,min=2,max=50,alpha"`
    Age  int    `json:"age" validate:"required,min=18,max=120"`
}
```

### Custom Validators

```go
model.RegisterGlobalFunc("strong_password", func(fieldName string, value interface{}, params map[string]interface{}) error {
    password, ok := value.(string)
    if !ok || len(password) < 8 {
        return model.NewValidationError(fieldName, value, "strong_password", "password must be at least 8 characters")
    }
    return nil
})
```

### Cross-Field Validators

```go
model.RegisterGlobalCrossFieldFunc("password_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
    password := structValue.FieldByName("Password").Interface().(string)
    confirmPassword := fieldValue.(string)
    if password != confirmPassword {
        return model.NewValidationError(fieldName, fieldValue, "password_match", "passwords do not match")
    }
    return nil
})
```

## Type Coercion

Automatic conversion between compatible types:

| Target | From | Examples |
|--------|------|----------|
| `int` | `string`, `float64` | `"42"` → `42`, `42.0` → `42` |
| `float64` | `string`, `int` | `"3.14"` → `3.14`, `42` → `42.0` |
| `bool` | `string`, `int` | `"true"` → `true`, `1` → `true` |
| `string` | Any | `42` → `"42"`, `true` → `"true"` |
| `time.Time` | `string`, `int` | RFC3339, Unix timestamps |

**Boolean coercion:**
- Truthy: `"true"`, `"yes"`, `"1"`, `"on"`, `1`, non-zero
- Falsy: `"false"`, `"no"`, `"0"`, `"off"`, `""`, `0`

**Time formats:** RFC3339, RFC3339Nano, Date only (`2023-01-15`), Unix timestamp (int/float)

## Error Types

### ParseError

```go
type ParseError struct {
    Field   string
    Value   interface{}
    Type    string
    Message string
}
```

Returned for type coercion failures.

### ValidationError

```go
type ValidationError struct {
    Field   string
    Value   interface{}
    Rule    string
    Message string
}
```

Returned for validation failures.

### Multiple Errors

Aggregated as: `"multiple errors: validation error on field 'ID': field is required; parse error on field 'Age': cannot convert string 'invalid' to integer"`

### Security: Error Messages

**Important:** Error messages include field values. Sanitize before logging/displaying:

```go
user, err := model.ParseInto[User](data)
if err != nil {
    log.Error("parse failed", "error", err)  // Internal: full details
    return errors.New("invalid request")      // External: sanitized
}
```

## Struct Tags

### JSON Tags

```go
type User struct {
    ID     int    `json:"id"`
    Name   string `json:"name,omitempty"`
    Hidden string `json:"-"`  // Ignored
}
```

### YAML Tags

```go
type Config struct {
    Port int    `yaml:"port" json:"port"`  // Both formats
    Host string `yaml:"host"`               // YAML only
}
```

Falls back to JSON tag if YAML tag is missing.

## Best Practices

### Performance

- Use `NewCachedParser` for repeated parsing
- Specify format explicitly when known
- Minimize validation rules
- Reuse parser instances

### Error Handling

```go
user, err := model.ParseInto[User](data)
if err != nil {
    if parseErr, ok := err.(*model.ParseError); ok {
        log.Printf("Parse error in %s: %s", parseErr.Field, parseErr.Message)
    }
    return err
}
```

### Memory Management

```go
parser := model.NewCachedParser[User](nil)
defer parser.Close()  // Cleanup resources

for data := range dataChannel {
    user, err := parser.Parse(data)
    // ... handle
}
```

### Struct Design

```go
type User struct {
    ID      int       `json:"id" validate:"required,min=1"`
    Email   string    `json:"email" validate:"required,email"`
    Name    string    `json:"name" validate:"required,min=2,max=50"`
    Age     *int      `json:"age,omitempty" validate:"min=0,max=150"` // Optional
    Created time.Time `json:"created_at"`
}
```

Use pointers for optional fields to distinguish nil from zero value.

## See Also

- [Type Reference](type-reference.md) - Supported types and limitations
- [Migration Guide](migration.md) - Switching from other libraries
- [Architecture](architecture.md) - Implementation details
- [Examples](../examples/) - Working code examples
