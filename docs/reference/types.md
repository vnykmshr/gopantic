# Type Reference

Complete guide to type support, coercion behavior, and limitations in gopantic.

## Supported Types

| Type | Support | Coercion | Notes |
|------|---------|----------|-------|
| Primitives (`string`, `int`, `float64`, `bool`) | Full | Yes | All validation tags work |
| Slices & Arrays (`[]T`, `[N]T`) | Full | Yes | Element validation supported |
| Nested Structs | Full | Yes | Recursive validation |
| Pointers (`*T`) | Full | Yes | Nil handling for optional fields |
| `time.Time` | Full | Yes | RFC3339, Unix timestamps |
| `json.RawMessage` | Full | Yes | Preserves raw JSON |
| Maps (`map[string]T`) | Partial | Limited | Structure only, no element validation |
| `interface{}` | Partial | Yes | Runtime detection, limited validation |
| Custom types with `UnmarshalJSON` | Compatible | N/A | Standard library patterns work |

## Type Coercion

gopantic automatically converts between compatible types:

```go
// String to number
type Product struct {
    Price float64 `json:"price"`
}
// {"price": "19.99"} → Product{Price: 19.99}

// Number to string
type Request struct {
    ID string `json:"id"`
}
// {"id": 12345} → Request{ID: "12345"}

// String to boolean
// "true", "false", "1", "0" all work

// Unix timestamp to time.Time
// 1704067200 → parsed time.Time
```

**Performance**: Coercion adds ~5-10% overhead. If input is already correctly typed:

```go
var req Request
json.Unmarshal(body, &req)  // No coercion
model.Validate(&req)         // Just validation
```

## json.RawMessage

Preserve raw JSON for flexible metadata or deferred parsing:

```go
type Account struct {
    ID          string          `json:"id" validate:"required"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

// Parse with validation
account, err := model.ParseInto[Account](input)

// Later: parse metadata as needed
var metadata map[string]interface{}
json.Unmarshal(account.MetadataRaw, &metadata)
```

Common use cases:

- PostgreSQL JSONB columns
- Plugin/extension systems
- Multi-tenant configurations
- Event payloads with varying schemas

## Nested Structs

Recursive validation works automatically:

```go
type Address struct {
    City    string `json:"city" validate:"required"`
    ZipCode string `json:"zip_code" validate:"len=5"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"`
}

// Both User and Address are validated
user, err := model.ParseInto[User](input)
```

Optional nested structs:

```go
type User struct {
    Name    string   `json:"name" validate:"required"`
    Address *Address `json:"address,omitempty"`  // Optional
}
```

## Slices and Arrays

```go
type UserList struct {
    Users []User `json:"users" validate:"required,min=1"`
}

// Each User element is validated
```

Primitive slices work too:

```go
type Tags struct {
    Items []string `json:"items" validate:"min=1"`
}
```

## Pointers (Optional Fields)

Use pointers to distinguish missing from zero values:

```go
type UpdateRequest struct {
    Name  *string `json:"name,omitempty"`   // nil = not provided
    Age   *int    `json:"age,omitempty"`    // 0 vs nil
    Email *string `json:"email,omitempty" validate:"email"`
}

// {"name": "Alice"} → Name="Alice", Age=nil, Email=nil
```

## Custom Types

Implement `UnmarshalJSON` for custom parsing:

```go
type CustomID string

func (c *CustomID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }
    if !strings.HasPrefix(s, "id_") {
        return errors.New("invalid ID format")
    }
    *c = CustomID(s)
    return nil
}

type Request struct {
    ID CustomID `json:"id" validate:"required"`
}
```

## Time Handling

Automatic support for:

- RFC3339: `"2024-01-01T12:00:00Z"`
- Unix timestamps: `1704067200`
- ISO 8601: `"2024-01-01T12:00:00+00:00"`

For custom formats, implement `UnmarshalJSON`:

```go
type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    var s string
    json.Unmarshal(data, &s)
    t, err := time.Parse("2006-01-02", s)
    if err != nil {
        return err
    }
    *ct = CustomTime(t)
    return nil
}
```

## Limitations and Workarounds

### Maps (Limited Validation)

**Problem**: Can't validate individual map values

```go
type Config struct {
    Settings map[string]string `json:"settings" validate:"required"`
    // Only validates existence/length, not individual values
}
```

**Workaround**: Use `json.RawMessage` + manual validation

```go
type Config struct {
    SettingsRaw json.RawMessage `json:"settings,omitempty"`
}

