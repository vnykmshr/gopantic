# Advanced Type Handling

This guide covers how gopantic handles complex and advanced types, including `json.RawMessage`, custom types, nested structures, and edge cases.

## Supported Field Types

gopantic provides comprehensive support for most Go types with intelligent type coercion and validation:

| Type | Validation Support | Type Coercion | Notes |
|------|-------------------|---------------|-------|
| **Primitives** | | | |
| `string`, `int`, `int64`, `int32`, `float64`, `float32`, `bool` | ✅ Full | ✅ Yes | All validation tags work. Strings coerced to numbers/bools automatically |
| **Slices & Arrays** | | | |
| `[]T` (any type) | ✅ Full | ✅ Yes | Element-level validation supported |
| `[N]T` (fixed arrays) | ✅ Full | ✅ Yes | Validated like slices |
| **Structs** | | | |
| Nested structs | ✅ Full | ✅ Yes | Recursive validation supported |
| `*Struct` (pointers) | ✅ Full | ✅ Yes | Nil handling, optional fields |
| **Special Types** | | | |
| `json.RawMessage` | ✅ Full | ✅ Yes | Preserves raw JSON for deferred parsing (v1.2.0+) |
| `time.Time` | ✅ Full | ✅ Yes | RFC3339, Unix timestamps, custom formats |
| **Maps** | | | |
| `map[string]T` | ⚠️ Partial | ⚠️ Limited | Structure validation only, no deep element validation |
| `map[K]V` (non-string keys) | ⚠️ Partial | ⚠️ Limited | Basic support, limited validation |
| **Interfaces** | | | |
| `interface{}` | ⚠️ Partial | ✅ Yes | Type detection at runtime, limited validation |
| Custom `UnmarshalJSON` | ✅ Compatible | N/A | Works seamlessly with standard library patterns |

### Legend
- ✅ **Full**: Complete support with all features
- ⚠️ **Partial**: Works but with limitations
- ❌ **Not Supported**: Use workarounds or standard library

## Type Coercion Examples

gopantic automatically converts between compatible types:

### String to Number
```go
type Product struct {
    Price float64 `json:"price" validate:"required,min=0.01"`
}

// Input: {"price": "19.99"}
// Result: Product{Price: 19.99}
```

### String to Boolean
```go
type Feature struct {
    Enabled bool `json:"enabled"`
}

// Input: {"enabled": "true"}  or {"enabled": "1"}
// Result: Feature{Enabled: true}
```

### Number to String
```go
type Request struct {
    ID string `json:"id" validate:"required"`
}

// Input: {"id": 12345}
// Result: Request{ID: "12345"}
```

### Unix Timestamp to Time
```go
type Event struct {
    Timestamp time.Time `json:"timestamp"`
}

// Input: {"timestamp": 1704067200}
// Result: Event{Timestamp: time.Time (parsed from Unix epoch)}
```

## json.RawMessage Support

**New in v1.2.0**: Full support for `json.RawMessage` fields, enabling flexible metadata storage and PostgreSQL JSONB integration.

### Basic Usage

```go
type Account struct {
    ID          string          `json:"id" validate:"required"`
    Name        string          `json:"name" validate:"required,min=2"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

input := []byte(`{
    "id": "acc_123",
    "name": "John Doe",
    "metadata": {"preferences": {"theme": "dark"}, "tags": ["vip"]}
}`)

account, err := model.ParseInto[Account](input)
// MetadataRaw preserves the raw JSON: {"preferences": {"theme": "dark"}, "tags": ["vip"]}
```

### Deferred Parsing Pattern

```go
type Request struct {
    Name        string          `json:"name" validate:"required"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

// Parse request with validation
req, err := model.ParseInto[Request](body)
if err != nil {
    return err
}

// Later: parse metadata based on context
var metadata map[string]interface{}
if len(req.MetadataRaw) > 0 {
    json.Unmarshal(req.MetadataRaw, &metadata)
}
```

### Multiple RawMessage Fields

```go
type ComplexRequest struct {
    Name      string          `json:"name" validate:"required"`
    Config    json.RawMessage `json:"config,omitempty"`
    Metadata  json.RawMessage `json:"metadata,omitempty"`
    ExtraData json.RawMessage `json:"extra_data,omitempty"`
}

// All RawMessage fields are preserved independently
req, err := model.ParseInto[ComplexRequest](input)
```

## Nested Struct Validation

gopantic recursively validates nested structs:

```go
type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    ZipCode string `json:"zip_code" validate:"required,length=5"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"`
}

input := []byte(`{
    "name": "Alice",
    "address": {
        "street": "123 Main St",
        "city": "Boston",
        "zip_code": "02101"
    }
}`)

user, err := model.ParseInto[User](input)
// Both User and Address fields are validated
```

### Nested Pointers

```go
type User struct {
    Name    string   `json:"name" validate:"required"`
    Address *Address `json:"address,omitempty"`  // Optional nested struct
}

// Works with nil addresses
input := []byte(`{"name": "Alice"}`)
user, err := model.ParseInto[User](input)  // user.Address == nil
```

## Slice and Array Validation

### Element Validation

```go
type UserList struct {
    Users []User `json:"users" validate:"required,min=1"`
}

// Validates each User in the slice
input := []byte(`{
    "users": [
        {"name": "Alice", "email": "alice@example.com"},
        {"name": "Bob", "email": "bob@example.com"}
    ]
}`)

