# Getting Started

This guide covers installation and basic usage of gopantic.

## Installation

```bash
go get github.com/vnykmshr/gopantic
```

## Basic Usage

### Define Your Struct

Use standard Go struct tags for JSON mapping and validation:

```go
type User struct {
    ID       int    `json:"id" validate:"required,min=1"`
    Username string `json:"username" validate:"required,min=3,max=50"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"min=0,max=150"`
    IsActive bool   `json:"is_active"`
}
```

### Parse Data

```go
import "github.com/vnykmshr/gopantic/pkg/model"

// Parse JSON with automatic type coercion and validation
data := []byte(`{"id": "123", "username": "alice", "email": "alice@example.com", "age": "25", "is_active": "true"}`)
user, err := model.ParseInto[User](data)
if err != nil {
    // Handle validation or parsing error
    log.Fatal(err)
}
```

Note how string values like `"123"` are automatically coerced to the appropriate Go types.

### YAML Support

YAML works the same way - format is detected automatically:

```go
yamlData := []byte(`
id: 456
username: bob
email: bob@example.com
age: 30
is_active: true
`)

user, err := model.ParseInto[User](yamlData)
```

### Explicit Format

If you know the format ahead of time:

```go
// Explicit JSON
user, err := model.ParseIntoWithFormat[User](data, model.FormatJSON)

// Explicit YAML
user, err := model.ParseIntoWithFormat[User](yamlData, model.FormatYAML)
```

## Validation

gopantic uses struct tags for validation. Common validators:

| Validator | Example | Description |
|-----------|---------|-------------|
| `required` | `validate:"required"` | Field must be non-zero |
| `min` | `validate:"min=5"` | Minimum value (numbers) or length (strings) |
| `max` | `validate:"max=100"` | Maximum value or length |
| `email` | `validate:"email"` | Valid email format |
| `oneof` | `validate:"oneof=draft published"` | Must be one of listed values |
| `len` | `validate:"len=10"` | Exact length |

Multiple validators are comma-separated:

```go
Email string `json:"email" validate:"required,email"`
Age   int    `json:"age" validate:"required,min=18,max=120"`
```

See [Validation Guide](guide/validation.md) for all validators.

## Error Handling

Errors provide detailed information:

```go
user, err := model.ParseInto[User](data)
if err != nil {
    // err.Error() includes field name and validation failure reason
    // Example: "field 'email' validation failed: invalid email format"
    log.Printf("Validation failed: %v", err)
}
```

!!! warning "Security Note"
    Error messages may contain field values. Do not expose raw errors to untrusted clients.

## Caching

For repeated parsing of identical data (e.g., retries, deduplication):

```go
// Create a cached parser
parser := model.NewCachedParser[User](nil) // nil = default config
defer parser.Close()

// First call parses, subsequent identical calls return cached result
user1, _ := parser.Parse(data)  // Cache miss - parses
user2, _ := parser.Parse(data)  // Cache hit - instant return
```

See [Caching Guide](guide/caching.md) for configuration options.

## Configuration

Global configuration with thread-safe accessors:

```go
// Set maximum input size (default: 10MB)
model.SetMaxInputSize(5 * 1024 * 1024) // 5MB

// Set maximum validation depth (default: 32)
model.SetMaxValidationDepth(16)

// Set maximum cache size (default: 1000)
model.SetMaxCacheSize(500)
```

## Next Steps

- [Validation Guide](guide/validation.md) - All validation options
- [Caching Guide](guide/caching.md) - Performance optimization
- [API Reference](reference/api.md) - Complete API documentation
