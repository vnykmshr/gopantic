package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Comprehensive time parsing test structures
type TimeContainer struct {
	Timestamp     time.Time  `json:"timestamp"`
	OptionalTime  *time.Time `json:"optional_time"`
	StartTime     time.Time  `json:"start_time"`
	EndTime       time.Time  `json:"end_time"`
	CreatedAt     time.Time  `json:"created_at"`
	LastSeen      time.Time  `json:"last_seen"`
	ProcessedTime time.Time  `json:"processed_time"`
}

// Helper function to create time pointers
func timePtr(t time.Time) *time.Time {
	return &t
}

// Helper function to parse time or fail test
func mustParseTimeComprehensive(t *testing.T, layout, value string) time.Time {
	parsed, err := time.Parse(layout, value)
	if err != nil {
		t.Fatalf("Failed to parse time %q with layout %q: %v", value, layout, err)
	}
	return parsed
}

func TestTimeParsingComprehensive_RFC3339Variants(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    TimeContainer
		wantErr bool
	}{
		{
			name: "RFC3339 UTC",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45Z",
				"start_time": "2023-01-01T00:00:00Z",
				"end_time": "2023-12-31T23:59:59Z"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:45Z"),
				StartTime: mustParseTimeComprehensive(t, time.RFC3339, "2023-01-01T00:00:00Z"),
				EndTime:   mustParseTimeComprehensive(t, time.RFC3339, "2023-12-31T23:59:59Z"),
			},
		},
		{
			name: "RFC3339 with positive timezone offset",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45+05:30",
				"start_time": "2023-01-01T00:00:00+08:00",
				"end_time": "2023-12-31T23:59:59+02:00"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:45+05:30"),
				StartTime: mustParseTimeComprehensive(t, time.RFC3339, "2023-01-01T00:00:00+08:00"),
				EndTime:   mustParseTimeComprehensive(t, time.RFC3339, "2023-12-31T23:59:59+02:00"),
			},
		},
		{
			name: "RFC3339 with negative timezone offset",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45-05:00",
				"start_time": "2023-01-01T00:00:00-08:00",
				"end_time": "2023-12-31T23:59:59-07:00"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:45-05:00"),
				StartTime: mustParseTimeComprehensive(t, time.RFC3339, "2023-01-01T00:00:00-08:00"),
				EndTime:   mustParseTimeComprehensive(t, time.RFC3339, "2023-12-31T23:59:59-07:00"),
			},
		},
		{
			name: "RFC3339Nano with full nanosecond precision",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45.123456789Z",
				"start_time": "2023-01-01T00:00:00.000000001Z",
				"end_time": "2023-12-31T23:59:59.999999999Z"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-12-25T10:30:45.123456789Z"),
				StartTime: mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-01-01T00:00:00.000000001Z"),
				EndTime:   mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-12-31T23:59:59.999999999Z"),
			},
		},
		{
			name: "RFC3339Nano with varying precision",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45.1Z",
				"start_time": "2023-01-01T00:00:00.12Z",
				"end_time": "2023-12-31T23:59:59.123456Z"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-12-25T10:30:45.1Z"),
				StartTime: mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-01-01T00:00:00.12Z"),
				EndTime:   mustParseTimeComprehensive(t, time.RFC3339Nano, "2023-12-31T23:59:59.123456Z"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeContainer](tt.input)
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

func TestTimeParsingComprehensive_ISO8601Formats(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    TimeContainer
		wantErr bool
	}{
		{
			name: "ISO 8601 without timezone",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45",
				"start_time": "2023-01-01T00:00:00",
				"end_time": "2023-12-31T23:59:59"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, "2006-01-02T15:04:05", "2023-12-25T10:30:45"),
				StartTime: mustParseTimeComprehensive(t, "2006-01-02T15:04:05", "2023-01-01T00:00:00"),
				EndTime:   mustParseTimeComprehensive(t, "2006-01-02T15:04:05", "2023-12-31T23:59:59"),
			},
		},
		{
			name: "ISO 8601 UTC explicit",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45Z",
				"start_time": "2023-01-01T00:00:00Z",
				"end_time": "2023-12-31T23:59:59Z"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, "2006-01-02T15:04:05Z", "2023-12-25T10:30:45Z"),
				StartTime: mustParseTimeComprehensive(t, "2006-01-02T15:04:05Z", "2023-01-01T00:00:00Z"),
				EndTime:   mustParseTimeComprehensive(t, "2006-01-02T15:04:05Z", "2023-12-31T23:59:59Z"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeContainer](tt.input)
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

