# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.1.0] - 2025-10-30

### Added

- **Validation metadata caching**: Struct validation rules now cached by type using `sync.Map` for 10-20% performance improvement on repeated parsing
- **Cache hit rate tracking**: `CachedParser.Stats()` now returns actual hit rate via atomic counters for better observability
- **Proactive cache cleanup**: Background goroutine periodically removes expired entries (configurable via `CleanupInterval`)
- **Input size protection**: Added `MaxInputSize` variable (default 10MB) to prevent DoS attacks from oversized inputs
- **Security policy**: Added SECURITY.md with vulnerability disclosure process and security best practices
- **Migration guide**: Comprehensive docs/migration.md covering migration from encoding/json, validator, and YAML libraries
- **Dependabot automation**: Weekly automated dependency updates for Go modules and GitHub Actions
- **`ClearValidationCache()` function**: Manual cache invalidation for dynamic validator registration scenarios
- **`Close()` method**: Properly stops background cleanup goroutine on CachedParser

### Changed

- **Optimized time parsing**: Heuristic-based format selection tries RFC3339/ISO 8601 first when 'T' detected at position 10
- **Cache configuration**: Added `CleanupInterval` field to `CacheConfig` (default: 30 minutes)
- **Documentation clarity**: Updated README and architecture docs to clarify cache use cases and limitations
- **golangci-lint v2 compatibility**: Updated configuration for golangci-lint v2 with version pinning for consistency

### Fixed

- **golangci-lint configuration**: Fixed YAML syntax and made compatible with v2.x
- **Test coverage reporting**: Added `-coverpkg=./pkg/...` flag to properly capture package coverage
- **CI cleanup**: Removed unused Redis service configurations from all GitHub Actions workflows
- **Linter version consistency**: Pinned golangci-lint to v2.5.0 in CI to match local development

### Performance

- 10-20% faster validation for repeated struct types via metadata caching
- Optimized time parsing for common RFC3339 timestamps
- Proactive cleanup prevents memory accumulation in long-running services

### Security

- DoS protection via configurable input size limits
- Documentation of error message content for secure handling of sensitive data
- Automated dependency vulnerability scanning

### Documentation

- New migration guide (200+ lines) covering common patterns
- Security policy with vulnerability reporting process
- Enhanced API docs with security considerations
- Clarified cache effectiveness and appropriate use cases

## [1.0.1] - 2025-01-16

### Fixed
- **GitHub Actions workflows** - Resolved tar cache extraction failures in nightly performance monitoring
- **Benchmark targets** - Corrected Make targets to point to correct tests directory structure
- **CI pipeline** - Fixed lint and security job failures, added proper SARIF upload permissions
- **Repository security** - Removed non-existent security scan actions and corrected gosec paths

### Changed
- **Documentation script** - Streamlined generate-docs.sh for cleaner, minimal output
- **Build artifacts** - Enhanced .gitignore for better profiling and benchmark file management
- **Code quality** - Removed TODO comments from production code for release readiness

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
- Compatible with Go 1.23+

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