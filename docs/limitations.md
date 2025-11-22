# Limitations and Known Issues

This document outlines the current limitations of gopantic and provides workarounds where available.

## Current Limitations

### 1. Map Field Validation

**Status**: Partial Support

Maps are parsed correctly but have limited validation capabilities:

```go
// ✅ Parsing works
type Config struct {
    Settings map[string]string `json:"settings"`
}

// ⚠️ Limited validation - can't validate individual map values
type Config struct {
    Settings map[string]string `json:"settings" validate:"required,min=1"`
    // Only validates map existence and length, not individual key/value pairs
}
```

**Workaround**: Use `json.RawMessage` with manual validation:

```go
type Config struct {
    SettingsRaw json.RawMessage `json:"settings,omitempty"`
}

func (c *Config) ValidateSettings() error {
    var settings map[string]string
    if err := json.Unmarshal(c.SettingsRaw, &settings); err != nil {
        return err
    }

    // Custom validation logic
    for key, value := range settings {
        if len(value) < 3 {
            return fmt.Errorf("setting %s: value too short", key)
        }
    }

    return nil
}
```

### 2. Interface{} Field Validation

**Status**: Partial Support

Fields typed as `interface{}` are parsed but cannot be validated beyond type checking:

```go
// ✅ Parsing works
type Flexible struct {
    Data interface{} `json:"data"`
}

// ❌ Validation not supported
type Flexible struct {
    Data interface{} `json:"data" validate:"required"`
    // "required" check works, but no deep validation
}
```

**Workaround**: Use type assertions and manual validation:

```go
func (f *Flexible) ValidateData() error {
    if f.Data == nil {
        return errors.New("data is required")
    }

    // Type-specific validation
    switch v := f.Data.(type) {
    case string:
        if len(v) < 3 {
            return errors.New("string data too short")
        }
    case map[string]interface{}:
        if len(v) == 0 {
            return errors.New("map data is empty")
        }
    default:
        return fmt.Errorf("unsupported data type: %T", v)
    }

    return nil
}
```

### 3. Circular References

**Status**: Not Supported

Circular struct references will cause infinite loops:

```go
// ❌ Don't do this
type Node struct {
    Value    string `json:"value"`
    Children []Node `json:"children"` // OK - not circular
    Parent   *Node  `json:"parent"`   // ❌ Creates circular reference
}
```

**Workaround**: Break circular references using IDs:

```go
// ✅ Use IDs instead
type Node struct {
    ID       string `json:"id"`
    ParentID string `json:"parent_id,omitempty"`
    Value    string `json:"value"`
    Children []Node `json:"children"`
}
```

### 4. Custom Validator Parameters with Special Characters

**Status**: Limited Support

Validator parameters that contain commas or equals signs require escaping:

```go
// ❌ Problem: comma in parameter
type User struct {
    Name string `json:"name" validate:"oneof=Alice,Bob,Charlie"`
    // Parser may misinterpret commas
}
```

**Workaround**: Use custom validators for complex parameter sets:

```go
model.RegisterGlobalFunc("valid_name", func(fieldName string, fieldValue interface{}, params map[string]interface{}) error {
    name := fieldValue.(string)
    validNames := []string{"Alice", "Bob", "Charlie"}

    for _, valid := range validNames {
        if name == valid {
            return nil
        }
    }

    return model.NewValidationError(fieldName, fieldValue, "valid_name", "name must be Alice, Bob, or Charlie")
})

type User struct {
    Name string `json:"name" validate:"required,valid_name"`
}
```

### 5. Deep Dive Validation for Nested Collections

**Status**: Limited Support

While nested structs are fully validated, deeply nested slices of maps have limitations:

```go
// ✅ Nested structs work perfectly
type Company struct {
    Departments []Department `json:"departments"`
}

type Department struct {
    Name      string   `json:"name" validate:"required"`
    Employees []Person `json:"employees"`
}

// ⚠️ Limited: Slice of maps
type Data struct {
    Records []map[string]interface{} `json:"records"`
    // Can't validate individual map contents
}
```

