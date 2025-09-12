package tests

import (
	"reflect"
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
