package tests

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

func TestEnhancedValidatorConfig_Defaults(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()

	// Test rate limiting defaults
	if config.RateLimit.RequestsPerSecond != 100 {
		t.Errorf("Expected RequestsPerSecond 100, got %d", config.RateLimit.RequestsPerSecond)
	}
	if config.RateLimit.BurstCapacity != 10 {
		t.Errorf("Expected BurstCapacity 10, got %d", config.RateLimit.BurstCapacity)
	}
	if config.RateLimit.Timeout != 5*time.Second {
		t.Errorf("Expected Timeout 5s, got %v", config.RateLimit.Timeout)
	}

	// Test cache defaults
	if config.Cache.TTL != 1*time.Hour {
		t.Errorf("Expected Cache TTL 1h, got %v", config.Cache.TTL)
	}
	if config.Cache.MaxEntries != 10000 {
		t.Errorf("Expected MaxEntries 10000, got %d", config.Cache.MaxEntries)
	}
	if config.Cache.Backend != "memory" {
		t.Errorf("Expected Backend memory, got %s", config.Cache.Backend)
	}

	// Test external service defaults
	if config.ExternalServices.RequestTimeout != 5*time.Second {
		t.Errorf("Expected RequestTimeout 5s, got %v", config.ExternalServices.RequestTimeout)
	}
	if config.ExternalServices.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries 3, got %d", config.ExternalServices.MaxRetries)
	}
	if !config.ExternalServices.GracefulDegradation {
		t.Error("Expected GracefulDegradation to be true")
	}
	if !config.ExternalServices.CostOptimization {
		t.Error("Expected CostOptimization to be true")
	}
}

func TestNewEnhancedValidator(t *testing.T) {
	tests := []struct {
		name        string
		config      *model.EnhancedValidatorConfig
		expectError bool
	}{
		{
			name:        "with nil config (uses defaults)",
			config:      nil,
			expectError: false,
		},
		{
			name:        "with default config",
			config:      model.DefaultEnhancedValidatorConfig(),
			expectError: false,
		},
		{
			name: "with custom config",
			config: func() *model.EnhancedValidatorConfig {
				config := model.DefaultEnhancedValidatorConfig()
				config.RateLimit.RequestsPerSecond = 50
				config.Cache.MaxEntries = 5000
				return config
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, err := model.NewEnhancedValidator(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && validator == nil {
				t.Error("Expected validator instance but got nil")
			}
		})
	}
}

func TestGetDefaultEnhancedValidator(t *testing.T) {
	// Test singleton behavior
	validator1, err1 := model.GetDefaultEnhancedValidator()
	validator2, err2 := model.GetDefaultEnhancedValidator()

	if err1 != nil {
		t.Errorf("Unexpected error on first call: %v", err1)
	}
	if err2 != nil {
		t.Errorf("Unexpected error on second call: %v", err2)
	}

	if validator1 == nil || validator2 == nil {
		t.Error("Expected validator instances but got nil")
	}

	// Should return the same instance (singleton)
	if validator1 != validator2 {
		t.Error("Expected same instance (singleton behavior)")
	}
}

func TestExternalEmailValidator_BasicValidation(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()
	validator, err := model.NewExternalEmailValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid email format",
			value:       "user@company.com",
			expectError: false,
		},
		{
			name:          "invalid type",
			value:         123,
			expectError:   true,
			errorContains: "must be a string",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: false, // Empty handled by required validator
		},
		{
			name:          "invalid email format",
			value:         "invalid-email",
			expectError:   true,
			errorContains: "invalid email format",
		},
		{
			name:          "malformed email",
			value:         "@example.com",
			expectError:   true,
			errorContains: "invalid email format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("email", tt.value)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestExternalEmailValidator_CostOptimization(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()
	config.ExternalServices.CostOptimization = true

	validator, err := model.NewExternalEmailValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	obviouslyInvalidEmails := []string{
		"test@example.com",
		"user@test.com",
		"admin@localhost",
		"fake@invalid",
		"dummy@fake",
		"user..double@example.com",
		".user@example.com",
		"user@example.com.",
	}

	for _, email := range obviouslyInvalidEmails {
		t.Run(fmt.Sprintf("obviously_invalid_%s", email), func(t *testing.T) {
			err := validator.Validate("email", email)

			// Should fail without external call due to cost optimization
			if err == nil {
				t.Error("Expected validation to fail for obviously invalid email")
			}
			if err != nil && !contains(err.Error(), "invalid") {
				t.Errorf("Expected error about invalid email, got: %v", err)
			}
		})
	}
}

func TestExternalEmailValidator_WithMockExternalService(t *testing.T) {
	// Create mock external service
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")

		// Mock service logic
		if email == "valid@domain.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"status": "valid", "email": "`+email+`"}`)
		} else if email == "invalid@domain.com" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, `{"status": "invalid", "email": "`+email+`"}`)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, `{"error": "service error"}`)
		}
	}))
	defer server.Close()

	// Configure validator with mock service
	config := model.DefaultEnhancedValidatorConfig()
	config.ExternalServices.EmailValidationURL = server.URL
	config.ExternalServices.GracefulDegradation = false // Disable for testing
	config.ExternalServices.CostOptimization = false    // Disable for testing

	validator, err := model.NewExternalEmailValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "external service validates as valid",
			email:       "valid@domain.com",
			expectError: false,
		},
		{
			name:        "external service validates as invalid",
			email:       "invalid@domain.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("email", tt.email)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestExternalEmailValidator_RateLimiting(t *testing.T) {
	// Create a slow mock service to test rate limiting
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond) // Simulate slow service
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "valid"}`)
	}))
	defer server.Close()

	// Configure very restrictive rate limiting
	config := model.DefaultEnhancedValidatorConfig()
	config.RateLimit.RequestsPerSecond = 2 // Very low limit
	config.RateLimit.BurstCapacity = 1     // Very low burst
	config.RateLimit.Timeout = 100 * time.Millisecond
	config.ExternalServices.EmailValidationURL = server.URL
	config.ExternalServices.GracefulDegradation = false // Disable for testing
	config.ExternalServices.CostOptimization = false    // Disable for testing

	validator, err := model.NewExternalEmailValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// First validation should succeed (within rate limit)
	err1 := validator.Validate("email", "test1@domain.com")
	if err1 != nil && !contains(err1.Error(), "valid") {
		t.Errorf("First validation should succeed or be cached, got: %v", err1)
	}

	// Immediate second validation should be rate limited
	err2 := validator.Validate("email", "test2@domain.com")
	if err2 == nil {
		t.Error("Second validation should be rate limited")
	}
	if err2 != nil && !contains(err2.Error(), "rate limit") {
		t.Errorf("Expected rate limit error, got: %v", err2)
	}
}

func TestExternalEmailValidator_Caching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"status": "valid"}`)
	}))
	defer server.Close()

	config := model.DefaultEnhancedValidatorConfig()
	config.ExternalServices.EmailValidationURL = server.URL
	config.ExternalServices.CostOptimization = false // Disable to test caching specifically
	config.Cache.TTL = 1 * time.Hour                 // Long TTL for testing

	validator, err := model.NewExternalEmailValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	email := "cached@domain.com"

	// First call should hit external service
	err1 := validator.Validate("email", email)
	if err1 != nil {
		t.Errorf("First validation failed: %v", err1)
	}

	initialCallCount := callCount

	// Second call should use cache (no additional external call)
	err2 := validator.Validate("email", email)
	if err2 != nil {
		t.Errorf("Second validation failed: %v", err2)
	}

	if callCount != initialCallCount {
		t.Errorf("Expected cached result (no additional calls), but call count increased from %d to %d", initialCallCount, callCount)
	}
}

