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

**Phase 1 (Current):**
- ✅ Basic JSON parsing into typed structs
- ✅ Type coercion for `int`, `float64`, `string`, `bool`
- ✅ Struct field mapping using `json` tags
- ✅ Comprehensive error handling and reporting
- ✅ Zero external dependencies

**Coming Soon:**
- 🔄 Struct tag validation (`required`, `min`, `max`, `email`)
- 🔄 YAML support
- 🔄 Nested struct parsing
- 🔄 Custom validators
- 🔄 Time parsing

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
    
    "github.com/vnykmshr/gopantic/pkg/model"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func main() {
    // JSON with mixed types (strings that should be numbers)
    raw := []byte(`{"id": "42", "name": "Alice", "email": "alice@example.com", "age": "28"}`)
    
    user, err := model.ParseInto[User](raw)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("%+v\n", user)
    // Output: {ID:42 Name:Alice Email:alice@example.com Age:28}
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

## Examples

Run the comprehensive example:

```bash
go run examples/basic/main.go
```

This demonstrates:
- Basic parsing and type coercion
- Mixed data types in JSON
- Boolean variations
- Error handling
- Missing field handling

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
2. 🔄 **Phase 2:** Validation Framework
3. 📋 **Phase 3:** Extended Type Support
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