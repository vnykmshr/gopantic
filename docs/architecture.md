# Architecture & Design

gopantic is designed for high-performance, type-safe parsing with minimal overhead. This document covers the core architecture, design decisions, and implementation details.

## Core Architecture

### Package Structure

```
pkg/model/
├── doc.go          → Package documentation
├── parse.go        → Main parsing logic and API (ParseInto, Validate)
├── format.go       → Format detection and parser abstraction
├── coerce.go       → Type coercion engine
├── validate.go     → Validation framework and registry
├── validators.go   → Built-in validators (required, min, max, etc.)
├── cache.go        → High-performance caching with FIFO eviction
├── config.go       → Thread-safe configuration accessors
└── errors.go       → Error types and aggregation
```

### Design Principles

1. **Type Safety First** - Leverages Go generics for compile-time type safety
2. **Zero-Cost Abstractions** - Optional features have no overhead when unused
3. **Performance by Default** - Optimized for common use cases
4. **Idiomatic Go** - Struct tags, interfaces, and familiar patterns

## Parsing Engine

### Generic API Design

The core parsing API uses Go generics for type safety:

```go
func ParseInto[T any](data []byte) (T, error)
func ParseIntoWithFormat[T any](data []byte, format Format) (T, error)
```

Benefits:
- Compile-time type checking
- No runtime type assertions
- Clean API without interface{} returns

### Format Abstraction

Format detection and parsing are abstracted through interfaces:

```go
type FormatParser interface {
    Parse(data []byte) (map[string]any, error)
    Format() Format
}

type Format int
const (
    FormatJSON Format = iota
    FormatYAML
)
```

Format Detection Algorithm:
1. Check for JSON markers (`{`, `[`)
2. Check for YAML markers (`---`, `:`)
3. Default to JSON for ambiguous cases
4. Performance: O(1) with early termination

## Type Coercion

### Coercion Strategy

Type coercion happens after parsing but before validation:

```
Raw Data → Parse to map[string]any → Type Coercion → Struct → Validation
```

### Supported Coercions

| Target | From | Algorithm |
|---------|-------|-----------|
| `int` | `string` | `strconv.Atoi()` with error handling |
| `float64` | `string` | `strconv.ParseFloat()` with precision |
| `bool` | `string` | Custom logic: `"true"`, `"yes"`, `"1"` → `true` |
| `bool` | `number` | `0` → `false`, non-zero → `true` |
| `time.Time` | `string` | Multiple format attempts (RFC3339, Unix) |
| `time.Time` | `number` | Unix timestamp conversion |

### Performance Optimization

- **Validation Metadata Caching** - Struct validation rules cached by type (via sync.Map)
- **Fast Path Detection** - Skip coercion for matching types
- **Minimal Allocations** - Reuse existing values when possible
- **Optimized Time Parsing** - Heuristic-based format selection for common cases

## Validation Framework

### Tag-Based Validation

Validation rules are specified using struct tags:

```go
type User struct {
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"min=18,max=120"`
}
```

### Validator Interface

All validators implement a common interface:

```go
type Validator interface {
    Validate(fieldName string, value interface{}) error
}
```

### Built-in Validators

| Validator | Description | Performance |
|-----------|-------------|-------------|
| `required` | Non-zero value check | O(1) |
| `min/max` | Numeric/length bounds | O(1) |
| `email` | Regex validation | O(n) - cached regex |
| `alpha` | Alphabetic characters | O(n) |
| `alphanum` | Alphanumeric characters | O(n) |
| `length` | Exact length check | O(1) |

### Error Aggregation

Multiple validation errors are collected and returned together:

```go
type ErrorList []error

func (el *ErrorList) Add(err error) {
    *el = append(*el, err)
}

func (el ErrorList) AsError() error {
    if len(el) == 0 {
        return nil
    }
    return fmt.Errorf("multiple errors: %s", strings.Join(errorStrings, "; "))
}
```

## Caching System

### Design Goals

- **Transparent**: Same API as non-cached parsing
- **Thread-Safe**: Concurrent access with minimal locking
- **Configurable**: TTL, max entries, namespacing
- **High Performance**: Content-based keys with SHA256 hashing

### Cache Architecture

```go
type CachedParser[T any] struct {
    cache       map[string]cacheEntry
    mu          sync.RWMutex
    config      *CacheConfig
    hits        uint64 // Atomic counter
    misses      uint64 // Atomic counter
    stopCleanup chan struct{}
}