func (c *Config) ValidateSettings() error {
    var settings map[string]string
    json.Unmarshal(c.SettingsRaw, &settings)

    for key, value := range settings {
        if len(value) < 3 {
            return fmt.Errorf("setting %s: value too short", key)
        }
    }
    return nil
}
```

### interface{} (Limited Validation)

**Problem**: Can't validate beyond type checking

```go
type Flexible struct {
    Data interface{} `json:"data" validate:"required"`
    // "required" works, but no deep validation
}
```

**Workaround**: Type assertions

```go
func (f *Flexible) ValidateData() error {
    switch v := f.Data.(type) {
    case string:
        if len(v) < 3 {
            return errors.New("string too short")
        }
    case map[string]interface{}:
        if len(v) == 0 {
            return errors.New("map empty")
        }
    }
    return nil
}
```

### Circular References

**Problem**: Causes infinite loops

```go
type Node struct {
    Children []Node `json:"children"` // OK
    Parent   *Node  `json:"parent"`   // Circular!
}
```

**Workaround**: Use IDs

```go
type Node struct {
    ID       string `json:"id"`
    ParentID string `json:"parent_id,omitempty"`
    Children []Node `json:"children"`
}
```

### Non-String Map Keys

**Problem**: JSON limitation

```go
type Data struct {
    Counts map[int]string `json:"counts"`
    // Keys serialized as strings: {"1": "one"}
}
```

**Workaround**: Use string keys or implement `UnmarshalJSON`

```go
type Data struct {
    Counts map[string]string `json:"counts"`
}
```

### Function/Channel Fields

Must be skipped:

```go
type Valid struct {
    Handler func()   `json:"-"`
    Ch      chan int `json:"-"`
    Name    string   `json:"name"`
}
```

## Edge Cases

### Zero Values vs Missing Fields

JSON can't distinguish:

```go
// {} and {"count": 0} both result in Count: 0
type Request struct {
    Count int `json:"count"`
}
```

Use pointers for optional fields:

```go
type Request struct {
    Count *int `json:"count,omitempty"`
}
// {} → Count: nil
// {"count": 0} → Count: &0
```

### Large json.RawMessage

`json.RawMessage` keeps entire JSON in memory. For large fields (>1MB), consider:

- External storage (S3, database BLOB)
- Streaming/chunked processing
- Compression

### Array Length Validation

Fixed-length array mismatches may not be caught in all cases:

```go
type Data struct {
    Scores [3]float64 `json:"scores"`
}
// {"scores": [1.0]} may not error (known limitation)
```

Use slices with length validation:

```go
type Data struct {
    Scores []float64 `json:"scores" validate:"required,len=3"`
}
```

## Performance Tips

1. **Skip coercion if types are correct**: Use `json.Unmarshal` + `model.Validate()`
2. **Use pointers sparingly**: Only for truly optional fields
3. **Avoid deep nesting**: Flatten structures when possible
4. **json.RawMessage for large/dynamic data**: More efficient than map[string]interface{}
5. **Cache parsed metadata**: Don't unmarshal RawMessage repeatedly

## Design Philosophy

**Type Coercion**: gopantic prioritizes developer convenience over strict typing. This reduces boilerplate when dealing with APIs that send `"123"` instead of `123`.

**Validation Tags**: Simple comma-separated syntax for common cases. Use custom validators for complex rules.

**Compatibility**: Works with standard library patterns (`UnmarshalJSON`, `json.RawMessage`) rather than replacing them.

## See Also

- [API Reference](api.md) - Complete API documentation
- [Migration Guide](../migration.md) - Switching from other libraries
- [Architecture](../architecture.md) - Implementation details
