package integration

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Real-world integration test structures
type E2EUser struct {
	ID        int       `json:"id" validate:"required,min=1"`
	Username  string    `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email     string    `json:"email" validate:"required,email"`
	FirstName string    `json:"first_name" validate:"required,alpha"`
	LastName  string    `json:"last_name" validate:"required,alpha"`
	Age       int       `json:"age" validate:"min=13,max=120"`
	IsActive  bool      `json:"is_active"`
	Profile   Profile   `json:"profile" validate:"required"`
	Settings  *Settings `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
}

type Profile struct {
	Bio       string   `json:"bio" validate:"max=500"`
	Website   *string  `json:"website"`
	Location  string   `json:"location"`
	Skills    []string `json:"skills"`
	Languages []string `json:"languages" validate:"required"`
}

type Settings struct {
	Theme         string                 `json:"theme" validate:"required"`
	Notifications map[string]interface{} `json:"notifications"`
	Privacy       PrivacySettings        `json:"privacy" validate:"required"`
}

type PrivacySettings struct {
	ProfileVisible bool `json:"profile_visible"`
	EmailVisible   bool `json:"email_visible"`
	ShowOnline     bool `json:"show_online"`
}

// API Response structure for testing
type APIResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      *T        `json:"data"`
	Error     *APIError `json:"error"`
	Meta      Meta      `json:"meta"`
	Timestamp time.Time `json:"timestamp"`
}

type APIError struct {
	Code    string                 `json:"code" validate:"required"`
	Message string                 `json:"message" validate:"required"`
	Details map[string]interface{} `json:"details"`
}

type Meta struct {
	RequestID   string `json:"request_id" validate:"required"`
	Version     string `json:"version" validate:"required"`
	ProcessTime int    `json:"process_time_ms" validate:"min=0"`
}

// Configuration structures
type DatabaseConfig struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Database string `json:"database" validate:"required"`
	SSL      bool   `json:"ssl"`
	Timeout  int    `json:"timeout" validate:"min=1000"`
}

type AppConfig struct {
	Name        string         `json:"name" validate:"required"`
	Version     string         `json:"version" validate:"required"`
	Environment string         `json:"environment" validate:"required"`
	Database    DatabaseConfig `json:"database" validate:"required"`
	Debug       bool           `json:"debug"`
	Features    []string       `json:"features"`
}

// Test comprehensive JSON parsing with validation
func TestIntegration_CompleteUserParsing(t *testing.T) {
	testCases := []struct {
		name        string
		jsonData    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid complete user",
			jsonData: `{
				"id": 123,
				"username": "johndoe",
				"email": "john@example.com",
				"first_name": "John",
				"last_name": "Doe",
				"age": 30,
				"is_active": true,
				"profile": {
					"bio": "Software developer and tech enthusiast",
					"website": "https://johndoe.dev",
					"location": "San Francisco, CA",
					"skills": ["Go", "Python", "JavaScript"],
					"languages": ["English", "Spanish"]
				},
				"settings": {
					"theme": "dark",
					"notifications": {
						"email": true,
						"push": false,
						"sms": true
					},
					"privacy": {
						"profile_visible": true,
						"email_visible": false,
						"show_online": true
					}
				},
				"created_at": "2023-01-15T10:30:00Z"
			}`,
			expectError: false,
		},
		{
			name: "User with missing required nested field",
			jsonData: `{
				"id": 124,
				"username": "janedoe",
				"email": "jane@example.com",
				"first_name": "Jane",
				"last_name": "Doe",
				"age": 25,
				"is_active": true,
				"profile": {
					"bio": "Designer and artist",
					"location": "New York, NY",
					"skills": ["Design", "Art"]
				},
				"created_at": "2023-02-20T14:45:00Z"
			}`,
			expectError: true,
			errorMsg:    "Languages",
		},
		{
			name: "User with invalid email",
			jsonData: `{
				"id": 125,
				"username": "invaliduser",
				"email": "not-an-email",
				"first_name": "Invalid",
				"last_name": "User",
				"age": 20,
				"is_active": true,
				"profile": {
					"bio": "Test user",
					"location": "Test City",
					"skills": ["Testing"],
					"languages": ["English"]
				},
				"created_at": "2023-03-10T09:00:00Z"
			}`,
			expectError: true,
			errorMsg:    "email",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := model.ParseInto[E2EUser]([]byte(tc.jsonData))

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tc.errorMsg != "" && !contains(err.Error(), tc.errorMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tc.errorMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else {
					// Validate parsing worked correctly
					if user.Username == "" {
						t.Error("Username should not be empty")
					}
					if user.Profile.Languages == nil {
						t.Error("Languages should not be nil")
					}
				}
			}
		})
	}
}

