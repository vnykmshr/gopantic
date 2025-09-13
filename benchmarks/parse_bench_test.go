package benchmarks

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Test structures for benchmarks
type SimpleUser struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

type ComplexUser struct {
	ID       int                    `json:"id" validate:"required,min=1"`
	Name     string                 `json:"name" validate:"required,min=2,max=50"`
	Email    string                 `json:"email" validate:"required,email"`
	Age      int                    `json:"age" validate:"min=0,max=150"`
	Active   bool                   `json:"active"`
	Tags     []string               `json:"tags"`
	Scores   []int                  `json:"scores"`
	Metadata map[string]interface{} `json:"metadata"`
	Address  Address                `json:"address" validate:"required"`
}

type Address struct {
	Street  string `json:"street" validate:"required"`
	City    string `json:"city" validate:"required"`
	Zip     string `json:"zip" validate:"required,min=5,max=10"`
	Country string `json:"country" validate:"required,alpha,min=2,max=3"`
}

// Benchmark data
var (
	simpleUserJSON = []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com",
		"active": true
	}`)

	complexUserJSON = []byte(`{
		"id": 456,
		"name": "Jane Smith",
		"email": "jane@example.com", 
		"age": 28,
		"active": true,
		"tags": ["admin", "user", "premium"],
		"scores": [85, 92, 78, 96],
		"metadata": {
			"last_login": "2023-12-01T10:00:00Z",
			"preferences": {
				"theme": "dark",
				"notifications": true
			}
		},
		"address": {
			"street": "123 Main St",
			"city": "New York", 
			"zip": "10001",
			"country": "USA"
		}
	}`)

	largeUserArrayJSON = func() []byte {
		users := make([]map[string]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			users[i] = map[string]interface{}{
				"id":     i + 1,
				"name":   "User " + string(rune('A'+i%26)),
				"email":  "user" + string(rune('0'+i%10)) + "@example.com",
				"active": i%2 == 0,
			}
		}
		data, _ := json.Marshal(users)
		return data
	}()
)

// Simple parsing benchmarks
func BenchmarkStdJSON_SimpleUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var user SimpleUser
		if err := json.Unmarshal(simpleUserJSON, &user); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_SimpleUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[SimpleUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_SimpleUser_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[SimpleUser](nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Complex parsing benchmarks with validation
func BenchmarkStdJSON_ComplexUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var user ComplexUser
		if err := json.Unmarshal(complexUserJSON, &user); err != nil {
			b.Fatal(err)
		}
		// Standard JSON doesn't include validation, so this is just parsing
	}
}

func BenchmarkGopantic_ComplexUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ComplexUser](complexUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_ComplexUser_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[ComplexUser](nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(complexUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Large array benchmarks
func BenchmarkStdJSON_LargeArray(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var users []SimpleUser
		if err := json.Unmarshal(largeUserArrayJSON, &users); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_LargeArray(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[[]SimpleUser](largeUserArrayJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation comparison benchmarks
func BenchmarkMemory_StdJSON_SimpleUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var user SimpleUser
		_ = json.Unmarshal(simpleUserJSON, &user)
		// Force allocation to prevent optimization
		_ = user.Name + user.Email
	}
}

func BenchmarkMemory_Gopantic_SimpleUser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		user, _ := model.ParseInto[SimpleUser](simpleUserJSON)
		// Force allocation to prevent optimization
		_ = user.Name + user.Email
	}
}

// Validation-specific benchmarks (gopantic advantage)
func BenchmarkValidation_ComplexUser_Success(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ComplexUser](complexUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkValidation_ComplexUser_Failure(b *testing.B) {
	invalidJSON := []byte(`{
		"id": 0,
		"name": "",
		"email": "invalid-email",
		"age": -5,
		"active": true,
		"tags": [],
		"scores": [],
		"metadata": {},
		"address": {
			"street": "",
			"city": "",
			"zip": "123",
			"country": "INVALID"
		}
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ComplexUser](invalidJSON)
		if err == nil {
			b.Fatal("Expected validation errors")
		}
	}
}

// YAML vs JSON parsing comparison
func BenchmarkYAML_ComplexUser(b *testing.B) {
	yamlData := []byte(`
id: 456
name: Jane Smith
email: jane@example.com
age: 28
active: true
tags: [admin, user, premium]
scores: [85, 92, 78, 96]
metadata:
  last_login: "2023-12-01T10:00:00Z"
  preferences:
    theme: dark
    notifications: true
address:
  street: "123 Main St"
  city: "New York"
  zip: "10001"
  country: "USA"
`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ComplexUser](yamlData)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Cache performance benchmarks
func BenchmarkCache_Hit_Ratio(b *testing.B) {
	parser, _ := model.NewCachedParser[SimpleUser](nil)

	// Warm up cache
	parser.Parse(simpleUserJSON)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCache_Miss_Pattern(b *testing.B) {
	parser, _ := model.NewCachedParser[SimpleUser](nil)

	// Create different JSON payloads to simulate cache misses
	variants := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		user := map[string]interface{}{
			"id":     i + 1,
			"name":   "User " + string(rune('A'+i)),
			"email":  "user" + string(rune('0'+i)) + "@example.com",
			"active": i%2 == 0,
		}
		data, _ := json.Marshal(user)
		variants[i] = data
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(variants[i%10])
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Cross-field validation benchmark
func BenchmarkCrossFieldValidation(b *testing.B) {
	type Registration struct {
		Email           string `json:"email" validate:"required,email"`
		Password        string `json:"password" validate:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" validate:"required,password_match"`
	}

	// Register cross-field validator for benchmark
	model.RegisterGlobalCrossFieldFunc("password_match", func(fieldName string, fieldValue interface{}, structValue reflect.Value, params map[string]interface{}) error {
		// Simple password match validation for benchmark
		return nil
	})

	registrationJSON := []byte(`{
		"email": "user@example.com",
		"password": "mypassword123",
		"confirm_password": "mypassword123"
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[Registration](registrationJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}
