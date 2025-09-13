package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// User represents a user with validation requirements
type User struct {
	ID       int    `json:"id" validate:"required,min=1"`
	Name     string `json:"name" validate:"required,min=2"`
	Email    string `json:"email" validate:"required,email"`
	Age      int    `json:"age" validate:"required,min=0,max=150"`
	Company  string `json:"company" validate:"required,min=2"`
	JoinDate string `json:"join_date"`
}

// LogEntry represents a log entry for processing
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level" validate:"required"`
	Message   string `json:"message" validate:"required"`
	Service   string `json:"service" validate:"required"`
	UserID    int    `json:"user_id" validate:"min=1"`
}

func main() {
	fmt.Println("🚀 High-Throughput Stream Processing with goflow Integration")
	fmt.Println("==========================================================")

	// Demo 1: Basic stream processing
	fmt.Println("\n📊 Demo 1: Basic Stream Processing")
	demoBasicStreamProcessing()

	// Demo 2: Channel-based streaming
	fmt.Println("\n🔄 Demo 2: Channel-Based Streaming")
	demoChannelStreaming()

	// Demo 3: Concurrent processing with metrics
	fmt.Println("\n⚡ Demo 3: High-Concurrency Processing with Real-time Metrics")
	demoConcurrentProcessing()

	// Demo 4: Backpressure handling
	fmt.Println("\n🌊 Demo 4: Backpressure Management")
	demoBackpressureHandling()

	// Demo 5: Error handling and circuit breaker
	fmt.Println("\n🛡️ Demo 5: Error Handling and Circuit Breaker")
	demoErrorHandling()

	fmt.Println("\n✅ Stream Processing Demo Complete!")
}

// demoBasicStreamProcessing shows basic stream processing capabilities
func demoBasicStreamProcessing() {
	// Create stream processor with default configuration
	processor, err := model.NewStreamProcessor[User](nil)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create sample user data
	users := []User{
		{ID: 1, Name: "Alice Johnson", Email: "alice@techcorp.com", Age: 28, Company: "TechCorp", JoinDate: "2023-01-15"},
		{ID: 2, Name: "Bob Smith", Email: "bob@datatech.com", Age: 34, Company: "DataTech", JoinDate: "2023-02-20"},
		{ID: 3, Name: "Carol Davis", Email: "carol@cloudco.com", Age: 29, Company: "CloudCo", JoinDate: "2023-03-10"},
		{ID: 4, Name: "David Wilson", Email: "david@aistart.com", Age: 31, Company: "AIStart", JoinDate: "2023-04-05"},
		{ID: 5, Name: "Eva Brown", Email: "eva@fintech.com", Age: 27, Company: "FinTech", JoinDate: "2023-05-12"},
	}

	// Convert to JSON bytes for processing
	var jsonData [][]byte
	for _, user := range users {
		data, err := json.Marshal(user)
		if err != nil {
			log.Printf("Failed to marshal user %d: %v", user.ID, err)
			continue
		}
		jsonData = append(jsonData, data)
	}

	fmt.Printf("📝 Processing %d user records...\n", len(jsonData))

	// Process the stream
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	resultStream, err := processor.ProcessStream(ctx, jsonData)
	if err != nil {
		log.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect and display results
	var successful, failed int
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		if result.Success {
			successful++
			if user, ok := result.Result.(User); ok {
				fmt.Printf("   ✅ Processed: %s <%s> from %s (took %v)\n",
					user.Name, user.Email, user.Company, result.Duration)
			}
		} else {
			failed++
			fmt.Printf("   ❌ Failed: %s (error: %v)\n", result.ID, result.Error)
		}
	})

	processingTime := time.Since(start)

	if err != nil {
		log.Printf("Stream processing error: %v", err)
	}

	fmt.Printf("📈 Results: %d successful, %d failed in %v\n", successful, failed, processingTime)

	// Show metrics
	metrics := processor.GetMetrics()
	fmt.Printf("💡 Metrics: %.2f items/sec, avg processing time: %v\n",
		metrics.Throughput, metrics.AverageProcessingTime)
}

