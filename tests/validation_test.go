package tests

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Test structs for validation
type ValidatedUser struct {
	ID       int    `json:"id" validate:"required,min=1"`
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"min=18,max=120"`
	Bio      string `json:"bio" validate:"max=500"`
	Name     string `json:"name" validate:"required,min=2,alpha"`
}

type ProductValidation struct {
	SKU         string  `json:"sku" validate:"required,length=8,alphanum"`
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Price       float64 `json:"price" validate:"required,min=0.01"`
	Description string  `json:"description" validate:"max=1000"`
}

type RegistrationForm struct {
	Username        string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	Terms           bool   `json:"terms" validate:"required"`
}

func TestParseInto_WithValidation_Success(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  ValidatedUser
	}{
		{
			name:  "valid user with all fields",
			input: []byte(`{"id":1,"username":"john123","email":"john@example.com","age":25,"bio":"Software developer","name":"John"}`),
			want: ValidatedUser{
				ID:       1,
				Username: "john123",
				Email:    "john@example.com",
				Age:      25,
				Bio:      "Software developer",
				Name:     "John",
			},
		},
		{
			name:  "valid user with minimum values",
			input: []byte(`{"id":1,"username":"abc","email":"a@b.co","age":18,"name":"Jo"}`),
			want: ValidatedUser{
				ID:       1,
				Username: "abc",
				Email:    "a@b.co",
				Age:      18,
				Bio:      "",
				Name:     "Jo",
			},
		},
		{
			name:  "valid user with maximum values",
			input: []byte(`{"id":999999,"username":"abcdefghij1234567890","email":"very.long.email.address@example.com","age":120,"name":"VeryLongNameButStillAlpha"}`),
			want: ValidatedUser{
				ID:       999999,
				Username: "abcdefghij1234567890",
				Email:    "very.long.email.address@example.com",
				Age:      120,
				Bio:      "",
				Name:     "VeryLongNameButStillAlpha",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[ValidatedUser](tt.input)
			if err != nil {
				t.Errorf("ParseInto() unexpected error = %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_WithValidation_Failures(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		wantErr       bool
		expectedError string
	}{
		{
			name:          "missing required field",
			input:         []byte(`{"username":"john123","email":"john@example.com"}`),
			wantErr:       true,
			expectedError: "field is required",
		},
		{
			name:          "invalid email format",
			input:         []byte(`{"id":1,"username":"john123","email":"invalid-email","age":25,"name":"John"}`),
			wantErr:       true,
			expectedError: "invalid email address format",
		},
		{
			name:          "username too short",
			input:         []byte(`{"id":1,"username":"ab","email":"john@example.com","age":25,"name":"John"}`),
			wantErr:       true,
			expectedError: "string length must be at least 3",
		},
		{
			name:          "username too long",
			input:         []byte(`{"id":1,"username":"thisusernameiswaytoolongtobevalid","email":"john@example.com","age":25,"name":"John"}`),
			wantErr:       true,
			expectedError: "string length must be at most 20",
		},
		{
			name:          "username not alphanumeric",
			input:         []byte(`{"id":1,"username":"john@123","email":"john@example.com","age":25,"name":"John"}`),
			wantErr:       true,
			expectedError: "alphanumeric characters",
		},
		{
			name:          "age too young",
			input:         []byte(`{"id":1,"username":"john123","email":"john@example.com","age":17,"name":"John"}`),
			wantErr:       true,
			expectedError: "value must be at least 18",
		},
		{
			name:          "age too old",
			input:         []byte(`{"id":1,"username":"john123","email":"john@example.com","age":121,"name":"John"}`),
			wantErr:       true,
			expectedError: "value must be at most 120",
		},
		{
			name:          "name not alphabetic",
			input:         []byte(`{"id":1,"username":"john123","email":"john@example.com","age":25,"name":"John123"}`),
			wantErr:       true,
			expectedError: "alphabetic characters",
		},
		{
			name:          "bio too long",
			input:         []byte(`{"id":1,"username":"john123","email":"john@example.com","age":25,"name":"John","bio":"` + strings.Repeat("a", 501) + `"}`),
			wantErr:       true,
			expectedError: "string length must be at most 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := model.ParseInto[ValidatedUser](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("ParseInto() error = %v, expected to contain %q", err, tt.expectedError)
				}
			}
		})
	}
}

func TestParseInto_ProductValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid product",
			input:   []byte(`{"sku":"ABC12345","name":"Widget","price":29.99,"description":"A useful widget"}`),
			wantErr: false,
		},
		{
			name:    "invalid SKU length",
			input:   []byte(`{"sku":"ABC123","name":"Widget","price":29.99}`),
			wantErr: true,
			errMsg:  "length must be exactly 8",
		},
		{
			name:    "invalid price (too low)",
			input:   []byte(`{"sku":"ABC12345","name":"Widget","price":0,"description":"Free widget"}`),
			wantErr: true,
			errMsg:  "value must be at least 0.01",
		},
		{
			name:    "SKU not alphanumeric",
			input:   []byte(`{"sku":"ABC-1234","name":"Widget","price":29.99}`),
			wantErr: true,
			errMsg:  "alphanumeric characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := model.ParseInto[ProductValidation](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseInto() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseInto_MultipleValidationErrors(t *testing.T) {
	// Test that multiple validation errors are aggregated
	input := []byte(`{"id":0,"username":"ab","email":"invalid","age":15,"name":"John123"}`)

	_, err := model.ParseInto[ValidatedUser](input)
	if err == nil {
		t.Fatalf("ParseInto() expected error but got none")
	}

	errStr := err.Error()
	expectedErrors := []string{
		"field is required",                // ID = 0 fails required validation
		"string length must be at least 3", // username "ab" too short
		"invalid email address format",     // "invalid" not an email
		"value must be at least 18",        // age 15 < 18
		"alphabetic characters",            // "John123" contains numbers
	}

	for _, expectedErr := range expectedErrors {
		if !strings.Contains(errStr, expectedErr) {
			t.Errorf("Error %q should contain %q", errStr, expectedErr)
		}
	}

	// Check that it's a multiple error format
	if !strings.Contains(errStr, "multiple errors:") {
		t.Errorf("Expected multiple errors format, got: %v", errStr)
	}
}

func TestParseInto_ValidationWithCoercion(t *testing.T) {
	// Test that validation works after type coercion
	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "string number coerced and validated",
			input:   []byte(`{"id":"42","username":"john123","email":"john@example.com","age":"25","name":"John"}`),
			wantErr: false,
		},
		{
			name:    "string number coerced but fails validation",
			input:   []byte(`{"id":"0","username":"john123","email":"john@example.com","age":"17","name":"John"}`),
			wantErr: true,
			errMsg:  "field is required", // ID = 0 fails required validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := model.ParseInto[ValidatedUser](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ParseInto() error = %v, expected to contain %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestParseInto_EmailValidator_Details(t *testing.T) {
	type EmailTest struct {
		Email string `json:"email" validate:"email"`
	}

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid simple email", "test@example.com", false},
		{"valid email with subdomain", "user@mail.example.com", false},
		{"valid email with numbers", "user123@example123.com", false},
		{"valid email with dots", "first.last@example.com", false},
		{"valid email with plus", "user+tag@example.com", false},
		{"valid email with dash", "user-name@example.com", false},

		{"invalid no @", "testexample.com", true},
		{"invalid multiple @", "test@@example.com", true},
		{"invalid no domain", "test@", true},
		{"invalid no local", "@example.com", true},
		{"invalid consecutive dots", "test..user@example.com", true},
		{"invalid start with dot", ".test@example.com", true},
		{"invalid end with dot", "test.@example.com", true},
		{"invalid domain start with dot", "test@.example.com", true},
		{"invalid domain end with dot", "test@example.com.", true},
		{"invalid too long", strings.Repeat("a", 250) + "@example.com", true},
		{"invalid no TLD", "test@localhost", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := []byte(`{"email":"` + tt.email + `"}`)
			_, err := model.ParseInto[EmailTest](input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Email validation for %q: error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

// Test individual validators
func TestRequiredValidator(t *testing.T) {
	validator := &model.RequiredValidator{}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{"nil value", nil, true},
		{"empty string", "", true},
		{"zero int", 0, true},
		{"zero float", 0.0, true},
		{"false bool", false, false}, // false is valid for required
		{"non-empty string", "hello", false},
		{"non-zero int", 42, false},
		{"non-zero float", 3.14, false},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("testField", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RequiredValidator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMinMaxValidators(t *testing.T) {
	minValidator := &model.MinValidator{Min: 5}
	maxValidator := &model.MaxValidator{Max: 10}

	tests := []struct {
		name       string
		value      interface{}
		minWantErr bool
		maxWantErr bool
	}{
		{"string length 3", "abc", true, false},
		{"string length 5", "hello", false, false},
		{"string length 8", "hello123", false, false},
		{"string length 12", "helloworldly", false, true},
		{"int 3", 3, true, false},
		{"int 5", 5, false, false},
		{"int 8", 8, false, false},
		{"int 12", 12, false, true},
		{"float 3.5", 3.5, true, false},
		{"float 7.5", 7.5, false, false},
		{"float 12.5", 12.5, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name+" min", func(t *testing.T) {
			err := minValidator.Validate("testField", tt.value)
			if (err != nil) != tt.minWantErr {
				t.Errorf("MinValidator.Validate() error = %v, wantErr %v", err, tt.minWantErr)
			}
		})

		t.Run(tt.name+" max", func(t *testing.T) {
			err := maxValidator.Validate("testField", tt.value)
			if (err != nil) != tt.maxWantErr {
				t.Errorf("MaxValidator.Validate() error = %v, wantErr %v", err, tt.maxWantErr)
			}
		})
	}
}

// Test structs for custom validators
type CustomValidatedUser struct {
	Username string `json:"username" validate:"required,contains=admin"`
	Password string `json:"password" validate:"required,password_strength"`
	Website  string `json:"website" validate:"url_format"`
}

func TestCustomValidatorFunction_Registration(t *testing.T) {
	// Create a new registry to avoid affecting other tests
	registry := model.NewValidatorRegistry()

	// Register a custom "contains" validator
	registry.RegisterFunc("contains", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "contains", "value must be a string")
		}

		substring, ok := params["value"].(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "contains", "contains validator requires a string parameter")
		}

		if !strings.Contains(str, substring) {
			return model.NewValidationError(fieldName, value, "contains",
				fmt.Sprintf("value must contain '%s'", substring))
		}
		return nil
	})

	// Test the custom validator
	validator := registry.Create("contains", map[string]interface{}{"value": "admin"})
	if validator == nil {
		t.Fatal("Expected custom validator to be created")
	}

	// Test valid case
	err := validator.Validate("username", "admin_user")
	if err != nil {
		t.Errorf("Expected no error for valid case, got: %v", err)
	}

	// Test invalid case
	err = validator.Validate("username", "regular_user")
	if err == nil {
		t.Error("Expected error for invalid case")
	} else if !strings.Contains(err.Error(), "must contain 'admin'") {
		t.Errorf("Expected error message to contain 'must contain 'admin'', got: %v", err)
	}
}

func TestCustomValidatorFunction_PasswordStrength(t *testing.T) {
	// Register a password strength validator globally
	model.RegisterGlobalFunc("password_strength", func(fieldName string, value interface{}, params map[string]interface{}) error {
		password, ok := value.(string)
		if !ok || password == "" {
			return nil // handled by required validator
		}

		if len(password) < 8 {
			return model.NewValidationError(fieldName, value, "password_strength", "password must be at least 8 characters")
		}

		hasUpper := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
		hasLower := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
		hasDigit := strings.ContainsAny(password, "0123456789")

		if !hasUpper || !hasLower || !hasDigit {
			return model.NewValidationError(fieldName, value, "password_strength", "password must contain uppercase, lowercase, and numeric characters")
		}

		return nil
	})

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid strong password",
			password: "MyPassword123",
			wantErr:  false,
		},
		{
			name:     "too short",
			password: "Pass1",
			wantErr:  true,
			errMsg:   "must be at least 8 characters",
		},
		{
			name:     "no uppercase",
			password: "mypassword123",
			wantErr:  true,
			errMsg:   "must contain uppercase, lowercase, and numeric characters",
		},
		{
			name:     "no lowercase",
			password: "MYPASSWORD123",
			wantErr:  true,
			errMsg:   "must contain uppercase, lowercase, and numeric characters",
		},
		{
			name:     "no digits",
			password: "MyPasswordAbc",
			wantErr:  true,
			errMsg:   "must contain uppercase, lowercase, and numeric characters",
		},
	}

	validator := model.GetDefaultRegistry().Create("password_strength", map[string]interface{}{})
	if validator == nil {
		t.Fatal("Expected password_strength validator to be created")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("password", tt.password)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestCustomValidatorFunction_WithStructParsing(t *testing.T) {
	// Register contains validator globally for this test
	model.RegisterGlobalFunc("contains", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "contains", "value must be a string")
		}

		substring, ok := params["value"].(string)
		if !ok {
			return model.NewValidationError(fieldName, value, "contains", "contains validator requires a string parameter")
		}

		if !strings.Contains(str, substring) {
			return model.NewValidationError(fieldName, value, "contains",
				fmt.Sprintf("value must contain '%s'", substring))
		}
		return nil
	})

	// Register URL format validator
	model.RegisterGlobalFunc("url_format", func(fieldName string, value interface{}, params map[string]interface{}) error {
		str, ok := value.(string)
		if !ok || str == "" {
			return nil // handled by required validator
		}

		// Simple URL validation
		if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
			return model.NewValidationError(fieldName, value, "url_format", "URL must start with http:// or https://")
		}

		if !strings.Contains(str, ".") {
			return model.NewValidationError(fieldName, value, "url_format", "URL must contain a domain")
		}

		return nil
	})

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid with all custom validators",
			input:   []byte(`{"username":"admin_user","password":"MyPassword123","website":"https://example.com"}`),
			wantErr: false,
		},
		{
			name:    "username missing 'admin'",
			input:   []byte(`{"username":"regular_user","password":"MyPassword123","website":"https://example.com"}`),
			wantErr: true,
			errMsg:  "must contain 'admin'",
		},
		{
			name:    "weak password",
			input:   []byte(`{"username":"admin_user","password":"weakpass","website":"https://example.com"}`),
			wantErr: true,
			errMsg:  "password must contain uppercase, lowercase, and numeric characters",
		},
		{
			name:    "invalid URL format",
			input:   []byte(`{"username":"admin_user","password":"MyPassword123","website":"not-a-url"}`),
			wantErr: true,
			errMsg:  "URL must start with http:// or https://",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := model.ParseInto[CustomValidatedUser](tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else {
					// Basic validation that parsing worked
					if user.Username == "" {
						t.Error("Expected username to be parsed")
					}
				}
			}
		})
	}
}

func TestValidatorRegistry_ListValidators(t *testing.T) {
	registry := model.NewValidatorRegistry()

	// Should have built-in validators
	validators := registry.ListValidators()
	expectedBuiltIns := []string{"required", "min", "max", "email", "length", "alpha", "alphanum"}

	for _, expected := range expectedBuiltIns {
		found := false
		for _, validator := range validators {
			if validator == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected built-in validator '%s' to be in list", expected)
		}
	}

	// Register a custom function
	registry.RegisterFunc("custom_test", func(fieldName string, value interface{}, params map[string]interface{}) error {
		return nil
	})

	// Should now include custom validator
	validators = registry.ListValidators()
	found := false
	for _, validator := range validators {
		if validator == "custom_test" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected custom validator 'custom_test' to be in list")
	}
}

func TestCustomValidatorFunction_ParameterHandling(t *testing.T) {
	registry := model.NewValidatorRegistry()

	// Register a validator that uses multiple parameters
	registry.RegisterFunc("range_check", func(fieldName string, value interface{}, params map[string]interface{}) error {
		num, ok := value.(float64)
		if !ok {
			// Try to convert from other numeric types
			switch v := value.(type) {
			case int:
				num = float64(v)
			case int64:
				num = float64(v)
			default:
				return model.NewValidationError(fieldName, value, "range_check", "value must be numeric")
			}
		}

		minValue, ok := params["min"].(float64)
		if !ok {
			minValue = 0 // default
		}

		maxValue, ok := params["max"].(float64)
		if !ok {
			maxValue = 100 // default
		}

		if num < minValue || num > maxValue {
			return model.NewValidationError(fieldName, value, "range_check",
				fmt.Sprintf("value must be between %g and %g", minValue, maxValue))
		}

		return nil
	})

	// Test with parameters parsed from tag (this would come from parseValidationRules)
	validator := registry.Create("range_check", map[string]interface{}{
		"min": float64(10),
		"max": float64(50),
	})

	if validator == nil {
		t.Fatal("Expected validator to be created")
	}

	// Test valid value
	err := validator.Validate("score", 25.0)
	if err != nil {
		t.Errorf("Expected no error for valid value, got: %v", err)
	}

	// Test invalid value (too low)
	err = validator.Validate("score", 5.0)
	if err == nil {
		t.Error("Expected error for value too low")
	} else if !strings.Contains(err.Error(), "must be between 10 and 50") {
		t.Errorf("Expected range error message, got: %v", err)
	}

	// Test invalid value (too high)
	err = validator.Validate("score", 75.0)
	if err == nil {
		t.Error("Expected error for value too high")
	} else if !strings.Contains(err.Error(), "must be between 10 and 50") {
		t.Errorf("Expected range error message, got: %v", err)
	}
}

// Test structs for cross-field validation
type PasswordRegistration struct {
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required,password_match"`
}

