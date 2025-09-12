package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Event represents an event with various time fields
type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	CreatedAt time.Time `json:"created_at"`
}

// LogEntry represents a log entry with timestamp
type LogEntry struct {
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	UserID    int       `json:"user_id"`
}

func main() {
	fmt.Println("🕐 Time Parsing Examples with gopantic")
	fmt.Println("=====================================")

	// Example 1: RFC3339 format (ISO 8601)
	fmt.Println("\n1. RFC3339 format (ISO 8601)")
	rfc3339JSON := []byte(`{
		"id": 1,
		"name": "Tech Conference",
		"start_time": "2023-12-25T10:30:00Z",
		"end_time": "2023-12-25T17:30:00Z",
		"created_at": "2023-11-01T09:15:30Z"
	}`)

	event1, err := model.ParseInto[Event](rfc3339JSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event: %s\n", event1.Name)
	fmt.Printf("Start: %s\n", event1.StartTime.Format(time.RFC3339))
	fmt.Printf("End:   %s\n", event1.EndTime.Format(time.RFC3339))
	fmt.Printf("Duration: %v\n", event1.EndTime.Sub(event1.StartTime))

	// Example 2: Unix timestamps
	fmt.Println("\n2. Unix timestamps")
	unixJSON := []byte(`{
		"id": 2,
		"name": "Workshop",
		"start_time": 1703505000,
		"end_time": 1703523000,
		"created_at": 1701417600
	}`)

	event2, err := model.ParseInto[Event](unixJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event: %s\n", event2.Name)
	fmt.Printf("Start: %s (Unix: %d)\n", event2.StartTime.Format(time.RFC3339), event2.StartTime.Unix())
	fmt.Printf("End:   %s (Unix: %d)\n", event2.EndTime.Format(time.RFC3339), event2.EndTime.Unix())

	// Example 3: Mixed formats with fractional seconds
	fmt.Println("\n3. Mixed formats with nanosecond precision")
	mixedJSON := []byte(`{
		"id": 3,
		"name": "Webinar",
		"start_time": "2023-12-25T14:30:00.123456789Z",
		"end_time": 1703523000.5,
		"created_at": "2023-12-01T10:00:00Z"
	}`)

	event3, err := model.ParseInto[Event](mixedJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event: %s\n", event3.Name)
	fmt.Printf("Start: %s (nanoseconds: %d)\n", event3.StartTime.Format(time.RFC3339Nano), event3.StartTime.Nanosecond())
	fmt.Printf("End:   %s (from float timestamp)\n", event3.EndTime.Format(time.RFC3339Nano))

	// Example 4: Date-only format
	fmt.Println("\n4. Date-only format (all-day events)")
	dateOnlyJSON := []byte(`{
		"id": 4,
		"name": "Holiday",
		"start_time": "2023-12-25",
		"end_time": "2023-12-26",
		"created_at": "2023-12-01"
	}`)

	event4, err := model.ParseInto[Event](dateOnlyJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event: %s\n", event4.Name)
	fmt.Printf("Start: %s (date only)\n", event4.StartTime.Format("2006-01-02"))
	fmt.Printf("End:   %s (date only)\n", event4.EndTime.Format("2006-01-02"))

	// Example 5: Various datetime formats
	fmt.Println("\n5. Various datetime formats")
	formats := []struct {
		name string
		json string
	}{
		{"ISO without timezone", `{"timestamp": "2023-12-25T10:30:00", "level": "INFO", "message": "Server started", "user_id": 1}`},
		{"Common format", `{"timestamp": "2023-12-25 10:30:00", "level": "WARN", "message": "High memory usage", "user_id": 2}`},
		{"Time only", `{"timestamp": "15:04:05", "level": "ERROR", "message": "Connection failed", "user_id": 3}`},
	}

	for _, format := range formats {
		fmt.Printf("\n%s:\n", format.name)
		entry, err := model.ParseInto[LogEntry]([]byte(format.json))
		if err != nil {
			log.Printf("Error parsing %s: %v", format.name, err)
			continue
		}
		fmt.Printf("  [%s] %s: %s (User: %d)\n",
			entry.Level,
			entry.Timestamp.Format("2006-01-02 15:04:05"),
			entry.Message,
			entry.UserID)
	}

	// Example 6: Handling null/missing values
	fmt.Println("\n6. Handling null/missing time values")
	nullJSON := []byte(`{
		"id": 6,
		"name": "Draft Event",
		"start_time": null,
		"created_at": "2023-12-01T09:00:00Z"
	}`)

	event6, err := model.ParseInto[Event](nullJSON)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Event: %s\n", event6.Name)
	fmt.Printf("Start: %v (zero time: %t)\n", event6.StartTime, event6.StartTime.IsZero())
	fmt.Printf("End:   %v (zero time: %t)\n", event6.EndTime, event6.EndTime.IsZero())
	fmt.Printf("Created: %s\n", event6.CreatedAt.Format(time.RFC3339))

	// Example 7: Error handling for invalid formats
	fmt.Println("\n7. Error handling for invalid time formats")
	invalidTimeFormats := []string{
		`{"timestamp": "not-a-date"}`,
		`{"timestamp": "2023-13-45"}`, // Invalid date
		`{"timestamp": "25:99:99"}`,   // Invalid time
		`{"timestamp": "invalid-unix-timestamp"}`,
	}

	for i, invalidJSON := range invalidTimeFormats {
		fmt.Printf("Invalid format %d: ", i+1)
		_, err := model.ParseInto[LogEntry]([]byte(invalidJSON))
		if err != nil {
			fmt.Printf("✓ Correctly rejected: %v\n", err)
		} else {
			fmt.Printf("✗ Unexpectedly accepted\n")
		}
	}

	fmt.Println("\n8. Supported time formats summary:")
	supportedFormats := []string{
		"RFC3339: 2006-01-02T15:04:05Z07:00",
		"RFC3339Nano: 2006-01-02T15:04:05.999999999Z07:00",
		"ISO 8601 UTC: 2006-01-02T15:04:05Z",
		"ISO 8601 (no timezone): 2006-01-02T15:04:05",
		"Common format: 2006-01-02 15:04:05",
		"Date only: 2006-01-02",
		"Time only: 15:04:05",
		"Unix timestamps: 1703505000 (int) or 1703505000.5 (float)",
	}

	for _, format := range supportedFormats {
		fmt.Printf("  • %s\n", format)
	}

	fmt.Println("\n✨ Time parsing with gopantic makes handling various timestamp formats effortless!")
}
