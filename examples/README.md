# gopantic Examples

Practical examples demonstrating gopantic's features for data parsing and validation.

## Quick Start

New to gopantic? Follow this learning path:

### 1. quickstart/
**Basic parsing and validation**

Learn the fundamentals:
- Parse JSON with automatic type coercion
- Add validation rules using struct tags
- Handle validation errors

Start here if you're new to gopantic.

### 2. yaml/
**YAML configuration parsing**

Working with configuration files:
- Parse YAML and JSON with the same API
- Automatic format detection
- Nested configuration structures
- Type coercion across formats

### 3. cache_demo/
**High-performance caching**

Optimize repeated parsing:
- Use CachedParser for identical inputs
- Monitor cache hit rates
- When caching helps (and when it doesn't)

### 4. cross_field_validation/
**Advanced validation patterns**

Custom validation logic:
- Cross-field validators
- Custom validation functions
- Complex business rules
- Password confirmation patterns

### 5. pointers/
**Optional fields with pointers**

Handle missing data:
- Distinguish between zero values and missing fields
- Optional vs required fields
- Nil pointer handling

### 6. api_validation/
**Real-world API validation**

Production patterns:
- HTTP request validation
- Error response formatting
- Integration with web frameworks
- Production-ready error handling

## Database Integration

For PostgreSQL JSONB integration, see:
- [docs/tutorials/postgresql-integration/](../docs/tutorials/postgresql-integration/)

Complete example with database integration, flexible metadata, and ORM patterns.

## Running Examples

```bash
# Run any example
cd quickstart
go run main.go

# Or from project root
go run examples/quickstart/main.go
```

## See Also

- [API Documentation](../docs/api.md) - Complete API reference
- [Type Reference](../docs/type-reference.md) - Supported types and limitations
- [Database Integration](../docs/database-integration.md) - PostgreSQL patterns