type DateRange struct {
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required,date_after=StartDate"`
}

func TestCrossFieldValidation_PasswordMatch(t *testing.T) {
	// Register password match validator
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

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "matching passwords",
			input:   []byte(`{"email":"user@example.com","password":"MyPassword123","confirm_password":"MyPassword123"}`),
			wantErr: false,
		},
		{
			name:    "non-matching passwords",
			input:   []byte(`{"email":"user@example.com","password":"MyPassword123","confirm_password":"DifferentPassword"}`),
			wantErr: true,
			errMsg:  "passwords do not match",
		},
		{
			name:    "empty confirm password",
			input:   []byte(`{"email":"user@example.com","password":"MyPassword123","confirm_password":""}`),
			wantErr: true,
			errMsg:  "field is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg, err := model.ParseInto[PasswordRegistration](tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else {
					if reg.Email == "" || reg.Password == "" {
						t.Error("Expected fields to be parsed successfully")
					}
				}
			}
		})
	}
}

func TestCrossFieldValidation_DateComparison(t *testing.T) {
	// Register date after validator
	model.RegisterGlobalCrossFieldFunc("date_after", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
		// Get the reference field name from parameters
		refFieldName, ok := params["value"].(string)
		if !ok {
			return model.NewValidationError(fieldName, fieldValue, "date_after", "date_after validator requires a field name parameter")
		}

		refField := structValue.FieldByName(refFieldName)
		if !refField.IsValid() {
			return model.NewValidationError(fieldName, fieldValue, "date_after", fmt.Sprintf("reference field '%s' not found", refFieldName))
		}

		endDateStr, ok := fieldValue.(string)
		if !ok {
			return model.NewValidationError(fieldName, fieldValue, "date_after", "value must be a string")
		}

		startDateStr := refField.Interface().(string)

		// Simple string comparison (in a real implementation, you'd parse dates)
		if endDateStr <= startDateStr {
			return model.NewValidationError(fieldName, fieldValue, "date_after", fmt.Sprintf("end date must be after start date (%s)", startDateStr))
		}

		return nil
	})

	tests := []struct {
		name    string
		input   []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid date range",
			input:   []byte(`{"start_date":"2023-01-01","end_date":"2023-12-31"}`),
			wantErr: false,
		},
		{
			name:    "end date before start date",
			input:   []byte(`{"start_date":"2023-12-31","end_date":"2023-01-01"}`),
			wantErr: true,
			errMsg:  "end date must be after start date",
		},
		{
			name:    "same dates",
			input:   []byte(`{"start_date":"2023-06-15","end_date":"2023-06-15"}`),
			wantErr: true,
			errMsg:  "end date must be after start date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dateRange, err := model.ParseInto[DateRange](tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else {
					if dateRange.StartDate == "" || dateRange.EndDate == "" {
						t.Error("Expected fields to be parsed successfully")
					}
				}
			}
		})
	}
}

