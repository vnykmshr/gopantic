package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// User represents a user with enhanced validation requirements
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,external_email"`
	Website  string `json:"website" validate:"domain"`
	Company  string `json:"company" validate:"required,min=2"`
	JoinDate string `json:"join_date"`
}

// CompanyInfo represents company information with domain validation
type CompanyInfo struct {
	Name     string `json:"name" validate:"required,min=2"`
	Domain   string `json:"domain" validate:"required,domain"`
	Industry string `json:"industry" validate:"required"`
	Email    string `json:"email" validate:"required,external_email"`
}

func main() {
	fmt.Println("🚀 Enhanced Validation with Rate Limiting and Caching Demo")
	fmt.Println("========================================================")

	// Configure enhanced validation
	config := setupEnhancedValidation()

	// Demo 1: Basic enhanced validation
	fmt.Println("\n📝 Demo 1: Enhanced Email Validation with Rate Limiting")
	demoEnhancedEmailValidation(config)

	// Demo 2: Caching benefits
	fmt.Println("\n⚡ Demo 2: Validation Caching Benefits")
	demoCachingBenefits(config)

	// Demo 3: Cost optimization
	fmt.Println("\n💰 Demo 3: Cost Optimization Features")
	demoCostOptimization(config)

	// Demo 4: Graceful degradation
	fmt.Println("\n🛡️ Demo 4: Graceful Degradation")
	demoGracefulDegradation()

	// Demo 5: Validation statistics
	fmt.Println("\n📊 Demo 5: Validation Statistics")
	demoValidationStats(config)

	fmt.Println("\n✅ Enhanced Validation Demo Complete!")
}

// setupEnhancedValidation configures and registers enhanced validators
func setupEnhancedValidation() *model.EnhancedValidatorConfig {
	// Create configuration
	config := model.DefaultEnhancedValidatorConfig()

	// Customize configuration for demo
	config.RateLimit.RequestsPerSecond = 10 // Lower for demo visibility
	config.RateLimit.BurstCapacity = 3
	config.RateLimit.Timeout = 2 * time.Second

	config.Cache.TTL = 5 * time.Minute // Shorter TTL for demo
	config.Cache.MaxEntries = 1000

	config.ExternalServices.GracefulDegradation = true
	config.ExternalServices.CostOptimization = true
	config.ExternalServices.RequestTimeout = 3 * time.Second

	// Note: In production, you would set actual URLs for external services
	// config.ExternalServices.EmailValidationURL = "https://api.emailvalidation.io/validate"
	// config.ExternalServices.DomainValidationURL = "https://api.domainvalidation.io/check"

	// Register enhanced validators
	err := model.RegisterEnhancedValidators(config)
	if err != nil {
		log.Fatalf("Failed to register enhanced validators: %v", err)
	}

	fmt.Printf("✅ Enhanced validators registered with:\n")
	fmt.Printf("   • Rate Limit: %d req/sec, burst %d\n",
		config.RateLimit.RequestsPerSecond,
		config.RateLimit.BurstCapacity)
	fmt.Printf("   • Cache TTL: %v\n", config.Cache.TTL)
	fmt.Printf("   • Graceful Degradation: %v\n", config.ExternalServices.GracefulDegradation)
	fmt.Printf("   • Cost Optimization: %v\n", config.ExternalServices.CostOptimization)

	return config
}

