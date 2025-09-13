# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-13

### Added
- **JSON/YAML parsing** with automatic format detection
- **Type coercion** - automatic conversion between compatible types (`"123"` → `123`, `"true"` → `true`)
- **Comprehensive validation** using struct tags (`validate:"required,email,min=5"`)
- **Cross-field validation** - validate fields against each other (password confirmation, field comparisons)
- **Built-in validators**: `required`, `min`, `max`, `email`, `alpha`, `alphanum`, `length`
- **Nested struct support** with full validation propagation
- **Array and slice parsing** with element validation
- **Time parsing** support (RFC3339, Unix timestamps, custom formats)
- **Pointer support** for optional fields (`*string`, `*int`)
- **High-performance caching** - 5-27x speedup for repeated parsing operations
- **Thread-safe** concurrent parsing operations
- **Generics support** for type-safe parsing with `ParseInto[T]()`
- **Structured error reporting** with field paths and detailed context
- **Custom validator registration** for domain-specific validation rules
- **Zero dependencies** (except optional YAML support via `gopkg.in/yaml.v3`)

### Features
- Single-function API: `model.ParseInto[T](data)` covers most use cases
- Automatic format detection (JSON vs YAML)
- Comprehensive example collection covering real-world scenarios
- Production-ready with extensive test coverage
- Compatible with Go 1.21+

### Examples
- Basic parsing with type coercion
- Validation examples with multiple error handling
- Cross-field validation (password confirmation, email differences)
- Time parsing with multiple format support
- YAML configuration parsing
- API request/response validation
- Pointer field handling for optional data
- High-performance caching usage

---

## Version Format
This project uses [Semantic Versioning](https://semver.org/):
- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality additions
- **PATCH** version for backwards-compatible bug fixes