func TestCrossFieldValidation_Registry(t *testing.T) {
	registry := model.NewValidatorRegistry()

	// Register a cross-field validator
	registry.RegisterCrossFieldFunc("field_equals", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
		refFieldName, ok := params["value"].(string)
		if !ok {
			return model.NewValidationError(fieldName, fieldValue, "field_equals", "field_equals validator requires a field name parameter")
		}

		refField := structValue.FieldByName(refFieldName)
		if !refField.IsValid() {
			return model.NewValidationError(fieldName, fieldValue, "field_equals", fmt.Sprintf("reference field '%s' not found", refFieldName))
		}

		if fieldValue != refField.Interface() {
			return model.NewValidationError(fieldName, fieldValue, "field_equals", fmt.Sprintf("field must equal %v", refField.Interface()))
		}

		return nil
	})

	// Test that the cross-field validator is created
	validator := registry.Create("field_equals", map[string]interface{}{"value": "SomeField"})
	if validator == nil {
		t.Fatal("Expected cross-field validator to be created")
	}

	// Test that it's recognized as a CrossFieldValidator
	crossFieldValidator, ok := validator.(*model.CrossFieldValidator)
	if !ok {
		t.Fatal("Expected CrossFieldValidator type")
	}

	if crossFieldValidator.Name() != "field_equals" {
		t.Errorf("Expected validator name 'field_equals', got '%s'", crossFieldValidator.Name())
	}

	// Test that ListValidators includes cross-field validators
	validators := registry.ListValidators()
	found := false
	for _, name := range validators {
		if name == "field_equals" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected cross-field validator to be in list")
	}
}

