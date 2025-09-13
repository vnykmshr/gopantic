package model

import (
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/vnykmshr/goflow/pkg/ratelimit/bucket"
	"github.com/vnykmshr/obcache-go/pkg/obcache"
)

// EnhancedValidatorConfig configures rate limiting and caching for external validation services
type EnhancedValidatorConfig struct {
	// Rate limiting configuration
	RateLimit struct {
		RequestsPerSecond int           // Maximum requests per second
		BurstCapacity     int           // Burst capacity for short bursts
		Timeout           time.Duration // Timeout for rate limit operations
	}

	// Caching configuration for validation results
	Cache struct {
		TTL         time.Duration // Time to live for cached validation results
		MaxEntries  int           // Maximum number of cached entries
		Backend     string        // "memory" or "redis"
		RedisConfig *RedisConfig  // Redis configuration if using Redis backend
	}

	// External service configuration
	ExternalServices struct {
		EmailValidationURL    string        // URL for email validation service
		DomainValidationURL   string        // URL for domain validation service
		RequestTimeout        time.Duration // Timeout for external service requests
		MaxRetries            int           // Maximum number of retries for failed requests
		BackoffMultiplier     float64       // Backoff multiplier for retries
		GracefulDegradation   bool          // Fall back to basic validation if service fails
		CostOptimization      bool          // Enable cost optimization features
		BatchValidation       bool          // Enable batch validation for multiple values
		CircuitBreakerEnabled bool          // Enable circuit breaker for external services
	}
}

// DefaultEnhancedValidatorConfig returns a default configuration for enhanced validators
func DefaultEnhancedValidatorConfig() *EnhancedValidatorConfig {
	config := &EnhancedValidatorConfig{}

	// Rate limiting defaults
	config.RateLimit.RequestsPerSecond = 100 // 100 requests per second
	config.RateLimit.BurstCapacity = 10      // Allow bursts of up to 10 requests
	config.RateLimit.Timeout = 5 * time.Second

	// Cache defaults
	config.Cache.TTL = 1 * time.Hour
	config.Cache.MaxEntries = 10000
	config.Cache.Backend = "memory"

	// External service defaults
	config.ExternalServices.RequestTimeout = 5 * time.Second
	config.ExternalServices.MaxRetries = 3
	config.ExternalServices.BackoffMultiplier = 2.0
	config.ExternalServices.GracefulDegradation = true
	config.ExternalServices.CostOptimization = true
	config.ExternalServices.BatchValidation = false
	config.ExternalServices.CircuitBreakerEnabled = true

	return config
}

// EnhancedValidator provides rate-limited and cached external validation capabilities
type EnhancedValidator struct {
	config      *EnhancedValidatorConfig
	rateLimiter bucket.Limiter
	cache       *obcache.Cache
	httpClient  *http.Client
}

// NewEnhancedValidator creates a new enhanced validator with rate limiting and caching
func NewEnhancedValidator(config *EnhancedValidatorConfig) (*EnhancedValidator, error) {
	if config == nil {
		config = DefaultEnhancedValidatorConfig()
	}

	// Create rate limiter
	rate := bucket.Limit(config.RateLimit.RequestsPerSecond)
	rateLimiter, err := bucket.NewSafe(rate, config.RateLimit.BurstCapacity)
	if err != nil {
		return nil, fmt.Errorf("failed to create rate limiter: %w", err)
	}

	// Create cache
	var cacheConfig *obcache.Config
	if config.Cache.Backend == "redis" && config.Cache.RedisConfig != nil {
		// Use Redis backend
		cacheConfig = obcache.NewRedisConfig(config.Cache.RedisConfig.Addr)
		if config.Cache.RedisConfig.Password != "" {
			cacheConfig.Redis.Password = config.Cache.RedisConfig.Password
		}
		cacheConfig.Redis.DB = config.Cache.RedisConfig.DB
	} else {
		// Use memory cache as default
		cacheConfig = obcache.NewDefaultConfig()
	}

	// Set TTL
	cacheConfig.DefaultTTL = config.Cache.TTL

	cache, err := obcache.New(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create cache: %w", err)
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.ExternalServices.RequestTimeout,
	}

	return &EnhancedValidator{
		config:      config,
		rateLimiter: rateLimiter,
		cache:       cache,
		httpClient:  httpClient,
	}, nil
}