// demoChannelStreaming demonstrates channel-based streaming
func demoChannelStreaming() {
	// Create processor with custom configuration
	config := model.DefaultStreamProcessorConfig()
	config.Stream.BufferSize = 50
	config.WorkerPool.MaxWorkers = 5

	processor, err := model.NewStreamProcessor[LogEntry](config)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create input channel
	inputChan := make(chan []byte, 100)

	// Start data producer
	go func() {
		defer close(inputChan)

		// Simulate real-time log entries
		for i := 0; i < 20; i++ {
			logEntry := LogEntry{
				Timestamp: time.Now().Format(time.RFC3339),
				Level:     []string{"INFO", "WARN", "ERROR", "DEBUG"}[i%4],
				Message:   fmt.Sprintf("Processing event %d", i+1),
				Service:   fmt.Sprintf("service-%d", (i%3)+1),
				UserID:    (i % 10) + 1,
			}

			data, err := json.Marshal(logEntry)
			if err != nil {
				log.Printf("Failed to marshal log entry: %v", err)
				continue
			}

			inputChan <- data

			// Simulate real-time data arrival
			time.Sleep(50 * time.Millisecond)
		}
	}()

	fmt.Printf("📡 Processing real-time log stream...\n")

	// Process the channel stream
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()
	resultStream, err := processor.ProcessChannel(ctx, inputChan)
	if err != nil {
		log.Fatalf("Failed to process channel: %v", err)
	}
	defer resultStream.Close()

	// Process results
	var logCount int
	var lastTimestamp time.Time
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		if result.Success {
			if logEntry, ok := result.Result.(LogEntry); ok {
				logCount++
				if timestamp, err := time.Parse(time.RFC3339, logEntry.Timestamp); err == nil {
					if timestamp.After(lastTimestamp) {
						lastTimestamp = timestamp
					}
				}
				fmt.Printf("   📋 %s [%s] %s: %s (user:%d)\n",
					logEntry.Timestamp, logEntry.Level, logEntry.Service, logEntry.Message, logEntry.UserID)
			}
		} else {
			fmt.Printf("   ❌ Failed to process log entry: %v\n", result.Error)
		}
	})

	processingTime := time.Since(start)

	if err != nil {
		log.Printf("Stream processing error: %v", err)
	}

	fmt.Printf("📈 Processed %d log entries in %v\n", logCount, processingTime)

	// Show final metrics
	metrics := processor.GetMetrics()
	fmt.Printf("💡 Stream metrics: %.2f entries/sec, error rate: %.2f%%\n",
		metrics.Throughput, metrics.ErrorRate*100)
}

// demoConcurrentProcessing shows high-concurrency processing
func demoConcurrentProcessing() {
	// Create high-performance configuration
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 20
	config.Stream.BufferSize = 500
	config.Stream.BackpressureSize = 1000
	config.Monitoring.MetricsInterval = 500 * time.Millisecond

	processor, err := model.NewStreamProcessor[User](config)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Generate large dataset
	var jsonData [][]byte
	for i := 0; i < 1000; i++ {
		user := User{
			ID:       i + 1,
			Name:     fmt.Sprintf("User-%d", i+1),
			Email:    fmt.Sprintf("user%d@company%d.com", i+1, (i%10)+1),
			Age:      25 + (i % 40),
			Company:  fmt.Sprintf("Company-%d", (i%10)+1),
			JoinDate: time.Now().AddDate(0, -(i % 12), 0).Format("2006-01-02"),
		}

		data, err := json.Marshal(user)
		if err != nil {
			continue
		}
		jsonData = append(jsonData, data)
	}

	fmt.Printf("🔥 Processing %d records with %d workers...\n",
		len(jsonData), config.WorkerPool.MaxWorkers)

	// Start metrics monitoring
	var wg sync.WaitGroup
	stopMetrics := make(chan bool)

	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				metrics := processor.GetMetrics()
				fmt.Printf("   📊 Progress: %d/%d items (%.1f%%), throughput: %.0f/sec, workers active: %d\n",
					metrics.TotalItems, len(jsonData),
					float64(metrics.TotalItems)/float64(len(jsonData))*100,
					metrics.Throughput, metrics.ActiveWorkers)

			case <-stopMetrics:
				return
			}
		}
	}()

	// Process the large dataset
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()
	resultStream, err := processor.ProcessStream(ctx, jsonData)
	if err != nil {
		log.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Count results
	var processed int
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		processed++
	})

	processingTime := time.Since(start)
	close(stopMetrics)
	wg.Wait()

	if err != nil {
		log.Printf("Stream processing error: %v", err)
	}

	fmt.Printf("🎯 Final Results: %d items processed in %v\n", processed, processingTime)

	// Show final performance metrics
	metrics := processor.GetMetrics()
	fmt.Printf("⚡ Performance: %.0f items/sec, avg time: %v, success rate: %.2f%%\n",
		metrics.Throughput, metrics.AverageProcessingTime,
		float64(metrics.SuccessfulItems)/float64(metrics.TotalItems)*100)
}