// Test API response parsing with generics
func TestIntegration_APIResponseParsing(t *testing.T) {
	t.Run("Successful API response", func(t *testing.T) {
		responseData := `{
			"success": true,
			"data": {
				"id": 456,
				"username": "apiuser",
				"email": "api@example.com",
				"first_name": "API",
				"last_name": "User",
				"age": 28,
				"is_active": true,
				"profile": {
					"bio": "API testing user",
					"location": "Cloud",
					"skills": ["API Testing"],
					"languages": ["English"]
				},
				"created_at": "2023-06-01T12:00:00Z"
			},
			"error": null,
			"meta": {
				"request_id": "req-12345",
				"version": "v1.0",
				"process_time_ms": 150
			},
			"timestamp": "2023-06-01T12:00:01Z"
		}`

		response, err := model.ParseInto[APIResponse[E2EUser]]([]byte(responseData))
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if !response.Success {
			t.Error("Expected success to be true")
		}
		if response.Data == nil {
			t.Fatal("Expected data to not be nil")
		}
		if response.Data.Username != "apiuser" {
			t.Errorf("Expected username 'apiuser', got '%s'", response.Data.Username)
		}
		if response.Meta.RequestID != "req-12345" {
			t.Errorf("Expected request ID 'req-12345', got '%s'", response.Meta.RequestID)
		}
	})

	t.Run("Error API response", func(t *testing.T) {
		errorResponseData := `{
			"success": false,
			"data": null,
			"error": {
				"code": "VALIDATION_ERROR",
				"message": "Invalid user data provided",
				"details": {
					"field": "email",
					"reason": "invalid format"
				}
			},
			"meta": {
				"request_id": "req-67890",
				"version": "v1.0",
				"process_time_ms": 25
			},
			"timestamp": "2023-06-01T12:01:00Z"
		}`

		response, err := model.ParseInto[APIResponse[E2EUser]]([]byte(errorResponseData))
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if response.Success {
			t.Error("Expected success to be false")
		}
		if response.Data != nil {
			t.Error("Expected data to be nil")
		}
		if response.Error == nil {
			t.Fatal("Expected error to not be nil")
		}
		if response.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("Expected error code 'VALIDATION_ERROR', got '%s'", response.Error.Code)
		}
	})
}

