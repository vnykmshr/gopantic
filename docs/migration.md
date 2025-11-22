# Migration Guide

This guide helps you migrate to gopantic from common Go parsing and validation libraries.

## From encoding/json + go-playground/validator

### Before (encoding/json + validator)

```go
import (
    "encoding/json"
    "github.com/go-playground/validator/v10"
)

type User struct {
    ID    int    `json:"id" validate:"required,min=1"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func parseUser(data []byte) (*User, error) {
    var user User

    // Step 1: Parse JSON
    if err := json.Unmarshal(data, &user); err != nil {
        return nil, err
    }

    // Step 2: Validate
    validate := validator.New()
    if err := validate.Struct(user); err != nil {
        return nil, err
    }

    return &user, nil
}
```

### After (gopantic)

```go
import "github.com/vnykmshr/gopantic/pkg/model"

type User struct {
    ID    int    `json:"id" validate:"required,min=1"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

func parseUser(data []byte) (User, error) {
    return model.ParseInto[User](data)
}
```

### Key Differences

1. **Single call**: Parsing and validation happen together
2. **Type coercion**: `{"id": "42"}` works automatically
3. **Generics**: Returns `User` directly, not `*User` or `interface{}`
4. **YAML support**: Same API works for YAML without changes

## From encoding/json (no validation)

### Before

```go
import "encoding/json"

type Config struct {
    Port int    `json:"port"`
    Host string `json:"host"`
}

func loadConfig(data []byte) (*Config, error) {
    var cfg Config
    if err := json.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    // Manual validation
    if cfg.Port < 1 || cfg.Port > 65535 {
        return nil, errors.New("invalid port")
    }
    if cfg.Host == "" {
        return nil, errors.New("host required")
    }

    return &cfg, nil
}
```

### After

```go
import "github.com/vnykmshr/gopantic/pkg/model"

type Config struct {
    Port int    `json:"port" validate:"required,min=1,max=65535"`
    Host string `json:"host" validate:"required"`
}

func loadConfig(data []byte) (Config, error) {
    return model.ParseInto[Config](data)
}
```

### Benefits

- Declarative validation via struct tags
- Validation errors include field names automatically
- Type coercion for free (`"8080"` → `8080`)

## From YAML libraries (gopkg.in/yaml.v3)

### Before

```go
import (
    "gopkg.in/yaml.v3"
    "github.com/go-playground/validator/v10"
)

type Config struct {
    Database struct {
        Host string `yaml:"host" validate:"required"`
        Port int    `yaml:"port" validate:"required,min=1"`
    } `yaml:"database"`
}

func loadConfig(data []byte) (*Config, error) {
    var cfg Config

    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    validate := validator.New()
    if err := validate.Struct(cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

### After

```go
import "github.com/vnykmshr/gopantic/pkg/model"

type Config struct {
    Database struct {
        Host string `yaml:"host" validate:"required"`
        Port int    `yaml:"port" validate:"required,min=1"`
    } `yaml:"database"`
}

func loadConfig(data []byte) (Config, error) {
    // Automatic YAML detection, or use ParseIntoWithFormat for explicit
    return model.ParseInto[Config](data)
}
```

## Migration Checklist

### 1. Update imports

```diff
- import "encoding/json"
- import "github.com/go-playground/validator/v10"
+ import "github.com/vnykmshr/gopantic/pkg/model"
```

### 2. Update validation tags

Most validator tags work as-is, but some differences:

| validator | gopantic | Notes |
|-----------|----------|-------|
| `required` | `required` | Same |
| `min=5` | `min=5` | Same |
| `max=100` | `max=100` | Same |
| `email` | `email` | Similar (simplified regex) |
| `len=8` | `length=8` | Different name |
| `eqfield=Password` | Custom | Use cross-field validators |
| `dive` | N/A | Nested validation automatic |

### 3. Handle type coercion

gopantic automatically coerces compatible types:

```go
// Before: Would fail with json.Unmarshal
{"age": "25"} // string "25"

// After: Works with gopantic
type User struct {
    Age int `json:"age"` // Automatically converts "25" → 25
}
```

If you have custom string-to-type conversion logic, you can often remove it.

### 4. Update error handling

```go
// Before: validator returns ValidationErrors
if err != nil {
    if validationErrs, ok := err.(validator.ValidationErrors); ok {
        for _, fieldErr := range validationErrs {
            fmt.Printf("Field: %s, Error: %s\n", fieldErr.Field(), fieldErr.Tag())
        }
    }
}

// After: gopantic returns structured errors
if err != nil {
    // Error message already formatted with field names
    log.Error(err) // "validation error on field 'Email': must be a valid email address"

    // Or handle structured errors
    if parseErr, ok := err.(*model.ParseError); ok {
        fmt.Printf("Field: %s, Type: %s\n", parseErr.Field, parseErr.Type)
    }
}
```

### 5. Handle pointer fields

```go
// Before: Use pointers for optional fields
type User struct {
    Name  string  `json:"name"`
    Phone *string `json:"phone,omitempty"` // Optional
}

// After: Same pattern works
type User struct {
    Name  string  `json:"name" validate:"required"`
    Phone *string `json:"phone"` // Optional, no required tag
}
```

## Common Gotchas

### 1. Struct tags order

gopantic tags can appear in any order:
```go
// Both work
`json:"name" validate:"required"`
`validate:"required" json:"name"`
```

### 2. Zero values vs required

```go
// This allows zero values (0, false, "")
type Config struct {
    Port int `json:"port"` // 0 is valid
}

// This requires non-zero values
type Config struct {
    Port int `json:"port" validate:"required"` // 0 fails validation
}
```

### 3. Nested struct validation

Validation is automatic for nested structs:

```go
type Address struct {
    City string `json:"city" validate:"required"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address"` // City is automatically validated
}
```

### 4. Custom validators

```go
// Before: validator.RegisterValidation
validate.RegisterValidation("strong_password", func(fl validator.FieldLevel) bool {
    password := fl.Field().String()
    return len(password) >= 8 && hasUpperLower(password)
})

// After: model.RegisterGlobalFunc
model.RegisterGlobalFunc("strong_password", func(fieldName string, value interface{}, params map[string]interface{}) error {
    password, ok := value.(string)
    if !ok || len(password) < 8 || !hasUpperLower(password) {
        return model.NewValidationError(fieldName, value, "strong_password",
            "password must be at least 8 characters with upper and lowercase letters")
    }
    return nil
})
```

## Performance Considerations

### Caching

If you're parsing the same payload repeatedly, use caching:

```go
// Create cached parser once
parser := model.NewCachedParser[User](nil)
defer parser.Close()

// Reuse for multiple parses
user1, _ := parser.Parse(data) // Cache miss
user2, _ := parser.Parse(data) // Cache hit - faster
```

### Input size limits

For untrusted input, configure max size:

```go
// Set global limit (default 10MB)
model.MaxInputSize = 5 * 1024 * 1024 // 5MB

// Or disable for large files
model.MaxInputSize = 0 // No limit
```

## Getting Help

If you encounter issues during migration:

1. Check the [API documentation](api.md) for specific function details
2. Review [examples](../examples/) for common patterns
3. File an issue at https://github.com/vnykmshr/gopantic/issues with:
   - Your current code (before migration)
   - The error or unexpected behavior
   - Go version and gopantic version

## Quick Reference

| Task | encoding/json + validator | gopantic |
|------|---------------------------|----------|
| Parse JSON | `json.Unmarshal()` + `validate.Struct()` | `model.ParseInto[T]()` |
| Parse YAML | `yaml.Unmarshal()` + `validate.Struct()` | `model.ParseInto[T]()` |
| Type coercion | Manual | Automatic |
| Validation | Separate step | Integrated |
| Custom validators | `RegisterValidation()` | `RegisterGlobalFunc()` |
| Error handling | `ValidationErrors` type | Structured error types |