func TestDomainValidator_BasicValidation(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()
	validator, err := model.NewDomainValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	tests := []struct {
		name          string
		value         interface{}
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid domain",
			value:       "example.com",
			expectError: false,
		},
		{
			name:        "valid subdomain",
			value:       "api.example.com",
			expectError: false,
		},
		{
			name:          "invalid type",
			value:         123,
			expectError:   true,
			errorContains: "must be a string",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: false, // Empty handled by required validator
		},
		{
			name:          "invalid domain format",
			value:         "invalid-domain",
			expectError:   true,
			errorContains: "invalid domain format",
		},
		{
			name:          "domain too long",
			value:         string(make([]byte, 300)) + ".com",
			expectError:   true,
			errorContains: "invalid domain format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("domain", tt.value)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorContains != "" {
				if !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

func TestRegisterEnhancedValidators(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()

	err := model.RegisterEnhancedValidators(config)
	if err != nil {
		t.Fatalf("Failed to register enhanced validators: %v", err)
	}

	// Test that validators are registered
	registry := model.GetDefaultRegistry()
	validators := registry.ListValidators()

	hasExternalEmail := false
	hasDomain := false

	for _, validator := range validators {
		if validator == "external_email" {
			hasExternalEmail = true
		}
		if validator == "domain" {
			hasDomain = true
		}
	}

	if !hasExternalEmail {
		t.Error("external_email validator not found in registry")
	}
	if !hasDomain {
		t.Error("domain validator not found in registry")
	}
}

func TestEnhancedValidator_GetValidationStats(t *testing.T) {
	config := model.DefaultEnhancedValidatorConfig()
	validator, err := model.NewEnhancedValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	stats := validator.GetValidationStats()

	// Check that stats contain expected keys
	if stats["rate_limiter"] == nil {
		t.Error("Expected rate_limiter stats")
	}
	if stats["cache"] == nil {
		t.Error("Expected cache stats")
	}
	if stats["external_services"] == nil {
		t.Error("Expected external_services stats")
	}

	// Check rate limiter stats
	rateLimiterStats := stats["rate_limiter"].(map[string]interface{})
	if rateLimiterStats["requests_per_second"] != 100 {
		t.Errorf("Expected requests_per_second 100, got %v", rateLimiterStats["requests_per_second"])
	}

	// Check cache stats
	cacheStats := stats["cache"].(map[string]interface{})
	if cacheStats["backend"] != "memory" {
		t.Errorf("Expected backend memory, got %v", cacheStats["backend"])
	}

	// Check external services stats
	externalStats := stats["external_services"].(map[string]interface{})
	if externalStats["graceful_degradation"] != true {
		t.Errorf("Expected graceful_degradation true, got %v", externalStats["graceful_degradation"])
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