// Test configuration parsing from JSON
func TestIntegration_ConfigurationParsing(t *testing.T) {
	configData := `{
		"name": "MyApp",
		"version": "1.2.3",
		"environment": "production",
		"database": {
			"host": "db.example.com",
			"port": 5432,
			"username": "app_user",
			"password": "secure_password123",
			"database": "myapp_db",
			"ssl": true,
			"timeout": 30000
		},
		"debug": false,
		"features": ["feature_a", "feature_b", "feature_c"]
	}`

	config, err := model.ParseInto[AppConfig]([]byte(configData))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Validate parsed config
	if config.Name != "MyApp" {
		t.Errorf("Expected name 'MyApp', got '%s'", config.Name)
	}
	if config.Database.Host != "db.example.com" {
		t.Errorf("Expected database host 'db.example.com', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 5432 {
		t.Errorf("Expected database port 5432, got %d", config.Database.Port)
	}
	if !config.Database.SSL {
		t.Error("Expected SSL to be true")
	}
	if len(config.Features) != 3 {
		t.Errorf("Expected 3 features, got %d", len(config.Features))
	}
}

// Test YAML format parsing
func TestIntegration_YAMLParsing(t *testing.T) {
	yamlData := `
name: TestApp
version: "2.0.0"
environment: staging
database:
  host: yaml-db.example.com
  port: 3306
  username: yaml_user
  password: yaml_password123
  database: yaml_app_db
  ssl: false
  timeout: 15000
debug: true
features:
  - yaml_feature_1
  - yaml_feature_2
`

	config, err := model.ParseInto[AppConfig]([]byte(yamlData))
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if config.Name != "TestApp" {
		t.Errorf("Expected name 'TestApp', got '%s'", config.Name)
	}
	if config.Database.Host != "yaml-db.example.com" {
		t.Errorf("Expected database host 'yaml-db.example.com', got '%s'", config.Database.Host)
	}
	if config.Database.Port != 3306 {
		t.Errorf("Expected database port 3306, got %d", config.Database.Port)
	}
	if config.Database.SSL {
		t.Error("Expected SSL to be false")
	}
	if !config.Debug {
		t.Error("Expected debug to be true")
	}
}

// Test caching functionality
func TestIntegration_CachedParsing(t *testing.T) {
	parser, err := model.NewCachedParser[E2EUser](nil)
	if err != nil {
		t.Fatalf("Failed to create cached parser: %v", err)
	}

	userData := `{
		"id": 789,
		"username": "cacheduser",
		"email": "cached@example.com",
		"first_name": "Cached",
		"last_name": "User",
		"age": 32,
		"is_active": true,
		"profile": {
			"bio": "Testing caching",
			"location": "Cache City",
			"skills": ["Caching"],
			"languages": ["Binary"]
		},
		"created_at": "2023-07-01T08:00:00Z"
	}`

	// First parse (cache miss)
	user1, err := parser.Parse([]byte(userData))
	if err != nil {
		t.Fatalf("First parse failed: %v", err)
	}

	// Second parse (cache hit)
	user2, err := parser.Parse([]byte(userData))
	if err != nil {
		t.Fatalf("Second parse failed: %v", err)
	}

	// Results should be identical
	if !reflect.DeepEqual(user1, user2) {
		t.Error("Cached result should be identical to original")
	}

	if user1.Username != "cacheduser" {
		t.Errorf("Expected username 'cacheduser', got '%s'", user1.Username)
	}
}

// Test type coercion across different types
func TestIntegration_TypeCoercion(t *testing.T) {
	// Test data with mixed types that need coercion
	mixedData := `{
		"id": "999",
		"username": "coercionuser",
		"email": "coercion@example.com",
		"first_name": "Type",
		"last_name": "Coercion",
		"age": "25",
		"is_active": "true",
		"profile": {
			"bio": "Testing type coercion",
			"location": "Coercion Town",
			"skills": ["Go", "Python", "Rust"],
			"languages": ["English"]
		},
		"created_at": "2023-07-01T08:00:00Z"
	}`

	user, err := model.ParseInto[E2EUser]([]byte(mixedData))
	if err != nil {
		t.Fatalf("Type coercion failed: %v", err)
	}

	if user.ID != 999 {
		t.Errorf("Expected ID 999, got %d", user.ID)
	}
	if user.Age != 25 {
		t.Errorf("Expected age 25, got %d", user.Age)
	}
	if !user.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

// Test cross-field validation
func TestIntegration_CrossFieldValidation(t *testing.T) {
	// Register a cross-field validator for testing
	model.RegisterGlobalCrossFieldFunc("profile_email_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
		emailField := structValue.FieldByName("Email")
		if !emailField.IsValid() {
			return model.NewValidationError(fieldName, fieldValue, "profile_email_match", "Email field not found")
		}

		email := emailField.Interface().(string)
		profileEmail, ok := fieldValue.(string)
		if !ok {
			return nil // Skip if not string
		}

		if email != profileEmail {
			return model.NewValidationError(fieldName, fieldValue, "profile_email_match", "profile email must match user email")
		}

		return nil
	})

	type UserWithCrossValidation struct {
		Email        string `json:"email" validate:"required,email"`
		ProfileEmail string `json:"profile_email" validate:"required,profile_email_match"`
	}

	t.Run("Valid cross-field data", func(t *testing.T) {
		validData := `{
			"email": "test@example.com",
			"profile_email": "test@example.com"
		}`

		user, err := model.ParseInto[UserWithCrossValidation]([]byte(validData))
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}

		if user.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", user.Email)
		}
	})

	t.Run("Invalid cross-field data", func(t *testing.T) {
		invalidData := `{
			"email": "test@example.com",
			"profile_email": "different@example.com"
		}`

		_, err := model.ParseInto[UserWithCrossValidation]([]byte(invalidData))
		if err == nil {
			t.Error("Expected cross-field validation error")
		} else if !contains(err.Error(), "must match") {
			t.Errorf("Expected 'must match' in error, got: %v", err)
		}
	})
}

// Test error reporting and serialization
func TestIntegration_ErrorReporting(t *testing.T) {
	invalidUserData := `{
		"id": 0,
		"username": "ab",
		"email": "invalid-email",
		"first_name": "",
		"last_name": "",
		"age": 150,
		"is_active": true,
		"profile": {
			"bio": "",
			"location": "",
			"skills": [],
			"languages": []
		},
		"created_at": "invalid-date"
	}`

	_, err := model.ParseInto[E2EUser]([]byte(invalidUserData))
	if err == nil {
		t.Fatal("Expected validation errors")
	}

	// Test error serialization if ErrorList
	if errorList, ok := err.(model.ErrorList); ok {
		// Test JSON serialization
		jsonData, jsonErr := errorList.ToJSON()
		if jsonErr != nil {
			t.Fatalf("Failed to serialize errors to JSON: %v", jsonErr)
		}

		// Verify JSON structure
		var errorReport model.StructuredErrorReport
		if err := json.Unmarshal(jsonData, &errorReport); err != nil {
			t.Fatalf("Failed to parse error JSON: %v", err)
		}

		if errorReport.Count == 0 {
			t.Error("Expected error count to be greater than 0")
		}

		if len(errorReport.Errors) == 0 {
			t.Error("Expected errors array to not be empty")
		}

		t.Logf("Error report: %d field errors found", errorReport.Count)
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && someMatch(s, substr)))
}

func someMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
