package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Essential benchmark structs
type BenchUser struct {
	ID        int       `json:"id" validate:"required,min=1"`
	Name      string    `json:"name" validate:"required,min=2"`
	Email     string    `json:"email" validate:"required,email"`
	Age       int       `json:"age" validate:"min=18,max=120"`
	CreatedAt time.Time `json:"created_at"`
	Active    bool      `json:"active"`
}

type BenchConfig struct {
	Name     string            `json:"name" validate:"required"`
	Debug    bool              `json:"debug"`
	Port     int               `json:"port" validate:"min=1000,max=65535"`
	Timeout  int               `json:"timeout" validate:"min=1"`
	Features []string          `json:"features"`
	Settings map[string]string `json:"settings"`
}

// Benchmark: Simple struct parsing vs standard library
func BenchmarkSimpleParsing_Gopantic(b *testing.B) {
	data := []byte(`{
		"id": 123,
		"name": "John Doe", 
		"email": "john@example.com",
		"age": 30,
		"created_at": "2023-01-01T12:00:00Z",
		"active": true
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSimpleParsing_StandardJSON(b *testing.B) {
	data := []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com", 
		"age": 30,
		"created_at": "2023-01-01T12:00:00Z",
		"active": true
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var user BenchUser
		err := json.Unmarshal(data, &user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Complex struct with validation
func BenchmarkComplexParsing_WithValidation(b *testing.B) {
	data := []byte(`{
		"name": "MyApp",
		"debug": true,
		"port": 8080,
		"timeout": 30,
		"features": ["auth", "logging", "metrics"],
		"settings": {
			"theme": "dark",
			"lang": "en"
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchConfig](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Array parsing performance
func BenchmarkArrayParsing(b *testing.B) {
	data := []byte(`[
		{"id": 1, "name": "Alice", "email": "alice@example.com", "age": 25, "created_at": "2023-01-01T12:00:00Z", "active": true},
		{"id": 2, "name": "Bob", "email": "bob@example.com", "age": 30, "created_at": "2023-01-02T12:00:00Z", "active": false},
		{"id": 3, "name": "Charlie", "email": "charlie@example.com", "age": 35, "created_at": "2023-01-03T12:00:00Z", "active": true}
	]`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[[]BenchUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Type coercion performance
func BenchmarkTypeCoercion(b *testing.B) {
	data := []byte(`{
		"id": "123",
		"name": "John Doe",
		"email": "john@example.com",
		"age": "30",
		"created_at": "1672574400",
		"active": "true"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Memory allocation
func BenchmarkMemoryAllocation(b *testing.B) {
	data := []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com", 
		"age": 30,
		"created_at": "2023-01-01T12:00:00Z",
		"active": true
	}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark: Cached parsing performance
func BenchmarkCachedParsing(b *testing.B) {
	data := []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30, 
		"created_at": "2023-01-01T12:00:00Z",
		"active": true
	}`)

	parser := model.NewCachedParser[BenchUser](nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
