package main

import (
	"fmt"
	"log"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// User represents a user in our system
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Product represents a product with various data types
type Product struct {
	ID       uint64  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"in_stock"`
	Discount float32 `json:"discount"`
}

// Config represents application configuration
type Config struct {
	MaxRetries int  `json:"max_retries"`
	Enabled    bool `json:"enabled"`
	Timeout    int  `json:"timeout_ms"`
	Debug      bool `json:"debug"`
}

func main() {
	fmt.Println("🚀 gopantic - Basic Parsing Examples")
	fmt.Println("=====================================")

	// Example 1: Basic User parsing with type coercion
	fmt.Println("\n1. User parsing with string-to-int coercion:")
	userJSON := `{"id": "42", "name": "Alice Johnson", "email": "alice@example.com", "age": "28"}`
	fmt.Printf("Input JSON: %s\n", userJSON)

	user, err := model.ParseInto[User]([]byte(userJSON))
	if err != nil {
		log.Printf("Error parsing user: %v", err)
	} else {
		fmt.Printf("Parsed User: %+v\n", user)
		fmt.Printf("- ID (int): %d\n", user.ID)
		fmt.Printf("- Name: %q\n", user.Name)
		fmt.Printf("- Email: %q\n", user.Email)
		fmt.Printf("- Age (int): %d\n", user.Age)
	}

	// Example 2: Product with mixed types and coercion
	fmt.Println("\n2. Product parsing with mixed type coercion:")
	productJSON := `{"id": "12345", "name": "Wireless Headphones", "price": "99.99", "in_stock": "true", "discount": 0.15}`
	fmt.Printf("Input JSON: %s\n", productJSON)

	product, err := model.ParseInto[Product]([]byte(productJSON))
	if err != nil {
		log.Printf("Error parsing product: %v", err)
	} else {
		fmt.Printf("Parsed Product: %+v\n", product)
		fmt.Printf("- ID (uint64): %d\n", product.ID)
		fmt.Printf("- Name: %q\n", product.Name)
		fmt.Printf("- Price (float64): $%.2f\n", product.Price)
		fmt.Printf("- InStock (bool): %t\n", product.InStock)
		fmt.Printf("- Discount (float32): %.2f%%\n", product.Discount*100)
	}

	// Example 3: Configuration with boolean coercion
	fmt.Println("\n3. Config parsing with boolean variations:")
	configJSON := `{"max_retries": "5", "enabled": "yes", "timeout_ms": 3000, "debug": 1}`
	fmt.Printf("Input JSON: %s\n", configJSON)

	config, err := model.ParseInto[Config]([]byte(configJSON))
	if err != nil {
		log.Printf("Error parsing config: %v", err)
	} else {
		fmt.Printf("Parsed Config: %+v\n", config)
		fmt.Printf("- MaxRetries (int): %d\n", config.MaxRetries)
		fmt.Printf("- Enabled (bool): %t\n", config.Enabled)
		fmt.Printf("- Timeout (int): %d ms\n", config.Timeout)
		fmt.Printf("- Debug (bool): %t\n", config.Debug)
	}

	// Example 4: Handling missing fields
	fmt.Println("\n4. Parsing with missing optional fields:")
	partialUserJSON := `{"id": 999, "name": "Bob"}`
	fmt.Printf("Input JSON: %s\n", partialUserJSON)

	partialUser, err := model.ParseInto[User]([]byte(partialUserJSON))
	if err != nil {
		log.Printf("Error parsing partial user: %v", err)
	} else {
		fmt.Printf("Parsed User: %+v\n", partialUser)
		fmt.Println("- Missing fields default to zero values")
		fmt.Printf("- Email: %q (empty string)\n", partialUser.Email)
		fmt.Printf("- Age: %d (zero value)\n", partialUser.Age)
	}

	// Example 5: Error handling
	fmt.Println("\n5. Error handling for invalid data:")
	invalidJSON := `{"id": "not-a-number", "name": "Invalid User"}`
	fmt.Printf("Input JSON: %s\n", invalidJSON)

	_, err = model.ParseInto[User]([]byte(invalidJSON))
	if err != nil {
		fmt.Printf("Expected error: %v\n", err)
		fmt.Println("✓ gopantic correctly caught the invalid type conversion")
	}

	// Example 6: Demonstrating various boolean coercions
	fmt.Println("\n6. Boolean coercion examples:")
	booleanExamples := []string{
		`{"enabled": "true"}`,
		`{"enabled": "false"}`,
		`{"enabled": "yes"}`,
		`{"enabled": "no"}`,
		`{"enabled": "1"}`,
		`{"enabled": "0"}`,
		`{"enabled": 1}`,
		`{"enabled": 0}`,
	}

	type BoolTest struct {
		Enabled bool `json:"enabled"`
	}

	for _, example := range booleanExamples {
		result, err := model.ParseInto[BoolTest]([]byte(example))
		if err != nil {
			fmt.Printf("  %s → ERROR: %v\n", example, err)
		} else {
			fmt.Printf("  %s → %t\n", example, result.Enabled)
		}
	}

	fmt.Println("\n✨ All examples completed!")
	fmt.Println("gopantic successfully parsed and coerced all the different data types!")
}