var defaultEnhancedValidator *EnhancedValidator

// GetDefaultEnhancedValidator returns a default enhanced validator instance.
// This is a convenience function for quick setup with reasonable defaults.
// The returned validator uses in-memory caching and standard rate limiting.
func GetDefaultEnhancedValidator() (*EnhancedValidator, error) {
	if defaultEnhancedValidator == nil {
		var err error
		defaultEnhancedValidator, err = NewEnhancedValidator(nil)
		if err != nil {
			return nil, err
		}
	}
	return defaultEnhancedValidator, nil
}

// ExternalEmailValidator validates email addresses using external services with rate limiting
type ExternalEmailValidator struct {
	enhancedValidator *EnhancedValidator
	fallbackToBasic   bool
}

// NewExternalEmailValidator creates a new external email validator
func NewExternalEmailValidator(config *EnhancedValidatorConfig) (*ExternalEmailValidator, error) {
	enhanced, err := NewEnhancedValidator(config)
	if err != nil {
		return nil, err
	}

	return &ExternalEmailValidator{
		enhancedValidator: enhanced,
		fallbackToBasic:   config != nil && config.ExternalServices.GracefulDegradation,
	}, nil
}

// Name returns the validator name
func (v *ExternalEmailValidator) Name() string {
	return "external_email"
}

