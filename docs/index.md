# gopantic

**Type-safe JSON/YAML parsing with validation for Go.**

Inspired by Python's Pydantic, gopantic provides automatic format detection, intelligent type coercion, and struct tag-based validation in a single `ParseInto[T]()` call.

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/vnykmshr/gopantic/pkg/model"
)

type User struct {
    ID    int    `json:"id" validate:"required,min=1"`
    Name  string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=0,max=150"`
}

func main() {
    data := []byte(`{
        "id": "123",
        "name": "Alice",
        "email": "alice@example.com",
        "age": "30"
    }`)

    // Parse with automatic format detection, type coercion, and validation
    user, err := model.ParseInto[User](data)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Parsed: %+v\n", user)
    // Output: Parsed: {ID:123 Name:Alice Email:alice@example.com Age:30}
    // Note: String "123" was coerced to int 123
}
```

## Key Features

| Feature | Description |
|---------|-------------|
| **Automatic Format Detection** | JSON or YAML, detected automatically |
| **Type Coercion** | `"123"` becomes `123`, `"true"` becomes `true` |
| **Validation** | Struct tags: `validate:"required,email,min=5"` |
| **Cross-Field Validation** | Compare fields: `validate:"gtfield=MinValue"` |
| **High Performance** | Optional caching for 5x+ speedup on repeated data |
| **Minimal Dependencies** | Only `gopkg.in/yaml.v3` for YAML support |

## Installation

```bash
go get github.com/vnykmshr/gopantic
```

Requires Go 1.23+.

## When to Use gopantic

**Good fit:**

- API request parsing with validation
- Configuration file loading (JSON/YAML)
- Data import/ETL with type conversion
- Any scenario where you want parsing + validation in one step

**Not ideal for:**

- High-frequency, latency-critical paths (use standard library)
- Simple JSON without validation needs
- Binary protocols

## Next Steps

- [Getting Started](getting-started.md) - Installation and basic usage
- [Validation Guide](guide/validation.md) - All validation options
- [API Reference](reference/api.md) - Complete API documentation
- [Migration Guide](migration.md) - Coming from other libraries

## License

MIT License - see [LICENSE](https://github.com/vnykmshr/gopantic/blob/main/LICENSE)