type CacheConfig struct {
    TTL             time.Duration // Entry lifetime (default: 1 hour)
    MaxEntries      int           // LRU eviction limit (default: 1000)
    CleanupInterval time.Duration // Background cleanup (default: 30 min)
}
```

**Features:**
- **Hit Rate Tracking**: Atomic counters for cache hits/misses
- **Proactive Cleanup**: Background goroutine removes expired entries
- **LRU Eviction**: Oldest entries removed when MaxEntries reached
- **Thread-Safe**: RWMutex for concurrent access

### Key Generation

Cache keys are generated using content-based SHA256 hashing:

```
key = sha256(data)[:16] + ":" + reflect.TypeOf(T).String()
```

**Benefits:**
- Deterministic keys for identical content
- Type-safe (different types don't collide)
- Efficient 16-byte prefix

**Limitations:**
- Even one byte difference invalidates cache
- Best for truly identical inputs (config files, retries)
- Limited benefit for unique API requests with varying data

### Cache Effectiveness

Benchmarks show significant speedups for **identical** inputs:

| Scenario | Uncached | Cached | Speedup |
|----------|----------|--------|---------|
| Simple JSON | 8.7μs | 1.5μs | 5.8x |
| Complex JSON | 27.6μs | 2.6μs | 10.5x |
| Simple YAML | 20.8μs | 1.5μs | 13.7x |
| Complex YAML | 69.4μs | 2.6μs | 27.2x |

**When to use caching:**
- Good fit: Parsing static configuration files repeatedly
- Good fit: Retrying identical failed requests
- Good fit: Processing duplicate messages in queues
- Poor fit: Parsing unique API requests (same schema, different data)
- Poor fit: Streaming different records from a file

**Monitoring:** Use `Stats()` to check hit rate and adjust strategy accordingly.

## Error Handling

### Structured Error Types

```go
type ParseError struct {
    Field   string
    Value   interface{}
    Type    string
    Message string
}

type ValidationError struct {
    Field   string
    Value   interface{}
    Rule    string
    Message string
}
```

### Error Context

Errors include full field paths for nested structures:

```
validation error on field "user.address.zip": length must be exactly 5 characters
```

## Memory Management

### Allocation Patterns

- **Struct Reuse** - Target struct allocated once
- **Slice Pre-allocation** - Collections sized by JSON array length
- **String Interning** - Common field names reused
- **Error Pooling** - Error objects reused in high-throughput scenarios

### GC Pressure

- **Minimal Heap Allocations** - Most work done on stack
- **No Intermediate Objects** - Direct parsing to target struct
- **Cache-Friendly** - Sequential access patterns where possible

## Benchmarks & Performance

### Comparison with Standard Library

| Operation | Standard JSON | gopantic | Overhead |
|-----------|--------------|----------|----------|
| Simple Parse | 1.8μs | 8.7μs | 4.8x |
| With Validation | N/A | 9.2μs | 5.1x |
| With Coercion | N/A | 8.9μs | 4.9x |

### Memory Usage

| Operation | Standard JSON | gopantic | Overhead |
|-----------|--------------|----------|----------|
| Simple Parse | 1.2KB | 3.8KB | 3.2x |
| Complex Parse | 4.5KB | 11.6KB | 2.6x |

### Caching Benefits

Caching provides substantial improvements for repeated parsing operations, making gopantic comparable to or faster than standard JSON for cache hit scenarios.

## Thread Safety

All public APIs are thread-safe:

- **ParseInto** - Stateless, fully concurrent
- **CachedParser** - RWMutex for cache access
- **Validators** - Stateless implementations
- **Format Detection** - No shared state

## Future Considerations

### Scalability

- **Parser Pooling** - Reuse parser instances for high throughput
- **Custom Allocators** - Memory pool for struct allocation
- **SIMD Optimization** - Vectorized validation operations

### Extensibility

- **Plugin System** - Runtime validator registration
- **Custom Formats** - Protocol buffer, MessagePack support
- **Streaming API** - Large dataset processing

This architecture provides a solid foundation for high-performance parsing while maintaining Go's principles of simplicity and explicitness.