// Validate validates an email address using external services with rate limiting and caching
func (v *ExternalEmailValidator) Validate(fieldName string, value interface{}) error {
	email, ok := value.(string)
	if !ok {
		return NewValidationError(fieldName, value, v.Name(), "value must be a string")
	}

	if email == "" {
		return nil // Empty values are handled by required validator
	}

	// Basic format validation first (no external call needed)
	if _, err := mail.ParseAddress(email); err != nil {
		return NewValidationError(fieldName, value, v.Name(), "invalid email format")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("email_validation:%s", email)
	if result, found := v.enhancedValidator.cache.Get(cacheKey); found {
		if result == "valid" {
			return nil
		}
		if errorMsg, ok := result.(string); ok && errorMsg != "valid" {
			return NewValidationError(fieldName, value, v.Name(), errorMsg)
		}
	}

	// Cost optimization: skip external validation for obvious invalid emails
	if v.enhancedValidator.config.ExternalServices.CostOptimization {
		if v.isObviouslyInvalid(email) {
			errorMsg := "email domain appears invalid"
			_ = v.enhancedValidator.cache.Set(cacheKey, errorMsg, v.enhancedValidator.config.Cache.TTL)
			return NewValidationError(fieldName, value, v.Name(), errorMsg)
		}
	}

	// Rate limiting check
	if !v.enhancedValidator.rateLimiter.Allow() {
		if v.fallbackToBasic {
			// Graceful degradation: fall back to basic validation
			return nil // Basic format check already passed
		}
		return NewValidationError(fieldName, value, v.Name(), "validation service temporarily unavailable (rate limit exceeded)")
	}

	// External validation
	if v.enhancedValidator.config.ExternalServices.EmailValidationURL != "" {
		if err := v.validateWithExternalService(email, cacheKey); err != nil {
			if v.fallbackToBasic {
				// Graceful degradation: external service failed, but basic validation passed
				return nil
			}
			return NewValidationError(fieldName, value, v.Name(), err.Error())
		}
	}

	// Cache successful validation
	_ = v.enhancedValidator.cache.Set(cacheKey, "valid", v.enhancedValidator.config.Cache.TTL)
	return nil
}

// validateWithExternalService performs the actual external validation call
func (v *ExternalEmailValidator) validateWithExternalService(email, cacheKey string) error {
	url := fmt.Sprintf("%s?email=%s", v.enhancedValidator.config.ExternalServices.EmailValidationURL, email)

	resp, err := v.enhancedValidator.httpClient.Get(url)
	if err != nil {
		// Cache negative result for shorter duration to retry sooner
		// Cache negative result for shorter duration to retry sooner
		_ = v.enhancedValidator.cache.Set(cacheKey, "external service error", v.enhancedValidator.config.Cache.TTL/10)
		return fmt.Errorf("external validation service error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("external validation failed with status %d", resp.StatusCode)
		// Cache negative result
		_ = v.enhancedValidator.cache.Set(cacheKey, errorMsg, v.enhancedValidator.config.Cache.TTL)
		return fmt.Errorf("%s", errorMsg)
	}

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read validation response: %w", err)
	}

	// Simple validation: if response contains "valid", consider it valid
	// In practice, you would parse JSON response according to your external service
	if strings.Contains(strings.ToLower(string(body)), "valid") {
		return nil
	}

	errorMsg := "email validation failed by external service"
	_ = v.enhancedValidator.cache.Set(cacheKey, errorMsg, v.enhancedValidator.config.Cache.TTL)
	return fmt.Errorf("%s", errorMsg)
}

// isObviouslyInvalid performs basic checks to avoid expensive external calls for obviously invalid emails
func (v *ExternalEmailValidator) isObviouslyInvalid(email string) bool {
	// Common invalid patterns - only exact domain matches
	invalidPatterns := []string{
		"@example.com",
		"@example.org",
		"@example.net",
		"@test.com",
		"@test.org",
		"@localhost",
		"@invalid",
		"@fake.com",
		"@dummy.com",
		"@fake",
		"@dummy",
		"@tempmail.com",
		"@mailinator.com",
		"@guerrillamail.com",
	}

	emailLower := strings.ToLower(email)
	for _, pattern := range invalidPatterns {
		if strings.HasSuffix(emailLower, pattern) {
			return true
		}
	}

	// Check for obviously invalid characters
	if strings.Contains(email, "..") || strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return true
	}

	return false
}

// DomainValidator validates domain names using external services with rate limiting
type DomainValidator struct {
	enhancedValidator *EnhancedValidator
	fallbackToBasic   bool
}

// NewDomainValidator creates a new domain validator with external service integration
func NewDomainValidator(config *EnhancedValidatorConfig) (*DomainValidator, error) {
	enhanced, err := NewEnhancedValidator(config)
	if err != nil {
		return nil, err
	}

	return &DomainValidator{
		enhancedValidator: enhanced,
		fallbackToBasic:   config != nil && config.ExternalServices.GracefulDegradation,
	}, nil
}

// Name returns the validator name
func (v *DomainValidator) Name() string {
	return "domain"
}

// Validate validates a domain name using external services with rate limiting and caching
func (v *DomainValidator) Validate(fieldName string, value interface{}) error {
	domain, ok := value.(string)
	if !ok {
		return NewValidationError(fieldName, value, v.Name(), "value must be a string")
	}

	if domain == "" {
		return nil // Empty values are handled by required validator
	}

	// Basic format validation first
	if !v.isValidDomainFormat(domain) {
		return NewValidationError(fieldName, value, v.Name(), "invalid domain format")
	}

	// Check cache first
	cacheKey := fmt.Sprintf("domain_validation:%s", domain)
	if result, found := v.enhancedValidator.cache.Get(cacheKey); found {
		if result == "valid" {
			return nil
		}
		if errorMsg, ok := result.(string); ok && errorMsg != "valid" {
			return NewValidationError(fieldName, value, v.Name(), errorMsg)
		}
	}

	// Rate limiting check
	if !v.enhancedValidator.rateLimiter.Allow() {
		if v.fallbackToBasic {
			// Graceful degradation: fall back to basic validation
			return nil // Basic format check already passed
		}
		return NewValidationError(fieldName, value, v.Name(), "validation service temporarily unavailable (rate limit exceeded)")
	}

	// External validation (placeholder - would integrate with actual domain validation service)
	if v.enhancedValidator.config.ExternalServices.DomainValidationURL != "" {
		if err := v.validateWithExternalService(domain, cacheKey); err != nil {
			if v.fallbackToBasic {
				// Graceful degradation
				return nil
			}
			return NewValidationError(fieldName, value, v.Name(), err.Error())
		}
	}

	// Cache successful validation
	_ = v.enhancedValidator.cache.Set(cacheKey, "valid", v.enhancedValidator.config.Cache.TTL)
	return nil
}

