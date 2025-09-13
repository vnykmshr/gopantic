# Documentation Overview

Welcome to the comprehensive documentation for gopantic - a high-performance Go library for parsing and validating data structures with Python Pydantic-inspired design.

## Quick Start

```go
import "github.com/vnykmshr/gopantic/pkg/model"

type User struct {
    ID   int    `json:"id" validate:"required,min=1"`
    Name string `json:"name" validate:"required,min=2"`
    Email string `json:"email" validate:"required,email"`
}

user, err := model.ParseInto[User](jsonData)
if err != nil {
    log.Fatal(err)
}
```

## Documentation Structure

### Core Guides

- **[Migration Guide](MIGRATION_GUIDE.md)** - Complete guide for migrating from standard library JSON parsing to gopantic
- **[Best Practices](BEST_PRACTICES.md)** - Production-ready patterns and recommendations
- **[API Reference](../pkg/model/)** - Complete GoDoc documentation for all public APIs

### Practical Resources

- **[Examples](../examples/)** - Working code examples for common use cases:
  - [Basic Usage](../examples/basic/) - Simple parsing and validation
  - [YAML Support](../examples/yaml_demo/) - YAML parsing examples
  - [Caching](../examples/cache_demo/) - Performance optimization with caching
  - [API Validation](../examples/api_validation/) - REST API validation server
  - [Config Parsing](../examples/config_parsing/) - Enterprise configuration management

- **[Benchmarks](../benchmarks/)** - Performance comparisons and optimization analysis
- **[Integration Tests](../integration/)** - End-to-end testing scenarios

## Key Features

### 🚀 Performance First
- **17-45% faster** than standard JSON parsing with optimizations
- **Struct info caching** reduces reflection overhead
- **Object pooling** minimizes memory allocations
- **Redis distributed caching** for multi-instance deployments

### ✅ Comprehensive Validation
- **Built-in validators**: required, min, max, email, length, alpha, alphanum
- **Custom validators**: Register your own validation functions
- **Cross-field validation**: Validate fields against other fields in the struct
- **Structured error reporting**: Machine-readable error details

### 🔄 Format Flexibility
- **JSON and YAML** support with automatic format detection
- **Type coercion**: Intelligent conversion between compatible types
- **Nested structures**: Full support for complex data hierarchies
- **Pointer types**: Optional fields with proper nil handling

### 📊 Production Ready
- **Comprehensive monitoring**: Built-in performance metrics
- **Health checks**: Ready-to-use health check endpoints
- **Error handling**: Structured error responses for APIs
- **Resource management**: Proper cleanup and lifecycle management

## Architecture Overview

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Input Data    │───▶│   Format Parser  │───▶│  Type Coercion  │
│  (JSON/YAML)    │    │ (JSON/YAML Auto  │    │   & Validation  │
└─────────────────┘    │    Detection)    │    └─────────────────┘
                       └──────────────────┘             │
                                                        ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Final Result  │◀───│   Error Handler  │◀───│   Struct Info   │
│   (Validated    │    │   (Aggregated    │    │     Cache       │
│    Struct)      │    │    Errors)       │    │ (Performance)   │
└─────────────────┘    └──────────────────┘    └─────────────────┘
```

## Performance Comparison

| Parsing Method | Speed | Memory | Use Case |
|----------------|-------|---------|----------|
| `ParseInto` | Baseline | Baseline | General purpose |
| `OptimizedParseIntoWithFormat` | +17-45% | -20% | High throughput |
| `PooledParseIntoWithFormat` | +10-30% | -50% | Memory constrained |
| `ParseIntoCached` | +200-1000%* | Variable | Repeated parsing |

*For cache hits

## Common Use Cases

### 1. REST API Validation

```go
// Automatic validation with structured error responses
func createUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    user, err := model.ParseInto[CreateUserRequest](body)
    if err != nil {
        writeValidationError(w, err)
        return
    }
    // Process validated user...
}
```

### 2. Configuration Management

```go
// Support both JSON and YAML configuration files
config, err := model.ParseInto[AppConfig](configData)
if err != nil {
    log.Fatalf("Invalid configuration: %v", err)
}
```

### 3. High-Performance Data Processing

```go
// Optimized parsing for data pipelines
for _, record := range records {
    item, err := model.OptimizedParseIntoWithFormat[DataItem](record, model.FormatJSON)
    if err != nil {
        continue
    }
    process(item)
}
```

### 4. Microservice Communication

```go
// Cached parsing for frequently exchanged message types
parser, _ := model.NewCachedParser[MessageType](config)
defer parser.Close()

message, err := parser.Parse(payload)
```

## Getting Started Path

1. **Start Here**: Read the [Migration Guide](MIGRATION_GUIDE.md) if migrating from standard library
2. **Learn Patterns**: Study the [Best Practices](BEST_PRACTICES.md) for production usage
3. **Try Examples**: Run code from the [examples](../examples/) directory
4. **Optimize**: Use [benchmarks](../benchmarks/) to measure and improve performance
5. **Deploy**: Follow production deployment guidelines in [Best Practices](BEST_PRACTICES.md)

## API Reference Quick Links

### Core Functions
- [`model.ParseInto[T]`](../pkg/model/parse.go) - Basic parsing with auto-detection
- [`model.ParseIntoWithFormat[T]`](../pkg/model/parse.go) - Format-specific parsing
- [`model.OptimizedParseIntoWithFormat[T]`](../pkg/model/optimization.go) - High-performance parsing
- [`model.PooledParseIntoWithFormat[T]`](../pkg/model/pool.go) - Memory-optimized parsing

### Caching
- [`model.NewCachedParser[T]`](../pkg/model/cache.go) - Create cached parser
- [`model.ParseIntoCached[T]`](../pkg/model/cache.go) - Global cached parsing
- [`model.DefaultCacheConfig()`](../pkg/model/cache.go) - Default cache configuration

### Validation
- [`model.RegisterGlobalFunc`](../pkg/model/validate.go) - Register custom validators
- [`model.RegisterGlobalCrossFieldFunc`](../pkg/model/validate.go) - Register cross-field validators

### Utilities
- [`model.DetectFormat`](../pkg/model/format.go) - Auto-detect data format
- [`model.GlobalMetrics`](../pkg/model/pool.go) - Performance monitoring

## Community and Support

- **GitHub Repository**: Source code, issues, and contributions
- **Examples**: Practical implementation patterns
- **Benchmarks**: Performance analysis and optimization guides
- **Documentation**: Comprehensive guides and API reference

## License

This project is licensed under the MIT License. See the [LICENSE](../LICENSE) file for details.

---

**Ready to get started?** Choose your path:
- 🚀 **New to gopantic**: Start with [Migration Guide](MIGRATION_GUIDE.md)
- 🏗️ **Building production apps**: Read [Best Practices](BEST_PRACTICES.md)
- 🔍 **Need specific examples**: Browse [examples](../examples/)
- 📈 **Performance optimization**: Check [benchmarks](../benchmarks/)