func TestTimeParsingComprehensive_CommonFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    TimeContainer
		wantErr bool
	}{
		{
			name: "Common database format",
			input: []byte(`{
				"timestamp": "2023-12-25 10:30:45",
				"start_time": "2023-01-01 00:00:00",
				"end_time": "2023-12-31 23:59:59"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, "2006-01-02 15:04:05", "2023-12-25 10:30:45"),
				StartTime: mustParseTimeComprehensive(t, "2006-01-02 15:04:05", "2023-01-01 00:00:00"),
				EndTime:   mustParseTimeComprehensive(t, "2006-01-02 15:04:05", "2023-12-31 23:59:59"),
			},
		},
		{
			name: "Date only format",
			input: []byte(`{
				"timestamp": "2023-12-25",
				"start_time": "2023-01-01",
				"end_time": "2023-12-31"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, "2006-01-02", "2023-12-25"),
				StartTime: mustParseTimeComprehensive(t, "2006-01-02", "2023-01-01"),
				EndTime:   mustParseTimeComprehensive(t, "2006-01-02", "2023-12-31"),
			},
		},
		{
			name: "Time only format",
			input: []byte(`{
				"timestamp": "10:30:45",
				"start_time": "00:00:00",
				"end_time": "23:59:59"
			}`),
			want: TimeContainer{
				Timestamp: mustParseTimeComprehensive(t, "15:04:05", "10:30:45"),
				StartTime: mustParseTimeComprehensive(t, "15:04:05", "00:00:00"),
				EndTime:   mustParseTimeComprehensive(t, "15:04:05", "23:59:59"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeContainer](tt.input)
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

func TestTimeParsingComprehensive_UnixTimestamps(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    TimeContainer
		wantErr bool
	}{
		{
			name: "Unix timestamp integers",
			input: []byte(`{
				"timestamp": 1703505045,
				"start_time": 1672531200,
				"end_time": 1704067199
			}`),
			want: TimeContainer{
				Timestamp: time.Unix(1703505045, 0),
				StartTime: time.Unix(1672531200, 0),
				EndTime:   time.Unix(1704067199, 0),
			},
		},
		{
			name: "Unix timestamp floats with fractional seconds",
			input: []byte(`{
				"timestamp": 1703505045.123,
				"start_time": 1672531200.456789,
				"end_time": 1704067199.999
			}`),
			want: TimeContainer{
				Timestamp: time.Unix(1703505045, 122999906), // Actual precision from float64 conversion
				StartTime: time.Unix(1672531200, 456789016), // Actual precision from float64 conversion
				EndTime:   time.Unix(1704067199, 999000072), // Actual precision from float64 conversion
			},
		},
		{
			name: "Unix timestamp edge cases",
			input: []byte(`{
				"timestamp": 0,
				"start_time": -1,
				"end_time": 2147483647
			}`),
			want: TimeContainer{
				Timestamp: time.Unix(0, 0),
				StartTime: time.Unix(-1, 0),
				EndTime:   time.Unix(2147483647, 0),
			},
		},
		{
			name: "Very large Unix timestamps (year 2038 and beyond)",
			input: []byte(`{
				"timestamp": 2147483648,
				"start_time": 4102444800,
				"end_time": 32503680000
			}`),
			want: TimeContainer{
				Timestamp: time.Unix(2147483648, 0),
				StartTime: time.Unix(4102444800, 0),
				EndTime:   time.Unix(32503680000, 0),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeContainer](tt.input)
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

func TestTimeParsingComprehensive_EdgeCasesAndErrors(t *testing.T) {
	type SimpleTime struct {
		Time time.Time `json:"time"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    SimpleTime
		wantErr bool
	}{
		{
			name:    "Empty string",
			input:   []byte(`{"time": ""}`),
			wantErr: true,
		},
		{
			name:    "Invalid date format",
			input:   []byte(`{"time": "2023-13-45"}`),
			wantErr: true,
		},
		{
			name:    "Invalid time format",
			input:   []byte(`{"time": "25:99:99"}`),
			wantErr: true,
		},
		{
			name:    "Invalid RFC3339 format",
			input:   []byte(`{"time": "2023-12-25T25:30:00Z"}`),
			wantErr: true,
		},
		{
			name:    "Invalid timezone format",
			input:   []byte(`{"time": "2023-12-25T10:30:00+25:00"}`),
			wantErr: true,
		},
		{
			name:    "Non-numeric Unix timestamp string",
			input:   []byte(`{"time": "not-a-number"}`),
			wantErr: true,
		},
		{
			name:    "Invalid fractional Unix timestamp",
			input:   []byte(`{"time": "123.45.67"}`),
			wantErr: true,
		},
		{
			name:    "Boolean value",
			input:   []byte(`{"time": true}`),
			wantErr: true,
		},
		{
			name:    "Array value",
			input:   []byte(`{"time": [2023, 12, 25]}`),
			wantErr: true,
		},
		{
			name:    "Object value",
			input:   []byte(`{"time": {"year": 2023, "month": 12, "day": 25}}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[SimpleTime](tt.input)
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

func TestTimeParsingComprehensive_BoundaryValues(t *testing.T) {
	type SimpleTime struct {
		Time time.Time `json:"time"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    SimpleTime
		wantErr bool
	}{
		{
			name:  "Leap year February 29",
			input: []byte(`{"time": "2024-02-29T12:00:00Z"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "2024-02-29T12:00:00Z"),
			},
		},
		{
			name:    "Invalid leap year February 29",
			input:   []byte(`{"time": "2023-02-29T12:00:00Z"}`),
			wantErr: true,
		},
		{
			name:  "Year 1 AD",
			input: []byte(`{"time": "0001-01-01T00:00:00Z"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "0001-01-01T00:00:00Z"),
			},
		},
		{
			name:  "Year 9999",
			input: []byte(`{"time": "9999-12-31T23:59:59Z"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "9999-12-31T23:59:59Z"),
			},
		},
		{
			name:  "End of month boundaries",
			input: []byte(`{"time": "2023-01-31T23:59:59Z"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "2023-01-31T23:59:59Z"),
			},
		},
		{
			name:    "Invalid day for month",
			input:   []byte(`{"time": "2023-02-30T12:00:00Z"}`),
			wantErr: true,
		},
		{
			name:  "Maximum timezone offset",
			input: []byte(`{"time": "2023-12-25T12:00:00+14:00"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T12:00:00+14:00"),
			},
		},
		{
			name:  "Minimum timezone offset",
			input: []byte(`{"time": "2023-12-25T12:00:00-12:00"}`),
			want: SimpleTime{
				Time: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T12:00:00-12:00"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[SimpleTime](tt.input)
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

func TestTimeParsingComprehensive_PointerTypes(t *testing.T) {
	type TimeWithOptional struct {
		Required time.Time  `json:"required"`
		Optional *time.Time `json:"optional"`
	}

	tests := []struct {
		name    string
		input   []byte
		want    TimeWithOptional
		wantErr bool
	}{
		{
			name: "Both times present",
			input: []byte(`{
				"required": "2023-12-25T10:30:00Z",
				"optional": "2023-12-26T11:45:00Z"
			}`),
			want: TimeWithOptional{
				Required: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:00Z"),
				Optional: timePtr(mustParseTimeComprehensive(t, time.RFC3339, "2023-12-26T11:45:00Z")),
			},
		},
		{
			name: "Optional time missing",
			input: []byte(`{
				"required": "2023-12-25T10:30:00Z"
			}`),
			want: TimeWithOptional{
				Required: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:00Z"),
				Optional: nil,
			},
		},
		{
			name: "Optional time null",
			input: []byte(`{
				"required": "2023-12-25T10:30:00Z",
				"optional": null
			}`),
			want: TimeWithOptional{
				Required: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:00Z"),
				Optional: nil,
			},
		},
		{
			name: "Optional with Unix timestamp",
			input: []byte(`{
				"required": "2023-12-25T10:30:00Z",
				"optional": 1703505000
			}`),
			want: TimeWithOptional{
				Required: mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:00Z"),
				Optional: timePtr(time.Unix(1703505000, 0)),
			},
		},
		{
			name:    "Required time missing",
			input:   []byte(`{"optional": "2023-12-25T10:30:00Z"}`),
			wantErr: false, // Missing non-pointer field gets zero value
			want: TimeWithOptional{
				Required: time.Time{}, // Zero time
				Optional: timePtr(mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:00Z")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeWithOptional](tt.input)
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

func TestTimeParsingComprehensive_MixedFormats(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    TimeContainer
		wantErr bool
	}{
		{
			name: "All different formats in one struct",
			input: []byte(`{
				"timestamp": "2023-12-25T10:30:45Z",
				"start_time": 1703505045,
				"end_time": "2023-12-25 15:30:45",
				"created_at": "2023-12-25",
				"last_seen": "10:30:45",
				"processed_time": 1703505045.123
			}`),
			want: TimeContainer{
				Timestamp:     mustParseTimeComprehensive(t, time.RFC3339, "2023-12-25T10:30:45Z"),
				StartTime:     time.Unix(1703505045, 0),
				EndTime:       mustParseTimeComprehensive(t, "2006-01-02 15:04:05", "2023-12-25 15:30:45"),
				CreatedAt:     mustParseTimeComprehensive(t, "2006-01-02", "2023-12-25"),
				LastSeen:      mustParseTimeComprehensive(t, "15:04:05", "10:30:45"),
				ProcessedTime: time.Unix(1703505045, 122999906), // Actual precision from float64 conversion
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := model.ParseInto[TimeContainer](tt.input)
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
