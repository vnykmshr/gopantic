# gopantic

**Practical JSON/YAML parsing with validation for Go.**

Inspired by Python's Pydantic, gopantic provides type-safe parsing, coercion, and validation with idiomatic Go APIs.

[![Go Reference](https://pkg.go.dev/badge/github.com/vnykmshr/gopantic.svg)](https://pkg.go.dev/github.com/vnykmshr/gopantic)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://github.com/vnykmshr/gopantic/actions/workflows/ci.yml/badge.svg)](https://github.com/vnykmshr/gopantic/actions/workflows/ci.yml)

## Quick Start

```bash
go get github.com/vnykmshr/gopantic
```

```go
package main

import (
    "fmt"
    "log"
    "github.com/vnykmshr/gopantic/pkg/model"
)

type User struct {
    ID    int    `json:"id" validate:"required,min=1"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}

func main() {
    raw := []byte(`{"id": "42", "name": "Alice", "email": "alice@example.com", "age": "28"}`)
    
    user, err := model.ParseInto[User](raw)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%+v\n", user) // {ID:42 Name:Alice Email:alice@example.com Age:28}
}
```

## Features

- **🔄 JSON/YAML parsing** with automatic format detection
- **⚡ Type coercion** (`"123"` → `123`, `"true"` → `true`) 
- **✅ Validation** using struct tags (`validate:"required,email,min=5"`)
- **🔗 Cross-field validation** (password confirmation, field comparisons)
- **🏗️ Built-in validators**: `required`, `min`, `max`, `email`, `alpha`, `alphanum`, `length`
- **📦 Nested structs** and arrays with full validation
- **⏰ Time parsing** (RFC3339, Unix timestamps, custom formats)
- **🎯 Pointer support** for optional fields (`*string`, `*int`)
- **🚀 High-performance caching** (5-27x speedup)
- **🔒 Thread-safe** concurrent parsing
- **🧩 Zero dependencies** (except optional YAML support)
- **🎨 Generics support** for type-safe parsing

## YAML Support

Works seamlessly with YAML - same API, automatic detection:

```go
yamlData := []byte(`
id: 42
name: Alice
email: alice@example.com
age: 28
`)

user, err := model.ParseInto[User](yamlData) // Automatic YAML detection
```

## Validation

```go
type Product struct {
    SKU   string  `json:"sku" validate:"required,length=8"`
    Price float64 `json:"price" validate:"required,min=0.01"`
}

// Multiple validation errors are aggregated
product, err := model.ParseInto[Product](invalidData)
// Error: "multiple errors: validation error on field 'SKU': ...; validation error on field 'Price': ..."
```

### Cross-Field Validation

Built-in support for cross-field validation - compare fields against each other:

```go
type UserRegistration struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,password_match"`
    Email           string `json:"email" validate:"required,email"`
    NotificationEmail string `json:"notification_email" validate:"email,email_different"`
}

// Register custom cross-field validators
model.RegisterGlobalCrossFieldFunc("password_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
    confirmPassword := fieldValue.(string)
    password := structValue.FieldByName("Password").String()
    if confirmPassword != password {
        return model.NewValidationError(fieldName, fieldValue, "password_match", "passwords do not match")
    }
    return nil
})
```

## Performance

High-performance caching for repeated parsing:

```go
parser := model.NewCachedParser[User](nil) // Simple in-memory cache
defer parser.Close()

user1, _ := parser.Parse(data) // Cache miss
user2, _ := parser.Parse(data) // Cache hit - 27x faster
```

## Why Choose gopantic?

### vs. Standard Library (`encoding/json`)
- ✅ **Built-in validation** - No separate validation step needed
- ✅ **Type coercion** - Handles `"123"` → `123` automatically  
- ✅ **Better errors** - Structured error reporting with field paths
- ✅ **YAML support** - Automatic format detection
- ✅ **Cross-field validation** - Compare fields against each other

### vs. Validation Libraries (`go-playground/validator`)
- ✅ **Integrated parsing** - Parse and validate in one step
- ✅ **Type coercion** - No manual string conversion needed
- ✅ **Format agnostic** - Works with JSON and YAML seamlessly
- ✅ **Generics support** - Type-safe with `ParseInto[T]()`
- ✅ **Performance** - Built-in caching for repeated operations

### vs. Code Generation (`easyjson`, `ffjson`)
- ✅ **Zero code generation** - No build step or generated files
- ✅ **Dynamic validation** - Runtime validation rule changes
- ✅ **Simpler workflow** - Standard Go development process
- ✅ **Faster iteration** - No regeneration on struct changes
- ✅ **Cross-field validation** - Complex validation logic support

### vs. Schema Libraries (`jsonschema`, `gojsonschema`)
- ✅ **Native Go structs** - Use existing struct definitions
- ✅ **Compile-time safety** - Type checking at compile time
- ✅ **Better performance** - Direct struct mapping vs. schema validation
- ✅ **IDE support** - Full autocompletion and refactoring
- ✅ **Integrated coercion** - Automatic type conversion

## Documentation

- [Architecture & Design](docs/architecture.md) - Implementation details and design decisions
- [API Reference](docs/api.md) - Complete API documentation  
- [Examples](examples/) - Practical usage examples

## License

MIT License - see [LICENSE](LICENSE) for details.