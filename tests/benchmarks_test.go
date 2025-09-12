package tests

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Benchmark structs for complex type parsing performance

type BenchmarkUser struct {
	ID        int        `json:"id" validate:"required,min=1"`
	Name      string     `json:"name" validate:"required,min=2,alpha"`
	Email     string     `json:"email" validate:"required,email"`
	Age       int        `json:"age" validate:"min=18,max=120"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type BenchmarkAddress struct {
	Street  string `json:"street" validate:"required,min=5"`
	City    string `json:"city" validate:"required,min=2"`
	Zip     string `json:"zip" validate:"required,length=5"`
	Country string `json:"country" validate:"required,length=2"`
}

type BenchmarkNestedUser struct {
	ID      int              `json:"id" validate:"required,min=1"`
	Name    string           `json:"name" validate:"required,min=2,alpha"`
	Email   string           `json:"email" validate:"required,email"`
	Age     int              `json:"age" validate:"min=18,max=120"`
	Address BenchmarkAddress `json:"address" validate:"required"`
	Tags    []string         `json:"tags"`
	Scores  []float64        `json:"scores"`
}

type BenchmarkDeepNested struct {
	User    BenchmarkNestedUser `json:"user" validate:"required"`
	Company struct {
		Name    string           `json:"name" validate:"required"`
		Address BenchmarkAddress `json:"address" validate:"required"`
	} `json:"company" validate:"required"`
}

type BenchmarkComplexStruct struct {
	Users     []BenchmarkNestedUser `json:"users" validate:"required"`
	Addresses []BenchmarkAddress    `json:"addresses"`
	Events    []struct {
		ID        int       `json:"id" validate:"required"`
		Name      string    `json:"name" validate:"required"`
		Timestamp time.Time `json:"timestamp"`
		Duration  *int      `json:"duration"`
	} `json:"events"`
	Settings struct {
		Enabled       bool     `json:"enabled"`
		MaxUsers      *int     `json:"max_users" validate:"min=1"`
		AllowedEmails []string `json:"allowed_emails"`
	} `json:"settings"`
}

// Basic parsing benchmarks

func BenchmarkParseInto_SimpleStruct(b *testing.B) {
	data := []byte(`{
		"id": "123",
		"name": "Alice",
		"email": "alice@example.com",
		"age": "30",
		"created_at": "2023-01-15T10:30:00Z"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_SimpleStructWithPointer(b *testing.B) {
	data := []byte(`{
		"id": "123",
		"name": "Alice",
		"email": "alice@example.com",
		"age": "30",
		"created_at": "2023-01-15T10:30:00Z",
		"updated_at": "2023-12-15T15:45:00Z"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Nested struct benchmarks

func BenchmarkParseInto_NestedStruct(b *testing.B) {
	data := []byte(`{
		"id": "456",
		"name": "Bob",
		"email": "bob@test.com",
		"age": "25",
		"address": {
			"street": "123 Main Street",
			"city": "Springfield",
			"zip": "12345",
			"country": "US"
		},
		"tags": ["developer", "golang", "backend"],
		"scores": ["85.5", "92.0", "78.5"]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkNestedUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_DeepNestedStruct(b *testing.B) {
	data := []byte(`{
		"user": {
			"id": "456",
			"name": "Bob",
			"email": "bob@test.com",
			"age": "25",
			"address": {
				"street": "123 Main Street",
				"city": "Springfield",
				"zip": "12345",
				"country": "US"
			},
			"tags": ["developer", "golang"],
			"scores": ["85.5", "92.0"]
		},
		"company": {
			"name": "Tech Corp",
			"address": {
				"street": "456 Business Ave",
				"city": "Metropolis", 
				"zip": "54321",
				"country": "US"
			}
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkDeepNested](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Array and slice benchmarks

func BenchmarkParseInto_LargeSlice(b *testing.B) {
	data := []byte(`{
		"users": [
			{
				"id": "1", "name": "Alice", "email": "alice@test.com", "age": "30",
				"address": {"street": "123 Main St", "city": "Town", "zip": "12345", "country": "US"},
				"tags": ["admin", "senior"], "scores": ["95.5", "88.0"]
			},
			{
				"id": "2", "name": "Bob", "email": "bob@test.com", "age": "25",
				"address": {"street": "456 Oak Ave", "city": "City", "zip": "67890", "country": "US"},
				"tags": ["developer"], "scores": ["82.5", "90.0"]
			},
			{
				"id": "3", "name": "Carol", "email": "carol@test.com", "age": "28",
				"address": {"street": "789 Pine Rd", "city": "Village", "zip": "11111", "country": "US"},
				"tags": ["designer", "lead"], "scores": ["91.0", "87.5"]
			}
		],
		"addresses": [
			{"street": "Corporate Blvd", "city": "Business", "zip": "99999", "country": "US"},
			{"street": "Tech Street", "city": "Innovation", "zip": "88888", "country": "US"}
		],
		"events": [
			{"id": "100", "name": "Launch", "timestamp": "2023-01-01T00:00:00Z"},
			{"id": "101", "name": "Update", "timestamp": "2023-06-01T12:00:00Z", "duration": "3600"},
			{"id": "102", "name": "Maintenance", "timestamp": "2023-12-01T02:00:00Z"}
		],
		"settings": {
			"enabled": true,
			"max_users": "1000",
			"allowed_emails": ["admin@company.com", "support@company.com"]
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkComplexStruct](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Time parsing benchmarks

func BenchmarkParseInto_TimeFields_RFC3339(b *testing.B) {
	data := []byte(`{
		"id": "1",
		"name": "Alice",
		"email": "alice@test.com",
		"age": "30",
		"created_at": "2023-01-15T10:30:45Z",
		"updated_at": "2023-12-25T15:45:30+05:45"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_TimeFields_UnixTimestamp(b *testing.B) {
	data := []byte(`{
		"id": "1",
		"name": "Alice",
		"email": "alice@test.com",
		"age": "30",
		"created_at": 1703505045,
		"updated_at": 1703508645.123
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Type coercion benchmarks

func BenchmarkParseInto_HeavyCoercion(b *testing.B) {
	data := []byte(`{
		"id": "789",
		"name": "Charlie",
		"email": "charlie@example.com",
		"age": "35",
		"created_at": 1703505045,
		"updated_at": 1703508645.456
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_MixedTypeCoercion(b *testing.B) {
	data := []byte(`{
		"users": [
			{
				"id": "1", "name": "Alice", "email": "alice@test.com", "age": "30",
				"address": {"street": "123 Main St", "city": "Town", "zip": "12345", "country": "US"},
				"tags": ["admin"], "scores": ["95.5"]
			}
		],
		"events": [
			{"id": "100", "name": "Event", "timestamp": 1703505045, "duration": "3600"}
		],
		"settings": {
			"enabled": "true",
			"max_users": "500"
		}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkComplexStruct](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Validation benchmarks

func BenchmarkParseInto_WithValidation(b *testing.B) {
	data := []byte(`{
		"id": "123",
		"name": "Alice",
		"email": "alice@example.com",
		"age": "30",
		"created_at": "2023-01-15T10:30:00Z"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_NestedValidation(b *testing.B) {
	data := []byte(`{
		"id": "456",
		"name": "Bob",
		"email": "bob@test.com",
		"age": "25",
		"address": {
			"street": "123 Main Street",
			"city": "Springfield",
			"zip": "12345",
			"country": "US"
		},
		"tags": ["developer"],
		"scores": ["85.5"]
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkNestedUser](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation benchmarks

func BenchmarkParseInto_MemoryAllocation(b *testing.B) {
	data := []byte(`{
		"users": [
			{
				"id": "1", "name": "Alice", "email": "alice@test.com", "age": "30",
				"address": {"street": "123 Main St", "city": "Town", "zip": "12345", "country": "US"},
				"tags": ["admin", "senior", "lead", "expert"], 
				"scores": ["95.5", "88.0", "92.5", "85.0", "90.5"]
			}
		],
		"events": [
			{"id": "100", "name": "Launch", "timestamp": "2023-01-01T00:00:00Z"},
			{"id": "101", "name": "Update", "timestamp": "2023-06-01T12:00:00Z", "duration": "3600"}
		],
		"settings": {
			"enabled": true,
			"max_users": "1000",
			"allowed_emails": ["admin@company.com", "support@company.com", "dev@company.com"]
		}
	}`)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchmarkComplexStruct](data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Error handling benchmarks

func BenchmarkParseInto_ValidationError(b *testing.B) {
	data := []byte(`{
		"id": "0",
		"name": "A",
		"email": "invalid-email",
		"age": "15",
		"created_at": "2023-01-15T10:30:00Z"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = model.ParseInto[BenchmarkUser](data)
		// Intentionally ignoring error since we're benchmarking error handling
	}
}

func BenchmarkParseInto_ParseError(b *testing.B) {
	data := []byte(`{
		"id": "not-a-number",
		"name": "Alice",
		"email": "alice@test.com",
		"age": "not-a-number",
		"created_at": "invalid-date"
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = model.ParseInto[BenchmarkUser](data)
		// Intentionally ignoring error since we're benchmarking error handling
	}
}

// Comparative benchmarks against different approaches

func BenchmarkParseInto_VsStandardJSON_Simple(b *testing.B) {
	data := []byte(`{
		"id": 123,
		"name": "Alice",
		"email": "alice@example.com",
		"age": 30,
		"created_at": "2023-01-15T10:30:00Z"
	}`)

	b.Run("gopantic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := model.ParseInto[BenchmarkUser](data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("standard_json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var user BenchmarkUser
			if err := json.Unmarshal(data, &user); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParseInto_VsStandardJSON_Complex(b *testing.B) {
	data := []byte(`{
		"user": {
			"id": 456,
			"name": "Bob",
			"email": "bob@test.com",
			"age": 25,
			"address": {
				"street": "123 Main Street",
				"city": "Springfield",
				"zip": "12345",
				"country": "US"
			},
			"tags": ["developer", "golang"],
			"scores": [85.5, 92.0]
		},
		"company": {
			"name": "Tech Corp",
			"address": {
				"street": "456 Business Ave",
				"city": "Metropolis", 
				"zip": "54321",
				"country": "US"
			}
		}
	}`)

	b.Run("gopantic", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := model.ParseInto[BenchmarkDeepNested](data)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("standard_json", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var nested BenchmarkDeepNested
			if err := json.Unmarshal(data, &nested); err != nil {
				b.Fatal(err)
			}
		}
	})
}
