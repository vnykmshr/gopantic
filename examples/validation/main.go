package main

import (
	"fmt"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// UserRegistration demonstrates comprehensive validation for a user registration form
type UserRegistration struct {
	Username        string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
	FullName        string `json:"full_name" validate:"required,min=2,alpha"`
	Age             int    `json:"age" validate:"required,min=18,max=120"`
	Bio             string `json:"bio" validate:"max=500"`
	Terms           bool   `json:"terms" validate:"required"`
}

// Product demonstrates validation for an e-commerce product
type Product struct {
	SKU         string  `json:"sku" validate:"required,length=8,alphanum"`
	Name        string  `json:"name" validate:"required,min=1,max=100"`
	Price       float64 `json:"price" validate:"required,min=0.01"`
	Category    string  `json:"category" validate:"required,alpha"`
	Description string  `json:"description" validate:"max=1000"`
	InStock     bool    `json:"in_stock"`
}

// APIKey demonstrates validation for API configuration
type APIKey struct {
	Name        string `json:"name" validate:"required,min=3,max=50,alphanum"`
	Key         string `json:"key" validate:"required,length=32,alphanum"`
	Permissions string `json:"permissions" validate:"required"`
	Active      bool   `json:"active"`
}

func main() {
	fmt.Println("gopantic - Validation Framework Examples")
	fmt.Println("============================================")

	// Example 1: Valid user registration
	fmt.Println("\n1. Valid User Registration:")
	validUserJSON := `{
		"username": "johndoe123",
		"email": "john.doe@example.com",
		"password": "SecurePass123",
		"confirm_password": "SecurePass123",
		"full_name": "John",
		"age": 28,
		"bio": "Software engineer passionate about Go",
		"terms": true
	}`

	user, err := model.ParseInto[UserRegistration]([]byte(validUserJSON))
	if err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Printf("User registered successfully: %s (%s)\n", user.Username, user.Email)
		fmt.Printf("   Full Name: %s, Age: %d\n", user.FullName, user.Age)
		fmt.Printf("   Bio: %q\n", user.Bio)
		fmt.Printf("   Terms Accepted: %t\n", user.Terms)
	}

	// Example 2: Invalid user registration (multiple validation errors)
	fmt.Println("\n2. Invalid User Registration (Multiple Errors):")
	invalidUserJSON := `{
		"username": "jd",
		"email": "invalid-email",
		"password": "weak",
		"confirm_password": "",
		"full_name": "John123",
		"age": 15,
		"bio": "` + generateLongString(600) + `",
		"terms": false
	}`

	_, err = model.ParseInto[UserRegistration]([]byte(invalidUserJSON))
	if err != nil {
		fmt.Printf("Expected validation errors:\n%v\n", err)
	} else {
		fmt.Println("Expected validation to fail!")
	}

	// Example 3: Valid product
	fmt.Println("\n3. Valid Product Creation:")
	validProductJSON := `{
		"sku": "WDG12345",
		"name": "Wireless Widget",
		"price": 99.99,
		"category": "Electronics",
		"description": "A high-quality wireless widget for all your widget needs",
		"in_stock": true
	}`

	product, err := model.ParseInto[Product]([]byte(validProductJSON))
	if err != nil {
		fmt.Printf("Product validation failed: %v\n", err)
	} else {
		fmt.Printf("Product created: %s - %s\n", product.SKU, product.Name)
		fmt.Printf("   Price: $%.2f, Category: %s\n", product.Price, product.Category)
		fmt.Printf("   In Stock: %t\n", product.InStock)
	}

	// Example 4: Invalid product
	fmt.Println("\n4. Invalid Product (SKU and Price Issues):")
	invalidProductJSON := `{
		"sku": "WDG-123",
		"name": "",
		"price": 0,
		"category": "Electronics123",
		"description": "A widget"
	}`

	_, err = model.ParseInto[Product]([]byte(invalidProductJSON))
	if err != nil {
		fmt.Printf("Expected product validation errors:\n%v\n", err)
	} else {
		fmt.Println("Expected product validation to fail!")
	}

	// Example 5: API Key validation
	fmt.Println("\n5. Valid API Key Configuration:")
	validAPIKeyJSON := `{
		"name": "prodapi1",
		"key": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
		"permissions": "read,write,admin",
		"active": true
	}`

	apiKey, err := model.ParseInto[APIKey]([]byte(validAPIKeyJSON))
	if err != nil {
		fmt.Printf("API Key validation failed: %v\n", err)
	} else {
		fmt.Printf("API Key created: %s\n", apiKey.Name)
		fmt.Printf("   Key: %s...\n", apiKey.Key[:8])
		fmt.Printf("   Permissions: %s, Active: %t\n", apiKey.Permissions, apiKey.Active)
	}

	// Example 6: Type coercion with validation
	fmt.Println("\n6. Type Coercion with Validation:")
	coercionJSON := `{
		"username": "alice123",
		"email": "alice@example.com",
		"password": "AlicePass123",
		"confirm_password": "AlicePass123",
		"full_name": "Alice",
		"age": "25",
		"bio": "Data scientist",
		"terms": "true"
	}`

	coercedUser, err := model.ParseInto[UserRegistration]([]byte(coercionJSON))
	if err != nil {
		fmt.Printf("Coercion with validation failed: %v\n", err)
	} else {
		fmt.Printf("User with coerced types: %s\n", coercedUser.Username)
		fmt.Printf("   Age (string->int): %d\n", coercedUser.Age)
		fmt.Printf("   Terms (string->bool): %t\n", coercedUser.Terms)
	}

	// Example 7: Demonstrating different validation rules
	fmt.Println("\n7. Validation Rule Demonstrations:")

	// Required validation
	fmt.Println("\n   Required Field Validation:")
	testRequiredValidation()

	// Min/Max validation
	fmt.Println("\n   Min/Max Length Validation:")
	testMinMaxValidation()

	// Email validation
	fmt.Println("\n   Email Format Validation:")
	testEmailValidation()

	// Alpha/Alphanum validation
	fmt.Println("\n   Alpha/Alphanumeric Validation:")
	testAlphaValidation()

	fmt.Println("\nAll validation examples completed!")
	fmt.Println("gopantic successfully demonstrated comprehensive validation capabilities!")
}