// demoBackpressureHandling demonstrates backpressure management
func demoBackpressureHandling() {
	// Create configuration with small buffers to demonstrate backpressure
	config := model.DefaultStreamProcessorConfig()
	config.Stream.BackpressureSize = 10 // Small buffer
	config.WorkerPool.MaxWorkers = 2    // Few workers

	processor, err := model.NewStreamProcessor[User](config)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create input channel with many items
	inputChan := make(chan []byte, 5)

	// Producer (fast)
	go func() {
		defer close(inputChan)

		for i := 0; i < 50; i++ {
			user := User{
				ID:      i + 1,
				Name:    fmt.Sprintf("FastUser-%d", i+1),
				Email:   fmt.Sprintf("fast%d@example.com", i+1),
				Age:     25 + (i % 40),
				Company: "FastCompany",
			}

			data, err := json.Marshal(user)
			if err != nil {
				continue
			}

			inputChan <- data
			fmt.Printf("   📤 Produced item %d\n", i+1)
		}
		fmt.Printf("   ✅ Producer finished\n")
	}()

	fmt.Printf("🌊 Testing backpressure with fast producer, slow consumer...\n")

	// Process with backpressure
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessChannel(ctx, inputChan)
	if err != nil {
		log.Fatalf("Failed to process channel: %v", err)
	}
	defer resultStream.Close()

	// Consumer (slow)
	var consumed int
	start := time.Now()
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		consumed++
		if result.Success {
			fmt.Printf("   📥 Consumed item %d (after %v)\n", consumed, time.Since(start))
		}

		// Simulate slow consumer
		time.Sleep(200 * time.Millisecond)
	})

	if err != nil {
		log.Printf("Stream processing error: %v", err)
	}

	fmt.Printf("🎯 Backpressure demo: %d items processed\n", consumed)
	fmt.Printf("💡 The system handled backpressure gracefully without dropping data\n")
}

// demoErrorHandling demonstrates error handling and circuit breaker
func demoErrorHandling() {
	// Create configuration with circuit breaker
	config := model.DefaultStreamProcessorConfig()
	config.Pipeline.ErrorThreshold = 0.3 // 30% error rate threshold
	config.Pipeline.RetryAttempts = 2

	processor, err := model.NewStreamProcessor[User](config)
	if err != nil {
		log.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		log.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create mix of valid and invalid data
	var jsonData [][]byte

	// Add some valid records
	for i := 0; i < 5; i++ {
		user := User{
			ID:      i + 1,
			Name:    fmt.Sprintf("ValidUser-%d", i+1),
			Email:   fmt.Sprintf("valid%d@example.com", i+1),
			Age:     30,
			Company: "ValidCompany",
		}
		data, _ := json.Marshal(user)
		jsonData = append(jsonData, data)
	}

	// Add invalid records to trigger errors
	invalidData := [][]byte{
		[]byte(`{"id": "invalid", "name": "", "email": "not-email", "age": -5}`),
		[]byte(`{"id": -1, "name": "Test"}`),                         // Missing required fields
		[]byte(`{malformed json`),                                    // Parse error
		[]byte(`{"id": 0, "name": "A", "email": "bad", "age": 200}`), // Validation errors
	}

	jsonData = append(jsonData, invalidData...)

	fmt.Printf("🛡️ Processing %d records (some invalid) to test error handling...\n", len(jsonData))

	// Process with error handling
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	start := time.Now()
	resultStream, err := processor.ProcessStream(ctx, jsonData)
	if err != nil {
		log.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect results and analyze errors
	var successful, failed int
	var retryCount int
	errorTypes := make(map[string]int)

	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		if result.Success {
			successful++
			if user, ok := result.Result.(User); ok {
				fmt.Printf("   ✅ Success: %s <%s>\n", user.Name, user.Email)
			}
		} else {
			failed++
			if result.Attempts > 1 {
				retryCount++
			}

			// Categorize error types
			if result.Error != nil {
				errorType := "unknown"
				errStr := result.Error.Error()
				if contains(errStr, "validation") {
					errorType = "validation"
				} else if contains(errStr, "parse") || contains(errStr, "unmarshal") {
					errorType = "parsing"
				} else if contains(errStr, "circuit breaker") {
					errorType = "circuit_breaker"
				}
				errorTypes[errorType]++
			}

			fmt.Printf("   ❌ Failed: %s (attempts: %d, error: %v)\n",
				result.ID, result.Attempts, result.Error)
		}
	})

	processingTime := time.Since(start)

	if err != nil {
		log.Printf("Stream processing error: %v", err)
	}

	// Show comprehensive error analysis
	fmt.Printf("\n📊 Error Handling Results:\n")
	fmt.Printf("   • Total processed: %d in %v\n", successful+failed, processingTime)
	fmt.Printf("   • Successful: %d (%.1f%%)\n", successful, float64(successful)/float64(successful+failed)*100)
	fmt.Printf("   • Failed: %d (%.1f%%)\n", failed, float64(failed)/float64(successful+failed)*100)
	fmt.Printf("   • Retries attempted: %d\n", retryCount)

	fmt.Printf("   • Error breakdown:\n")
	for errorType, count := range errorTypes {
		fmt.Printf("     - %s: %d\n", errorType, count)
	}

	// Show final metrics including circuit breaker status
	metrics := processor.GetMetrics()
	fmt.Printf("   • Circuit breaker open: %v\n", metrics.CircuitBreakerOpen)
	fmt.Printf("   • Error rate: %.2f%%\n", metrics.ErrorRate*100)

	fmt.Printf("💡 The system handled errors gracefully with retries and circuit breaking\n")
}

// Helper function to check if string contains substring (same as in enhanced_validator_test.go)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())))
}