func TestCrossFieldValidation_ErrorHandling(t *testing.T) {
	registry := model.NewValidatorRegistry()

	// Register a cross-field validator that can fail
	registry.RegisterCrossFieldFunc("test_error", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
		return model.NewValidationError(fieldName, fieldValue, "test_error", "test error message")
	})

	validator := registry.Create("test_error", map[string]interface{}{})
	crossFieldValidator := validator.(*model.CrossFieldValidator)

	// Test direct validation (should fail with context error)
	err := validator.Validate("testField", "testValue")
	if err == nil {
		t.Error("Expected error when calling Validate directly on cross-field validator")
	} else if !strings.Contains(err.Error(), "cross-field validator requires full struct context") {
		t.Errorf("Expected context error, got: %v", err)
	}

	// Test struct validation (should work)
	testStruct := struct {
		TestField string
	}{TestField: "testValue"}
	structValue := reflect.ValueOf(testStruct)

	err = crossFieldValidator.ValidateWithStruct("TestField", "testValue", structValue)
	if err == nil {
		t.Error("Expected test error from cross-field validator")
	} else if !strings.Contains(err.Error(), "test error message") {
		t.Errorf("Expected test error message, got: %v", err)
	}
}

// Test structs for enhanced error reporting
type EnhancedUser struct {
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func TestEnhancedErrorReporting_StructuredErrors(t *testing.T) {
	tests := []struct {
		name                string
		input               []byte
		wantErr             bool
		expectedFieldsCount int
		expectedFields      []string
	}{
		{
			name:    "valid user",
			input:   []byte(`{"name":"John","email":"john@example.com"}`),
			wantErr: false,
		},
		{
			name:                "multiple validation errors",
			input:               []byte(`{"name":"J","email":"invalid-email"}`),
			wantErr:             true,
			expectedFieldsCount: 2,
			expectedFields:      []string{"name", "email"},
		},
		{
			name:                "missing required fields",
			input:               []byte(`{}`),
			wantErr:             true,
			expectedFieldsCount: 2,
			expectedFields:      []string{"name", "email"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := model.ParseInto[EnhancedUser](tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error but got none")
					return
				}

				// Test ErrorList functionality
				if errorList, ok := err.(model.ErrorList); ok {
					validationErrors := errorList.ValidationErrors()
					if len(validationErrors) < tt.expectedFieldsCount {
						t.Errorf("Expected at least %d validation errors, got %d", tt.expectedFieldsCount, len(validationErrors))
					}

					// Test structured report
					report := errorList.ToStructuredReport()
					if report.Count != len(report.Errors) {
						t.Errorf("Report count mismatch: count=%d, errors=%d", report.Count, len(report.Errors))
					}

					// Test JSON serialization
					jsonBytes, jsonErr := errorList.ToJSON()
					if jsonErr != nil {
						t.Errorf("Failed to serialize errors to JSON: %v", jsonErr)
					} else if len(jsonBytes) == 0 {
						t.Error("JSON serialization resulted in empty bytes")
					}

					// Test field grouping
					fieldGroups := errorList.GroupByField()
					if len(fieldGroups) == 0 {
						t.Error("Expected field groups but got none")
					}
				} else {
					t.Errorf("Expected ErrorList type, got %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				} else {
					if user.Name == "" || user.Email == "" {
						t.Error("Expected fields to be parsed successfully")
					}
				}
			}
		})
	}
}

func TestEnhancedErrorReporting_ValidationErrorDetails(t *testing.T) {
	// Test ValidationError with details
	details := map[string]interface{}{
		"min_length":    8,
		"actual_length": 3,
		"suggestions":   []string{"use uppercase letters", "add numbers"},
	}

	err := model.NewValidationErrorWithDetails("password", "user.password", "abc", "password_strength", "password too weak", details)

	if err.Field != "password" {
		t.Errorf("Expected field 'password', got '%s'", err.Field)
	}

	if err.FieldPath != "user.password" {
		t.Errorf("Expected field path 'user.password', got '%s'", err.FieldPath)
	}

	if err.Rule != "password_strength" {
		t.Errorf("Expected rule 'password_strength', got '%s'", err.Rule)
	}

	if len(err.Details) != 3 {
		t.Errorf("Expected 3 detail entries, got %d", len(err.Details))
	}

	if err.Details["min_length"].(int) != 8 {
		t.Errorf("Expected min_length detail to be 8, got %v", err.Details["min_length"])
	}

	// Test error message uses field path
	errorMsg := err.Error()
	if !strings.Contains(errorMsg, "user.password") {
		t.Errorf("Expected error message to contain field path 'user.password', got: %s", errorMsg)
	}
}

func TestEnhancedErrorReporting_JSONSerialization(t *testing.T) {
	// Create an ErrorList with multiple validation errors
	var errorList model.ErrorList

	err1 := model.NewValidationErrorWithPath("email", "user.email", "invalid", "email", "invalid email format")
	err2 := model.NewValidationErrorWithDetails("password", "user.password", "weak", "min", "password too short", map[string]interface{}{
		"required_length": 8,
		"actual_length":   4,
	})
	err3 := model.NewValidationErrorWithPath("age", "user.age", 15, "min", "age must be at least 18")

	errorList.Add(err1)
	errorList.Add(err2)
	errorList.Add(err3)

	// Test JSON serialization
	jsonBytes, err := errorList.ToJSON()
	if err != nil {
		t.Fatalf("Failed to serialize errors to JSON: %v", err)
	}

	// Parse the JSON back to verify structure
	var report model.StructuredErrorReport
	if err := json.Unmarshal(jsonBytes, &report); err != nil {
		t.Fatalf("Failed to parse serialized JSON: %v", err)
	}

	if report.Count != 3 {
		t.Errorf("Expected count 3, got %d", report.Count)
	}

	if len(report.Errors) != 3 {
		t.Errorf("Expected 3 error entries, got %d", len(report.Errors))
	}

	// Check that field paths are preserved
	fieldPaths := make(map[string]bool)
	for _, fieldError := range report.Errors {
		fieldPaths[fieldError.FieldPath] = true

		if len(fieldError.Errors) == 0 {
			t.Errorf("Expected validation errors for field %s", fieldError.FieldPath)
		}

		// Check that details are preserved for the password field
		if fieldError.FieldPath == "user.password" {
			for _, validationError := range fieldError.Errors {
				if validationError.Rule == "min" {
					if validationError.Details == nil {
						t.Error("Expected details for password validation error")
					} else if validationError.Details["required_length"].(float64) != 8 {
						t.Error("Expected required_length detail to be preserved")
					}
				}
			}
		}
	}

	expectedPaths := []string{"user.email", "user.password", "user.age"}
	for _, expectedPath := range expectedPaths {
		if !fieldPaths[expectedPath] {
			t.Errorf("Expected field path '%s' in serialized report", expectedPath)
		}
	}
}

func TestEnhancedErrorReporting_FieldGrouping(t *testing.T) {
	var errorList model.ErrorList

	// Add multiple errors for the same field
	err1 := model.NewValidationErrorWithPath("username", "user.username", "", "required", "field is required")
	err2 := model.NewValidationErrorWithPath("username", "user.username", "", "min", "minimum length is 3")

	// Add error for different field
	err3 := model.NewValidationErrorWithPath("email", "user.email", "invalid", "email", "invalid email format")

	errorList.Add(err1)
	errorList.Add(err2)
	errorList.Add(err3)

	// Test field grouping
	groups := errorList.GroupByField()

	if len(groups) != 2 {
		t.Errorf("Expected 2 field groups, got %d", len(groups))
	}

	usernameErrors, exists := groups["user.username"]
	if !exists {
		t.Error("Expected 'user.username' field group")
	} else if len(usernameErrors) != 2 {
		t.Errorf("Expected 2 errors for 'user.username', got %d", len(usernameErrors))
	}

	emailErrors, exists := groups["user.email"]
	if !exists {
		t.Error("Expected 'user.email' field group")
	} else if len(emailErrors) != 1 {
		t.Errorf("Expected 1 error for 'user.email', got %d", len(emailErrors))
	}
}