func testRequiredValidation() {
	type RequiredTest struct {
		Name string `json:"name" validate:"required"`
	}

	tests := []string{
		`{"name": "John"}`, // Valid
		`{"name": ""}`,     // Invalid - empty
		`{}`,               // Invalid - missing
	}

	for _, test := range tests {
		_, err := model.ParseInto[RequiredTest]([]byte(test))
		if err != nil {
			fmt.Printf("     %s -> %v\n", test, err)
		} else {
			fmt.Printf("     %s -> Valid\n", test)
		}
	}
}

func testMinMaxValidation() {
	type LengthTest struct {
		Username string `json:"username" validate:"min=3,max=10"`
	}

	tests := []string{
		`{"username": "jo"}`,               // Too short
		`{"username": "john"}`,             // Valid
		`{"username": "johnsmith"}`,        // Valid
		`{"username": "verylongusername"}`, // Too long
	}

	for _, test := range tests {
		_, err := model.ParseInto[LengthTest]([]byte(test))
		if err != nil {
			fmt.Printf("     %s -> %v\n", test, err)
		} else {
			fmt.Printf("     %s -> Valid\n", test)
		}
	}
}

func testEmailValidation() {
	type EmailTest struct {
		Email string `json:"email" validate:"email"`
	}

	tests := []string{
		`{"email": "user@example.com"}`,      // Valid
		`{"email": "user.name@example.com"}`, // Valid
		`{"email": "invalid-email"}`,         // Invalid
		`{"email": "user@"}`,                 // Invalid
		`{"email": "@example.com"}`,          // Invalid
	}

	for _, test := range tests {
		_, err := model.ParseInto[EmailTest]([]byte(test))
		if err != nil {
			fmt.Printf("     %s -> Invalid email\n", test)
		} else {
			fmt.Printf("     %s -> Valid email\n", test)
		}
	}
}

func testAlphaValidation() {
	type AlphaTest struct {
		Name string `json:"name" validate:"alpha"`
		Code string `json:"code" validate:"alphanum"`
	}

	tests := []string{
		`{"name": "John", "code": "ABC123"}`,    // Valid
		`{"name": "John123", "code": "ABC123"}`, // Invalid name
		`{"name": "John", "code": "ABC-123"}`,   // Invalid code
	}

	for _, test := range tests {
		_, err := model.ParseInto[AlphaTest]([]byte(test))
		if err != nil {
			fmt.Printf("     %s -> %v\n", test, err)
		} else {
			fmt.Printf("     %s -> Valid\n", test)
		}
	}
}

func generateLongString(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "a"
	}
	return result
}
