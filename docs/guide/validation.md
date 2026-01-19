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

### Presence

| Validator | Description | Example |
|-----------|-------------|---------|
| `required` | Field must be non-zero value | `validate:"required"` |

```go
Name string `json:"name" validate:"required"`
// "" fails, "Alice" passes
```

### Comparison

| Validator | Description | Example |
|-----------|-------------|---------|
| `min` | Minimum value/length | `validate:"min=5"` |
| `max` | Maximum value/length | `validate:"max=100"` |
| `len` | Exact length | `validate:"len=10"` |
| `eq` | Equals value | `validate:"eq=active"` |
| `ne` | Not equals value | `validate:"ne=deleted"` |
| `gt` | Greater than | `validate:"gt=0"` |
| `gte` | Greater than or equal | `validate:"gte=1"` |
| `lt` | Less than | `validate:"lt=100"` |
| `lte` | Less than or equal | `validate:"lte=99"` |

For strings, `min`/`max`/`len` check length. For numbers, they check value.

```go
Age     int    `json:"age" validate:"min=0,max=150"`      // 0 <= age <= 150
Name    string `json:"name" validate:"min=2,max=50"`      // 2 <= len(name) <= 50
Status  string `json:"status" validate:"eq=active"`       // must be "active"
Balance int    `json:"balance" validate:"gte=0"`          // must be >= 0
```

### String Formats

| Validator | Description | Example |
|-----------|-------------|---------|
| `email` | Valid email format | `validate:"email"` |
| `url` | Valid URL | `validate:"url"` |
| `uuid` | Valid UUID | `validate:"uuid"` |
| `alpha` | Letters only | `validate:"alpha"` |
| `alphanum` | Letters and numbers only | `validate:"alphanum"` |
| `numeric` | Numeric string | `validate:"numeric"` |

```go
Email   string `json:"email" validate:"required,email"`
Website string `json:"website" validate:"url"`
ID      string `json:"id" validate:"uuid"`
Code    string `json:"code" validate:"alphanum,len=6"`
```

### Choice

| Validator | Description | Example |
|-----------|-------------|---------|
| `oneof` | Must be one of listed values | `validate:"oneof=draft published archived"` |

```go
Status string `json:"status" validate:"required,oneof=draft published archived"`
// Only "draft", "published", or "archived" are valid
```

### Slice/Array

| Validator | Description | Example |
|-----------|-------------|---------|
| `min` | Minimum length | `validate:"min=1"` |
| `max` | Maximum length | `validate:"max=10"` |
| `dive` | Validate each element | `validate:"dive,required"` |

```go
Tags []string `json:"tags" validate:"min=1,max=5"`           // 1-5 tags
IDs  []int    `json:"ids" validate:"required,dive,min=1"`    // each ID >= 1
```

## Cross-Field Validation

Compare one field against another:

| Validator | Description | Example |
|-----------|-------------|---------|
| `eqfield` | Equal to other field | `validate:"eqfield=ConfirmPassword"` |
| `nefield` | Not equal to other field | `validate:"nefield=OldPassword"` |
| `gtfield` | Greater than other field | `validate:"gtfield=MinValue"` |
| `gtefield` | Greater than or equal to other field | `validate:"gtefield=StartDate"` |
| `ltfield` | Less than other field | `validate:"ltfield=MaxValue"` |
| `ltefield` | Less than or equal to other field | `validate:"ltefield=EndDate"` |

```go
type Registration struct {
    Password        string `json:"password" validate:"required,min=8"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type DateRange struct {
    StartDate time.Time `json:"start_date" validate:"required"`
    EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
}

type PriceRange struct {
    MinPrice float64 `json:"min_price" validate:"min=0"`
    MaxPrice float64 `json:"max_price" validate:"gtefield=MinPrice"`
}
```

## Nested Struct Validation

Nested structs are validated automatically:

```go
type Address struct {
    Street  string `json:"street" validate:"required"`
    City    string `json:"city" validate:"required"`
    ZipCode string `json:"zip_code" validate:"required,len=5"`
}

type User struct {
    Name    string  `json:"name" validate:"required"`
    Address Address `json:"address"`  // Nested struct is validated
}
```

## Custom Validators

Register custom validation functions:

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
    // "field 'email' validation failed: invalid email format"
    // "field 'age' validation failed: min validation failed, got 5, expected >= 18"
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

## Tips

1. **Order matters**: `validate:"required,email"` checks required first
2. **Empty strings**: An empty string passes `min=0` but fails `required`
3. **Nil slices**: A nil slice fails `required` but passes `min=0`
4. **Performance**: Validation metadata is cached per type

## Common Patterns

### Optional Email

```go
Email string `json:"email" validate:"omitempty,email"`
// Empty is OK, but if provided must be valid email
```

### Password Complexity

```go
Password string `json:"password" validate:"required,min=8,max=72"`
```

### Phone Number

```go
Phone string `json:"phone" validate:"required,numeric,len=10"`
```

### Date Range

```go
StartDate time.Time `json:"start_date" validate:"required"`
EndDate   time.Time `json:"end_date" validate:"required,gtfield=StartDate"`
```
