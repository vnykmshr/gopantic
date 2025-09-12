package tests

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Test structs
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Product struct {
	ID       uint64  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	InStock  bool    `json:"in_stock"`
	Discount float32 `json:"discount"`
}

type Settings struct {
	MaxRetries int  `json:"max_retries"`
	Enabled    bool `json:"enabled"`
	Timeout    int  `json:"timeout"`
}

func TestParseInto_User(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    User
		wantErr bool
	}{
		{
			name:  "valid user with string ID (coercion)",
			input: []byte(`{"id":"123", "name":"Alice", "email":"alice@example.com", "age":"25"}`),
			want:  User{ID: 123, Name: "Alice", Email: "alice@example.com", Age: 25},
		},
		{
			name:  "valid user with numeric ID",
			input: []byte(`{"id":456, "name":"Bob", "email":"bob@example.com", "age":30}`),
			want:  User{ID: 456, Name: "Bob", Email: "bob@example.com", Age: 30},
		},
		{
			name:  "missing optional fields",
			input: []byte(`{"id":789, "name":"Charlie"}`),
			want:  User{ID: 789, Name: "Charlie", Email: "", Age: 0},
		},
		{
			name:  "numeric values as strings",
			input: []byte(`{"id":"999", "name":"David", "email":"david@example.com", "age":"35"}`),
			want:  User{ID: 999, Name: "David", Email: "david@example.com", Age: 35},
		},
		{
			name:    "invalid JSON",
			input:   []byte(`{"id":"123", "name":"Alice"`), // missing closing brace
			wantErr: true,
		},
		{
			name:    "invalid ID coercion",
			input:   []byte(`{"id":"not-a-number", "name":"Alice"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[User](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_Product(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Product
		wantErr bool
	}{
		{
			name:  "valid product with mixed types",
			input: []byte(`{"id":"100", "name":"Widget", "price":"29.99", "in_stock":"true", "discount":0.1}`),
			want:  Product{ID: 100, Name: "Widget", Price: 29.99, InStock: true, Discount: 0.1},
		},
		{
			name:  "boolean coercion from numbers",
			input: []byte(`{"id":200, "name":"Gadget", "price":49.99, "in_stock":1, "discount":"0.15"}`),
			want:  Product{ID: 200, Name: "Gadget", Price: 49.99, InStock: true, Discount: 0.15},
		},
		{
			name:  "boolean false from zero",
			input: []byte(`{"id":300, "name":"Tool", "price":19.99, "in_stock":0, "discount":0}`),
			want:  Product{ID: 300, Name: "Tool", Price: 19.99, InStock: false, Discount: 0},
		},
		{
			name:    "negative uint (should fail)",
			input:   []byte(`{"id":-1, "name":"Invalid"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Product](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_BooleanCoercion(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Settings
		wantErr bool
	}{
		{
			name:  "boolean from string true values",
			input: []byte(`{"max_retries":"3", "enabled":"true", "timeout":5000}`),
			want:  Settings{MaxRetries: 3, Enabled: true, Timeout: 5000},
		},
		{
			name:  "boolean from string false values",
			input: []byte(`{"max_retries":5, "enabled":"false", "timeout":"1000"}`),
			want:  Settings{MaxRetries: 5, Enabled: false, Timeout: 1000},
		},
		{
			name:  "boolean from numbers",
			input: []byte(`{"max_retries":"10", "enabled":1, "timeout":"2000"}`),
			want:  Settings{MaxRetries: 10, Enabled: true, Timeout: 2000},
		},
		{
			name:  "boolean variations",
			input: []byte(`{"max_retries":7, "enabled":"yes", "timeout":3000}`),
			want:  Settings{MaxRetries: 7, Enabled: true, Timeout: 3000},
		},
		{
			name:    "invalid boolean string",
			input:   []byte(`{"enabled":"maybe"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Settings](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_EmptyAndNilValues(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
		want  User
	}{
		{
			name:  "empty JSON object",
			input: []byte(`{}`),
			want:  User{ID: 0, Name: "", Email: "", Age: 0},
		},
		{
			name:  "null values in JSON",
			input: []byte(`{"id":null, "name":null, "email":"test@example.com", "age":null}`),
			want:  User{ID: 0, Name: "", Email: "test@example.com", Age: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[User](tt.input)
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

// Test struct with JSON tags
type TaggedStruct struct {
	RealName    string `json:"real_name"`
	DisplayName string `json:"display_name"`
	Hidden      string `json:"-"`               // Should be ignored
	Count       int    `json:"count,omitempty"` // Should work with options
}

func TestParseInto_JSONTags(t *testing.T) {
	input := []byte(`{"real_name":"John", "display_name":"Johnny", "hidden":"secret", "count":"42"}`)
	want := TaggedStruct{
		RealName:    "John",
		DisplayName: "Johnny",
		Hidden:      "", // Should remain empty due to json:"-"
		Count:       42,
	}

	got, err := model.ParseInto[TaggedStruct](input)
	if err != nil {
		t.Errorf("ParseInto() unexpected error = %v", err)
		return
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ParseInto() = %v, want %v", got, want)
	}
}

// Test struct with time.Time fields
type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
}

func TestParseInto_TimeFields(t *testing.T) {
	// Helper function to create time from string
	parseTime := func(s string) time.Time {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			panic(err)
		}
		return t
	}

	tests := []struct {
		name    string
		input   []byte
		want    Event
		wantErr bool
	}{
		{
			name: "RFC3339 format",
			input: []byte(`{
				"id": 1,
				"name": "Conference",
				"start_time": "2023-12-25T10:30:00Z",
				"end_time": "2023-12-25T15:30:00Z",
				"created_at": "2023-12-01T09:00:00Z"
			}`),
			want: Event{
				ID:        1,
				Name:      "Conference",
				StartTime: parseTime("2023-12-25T10:30:00Z"),
				EndTime:   parseTime("2023-12-25T15:30:00Z"),
				CreatedAt: parseTime("2023-12-01T09:00:00Z"),
			},
		},
		{
			name: "RFC3339Nano format",
			input: []byte(`{
				"id": 2,
				"name": "Meeting",
				"start_time": "2023-12-25T10:30:00.123456789Z",
				"end_time": "2023-12-25T11:30:00.987654321Z",
				"created_at": "2023-12-01T09:00:00.000000001Z"
			}`),
			want: Event{
				ID:        2,
				Name:      "Meeting",
				StartTime: mustParseTime(t, time.RFC3339Nano, "2023-12-25T10:30:00.123456789Z"),
				EndTime:   mustParseTime(t, time.RFC3339Nano, "2023-12-25T11:30:00.987654321Z"),
				CreatedAt: mustParseTime(t, time.RFC3339Nano, "2023-12-01T09:00:00.000000001Z"),
			},
		},
		{
			name: "Unix timestamps as integers",
			input: []byte(`{
				"id": 3,
				"name": "Workshop",
				"start_time": 1703505000,
				"end_time": 1703523000,
				"created_at": 1701417600
			}`),
			want: Event{
				ID:        3,
				Name:      "Workshop",
				StartTime: time.Unix(1703505000, 0),
				EndTime:   time.Unix(1703523000, 0),
				CreatedAt: time.Unix(1701417600, 0),
			},
		},
		{
			name: "Unix timestamps as floats",
			input: []byte(`{
				"id": 4,
				"name": "Seminar",
				"start_time": 1703505000.5,
				"end_time": 1703523000.123,
				"created_at": 1701417600.999
			}`),
			want: Event{
				ID:        4,
				Name:      "Seminar",
				StartTime: time.Unix(1703505000, 500000000),
				EndTime:   time.Unix(1703523000, 122999906), // Adjusted for float precision
				CreatedAt: time.Unix(1701417600, 999000072), // Adjusted for float precision
			},
		},
		{
			name: "Date only format",
			input: []byte(`{
				"id": 5,
				"name": "All Day Event",
				"start_time": "2023-12-25",
				"end_time": "2023-12-26",
				"created_at": "2023-12-01"
			}`),
			want: Event{
				ID:        5,
				Name:      "All Day Event",
				StartTime: mustParseTime(t, "2006-01-02", "2023-12-25"),
				EndTime:   mustParseTime(t, "2006-01-02", "2023-12-26"),
				CreatedAt: mustParseTime(t, "2006-01-02", "2023-12-01"),
			},
		},
		{
			name: "Mixed date formats",
			input: []byte(`{
				"id": 6,
				"name": "Mixed Event",
				"start_time": "2023-12-25T10:30:00",
				"end_time": "2023-12-25 15:30:00",
				"created_at": "2023-12-01T09:00:00Z"
			}`),
			want: Event{
				ID:        6,
				Name:      "Mixed Event",
				StartTime: mustParseTime(t, "2006-01-02T15:04:05", "2023-12-25T10:30:00"),
				EndTime:   mustParseTime(t, "2006-01-02 15:04:05", "2023-12-25 15:30:00"),
				CreatedAt: parseTime("2023-12-01T09:00:00Z"),
			},
		},
		{
			name: "Null time values (should become zero time)",
			input: []byte(`{
				"id": 7,
				"name": "Event with nulls",
				"start_time": null,
				"end_time": null,
				"created_at": "2023-12-01T09:00:00Z"
			}`),
			want: Event{
				ID:        7,
				Name:      "Event with nulls",
				StartTime: time.Time{},
				EndTime:   time.Time{},
				CreatedAt: parseTime("2023-12-01T09:00:00Z"),
			},
		},
		{
			name:    "Invalid time format",
			input:   []byte(`{"id": 8, "name": "Invalid", "start_time": "not-a-date"}`),
			wantErr: true,
		},
		{
			name:    "Invalid Unix timestamp",
			input:   []byte(`{"id": 9, "name": "Invalid", "start_time": "invalid-timestamp"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Event](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function to parse time with specific format
func mustParseTime(t *testing.T, format, value string) time.Time {
	parsed, err := time.Parse(format, value)
	if err != nil {
		t.Fatalf("Failed to parse time %q with format %q: %v", value, format, err)
	}
	return parsed
}

func TestParseInto_TimeCoercion_EdgeCases(t *testing.T) {
	type TimeTest struct {
		Timestamp time.Time `json:"timestamp"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    TimeTest
		wantErr bool
	}{
		{
			name:  "Time only format (today's date)",
			input: []byte(`{"timestamp": "15:04:05"}`),
			want: TimeTest{
				Timestamp: mustParseTime(t, "15:04:05", "15:04:05"),
			},
		},
		{
			name:  "Zero Unix timestamp",
			input: []byte(`{"timestamp": 0}`),
			want: TimeTest{
				Timestamp: time.Unix(0, 0),
			},
		},
		{
			name:  "Negative Unix timestamp",
			input: []byte(`{"timestamp": -1}`),
			want: TimeTest{
				Timestamp: time.Unix(-1, 0),
			},
		},
		{
			name:    "Empty string",
			input:   []byte(`{"timestamp": ""}`),
			wantErr: true,
		},
		{
			name:    "Invalid number format",
			input:   []byte(`{"timestamp": "12.34.56"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeTest](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test struct with slices and arrays
type CollectionStruct struct {
	IDs      []int       `json:"ids"`
	Tags     []string    `json:"tags"`
	Scores   [3]float64  `json:"scores"`
	Features []bool      `json:"features"`
	Weights  []float32   `json:"weights"`
	Names    [2]string   `json:"names"`
}

func TestParseInto_SlicesAndArrays(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    CollectionStruct
		wantErr bool
	}{
		{
			name: "valid arrays and slices",
			input: []byte(`{
				"ids": [1, 2, 3],
				"tags": ["go", "json", "validation"],
				"scores": [85.5, 92.0, 78.5],
				"features": [true, false, true],
				"weights": [1.2, 3.4, 5.6],
				"names": ["Alice", "Bob"]
			}`),
			want: CollectionStruct{
				IDs:      []int{1, 2, 3},
				Tags:     []string{"go", "json", "validation"},
				Scores:   [3]float64{85.5, 92.0, 78.5},
				Features: []bool{true, false, true},
				Weights:  []float32{1.2, 3.4, 5.6},
				Names:    [2]string{"Alice", "Bob"},
			},
			wantErr: false,
		},
		{
			name: "coercion from strings to numbers",
			input: []byte(`{
				"ids": ["1", "2", "3"],
				"tags": ["go"],
				"scores": ["85.5", "92.0", "78.5"],
				"features": ["true", "false", "true"],
				"weights": ["1.2", "3.4"],
				"names": ["Alice", "Bob"]
			}`),
			want: CollectionStruct{
				IDs:      []int{1, 2, 3},
				Tags:     []string{"go"},
				Scores:   [3]float64{85.5, 92.0, 78.5},
				Features: []bool{true, false, true},
				Weights:  []float32{1.2, 3.4},
				Names:    [2]string{"Alice", "Bob"},
			},
			wantErr: false,
		},
		{
			name: "empty arrays and slices",
			input: []byte(`{
				"ids": [],
				"tags": [],
				"scores": [0.0, 0.0, 0.0],
				"features": [],
				"weights": [],
				"names": ["", ""]
			}`),
			want: CollectionStruct{
				IDs:      []int{},
				Tags:     []string{},
				Scores:   [3]float64{0.0, 0.0, 0.0},
				Features: []bool{},
				Weights:  []float32{},
				Names:    [2]string{"", ""},
			},
			wantErr: false,
		},
		{
			name: "null slices (should become empty)",
			input: []byte(`{
				"ids": null,
				"tags": null,
				"scores": [1.0, 2.0, 3.0],
				"features": null,
				"weights": null,
				"names": ["Test", "User"]
			}`),
			want: CollectionStruct{
				IDs:      []int{},
				Tags:     []string{},
				Scores:   [3]float64{1.0, 2.0, 3.0},
				Features: []bool{},
				Weights:  []float32{},
				Names:    [2]string{"Test", "User"},
			},
			wantErr: false,
		},
		{
			name:    "array length mismatch",
			input:   []byte(`{"scores": [85.5, 92.0]}`), // Only 2 elements for [3]float64
			wantErr: true,
		},
		{
			name:    "array too many elements",
			input:   []byte(`{"scores": [85.5, 92.0, 78.5, 99.0]}`), // 4 elements for [3]float64
			wantErr: true,
		},
		{
			name:    "invalid element coercion in slice",
			input:   []byte(`{"ids": [1, "invalid", 3]}`),
			wantErr: true,
		},
		{
			name:    "invalid element coercion in array",
			input:   []byte(`{"scores": [85.5, "invalid", 78.5]}`),
			wantErr: true,
		},
		{
			name:    "wrong array length for names",
			input:   []byte(`{"names": ["Alice"]}`), // Only 1 element for [2]string
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[CollectionStruct](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_SliceCoercion_EdgeCases(t *testing.T) {
	type SliceTest struct {
		Numbers []int    `json:"numbers"`
		Bools   []bool   `json:"bools"`
		Strings []string `json:"strings"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    SliceTest
		wantErr bool
	}{
		{
			name: "mixed number coercion",
			input: []byte(`{
				"numbers": [1, "2", 3.0, "4"],
				"bools": [true, "false", 1, "0"],
				"strings": [123, true, 45.67]
			}`),
			want: SliceTest{
				Numbers: []int{1, 2, 3, 4},
				Bools:   []bool{true, false, true, false},
				Strings: []string{"123", "true", "45.67"},
			},
			wantErr: false,
		},
		{
			name: "boolean coercion variations",
			input: []byte(`{
				"bools": ["true", "false", "1", "0", "yes", "no", "on", "off"]
			}`),
			want: SliceTest{
				Numbers: []int{},
				Bools: []bool{true, false, true, false, true, false, true, false},
				Strings: []string{},
			},
			wantErr: false,
		},
		{
			name:    "invalid boolean in slice",
			input:   []byte(`{"bools": ["true", "maybe", "false"]}`),
			wantErr: true,
		},
		{
			name:    "non-array value for slice",
			input:   []byte(`{"numbers": "not an array"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[SliceTest](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_ArrayCoercion_EdgeCases(t *testing.T) {
	type ArrayTest struct {
		Numbers [3]int     `json:"numbers"`
		Bools   [2]bool    `json:"bools"`
		Strings [1]string  `json:"strings"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    ArrayTest
		wantErr bool
	}{
		{
			name: "exact length arrays with coercion",
			input: []byte(`{
				"numbers": ["1", "2", "3"],
				"bools": [1, 0],
				"strings": [42]
			}`),
			want: ArrayTest{
				Numbers: [3]int{1, 2, 3},
				Bools:   [2]bool{true, false},
				Strings: [1]string{"42"},
			},
			wantErr: false,
		},
		{
			name:    "array too short",
			input:   []byte(`{"numbers": [1, 2]}`), // Missing 1 element
			wantErr: true,
		},
		{
			name:    "array too long",
			input:   []byte(`{"numbers": [1, 2, 3, 4]}`), // 1 extra element
			wantErr: true,
		},
		{
			name:    "empty array when expecting elements",
			input:   []byte(`{"numbers": []}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[ArrayTest](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test structs for nested parsing
type Address struct {
	Street string `json:"street" validate:"required"`
	City   string `json:"city" validate:"required"`
	Zip    string `json:"zip" validate:"required"`
}

type Person struct {
	Name    string  `json:"name" validate:"required"`
	Age     int     `json:"age" validate:"min=0"`
	Address Address `json:"address" validate:"required"`
}

func TestParseInto_NestedStructs(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Person
		wantErr bool
	}{
		{
			name: "valid nested struct",
			input: []byte(`{
				"name": "John Doe",
				"age": 30,
				"address": {
					"street": "123 Main St",
					"city": "New York",
					"zip": "10001"
				}
			}`),
			want: Person{
				Name: "John Doe",
				Age:  30,
				Address: Address{
					Street: "123 Main St",
					City:   "New York",
					Zip:    "10001",
				},
			},
			wantErr: false,
		},
		{
			name: "nested struct with type coercion",
			input: []byte(`{
				"name": "Jane Smith",
				"age": "25",
				"address": {
					"street": "456 Oak Ave",
					"city": "Boston",
					"zip": "02101"
				}
			}`),
			want: Person{
				Name: "Jane Smith",
				Age:  25,
				Address: Address{
					Street: "456 Oak Ave",
					City:   "Boston",
					Zip:    "02101",
				},
			},
			wantErr: false,
		},
		{
			name:    "null nested struct with required fields should fail",
			input:   []byte(`{"name": "Bob Wilson", "age": 35, "address": null}`),
			wantErr: true,
		},
		{
			name:    "invalid nested struct type",
			input:   []byte(`{"name": "Test", "age": 30, "address": "not an object"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Person](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_NestedStructValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantErr     bool
		errorFields []string
	}{
		{
			name: "missing required nested field",
			input: []byte(`{
				"name": "John Doe",
				"age": 30,
				"address": {
					"city": "New York",
					"zip": "10001"
				}
			}`),
			wantErr:     true,
			errorFields: []string{"Address.Street"},
		},
		{
			name: "multiple missing nested fields",
			input: []byte(`{
				"name": "Jane Smith",
				"age": 25,
				"address": {
					"street": "123 Main St"
				}
			}`),
			wantErr:     true,
			errorFields: []string{"Address.City", "Address.Zip"},
		},
		{
			name: "invalid age with missing nested fields",
			input: []byte(`{
				"name": "Bob Wilson",
				"age": -5,
				"address": {}
			}`),
			wantErr:     true,
			errorFields: []string{"Age", "Address.Street", "Address.City", "Address.Zip"},
		},
		{
			name:    "missing required top-level field",
			input:   []byte(`{"age": 30, "address": {"street": "123 Main St", "city": "New York", "zip": "10001"}}`),
			wantErr: true,
		},
		{
			name:    "missing required nested struct",
			input:   []byte(`{"name": "Test User", "age": 30}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Person](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				errorString := err.Error()
				t.Logf("Error message: %s", errorString)
				
				for _, expectedField := range tt.errorFields {
					if !strings.Contains(errorString, expectedField) {
						t.Errorf("Expected error to contain field path %q, but got: %s", expectedField, errorString)
					}
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(got, Person{}) {
				// For error cases, we don't check the actual struct value
			}
		})
	}
}

// Test structs for deeper nested parsing (3+ levels)
type Country struct {
	Name string `json:"name" validate:"required"`
	Code string `json:"code" validate:"required,length=2"`
}

type ExtendedAddress struct {
	Street  string  `json:"street" validate:"required"`
	City    string  `json:"city" validate:"required"`
	Zip     string  `json:"zip" validate:"required"`
	Country Country `json:"country" validate:"required"`
}

type Company struct {
	Name        string          `json:"name" validate:"required"`
	Address     ExtendedAddress `json:"address" validate:"required"`
	EmployeeIDs []int           `json:"employee_ids"`
}

type Employee struct {
	Name    string  `json:"name" validate:"required"`
	Age     int     `json:"age" validate:"min=18"`
	Company Company `json:"company" validate:"required"`
}

func TestParseInto_DeepNestedStructs(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    Employee
		wantErr bool
	}{
		{
			name: "valid deeply nested struct",
			input: []byte(`{
				"name": "John Doe",
				"age": 30,
				"company": {
					"name": "TechCorp",
					"address": {
						"street": "123 Tech Ave",
						"city": "San Francisco",
						"zip": "94105",
						"country": {
							"name": "United States",
							"code": "US"
						}
					},
					"employee_ids": [1, 2, 3]
				}
			}`),
			want: Employee{
				Name: "John Doe",
				Age:  30,
				Company: Company{
					Name: "TechCorp",
					Address: ExtendedAddress{
						Street: "123 Tech Ave",
						City:   "San Francisco",
						Zip:    "94105",
						Country: Country{
							Name: "United States",
							Code: "US",
						},
					},
					EmployeeIDs: []int{1, 2, 3},
				},
			},
			wantErr: false,
		},
		{
			name: "nested struct with type coercion",
			input: []byte(`{
				"name": "Jane Smith",
				"age": "25",
				"company": {
					"name": "StartupInc",
					"address": {
						"street": "456 Innovation Blvd",
						"city": "Austin",
						"zip": "78701",
						"country": {
							"name": "United States",
							"code": "US"
						}
					},
					"employee_ids": ["1", "2", "3"]
				}
			}`),
			want: Employee{
				Name: "Jane Smith",
				Age:  25,
				Company: Company{
					Name: "StartupInc",
					Address: ExtendedAddress{
						Street: "456 Innovation Blvd",
						City:   "Austin",
						Zip:    "78701",
						Country: Country{
							Name: "United States",
							Code: "US",
						},
					},
					EmployeeIDs: []int{1, 2, 3},
				},
			},
			wantErr: false,
		},
		{
			name:    "null nested objects with required fields should fail",
			input:   []byte(`{"name": "Bob Wilson", "age": 35, "company": {"name": "RemoteCorp", "address": {"street": "789 Remote St", "city": "Denver", "zip": "80202", "country": null}, "employee_ids": null}}`),
			wantErr: true,
		},
		{
			name:    "invalid nested struct type",
			input:   []byte(`{"name": "Test", "age": 30, "company": "not an object"}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Employee](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseInto_DeepNestedValidation(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantErr     bool
		errorFields []string
	}{
		{
			name: "missing deep nested field",
			input: []byte(`{
				"name": "John Doe",
				"age": 30,
				"company": {
					"name": "TechCorp",
					"address": {
						"street": "123 Tech Ave",
						"city": "San Francisco",
						"zip": "94105",
						"country": {
							"name": "United States"
						}
					}
				}
			}`),
			wantErr:     true,
			errorFields: []string{"Company.Address.Country.Code"},
		},
		{
			name: "invalid country code length",
			input: []byte(`{
				"name": "Jane Smith",
				"age": 25,
				"company": {
					"name": "StartupInc",
					"address": {
						"street": "456 Innovation Blvd",
						"city": "Austin",
						"zip": "78701",
						"country": {
							"name": "United States",
							"code": "USA"
						}
					}
				}
			}`),
			wantErr:     true,
			errorFields: []string{"Company.Address.Country.Code"},
		},
		{
			name: "multiple deep validation errors",
			input: []byte(`{
				"name": "Bob Wilson",
				"age": 15,
				"company": {
					"address": {
						"city": "Denver",
						"country": {
							"code": "INVALID"
						}
					}
				}
			}`),
			wantErr: true,
			errorFields: []string{
				"Age",
				"Company.Name",
				"Company.Address.Street",
				"Company.Address.Zip",
				"Company.Address.Country.Name",
				"Company.Address.Country.Code",
			},
		},
		{
			name:    "missing entire nested structure",
			input:   []byte(`{"name": "Test User", "age": 30}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[Employee](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				errorString := err.Error()
				t.Logf("Error message: %s", errorString)
				
				for _, expectedField := range tt.errorFields {
					if !strings.Contains(errorString, expectedField) {
						t.Errorf("Expected error to contain field path %q, but got: %s", expectedField, errorString)
					}
				}
			}
			if !tt.wantErr && !reflect.DeepEqual(got, Employee{}) {
				// For error cases, we don't check the actual struct value
			}
		})
	}
}

func TestParseInto_NestedStructEdgeCases(t *testing.T) {
	// Test struct with nested slices of structs
	type NestedSliceStruct struct {
		Name      string    `json:"name" validate:"required"`
		Addresses []Address `json:"addresses"`
		Tags      []string  `json:"tags"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    NestedSliceStruct
		wantErr bool
	}{
		{
			name: "nested slice of structs with valid data",
			input: []byte(`{
				"name": "Multi Location User",
				"addresses": [
					{
						"street": "123 Main St",
						"city": "New York",
						"zip": "10001"
					},
					{
						"street": "456 Oak Ave",
						"city": "Boston",
						"zip": "02101"
					}
				],
				"tags": ["vip", "premium"]
			}`),
			want: NestedSliceStruct{
				Name: "Multi Location User",
				Addresses: []Address{
					{Street: "123 Main St", City: "New York", Zip: "10001"},
					{Street: "456 Oak Ave", City: "Boston", Zip: "02101"},
				},
				Tags: []string{"vip", "premium"},
			},
			wantErr: false,
		},
		{
			name: "empty nested slice",
			input: []byte(`{
				"name": "Single Location User",
				"addresses": [],
				"tags": []
			}`),
			want: NestedSliceStruct{
				Name:      "Single Location User",
				Addresses: []Address{},
				Tags:      []string{},
			},
			wantErr: false,
		},
		{
			name: "null nested slices",
			input: []byte(`{
				"name": "Minimal User",
				"addresses": null,
				"tags": null
			}`),
			want: NestedSliceStruct{
				Name:      "Minimal User",
				Addresses: []Address{},
				Tags:      []string{},
			},
			wantErr: false,
		},
		{
			name: "nested struct coercion in slice",
			input: []byte(`{
				"name": "Coercion User",
				"addresses": [
					{
						"street": "123 Main St",
						"city": "New York",
						"zip": 10001
					}
				],
				"tags": [123, true, 45.67]
			}`),
			want: NestedSliceStruct{
				Name: "Coercion User",
				Addresses: []Address{
					{Street: "123 Main St", City: "New York", Zip: "10001"},
				},
				Tags: []string{"123", "true", "45.67"},
			},
			wantErr: false,
		},
		{
			name:    "invalid nested struct in slice",
			input:   []byte(`{"name": "Invalid User", "addresses": ["not an object"]}`),
			wantErr: true,
		},
		{
			name:    "missing required field in nested slice struct",
			input:   []byte(`{"name": "Invalid User", "addresses": [{"city": "New York", "zip": "10001"}]}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[NestedSliceStruct](tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseInto() = %v, want %v", got, tt.want)
			}
		})
	}
}