userList, err := model.ParseInto[UserList](input)
```

### Primitive Slices

```go
type Tags struct {
    Items []string `json:"items" validate:"required,min=1"`
}

// Input: {"items": ["tag1", "tag2"]}
// Validates slice length, each element parsed
```

## Map Handling

Maps have limited validation support but are parsed correctly:

### Supported Map Patterns

```go
type Config struct {
    Settings map[string]string `json:"settings"`
}

// ✅ Parsing works
input := []byte(`{"settings": {"key1": "value1", "key2": "value2"}}`)
config, err := model.ParseInto[Config](input)

// ⚠️ Limited validation - can't validate individual map values
```

### Workaround: Use json.RawMessage

For complex map validation, use `json.RawMessage` with manual validation:

```go
type Request struct {
    Name        string          `json:"name" validate:"required"`
    SettingsRaw json.RawMessage `json:"settings,omitempty"`
}

func (r *Request) ValidateSettings() error {
    var settings map[string]interface{}
    if err := json.Unmarshal(r.SettingsRaw, &settings); err != nil {
        return err
    }

    // Custom validation logic
    if env, ok := settings["environment"].(string); ok {
        if env != "prod" && env != "dev" {
            return errors.New("invalid environment")
        }
    }

    return nil
}
```

## Custom Types with UnmarshalJSON

gopantic works seamlessly with types implementing `json.Unmarshaler`:

```go
type CustomID string

func (c *CustomID) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }

    // Custom parsing logic
    if !strings.HasPrefix(s, "id_") {
        return errors.New("invalid ID format")
    }

    *c = CustomID(s)
    return nil
}

type Request struct {
    ID CustomID `json:"id" validate:"required"`
}

// gopantic uses your custom UnmarshalJSON method
req, err := model.ParseInto[Request]([]byte(`{"id": "id_12345"}`))
```

## Pointer Fields (Optional Values)

Use pointers for optional fields that need to distinguish between zero values and absence:

```go
type UpdateRequest struct {
    Name  *string `json:"name,omitempty"`   // nil = not provided, "" = empty string
    Age   *int    `json:"age,omitempty"`    // nil = not provided, 0 = zero
    Email *string `json:"email,omitempty" validate:"email"`  // Validated only if present
}

// Partial update example
input := []byte(`{"name": "Alice"}`)
req, err := model.ParseInto[UpdateRequest](input)
// req.Name = "Alice", req.Age = nil, req.Email = nil
```

## Interface{} Fields

Fields typed as `interface{}` are parsed but have limited validation:

```go
type Flexible struct {
    Name string      `json:"name" validate:"required"`
    Data interface{} `json:"data"`  // Any JSON value accepted
}

// Works with any valid JSON
input := []byte(`{"name": "test", "data": {"nested": "object"}}`)
flex, err := model.ParseInto[Flexible](input)
```

## Time Handling

gopantic supports multiple time formats:

### Supported Formats

```go
type Event struct {
    Timestamp time.Time `json:"timestamp"`
}

// RFC3339
{"timestamp": "2024-01-01T12:00:00Z"}

// Unix timestamp (seconds)
{"timestamp": 1704067200}

// ISO 8601
{"timestamp": "2024-01-01T12:00:00+00:00"}
```

### Custom Time Formats

For custom formats, implement `UnmarshalJSON`:

```go
type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    var s string
    if err := json.Unmarshal(data, &s); err != nil {
        return err
    }

    t, err := time.Parse("2006-01-02", s)  // Custom format
    if err != nil {
        return err
    }

    *ct = CustomTime(t)
    return nil
}
```

## Edge Cases and Limitations

### Maps with Non-String Keys

Limited support. Consider using `json.RawMessage` instead:

```go
// ⚠️ Limited support
type Request struct {
    Data map[int]string `json:"data"`
}

// ✅ Better approach
type Request struct {
    DataRaw json.RawMessage `json:"data"`
}
```

### Circular References

Not supported. Avoid circular struct references:

```go
// ❌ Don't do this
type Node struct {
    Value    string `json:"value"`
    Children []Node `json:"children"`  // OK
    Parent   *Node  `json:"parent"`    // ❌ Circular reference
}
```

### Function and Channel Fields

Not supported (not JSON-serializable):

```go
// ❌ Won't work
type Invalid struct {
    Handler func() `json:"-"`  // Must skip with json:"-"
    Ch      chan int `json:"-"`
}
```

## Performance Considerations

### Type Coercion Overhead

Type coercion adds minimal overhead (~5-10% vs standard `json.Unmarshal`). For hot paths, consider:

```go
// If you know types are correct in input, use standard library
var req Request
json.Unmarshal(body, &req)
model.Validate(&req)  // Validate only

// vs.

// All-in-one (parsing + coercion + validation)
req, err := model.ParseInto[Request](body)
```

### Large json.RawMessage Fields

`json.RawMessage` preserves the entire raw JSON in memory. For very large metadata fields, consider streaming or external storage.

## See Also

- [Database Integration Guide](database-integration.md) - PostgreSQL JSONB patterns
- [API Reference](api.md) - Complete API documentation
- [Migration Guide](migration.md) - Migrating from encoding/json
- [Limitations](limitations.md) - Current limitations and workarounds