**Workaround**: Define proper struct types:

```go
// ✅ Better approach
type Data struct {
    Records []Record `json:"records"`
}

type Record struct {
    ID    string `json:"id" validate:"required"`
    Value string `json:"value" validate:"required"`
}
```

### 6. Non-String Map Keys

**Status**: Limited Support

Maps with non-string keys have limited JSON support (this is a JSON limitation, not gopantic-specific):

```go
// ⚠️ JSON doesn't support non-string keys natively
type Data struct {
    Counts map[int]string `json:"counts"`
}

// JSON will serialize keys as strings: {"1": "one", "2": "two"}
```

**Workaround**: Use string keys or custom marshal/unmarshal:

```go
// Option 1: Use string keys
type Data struct {
    Counts map[string]string `json:"counts"`
}

// Option 2: Custom type with UnmarshalJSON
type IntKeyMap map[int]string

func (m *IntKeyMap) UnmarshalJSON(data []byte) error {
    var strMap map[string]string
    if err := json.Unmarshal(data, &strMap); err != nil {
        return err
    }

    *m = make(map[int]string)
    for k, v := range strMap {
        intKey, err := strconv.Atoi(k)
        if err != nil {
            return err
        }
        (*m)[intKey] = v
    }

    return nil
}
```

### 7. Function and Channel Fields

**Status**: Not Supported

Functions and channels are not JSON-serializable:

```go
// ❌ Won't work
type Invalid struct {
    Handler func()   `json:"handler"`
    Ch      chan int `json:"ch"`
}
```

**Workaround**: Skip these fields with `json:"-"`:

```go
// ✅ Correct approach
type Valid struct {
    Handler func()   `json:"-"` // Skipped by JSON
    Ch      chan int `json:"-"` // Skipped by JSON
    Name    string   `json:"name"`
}
```

## Previously Fixed Limitations

### ✅ json.RawMessage Support (Fixed in v1.2.0)

**Previous Issue**: `json.RawMessage` fields caused "cannot coerce map to slice" errors.

**Status**: Fully Supported (as of v1.2.0)

```go
// ✅ Now works perfectly
type Request struct {
    Name        string          `json:"name" validate:"required"`
    MetadataRaw json.RawMessage `json:"metadata,omitempty"`
}

req, err := model.ParseInto[Request](body)
// MetadataRaw is preserved correctly
```

See [Issue #10](https://github.com/vnykmshr/gopantic/issues/10) for details.

### ✅ Nested Struct Validation (Fixed in v1.2.0)

**Previous Issue**: Validation was not applied recursively to nested structs.

**Status**: Fully Supported (as of v1.2.0)

```go
// ✅ Nested validation now works
type Address struct {
    Street string `json:"street" validate:"required"`
    City   string `json:"city" validate:"required"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address" validate:"required"`
}

// Both User and Address fields are validated
user, err := model.ParseInto[User](input)
```

### ✅ Standalone Validation (Added in v1.2.0)

**Previous Issue**: No way to validate structs independently of parsing.

**Status**: Fully Supported (as of v1.2.0)

```go
// ✅ Validate structs from any source
var user User
json.Unmarshal(body, &user)
err := model.Validate(&user) // Independent validation
```

See [Issue #11](https://github.com/vnykmshr/gopantic/issues/11) for details.

## Design Trade-offs

### Type Coercion vs. Strict Typing

gopantic prioritizes developer ergonomics by automatically coercing compatible types (`"123"` → `123`). This differs from strict validation libraries that reject type mismatches.

**Why**: API clients often send numbers as strings (form data, URL params). Automatic coercion reduces boilerplate.

**Alternative**: Use standard `json.Unmarshal` + `model.Validate()` for strict typing:

```go
// Strict typing (no coercion)
var req Request
if err := json.Unmarshal(body, &req); err != nil {
    return err  // Fails on type mismatch
}

