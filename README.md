# gopantic

**A typed data parsing and validation library for Go**, inspired by [Pydantic](https://docs.pydantic.dev/).

`gopantic` is not a 1:1 clone of Pydantic. It borrows Pydantic's best ideas — parsing, coercion, validation — but keeps the implementation **practical, idiomatic, and concise** for Go.

[![Go Version](https://img.shields.io/badge/go-1.22+-blue.svg)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![CI](https://github.com/vnykmshr/gopantic/actions/workflows/ci.yml/badge.svg)](https://github.com/vnykmshr/gopantic/actions/workflows/ci.yml)

## Vision

- Keep it **idiomatic Go** → APIs should feel natural in Go (struct tags, helpers, minimal magic)
- Prefer **practical features over full parity** → No attempt to replicate every Pydantic feature, only those useful in Go
- Focus on **parsing + validation** → Core scope is JSON/YAML → typed Go structs with coercion + validation

## Features ✨

**Phase 1, 2 & 3 (Current):**
- ✅ Basic JSON parsing into typed structs
- ✅ Type coercion for `int`, `float64`, `string`, `bool`
- ✅ Struct field mapping using `json` tags
- ✅ **Validation framework with struct tags**
- ✅ **Built-in validators: `required`, `min`, `max`, `email`, `alpha`, `alphanum`, `length`**
- ✅ **Error aggregation with detailed field-level reporting**
- ✅ **Nested struct parsing and validation with field paths**
- ✅ **Time parsing with multiple formats (RFC3339, Unix timestamps, custom formats)**
- ✅ **Slice and array parsing with element validation**
- ✅ Comprehensive error handling and reporting
- ✅ Zero external dependencies

**Coming Soon:**
- 🔄 YAML support
- 🔄 Pointer type handling
- 🔄 Custom validators
- 🔄 Cross-field validation
- 🔄 Advanced validation features

## Installation

```bash
go get github.com/vnykmshr/gopantic
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/vnykmshr/gopantic/pkg/model"
)

type Address struct {
    Street string `json:"street" validate:"required,min=5"`
    City   string `json:"city" validate:"required,min=2"`
    Zip    string `json:"zip" validate:"required,length=5"`
}

type User struct {
    ID        int       `json:"id" validate:"required,min=1"`
    Name      string    `json:"name" validate:"required,min=2,alpha"`
    Email     string    `json:"email" validate:"required,email"`
    Age       int       `json:"age" validate:"min=18,max=120"`
    Address   Address   `json:"address" validate:"required"`
    CreatedAt time.Time `json:"created_at"`
}

func main() {
    // JSON with nested structs, time parsing, and mixed types
    raw := []byte(`{
        "id": "42", 
        "name": "Alice", 
        "email": "alice@example.com", 
        "age": "28",
        "address": {
            "street": "123 Main St",
            "city": "Springfield", 
            "zip": "12345"
        },
        "created_at": "2023-01-15T10:30:00Z"
    }`)
    
    user, err := model.ParseInto[User](raw)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%+v\n", user)
    // Output: {ID:42 Name:Alice Email:alice@example.com Age:28 Address:{Street:123 Main St City:Springfield Zip:12345} CreatedAt:2023-01-15 10:30:00 +0000 UTC}
}
```

## Type Coercion

gopantic automatically coerces between compatible types:

```go
type Product struct {
    ID       uint64  `json:"id"`
    Name     string  `json:"name"`
    Price    float64 `json:"price"`
    InStock  bool    `json:"in_stock"`
}

// All of these work:
raw1 := []byte(`{"id": "123", "price": "29.99", "in_stock": "true"}`)
raw2 := []byte(`{"id": 123, "price": 29.99, "in_stock": 1}`)
raw3 := []byte(`{"id": "123", "price": "29.99", "in_stock": "yes"}`)

product, _ := model.ParseInto[Product](raw1)
// Works! Strings coerced to appropriate types
```

### Supported Coercions

| Target Type | From Types | Examples |
|-------------|------------|----------|
| `string` | `int`, `float`, `bool` | `42` → `"42"`, `true` → `"true"` |
| `int`/`uint` | `string`, `float`, `bool` | `"42"` → `42`, `3.14` → `3` |
| `float` | `string`, `int`, `bool` | `"3.14"` → `3.14`, `42` → `42.0` |
| `bool` | `string`, `int`, `float` | `"true"`, `"yes"`, `"1"`, `1` → `true` |

### Boolean Coercion

gopantic supports flexible boolean parsing:

- **Truthy:** `"true"`, `"yes"`, `"1"`, `"on"`, `1`, non-zero numbers
- **Falsy:** `"false"`, `"no"`, `"0"`, `"off"`, `""`, `0`, zero values

## Validation Framework

gopantic includes a powerful validation system using struct tags:

```go
type UserRegistration struct {
    Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Bio      string `json:"bio" validate:"max=500"`
}

// Validation happens automatically during parsing
user, err := model.ParseInto[UserRegistration](jsonData)
if err != nil {
    // err contains detailed validation failures:
    // "multiple errors: validation error on field "Username": string length must be at least 3 characters; ..."
}
```

### Built-in Validators

| Validator | Description | Example |
|-----------|-------------|---------|
| `required` | Field must have non-zero value | `validate:"required"` |
| `min=N` | Minimum value/length | `validate:"min=3"` |
| `max=N` | Maximum value/length | `validate:"max=100"` |
| `length=N` | Exact length | `validate:"length=8"` |
| `email` | Valid email format | `validate:"email"` |
| `alpha` | Alphabetic characters only | `validate:"alpha"` |
| `alphanum` | Alphanumeric characters only | `validate:"alphanum"` |

### Validation Features

- **Error Aggregation**: All validation errors reported together
- **Type-aware**: Different validation logic for strings, numbers, arrays
- **Post-coercion**: Validation runs after type coercion
- **Detailed Messages**: Clear, actionable error messages
- **Zero Overhead**: No validation impact when tags aren't used

## Examples

Run the comprehensive examples:

```bash
go run examples/basic/main.go      # Basic parsing and coercion
go run examples/validation/main.go # Validation framework demo
go run examples/time_parsing/main.go # Time parsing with multiple formats
```

These demonstrate:
- Basic parsing and type coercion
- Validation with multiple rules
- Nested struct parsing with validation
- Time parsing (RFC3339, Unix timestamps, custom formats)
- Slice and array parsing with element validation
- Error handling and aggregation with field paths
- Mixed data types in JSON
- Boolean variations
- Real-world use cases

## Error Handling

gopantic provides detailed error messages for debugging:

```go
raw := []byte(`{"id": "not-a-number", "name": "Alice"}`)
_, err := model.ParseInto[User](raw)

// Error: parse error on field "id": cannot parse string "not-a-number" as integer
```

Multiple errors are aggregated:

```go
raw := []byte(`{"id": "bad", "age": "also-bad"}`)
_, err := model.ParseInto[User](raw)

// Error: multiple errors: parse error on field "id": ...; parse error on field "age": ...
```

## Development

### Prerequisites

- Go 1.22 or later
- Make (optional, for development commands)

### Setup

```bash
git clone https://github.com/vnykmshr/gopantic.git
cd gopantic
make init  # Install development tools and setup git hooks
```

### Development Commands

```bash
make check      # Fast development cycle (fmt, vet, lint, test)
make audit      # Comprehensive quality and security audit
make test       # Run tests with coverage
make examples   # Run all examples
make help       # Show all available commands
```

### Quality Gates

This project maintains high code quality with:

- **Linting:** golangci-lint with comprehensive rules
- **Testing:** >90% code coverage requirement
- **Security:** gosec vulnerability scanning
- **Complexity:** gocyclo complexity analysis
- **Dead Code:** deadcode detection

## Roadmap

See our [comprehensive implementation plan](todos/todos.md) with 6 phases:

1. ✅ **Phase 1:** Core Foundation & Basic Parsing
2. ✅ **Phase 2:** Validation Framework
3. 🔄 **Phase 3:** Extended Type Support (75% complete - nested structs, time parsing, arrays/slices done)
4. 📋 **Phase 4:** YAML Support
5. 📋 **Phase 5:** Advanced Validation
6. 📋 **Phase 6:** Performance & Polish

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Steps

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes following our coding standards
4. Run quality checks: `make check`
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) for details.

## Inspiration

This project is inspired by [Pydantic](https://docs.pydantic.dev/) but designed specifically for Go's type system and idioms. We focus on practical features that provide value in Go applications while maintaining simplicity and performance.

---

**gopantic** - Parse with confidence, validate with ease! 🚀