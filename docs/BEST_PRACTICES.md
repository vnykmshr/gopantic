# Best Practices Guide

This guide provides recommendations for effectively using gopantic in production applications.

## Table of Contents

- [Performance Best Practices](#performance-best-practices)
- [Validation Best Practices](#validation-best-practices)
- [Error Handling Best Practices](#error-handling-best-practices)
- [Caching Best Practices](#caching-best-practices)
- [Production Deployment](#production-deployment)
- [Security Considerations](#security-considerations)
- [Testing Strategies](#testing-strategies)

## Performance Best Practices

### Choose the Right Parsing Method

```go
// ✅ Good: Basic usage for simple scenarios
user, err := model.ParseInto[User](data)

// ✅ Good: Optimized parsing for high-throughput
user, err := model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)

// ✅ Good: Pooled parsing for memory-constrained environments
user, err := model.PooledParseIntoWithFormat[User](data, model.FormatJSON)

// ✅ Good: Cached parsing for repeated operations
user, err := model.ParseIntoCached[User](data)
```

### Performance Monitoring

```go
// Monitor performance metrics
stats := model.GlobalMetrics.GetStats()
fmt.Printf("Parse operations: %d\n", stats.ParseCount)
fmt.Printf("Average time: %.2f ms\n", stats.AvgTimeNs/1_000_000)
fmt.Printf("Cache hit rate: %.2f%%\n", 
    float64(stats.CacheHits)/float64(stats.CacheHits+stats.CacheMisses)*100)
```

### Struct Design for Performance

```go
// ✅ Good: Efficient struct design
type User struct {
    ID       int       `json:"id" validate:"required,min=1"`
    Name     string    `json:"name" validate:"required,min=2,max=100"`
    Email    string    `json:"email" validate:"required,email"`
    Active   bool      `json:"active"`
    Created  time.Time `json:"created_at"`
}

// ❌ Avoid: Overly complex nested structures in hot paths
type OverComplexUser struct {
    Profile struct {
        Details struct {
            PersonalInfo struct {
                // ... deeply nested structures slow down parsing
            } `json:"personal_info"`
        } `json:"details"`
    } `json:"profile"`
}
```

### Batch Processing Optimization

```go
// ✅ Good: Use caching for batch processing
parser, err := model.NewCachedParser[User](model.DefaultCacheConfig())
if err != nil {
    return err
}
defer parser.Close()

for _, data := range largeBatch {
    user, err := parser.Parse(data)
    if err != nil {
        // Handle error
        continue
    }
    // Process user
}
```

## Validation Best Practices

### Effective Validation Tags

```go
type User struct {
    // ✅ Good: Clear, specific validation rules
    ID       int     `json:"id" validate:"required,min=1"`
    Username string  `json:"username" validate:"required,alphanum,min=3,max=30"`
    Email    string  `json:"email" validate:"required,email"`
    Age      int     `json:"age" validate:"min=13,max=120"`
    
    // ✅ Good: Optional fields with validation when present
    Phone    *string `json:"phone,omitempty" validate:"omitempty,min=10"`
    
    // ✅ Good: Use appropriate length constraints
    Bio      string  `json:"bio" validate:"max=500"`
}

// ❌ Avoid: Overly restrictive or vague validation
type BadUser struct {
    ID   int    `json:"id" validate:"required,min=1,max=1000000"` // Too specific
    Name string `json:"name" validate:"required"`                 // Too vague
}
```

### Custom Validation Functions

```go
// ✅ Good: Register reusable custom validators
func init() {
    model.RegisterGlobalFunc("strong_password", func(fieldName string, value interface{}, params map[string]interface{}) error {
        password, ok := value.(string)
        if !ok || password == "" {
            return nil // Let required validator handle empty values
        }
        
        if len(password) < 8 {
            return model.NewValidationError(fieldName, value, "strong_password", 
                "password must be at least 8 characters")
        }
        
        hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
        hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
        hasDigit := strings.ContainsAny(password, "0123456789")
        
        if !hasUpper || !hasLower || !hasDigit {
            return model.NewValidationError(fieldName, value, "strong_password",
                "password must contain uppercase, lowercase, and numeric characters")
        }
        
        return nil
    })
}

type RegistrationRequest struct {
    Password string `json:"password" validate:"required,strong_password"`
}
```

### Cross-Field Validation

```go
// ✅ Good: Use cross-field validation for related fields
func init() {
    model.RegisterGlobalCrossFieldFunc("password_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
        passwordField := structValue.FieldByName("Password")
        if !passwordField.IsValid() {
            return model.NewValidationError(fieldName, fieldValue, "password_match", "Password field not found")
        }
        
        password := passwordField.Interface().(string)
        confirmPassword, ok := fieldValue.(string)
        if !ok {
            return model.NewValidationError(fieldName, fieldValue, "password_match", "value must be a string")
        }
        
        if password != confirmPassword {
            return model.NewValidationError(fieldName, fieldValue, "password_match", "passwords do not match")
        }
        
        return nil
    })
}

type PasswordChangeRequest struct {
    Password        string `json:"password" validate:"required,strong_password"`
    ConfirmPassword string `json:"confirm_password" validate:"required,password_match"`
}
```

## Error Handling Best Practices

### Structured Error Handling

```go
// ✅ Good: Handle different error types appropriately
func parseUserData(data []byte) (*User, error) {
    user, err := model.ParseInto[User](data)
    if err != nil {
        if errorList, ok := err.(model.ErrorList); ok {
            // Handle validation errors
            return nil, handleValidationErrors(errorList)
        }
        // Handle parsing errors
        return nil, fmt.Errorf("failed to parse user data: %w", err)
    }
    return &user, nil
}

func handleValidationErrors(errors model.ErrorList) error {
    // Group errors by field for better user experience
    grouped := errors.GroupByField()
    
    var messages []string
    for fieldPath, fieldErrors := range grouped {
        for _, e := range fieldErrors {
            if validationErr, ok := e.(*model.ValidationError); ok {
                messages = append(messages, fmt.Sprintf("%s: %s", fieldPath, validationErr.Message))
            }
        }
    }
    
    return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}
```

### API Error Responses

```go
// ✅ Good: Convert errors to structured API responses
func handleAPIError(w http.ResponseWriter, err error) {
    if errorList, ok := err.(model.ErrorList); ok {
        // Convert to structured error report
        if jsonData, jsonErr := errorList.ToJSON(); jsonErr == nil {
            var errorReport model.StructuredErrorReport
            if parseErr := json.Unmarshal(jsonData, &errorReport); parseErr == nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusBadRequest)
                json.NewEncoder(w).Encode(map[string]interface{}{
                    "error": "validation_failed",
                    "details": errorReport,
                })
                return
            }
        }
    }
    
    // Fallback for other errors
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "internal_server_error",
        "message": "Failed to process request",
    })
}
```

## Caching Best Practices

### Cache Configuration

```go
// ✅ Good: Configure cache based on your use case
func setupCache() (*model.CachedParser[User], error) {
    config := &model.CacheConfig{
        TTL:                time.Hour,              // Reasonable TTL
        MaxEntries:         10000,                  // Based on memory constraints
        CompressionEnabled: true,                   // Enable for large objects
        Namespace:          "myapp:users",          // Avoid key collisions
        Backend:            model.CacheBackendMemory, // Start with memory
    }
    
    return model.NewCachedParser[User](config)
}

// ✅ Good: Use Redis for distributed applications
func setupDistributedCache() (*model.CachedParser[User], error) {
    config := &model.CacheConfig{
        TTL:                30 * time.Minute,
        CompressionEnabled: true,
        Namespace:          "myapp:users",
        Backend:            model.CacheBackendRedis,
        RedisConfig: &model.RedisConfig{
            Addr:      "redis:6379",
            Password:  os.Getenv("REDIS_PASSWORD"),
            DB:        1,
            KeyPrefix: "gopantic:",
        },
    }
    
    return model.NewCachedParser[User](config)
}
```

### Cache Lifecycle Management

```go
// ✅ Good: Properly manage cache lifecycle
type UserService struct {
    parser *model.CachedParser[User]
}

func NewUserService() (*UserService, error) {
    parser, err := setupCache()
    if err != nil {
        return nil, err
    }
    
    return &UserService{parser: parser}, nil
}

func (s *UserService) Close() error {
    return s.parser.Close()
}

func (s *UserService) ParseUser(data []byte) (User, error) {
    return s.parser.Parse(data)
}

// ✅ Good: Use in main function
func main() {
    service, err := NewUserService()
    if err != nil {
        log.Fatal(err)
    }
    defer service.Close() // Always close resources
    
    // Use service...
}
```

## Production Deployment

### Environment-Specific Configuration

```go
// ✅ Good: Environment-aware configuration
type Config struct {
    Environment string
    Cache       CacheConfig
}

type CacheConfig struct {
    Enabled bool
    Backend string
    Redis   RedisConfig
}

func loadConfig() Config {
    env := os.Getenv("ENVIRONMENT")
    
    switch env {
    case "production":
        return Config{
            Environment: env,
            Cache: CacheConfig{
                Enabled: true,
                Backend: "redis",
                Redis: RedisConfig{
                    Addr:     os.Getenv("REDIS_URL"),
                    Password: os.Getenv("REDIS_PASSWORD"),
                },
            },
        }
    case "staging":
        return Config{
            Environment: env,
            Cache: CacheConfig{
                Enabled: true,
                Backend: "memory",
            },
        }
    default: // development
        return Config{
            Environment: "development",
            Cache: CacheConfig{
                Enabled: false,
            },
        }
    }
}
```

### Health Checks

```go
// ✅ Good: Include gopantic metrics in health checks
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    stats := model.GlobalMetrics.GetStats()
    
    health := map[string]interface{}{
        "status": "healthy",
        "parsing": map[string]interface{}{
            "total_operations": stats.ParseCount,
            "error_rate":      float64(stats.ErrorCount) / float64(stats.ParseCount),
            "avg_time_ms":     stats.AvgTimeNs / 1_000_000,
        },
    }
    
    if stats.CacheHits+stats.CacheMisses > 0 {
        health["cache"] = map[string]interface{}{
            "hit_rate": float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses),
            "hits":     stats.CacheHits,
            "misses":   stats.CacheMisses,
        }
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

### Monitoring and Alerting

```go
// ✅ Good: Set up monitoring for key metrics
func setupMonitoring() {
    ticker := time.NewTicker(1 * time.Minute)
    go func() {
        for range ticker.C {
            stats := model.GlobalMetrics.GetStats()
            
            // Log metrics for monitoring systems
            log.Printf("gopantic_parse_count=%d gopantic_error_count=%d gopantic_avg_time_ms=%.2f",
                stats.ParseCount, stats.ErrorCount, stats.AvgTimeNs/1_000_000)
            
            // Alert on high error rates
            if stats.ParseCount > 100 {
                errorRate := float64(stats.ErrorCount) / float64(stats.ParseCount)
                if errorRate > 0.05 { // 5% error rate threshold
                    log.Printf("HIGH ERROR RATE: %.2f%% (%d/%d)",
                        errorRate*100, stats.ErrorCount, stats.ParseCount)
                }
            }
        }
    }()
}
```

## Security Considerations

### Input Validation

```go
// ✅ Good: Validate input size and structure
func parseUserInput(data []byte) (User, error) {
    // Limit input size
    const maxInputSize = 1024 * 1024 // 1MB
    if len(data) > maxInputSize {
        return User{}, fmt.Errorf("input too large: %d bytes", len(data))
    }
    
    // Parse with validation
    user, err := model.ParseInto[User](data)
    if err != nil {
        // Don't expose internal parsing details
        return User{}, fmt.Errorf("invalid input format")
    }
    
    return user, nil
}
```

### Sanitization

```go
// ✅ Good: Sanitize sensitive fields
type User struct {
    ID       int    `json:"id" validate:"required,min=1"`
    Username string `json:"username" validate:"required,alphanum,min=3,max=30"`
    Email    string `json:"email" validate:"required,email"`
}

func (u *User) Sanitize() {
    u.Username = strings.TrimSpace(u.Username)
    u.Email = strings.ToLower(strings.TrimSpace(u.Email))
}

func parseAndSanitizeUser(data []byte) (User, error) {
    user, err := model.ParseInto[User](data)
    if err != nil {
        return User{}, err
    }
    
    user.Sanitize()
    return user, nil
}
```

## Testing Strategies

### Unit Testing

```go
// ✅ Good: Test parsing with various inputs
func TestUserParsing(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    User
        wantErr bool
    }{
        {
            name:  "valid user",
            input: `{"id": 1, "username": "john", "email": "john@example.com"}`,
            want:  User{ID: 1, Username: "john", Email: "john@example.com"},
        },
        {
            name:    "invalid email",
            input:   `{"id": 1, "username": "john", "email": "invalid"}`,
            wantErr: true,
        },
        {
            name:    "missing required field",
            input:   `{"username": "john"}`,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := model.ParseInto[User]([]byte(tt.input))
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && got != tt.want {
                t.Errorf("ParseInto() got = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Benchmark Testing

```go
// ✅ Good: Benchmark different parsing methods
func BenchmarkUserParsing(b *testing.B) {
    data := []byte(`{"id": 1, "username": "john", "email": "john@example.com"}`)
    
    b.Run("Basic", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _, err := model.ParseInto[User](data)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("Optimized", func(b *testing.B) {
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _, err := model.OptimizedParseIntoWithFormat[User](data, model.FormatJSON)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("Cached", func(b *testing.B) {
        parser, _ := model.NewCachedParser[User](model.DefaultCacheConfig())
        defer parser.Close()
        
        b.ResetTimer()
        b.ReportAllocs()
        for i := 0; i < b.N; i++ {
            _, err := parser.Parse(data)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

### Integration Testing

```go
// ✅ Good: Test with realistic data and scenarios
func TestUserServiceIntegration(t *testing.T) {
    // Setup test cache
    config := model.DefaultCacheConfig()
    config.TTL = time.Minute
    
    service, err := setupUserServiceWithConfig(config)
    if err != nil {
        t.Fatal(err)
    }
    defer service.Close()
    
    // Test realistic scenarios
    userData := []byte(`{
        "id": 1,
        "username": "integration_test",
        "email": "test@example.com",
        "profile": {
            "first_name": "Test",
            "last_name": "User"
        }
    }`)
    
    // First parse (cache miss)
    user1, err := service.ParseUser(userData)
    if err != nil {
        t.Fatal(err)
    }
    
    // Second parse (cache hit)
    user2, err := service.ParseUser(userData)
    if err != nil {
        t.Fatal(err)
    }
    
    if user1 != user2 {
        t.Error("Cached result differs from original")
    }
    
    // Verify cache metrics
    stats := model.GlobalMetrics.GetStats()
    if stats.CacheHits < 1 {
        t.Error("Expected at least one cache hit")
    }
}
```

## Anti-Patterns to Avoid

### ❌ Don't: Ignore Error Types

```go
// ❌ Bad: Generic error handling
user, err := model.ParseInto[User](data)
if err != nil {
    return fmt.Errorf("parse failed: %v", err)
}

// ✅ Good: Specific error handling
user, err := model.ParseInto[User](data)
if err != nil {
    if errorList, ok := err.(model.ErrorList); ok {
        return handleValidationErrors(errorList)
    }
    return fmt.Errorf("parse failed: %w", err)
}
```

### ❌ Don't: Overuse Caching

```go
// ❌ Bad: Caching everything
parser1, _ := model.NewCachedParser[SmallStruct](config)
parser2, _ := model.NewCachedParser[AnotherSmallStruct](config)
// ... dozens of cached parsers

// ✅ Good: Cache selectively
// Only cache frequently used or expensive-to-parse types
parser, _ := model.NewCachedParser[ExpensiveStruct](config)
```

### ❌ Don't: Forget Resource Cleanup

```go
// ❌ Bad: Resource leak
func processData() {
    parser, _ := model.NewCachedParser[User](config)
    // Forgot to call parser.Close()
}

// ✅ Good: Proper cleanup
func processData() {
    parser, err := model.NewCachedParser[User](config)
    if err != nil {
        return err
    }
    defer parser.Close()
    // Process data...
}
```

By following these best practices, you'll get the most out of gopantic while maintaining high performance, security, and maintainability in your applications.