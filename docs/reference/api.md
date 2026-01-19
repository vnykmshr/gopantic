# API Reference

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
func NewCachedParser[T any](config *CacheConfig) *CachedParser[T]
```

Creates a cached parser instance.

```go
config := &model.CacheConfig{
    TTL:             5 * time.Minute,
    MaxEntries:      1000,
    CleanupInterval: 2 * time.Minute,
}
parser := model.NewCachedParser[User](config)
defer parser.Close()
```

### CachedParser Methods

```go
func (cp *CachedParser[T]) Parse(data []byte) (T, error)
func (cp *CachedParser[T]) ParseWithFormat(data []byte, format Format) (T, error)
func (cp *CachedParser[T]) Stats() (size, maxSize int, hitRate float64)
func (cp *CachedParser[T]) ClearCache()
func (cp *CachedParser[T]) Close()
```

### DefaultCacheConfig

```go
func DefaultCacheConfig() *CacheConfig
```

Returns sensible defaults:

- TTL: 1 hour
- MaxEntries: 1000
- CleanupInterval: 30 minutes

## Configuration

### CacheConfig

```go
type CacheConfig struct {
    TTL             time.Duration // Time to live for cached entries (default: 1 hour)
    MaxEntries      int           // Maximum number of cached entries (default: 1000)
    CleanupInterval time.Duration // How often to run cleanup (default: TTL/2, 0 to disable)
}
```

### Global Configuration

Package-level configuration variables:

```go
var MaxInputSize = 10 * 1024 * 1024  // Max 10MB input (0 = unlimited)
var MaxCacheSize = 1000               // Max validation metadata cache (0 = unlimited)
var MaxValidationDepth = 32           // Max nested struct depth
```

**Warning:** Direct modification of these variables is NOT thread-safe. Use the Get/Set functions for concurrent access.

### Thread-Safe Accessors

```go
// MaxInputSize
func GetMaxInputSize() int
func SetMaxInputSize(size int)

// MaxCacheSize
func GetMaxCacheSize() int
func SetMaxCacheSize(size int)

// MaxValidationDepth
func GetMaxValidationDepth() int
func SetMaxValidationDepth(depth int)
```

Example:

```go
// Safe for concurrent use
model.SetMaxInputSize(5 * 1024 * 1024)  // 5MB
size := model.GetMaxInputSize()
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
| `len=N` | String, Slice | Exact length | `validate:"len=8"` |
| `email` | String | Valid email format | `validate:"email"` |
| `url` | String | Valid URL | `validate:"url"` |
| `uuid` | String | Valid UUID | `validate:"uuid"` |
| `alpha` | String | Alphabetic only | `validate:"alpha"` |
| `alphanum` | String | Alphanumeric only | `validate:"alphanum"` |
| `oneof` | String | One of listed values | `validate:"oneof=a b c"` |

### Cross-Field Validators

| Tag | Description | Example |
|-----|-------------|---------|
| `eqfield=F` | Equal to field F | `validate:"eqfield=Password"` |
| `nefield=F` | Not equal to field F | `validate:"nefield=OldPassword"` |
| `gtfield=F` | Greater than field F | `validate:"gtfield=Min"` |
| `gtefield=F` | Greater than or equal | `validate:"gtefield=Start"` |
| `ltfield=F` | Less than field F | `validate:"ltfield=Max"` |
| `ltefield=F` | Less than or equal | `validate:"ltefield=End"` |

### Custom Validators

```go
model.RegisterGlobalFunc("is_even", func(fieldName string, value interface{}, params map[string]interface{}) error {
    num, ok := value.(int)
    if !ok {
        return nil
    }
    if num%2 != 0 {
        return model.NewValidationError(fieldName, value, "is_even", "must be even")
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

Multiple errors are aggregated:

```
multiple errors: validation error on field 'ID': field is required; parse error on field 'Age': cannot convert string 'invalid' to integer
```

**Security note:** Error messages include field values. Sanitize before logging or returning to clients.

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

## See Also

- [Types Reference](types.md) - Supported types and limitations
- [Migration Guide](../migration.md) - Switching from other libraries
- [Architecture](../architecture.md) - Implementation details
