package tests

import (
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
