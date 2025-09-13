# Migration Guide

This guide helps you migrate from standard library JSON parsing to gopantic, and between different versions of gopantic.

## Table of Contents

- [Migrating from Standard Library JSON](#migrating-from-standard-library-json)
- [Performance Optimization Migration](#performance-optimization-migration)
- [Caching Migration](#caching-migration)
- [Version-Specific Migrations](#version-specific-migrations)
- [Common Migration Patterns](#common-migration-patterns)

## Migrating from Standard Library JSON

### Basic Parsing Migration

**Before (Standard Library):**
```go
import "encoding/json"

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Email string `json:"email"`
}

var user User
if err := json.Unmarshal(data, &user); err != nil {
    return err
}
```

**After (gopantic):**
```go
import "github.com/vnykmshr/gopantic/pkg/model"

type User struct {
    ID   int    `json:"id" validate:"required,min=1"`
    Name string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

user, err := model.ParseInto[User](data)
if err != nil {
    return err
}
```

### Key Benefits of Migration

1. **Automatic Validation**: Add `validate` tags for built-in validation
2. **Type Coercion**: Automatic conversion between compatible types
3. **Better Error Messages**: Structured error reporting with field paths
4. **YAML Support**: Parse YAML with the same API
5. **Performance**: Optional caching and optimization features

### Validation Migration

**Before (Manual Validation):**
```go
var user User
if err := json.Unmarshal(data, &user); err != nil {
    return err
}

// Manual validation
if user.ID <= 0 {
    return errors.New("ID must be positive")
}
if len(user.Name) < 2 {
    return errors.New("name too short")
}
if !isValidEmail(user.Email) {
    return errors.New("invalid email")
}
```

**After (gopantic with Validation):**
```go
type User struct {
    ID   int    `json:"id" validate:"required,min=1"`
    Name string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

user, err := model.ParseInto[User](data)
// Validation happens automatically during parsing
if err != nil {
    // Handle structured validation errors
    if errorList, ok := err.(model.ErrorList); ok {
        for _, e := range errorList {
            fmt.Printf("Validation error: %v\n", e)
        }
    }
    return err
}
```

## Performance Optimization Migration

### From Basic to Optimized Parsing

**Basic Usage:**
```go
user, err := model.ParseInto[User](data)
```

**Optimized for High-Throughput:**
```go
user, err := model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)
```

**Benefits:**
- 17-45% performance improvement
- Cached reflection operations
- Reduced memory allocations

### Memory-Optimized Parsing

**For Memory-Constrained Environments:**
```go
user, err := model.PooledParseIntoWithFormat[User](data, model.FormatJSON)
```

**Benefits:**
- Object pooling reduces GC pressure
- Reuses allocated memory
- Ideal for high-frequency parsing

## Caching Migration

### Adding Caching to Existing Code

**Before (No Caching):**
```go
for _, data := range datasets {
    user, err := model.ParseInto[User](data)
    // Process user...
}
```

**After (With Caching):**
```go
// Option 1: Global cache functions
for _, data := range datasets {
    user, err := model.ParseIntoCached[User](data)
    // Process user...
}

// Option 2: Explicit cache configuration
parser, err := model.NewCachedParser[User](model.DefaultCacheConfig())
if err != nil {
    return err
}
defer parser.Close()

for _, data := range datasets {
    user, err := parser.Parse(data)
    // Process user...
}
```

### Redis Distributed Caching

**For Multi-Instance Applications:**
```go
config := model.DefaultRedisCacheConfig("redis:6379")
parser, err := model.NewCachedParser[User](config)
if err != nil {
    return err
}
defer parser.Close()

user, err := parser.Parse(data)
```

## Version-Specific Migrations

### Upgrading to v1.0+

#### Breaking Changes
1. **Error Types**: `ErrorList` now implements better serialization
2. **Cache Configuration**: New `CacheConfig` structure
3. **Validation Tags**: Enhanced validation tag syntax

#### Migration Steps

**Update Error Handling:**
```go
// Before v1.0
if err != nil {
    fmt.Printf("Error: %v", err)
}

// After v1.0
if err != nil {
    if errorList, ok := err.(model.ErrorList); ok {
        jsonData, _ := errorList.ToJSON()
        fmt.Printf("Structured errors: %s", jsonData)
    }
}
```

**Update Cache Configuration:**
```go
// Before v1.0
parser := model.NewCachedParser[User]()

// After v1.0
config := model.DefaultCacheConfig()
parser, err := model.NewCachedParser[User](config)
if err != nil {
    return err
}
defer parser.Close()
```

## Common Migration Patterns

### Pattern 1: Gradual Migration

Migrate your codebase incrementally by replacing JSON parsing calls:

```go
// Step 1: Replace basic parsing
// json.Unmarshal(data, &user) → model.ParseInto[User](data)

// Step 2: Add validation tags
type User struct {
    ID   int    `json:"id" validate:"required,min=1"`
    Name string `json:"name" validate:"required"`
}

// Step 3: Add caching for frequently parsed types
user, err := model.ParseIntoCached[User](data)

// Step 4: Use optimized parsing for high-throughput scenarios
user, err := model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)
```

### Pattern 2: Configuration-Based Migration

Create a configuration system for different parsing strategies:

```go
type ParsingConfig struct {
    UseCache     bool
    UseOptimized bool
    UsePooling   bool
}

func parseUser(data []byte, config ParsingConfig) (User, error) {
    switch {
    case config.UsePooling:
        return model.PooledParseIntoWithFormat[User](data, model.FormatJSON)
    case config.UseOptimized:
        return model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)
    case config.UseCache:
        return model.ParseIntoCached[User](data)
    default:
        return model.ParseInto[User](data)
    }
}
```

### Pattern 3: Error Handling Migration

Enhance your error handling to take advantage of structured errors:

```go
func handleParsingError(err error) {
    if err == nil {
        return
    }

    if errorList, ok := err.(model.ErrorList); ok {
        // Group errors by field
        grouped := errorList.GroupByField()
        for fieldPath, errors := range grouped {
            fmt.Printf("Field %s has %d validation errors:\n", fieldPath, len(errors))
            for _, e := range errors {
                if validationErr, ok := e.(*model.ValidationError); ok {
                    fmt.Printf("  - %s: %s\n", validationErr.Rule, validationErr.Message)
                }
            }
        }
    } else {
        fmt.Printf("Parse error: %v\n", err)
    }
}
```

## Migration Checklist

### Pre-Migration
- [ ] Identify all JSON parsing locations in your codebase
- [ ] Document current validation logic
- [ ] Measure baseline performance for critical paths
- [ ] Plan gradual rollout strategy

### During Migration
- [ ] Add gopantic dependency: `go get github.com/vnykmshr/gopantic`
- [ ] Replace `json.Unmarshal` calls with `model.ParseInto`
- [ ] Add validation tags to struct definitions
- [ ] Update error handling to use structured errors
- [ ] Add caching for frequently parsed types
- [ ] Use optimized parsing for performance-critical paths

### Post-Migration
- [ ] Remove manual validation code
- [ ] Measure performance improvements
- [ ] Update documentation and examples
- [ ] Monitor error rates and performance metrics
- [ ] Consider adding YAML support where beneficial

## Performance Considerations

### Choosing the Right Parsing Method

| Use Case | Recommended Method | Performance Gain |
|----------|-------------------|------------------|
| Basic parsing | `ParseInto` | Baseline |
| High-throughput | `OptimizedParseIntoWithFormat` | 17-45% faster |
| Memory-constrained | `PooledParseIntoWithFormat` | Reduced GC pressure |
| Repeated parsing | `ParseIntoCached` | 2-10x faster for cache hits |
| Distributed systems | Redis-backed cache | Shared cache benefits |

### Memory Usage Guidelines

1. **Use pooling** for applications with high parsing frequency
2. **Use caching** for applications that parse the same data repeatedly
3. **Use optimization** for applications with strict performance requirements
4. **Monitor metrics** using `model.GlobalMetrics.GetStats()`

## Troubleshooting Common Issues

### Issue 1: Validation Errors After Migration

**Problem:** Existing data fails validation with new tags.

**Solution:** Start with lenient validation and gradually tighten:
```go
// Start with basic validation
type User struct {
    ID   int    `json:"id" validate:"min=1"`  // Allow some flexibility
    Name string `json:"name" validate:"min=1"` // Start with min=1, not min=2
}

// Later tighten validation
type User struct {
    ID   int    `json:"id" validate:"required,min=1"`
    Name string `json:"name" validate:"required,min=2"`
}
```

### Issue 2: Performance Regression

**Problem:** Parsing is slower after migration.

**Solution:** Use appropriate optimization level:
```go
// For high-throughput scenarios
user, err := model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)

// For repeated parsing
user, err := model.ParseIntoCached[User](data)
```

### Issue 3: Type Coercion Issues

**Problem:** Values are not coerced as expected.

**Solution:** Understand coercion rules and adjust types:
```go
// gopantic converts "123" string to int automatically
type User struct {
    ID   int    `json:"id"`     // Accepts both int and string
    Age  *int   `json:"age"`    // Use pointer for optional fields
}
```

## Getting Help

- **Documentation**: Read the full documentation in `/docs/`
- **Examples**: Check practical examples in `/examples/`
- **Performance**: See benchmarks in `/benchmarks/`
- **Issues**: Report issues on GitHub with migration context
- **Best Practices**: Follow the patterns in `BEST_PRACTICES.md`