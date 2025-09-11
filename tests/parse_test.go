package tests

import (
	"reflect"
	"testing"

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