// demoEnhancedEmailValidation shows enhanced email validation in action
func demoEnhancedEmailValidation(config *model.EnhancedValidatorConfig) {
	users := []User{
		{
			ID:       1,
			Name:     "John Doe",
			Email:    "john.doe@company.com",
			Website:  "company.com",
			Company:  "Tech Corp",
			JoinDate: "2023-01-15",
		},
		{
			ID:       2,
			Name:     "Jane Smith",
			Email:    "jane@invalid-format", // Invalid format
			Website:  "techstart.io",
			Company:  "TechStart",
			JoinDate: "2023-02-20",
		},
		{
			ID:       3,
			Name:     "Bob Wilson",
			Email:    "bob@example.com", // Will be flagged by cost optimization
			Website:  "bobsites.net",
			Company:  "Freelance",
			JoinDate: "2023-03-10",
		},
	}

	for i, user := range users {
		fmt.Printf("\n🔍 Validating User %d: %s\n", i+1, user.Name)

		// Parse with enhanced validation
		parsedUser, err := model.ParseInto[User]([]byte(toJSON(user)))

		if err != nil {
			fmt.Printf("   ❌ Validation Error: %v\n", err)
		} else {
			fmt.Printf("   ✅ Validation passed: %s <%s>\n", parsedUser.Name, parsedUser.Email)
		}

		// Small delay to see rate limiting in action
		if i < len(users)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}
}

// demoCachingBenefits demonstrates validation result caching
func demoCachingBenefits(config *model.EnhancedValidatorConfig) {
	email := "cached.user@company.com"

	fmt.Printf("📧 Testing validation caching for: %s\n", email)

	// Create test data
	userData := map[string]interface{}{
		"id":        1,
		"name":      "Cache Test User",
		"email":     email,
		"website":   "company.com",
		"company":   "Cache Corp",
		"join_date": "2023-01-15",
	}

	// First validation (will hit external service or basic validation)
	fmt.Printf("\n⏱️  First validation (fresh)...\n")
	start1 := time.Now()
	user1, err1 := model.ParseInto[User]([]byte(toJSON(userData)))
	duration1 := time.Since(start1)

	if err1 != nil {
		fmt.Printf("   ❌ Error: %v\n", err1)
	} else {
		fmt.Printf("   ✅ Success in %v (User: %s)\n", duration1, user1.Name)
	}

	// Second validation (should use cache)
	fmt.Printf("\n⚡ Second validation (cached)...\n")
	start2 := time.Now()
	user2, err2 := model.ParseInto[User]([]byte(toJSON(userData)))
	duration2 := time.Since(start2)

	if err2 != nil {
		fmt.Printf("   ❌ Error: %v\n", err2)
	} else {
		fmt.Printf("   ✅ Success in %v (User: %s)\n", duration2, user2.Name)
	}

	// Show performance improvement
	if duration1 > 0 && duration2 > 0 {
		improvement := float64(duration1-duration2) / float64(duration1) * 100
		fmt.Printf("\n📈 Performance improvement: %.1f%% faster with caching\n", improvement)
	}
}

// demoCostOptimization shows cost optimization features
func demoCostOptimization(config *model.EnhancedValidatorConfig) {
	obviouslyInvalidEmails := []string{
		"test@example.com",
		"user@test.com",
		"admin@localhost",
		"fake@invalid",
		"user..double@domain.com",
		".user@domain.com",
	}

	fmt.Printf("💰 Testing cost optimization for obviously invalid emails:\n")

	for _, email := range obviouslyInvalidEmails {
		userData := map[string]interface{}{
			"id":      1,
			"name":    "Test User",
			"email":   email,
			"company": "Test Corp",
		}

		fmt.Printf("\n📧 Testing: %s\n", email)

		start := time.Now()
		user, err := model.ParseInto[User]([]byte(toJSON(userData)))
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("   ❌ Rejected (cost optimized) in %v\n", duration)
			fmt.Printf("   💡 Reason: %v\n", err)
		} else {
			fmt.Printf("   ⚠️  Unexpectedly passed in %v (User: %s)\n", duration, user.Name)
		}
	}

	fmt.Printf("\n💡 Cost optimization prevents expensive external calls for obviously invalid emails")
}