// isValidDomainFormat performs basic domain format validation
func (v *DomainValidator) isValidDomainFormat(domain string) bool {
	// Basic domain regex - this is a simplified version
	domainRegex := regexp.MustCompile(`^([a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)
	return domainRegex.MatchString(domain) && len(domain) <= 253
}

// validateWithExternalService performs external domain validation
func (v *DomainValidator) validateWithExternalService(domain, cacheKey string) error {
	url := fmt.Sprintf("%s?domain=%s", v.enhancedValidator.config.ExternalServices.DomainValidationURL, domain)

	resp, err := v.enhancedValidator.httpClient.Get(url)
	if err != nil {
		// Cache negative result for shorter duration to retry sooner
		_ = v.enhancedValidator.cache.Set(cacheKey, "external service error", v.enhancedValidator.config.Cache.TTL/10)
		return fmt.Errorf("external validation service error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("external validation failed with status %d", resp.StatusCode)
		_ = v.enhancedValidator.cache.Set(cacheKey, errorMsg, v.enhancedValidator.config.Cache.TTL)
		return fmt.Errorf("%s", errorMsg)
	}

	// Cache successful validation
	return nil
}

// RegisterEnhancedValidators registers the enhanced validators with the global registry
func RegisterEnhancedValidators(config *EnhancedValidatorConfig) error {
	// Register external email validator
	RegisterGlobalFunc("external_email", func(fieldName string, value interface{}, params map[string]interface{}) error {
		validator, err := NewExternalEmailValidator(config)
		if err != nil {
			return fmt.Errorf("failed to create external email validator: %w", err)
		}
		return validator.Validate(fieldName, value)
	})

	// Register domain validator
	RegisterGlobalFunc("domain", func(fieldName string, value interface{}, params map[string]interface{}) error {
		validator, err := NewDomainValidator(config)
		if err != nil {
			return fmt.Errorf("failed to create domain validator: %w", err)
		}
		return validator.Validate(fieldName, value)
	})

	return nil
}

// GetValidationStats returns statistics about the enhanced validator usage
func (v *EnhancedValidator) GetValidationStats() map[string]interface{} {
	stats := make(map[string]interface{})

	// Rate limiter stats
	stats["rate_limiter"] = map[string]interface{}{
		"requests_per_second": v.config.RateLimit.RequestsPerSecond,
		"burst_capacity":      v.config.RateLimit.BurstCapacity,
		"timeout":             v.config.RateLimit.Timeout.String(),
	}

	// Cache stats (if available)
	stats["cache"] = map[string]interface{}{
		"backend":     v.config.Cache.Backend,
		"ttl":         v.config.Cache.TTL.String(),
		"max_entries": v.config.Cache.MaxEntries,
	}

	// External service config
	stats["external_services"] = map[string]interface{}{
		"graceful_degradation": v.config.ExternalServices.GracefulDegradation,
		"cost_optimization":    v.config.ExternalServices.CostOptimization,
		"request_timeout":      v.config.ExternalServices.RequestTimeout.String(),
		"max_retries":          v.config.ExternalServices.MaxRetries,
	}

	return stats
}
