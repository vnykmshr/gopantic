# Validation Guide

## Basic Syntax

Use the `validate` struct tag:

```go
type User struct {
    Name  string `json:"name" validate:"required,min=2,max=50"`
    Email string `json:"email" validate:"required,email"`
}
```

Multiple validators are comma-separated. Validators are applied in order.

## Built-in Validators

gopantic includes these built-in validators:

### Presence

| Validator | Description | Example |
|-----------|-------------|---------|
| `required` | Field must be non-zero value | `validate:"required"` |

```go
Name string `json:"name" validate:"required"`
// "" fails, "Alice" passes
```

### Range

| Validator | Description | Example |
|-----------|-------------|---------|
| `min` | Minimum value/length | `validate:"min=5"` |
| `max` | Maximum value/length | `validate:"max=100"` |
| `length` | Exact length (strings only) | `validate:"length=10"` |

For strings, `min`/`max` check length. For numbers, they check value.

```go
Age     int    `json:"age" validate:"min=0,max=150"`      // 0 <= age <= 150
Name    string `json:"name" validate:"min=2,max=50"`      // 2 <= len(name) <= 50
Code    string `json:"code" validate:"length=6"`          // len(code) == 6
```

### String Formats

| Validator | Description | Example |
|-----------|-------------|---------|
| `email` | Valid email format | `validate:"email"` |
| `alpha` | Letters only (a-zA-Z) | `validate:"alpha"` |
| `alphanum` | Letters and numbers only | `validate:"alphanum"` |

```go
Email   string `json:"email" validate:"required,email"`
Country string `json:"country" validate:"alpha"`
Code    string `json:"code" validate:"alphanum"`
```

## Nested Struct Validation

Nested structs are validated automatically:

```go
type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    ZipCode string `json:"zip_code" validate:"required,alphanum"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address"`  // Nested struct is validated
}
```

## Slice Validation

Slices can be validated for length:

```go
Tags []string `json:"tags" validate:"min=1,max=5"`  // 1-5 items
```

## Custom Validators

Register custom validation functions for domain-specific rules:

```go
model.RegisterGlobalFunc("is_even", func(fieldName string, value interface{}, params map[string]interface{}) error {
    num, ok := value.(int)
    if !ok {
        return nil // Let type validation handle this
    }
    if num%2 != 0 {
        return model.NewValidationError(fieldName, value, "is_even", "must be an even number")
    }
    return nil
})

type Numbers struct {
    EvenNumber int `json:"even_number" validate:"required,is_even"`
}
```

### Custom Cross-Field Validators

For validations that compare fields against each other:

```go
model.RegisterGlobalCrossFieldFunc("password_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
    confirmPassword, ok := fieldValue.(string)
    if !ok {
        return model.NewValidationError(fieldName, fieldValue, "password_match", "must be a string")
    }

    password := structValue.FieldByName("Password").String()
    if confirmPassword != password {
        return model.NewValidationError(fieldName, fieldValue, "password_match", "passwords do not match")
    }
    return nil
})

type Registration struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,password_match"`
}
```

## Validation Errors

Errors include field names and failure reasons:

```go
user, err := model.ParseInto[User](data)
if err != nil {
    // "validation error on field 'email': invalid email format"
    // "validation error on field 'age': value 5 is less than minimum 18"
}
```

### Error Types

```go
// Check for specific error types
var parseErr *model.ParseError
if errors.As(err, &parseErr) {
    // JSON/YAML parsing failed
}

var validErr *model.ValidationError
if errors.As(err, &validErr) {
    // Validation rule failed
}
```

### Sensitive Field Protection

Sensitive field values are automatically redacted in error output:

```go
type Login struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required,min=8"`
}

// If password validation fails, the error won't contain the actual password
// The value will show as "[REDACTED]" in error reports
```

## Tips

1. **Order matters**: `validate:"required,email"` checks required first
2. **Empty strings**: An empty string passes `min=0` but fails `required`
3. **Nil slices**: A nil slice fails `required` but passes `min=0`
4. **Performance**: Validation metadata is cached per type

## Common Patterns

### Password with Confirmation

```go
// Register a custom cross-field validator
model.RegisterGlobalCrossFieldFunc("eqfield", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
    otherField := params["value"].(string)
    otherValue := structValue.FieldByName(otherField)
    if !otherValue.IsValid() {
        return model.NewValidationError(fieldName, fieldValue, "eqfield", "comparison field not found")
    }
    if fieldValue != otherValue.Interface() {
        return model.NewValidationError(fieldName, fieldValue, "eqfield", "fields do not match")
    }
    return nil
})

type Registration struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}
```

### Optional with Format

```go
// Use custom validator for optional + format
model.RegisterGlobalFunc("optional_email", func(fieldName string, value interface{}, params map[string]interface{}) error {
    str, ok := value.(string)
    if !ok || str == "" {
        return nil // Empty is OK
    }
    // Validate email format if non-empty
    if !strings.Contains(str, "@") {
        return model.NewValidationError(fieldName, value, "optional_email", "invalid email format")
    }
    return nil
})

Email string `json:"email" validate:"optional_email"`
```

### Phone Number

```go
Phone string `json:"phone" validate:"required,alphanum,min=10,max=15"`
```

## Extending Validators

For validators not included by default (like `url`, `uuid`, `oneof`), register custom implementations:

```go
// Example: oneof validator
model.RegisterGlobalFunc("oneof", func(fieldName string, value interface{}, params map[string]interface{}) error {
    str, ok := value.(string)
    if !ok {
        return nil
    }
    allowed := strings.Split(params["value"].(string), " ")
    for _, v := range allowed {
        if str == v {
            return nil
        }
    }
    return model.NewValidationError(fieldName, value, "oneof",
        fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
})

Status string `json:"status" validate:"oneof=draft published archived"`
```