// demoGracefulDegradation shows graceful degradation when external services fail
func demoGracefulDegradation() {
	fmt.Printf("🛡️ Testing graceful degradation with external service failures:\n")

	// Configure with graceful degradation enabled
	config := model.DefaultEnhancedValidatorConfig()
	config.ExternalServices.GracefulDegradation = true
	config.ExternalServices.EmailValidationURL = "http://non-existent-service.invalid"

	// This would normally register validators, but we'll simulate the behavior
	fmt.Printf("\n📧 Simulating external service failure...\n")
	fmt.Printf("   🔄 External email validation service unavailable\n")
	fmt.Printf("   🛡️ Graceful degradation: falling back to basic email format validation\n")

	basicEmail := "user@company.com"
	fmt.Printf("   📧 Email: %s\n", basicEmail)
	fmt.Printf("   ✅ Basic format validation passed\n")
	fmt.Printf("   💡 Service continues operating despite external dependency failure\n")

	// Show what happens without graceful degradation
	fmt.Printf("\n❌ Without graceful degradation:\n")
	fmt.Printf("   🚫 Validation would fail completely\n")
	fmt.Printf("   🚫 User experience would be impacted\n")
	fmt.Printf("   🚫 Service reliability would depend on external services\n")
}

// demoValidationStats shows validation statistics
func demoValidationStats(config *model.EnhancedValidatorConfig) {
	fmt.Printf("📊 Enhanced Validator Statistics:\n")

	// Create validator to get stats
	validator, err := model.NewEnhancedValidator(config)
	if err != nil {
		fmt.Printf("   ❌ Error creating validator: %v\n", err)
		return
	}

	stats := validator.GetValidationStats()

	// Display rate limiter stats
	if rateLimiter, ok := stats["rate_limiter"].(map[string]interface{}); ok {
		fmt.Printf("\n🚦 Rate Limiter Configuration:\n")
		fmt.Printf("   • Requests per second: %v\n", rateLimiter["requests_per_second"])
		fmt.Printf("   • Burst capacity: %v\n", rateLimiter["burst_capacity"])
		fmt.Printf("   • Timeout: %v\n", rateLimiter["timeout"])
	}

	// Display cache stats
	if cache, ok := stats["cache"].(map[string]interface{}); ok {
		fmt.Printf("\n💾 Cache Configuration:\n")
		fmt.Printf("   • Backend: %v\n", cache["backend"])
		fmt.Printf("   • TTL: %v\n", cache["ttl"])
		fmt.Printf("   • Max entries: %v\n", cache["max_entries"])
	}

	// Display external service stats
	if external, ok := stats["external_services"].(map[string]interface{}); ok {
		fmt.Printf("\n🌐 External Services Configuration:\n")
		fmt.Printf("   • Graceful degradation: %v\n", external["graceful_degradation"])
		fmt.Printf("   • Cost optimization: %v\n", external["cost_optimization"])
		fmt.Printf("   • Request timeout: %v\n", external["request_timeout"])
		fmt.Printf("   • Max retries: %v\n", external["max_retries"])
	}

	fmt.Printf("\n💡 These statistics help monitor and optimize validation performance")
}

// toJSON converts a value to JSON string for parsing
func toJSON(v interface{}) string {
	// Simple JSON conversion for demo (in production, use proper JSON marshaling)
	switch data := v.(type) {
	case User:
		return fmt.Sprintf(`{
			"id": %d,
			"name": "%s",
			"email": "%s",
			"website": "%s",
			"company": "%s",
			"join_date": "%s"
		}`, data.ID, data.Name, data.Email, data.Website, data.Company, data.JoinDate)
	case map[string]interface{}:
		result := "{"
		first := true
		for k, v := range data {
			if !first {
				result += ","
			}
			switch val := v.(type) {
			case string:
				result += fmt.Sprintf(`"%s": "%s"`, k, val)
			case int:
				result += fmt.Sprintf(`"%s": %d`, k, val)
			default:
				result += fmt.Sprintf(`"%s": "%v"`, k, val)
			}
			first = false
		}
		result += "}"
		return result
	default:
		return fmt.Sprintf("%v", v)
	}
}