// Validation only
if err := model.Validate(&req); err != nil {
    return err
}
```

### Validation Tag Syntax

gopantic uses comma-separated validation tags (`validate:"required,min=5"`). This is simpler than some alternatives but limits parameter flexibility.

**Why**: Easier to parse and less error-prone for common cases.

**Alternative**: Use custom validators for complex rules:

```go
model.RegisterGlobalFunc("complex_rule", func(...) error {
    // Complex validation logic here
})
```

## Performance Characteristics

### Type Coercion Overhead

Automatic type coercion adds ~5-10% overhead vs. standard `json.Unmarshal`.

**When it matters**: High-throughput APIs (>10k req/s) with correctly-typed inputs.

**Mitigation**: Use the hybrid approach for hot paths:

```go
// Hot path: skip coercion
var req Request
json.Unmarshal(body, &req)
model.Validate(&req)

// Cold path: full parsing with coercion
req, err := model.ParseInto[Request](body)
```

### Validation Metadata Caching

Validation rules are cached per struct type using `sync.Map`. First validation per type is slower (~10-20%), subsequent validations benefit from cached metadata.

**Impact**: Negligible for long-running services. May be noticeable in short-lived CLI tools.

### Large json.RawMessage Fields

`json.RawMessage` preserves the entire raw JSON in memory. For very large metadata fields (>1MB), consider:

1. External storage (S3, disk)
2. Streaming/chunked processing
3. Database BLOB columns instead of JSONB

## Comparison with Other Libraries

### vs. go-playground/validator

**gopantic advantages**:
- Integrated parsing + validation
- Automatic type coercion
- Format detection (JSON/YAML)

**go-playground/validator advantages**:
- More built-in validators
- More flexible tag syntax
- Deeper map/slice validation

**Best for**: gopantic is best for API request validation; go-playground/validator is best for pure validation logic.

### vs. encoding/json

**gopantic advantages**:
- Built-in validation
- Type coercion
- Better error messages

**encoding/json advantages**:
- Standard library (no dependencies)
- Maximum performance
- Wider ecosystem support

**Best for**: gopantic is best for APIs with validation needs; encoding/json is best for pure parsing.

## Known Edge Cases

### Time Parsing Ambiguity

When multiple time formats could match, gopantic uses heuristics:

```go
// "2024-01-01T12:00:00Z" - parsed as RFC3339
// "1704067200" - parsed as Unix timestamp
// "2024-01-01" - may fail (use custom time type)
```

For custom formats, implement `UnmarshalJSON`:

```go
type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(data []byte) error {
    // Custom parsing logic
}
```

### Zero Values vs. Omitted Fields

JSON doesn't distinguish between zero values and omitted fields:

```go
type Request struct {
    Count int `json:"count"`
}

// Input: {}
// Result: Request{Count: 0}

// Input: {"count": 0}
// Result: Request{Count: 0}
// Both are indistinguishable
```

**Workaround**: Use pointers for optional fields:

```go
type Request struct {
    Count *int `json:"count,omitempty"`
}

// Input: {}
// Result: Request{Count: nil}

// Input: {"count": 0}
// Result: Request{Count: &0}
```

## Future Improvements

Potential enhancements under consideration:

1. **Deep map validation** - Validate map[string]T with element-level rules
2. **Conditional validation** - Validate field A only if field B has specific value
3. **Custom error formats** - Pluggable error formatters (JSON, XML, etc.)
4. **Validation groups** - Apply different validation rules based on context
5. **Async validation** - Support for validators that make external calls

See [GitHub Issues](https://github.com/vnykmshr/gopantic/issues) for active discussions.

## Reporting Issues

If you encounter unexpected behavior:

1. Check this limitations document
2. Review [GitHub Issues](https://github.com/vnykmshr/gopantic/issues)
3. Open a new issue with:
   - Go version
   - gopantic version
   - Minimal reproduction code
   - Expected vs. actual behavior

## See Also

- [Advanced Type Handling](advanced-types.md) - Complex type patterns
- [Database Integration](database-integration.md) - PostgreSQL JSONB patterns
- [API Reference](api.md) - Complete API documentation
- [Migration Guide](migration.md) - Migrating from other libraries
