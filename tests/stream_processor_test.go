package tests

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// TestStruct for stream processing testing
type StreamTestStruct struct {
	ID    int    `json:"id" validate:"required,min=1"`
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
	Value int    `json:"value" validate:"min=0"`
}

func TestStreamProcessorConfig_Defaults(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()

	// Test worker pool defaults
	if config.WorkerPool.MaxWorkers != 10 {
		t.Errorf("Expected MaxWorkers 10, got %d", config.WorkerPool.MaxWorkers)
	}
	if config.WorkerPool.QueueSize != 1000 {
		t.Errorf("Expected QueueSize 1000, got %d", config.WorkerPool.QueueSize)
	}
	if config.WorkerPool.IdleTimeout != 30*time.Second {
		t.Errorf("Expected IdleTimeout 30s, got %v", config.WorkerPool.IdleTimeout)
	}

	// Test stream defaults
	if config.Stream.BufferSize != 100 {
		t.Errorf("Expected BufferSize 100, got %d", config.Stream.BufferSize)
	}
	if config.Stream.BatchSize != 50 {
		t.Errorf("Expected BatchSize 50, got %d", config.Stream.BatchSize)
	}
	if config.Stream.BackpressureSize != 200 {
		t.Errorf("Expected BackpressureSize 200, got %d", config.Stream.BackpressureSize)
	}

	// Test pipeline defaults
	if config.Pipeline.RetryAttempts != 3 {
		t.Errorf("Expected RetryAttempts 3, got %d", config.Pipeline.RetryAttempts)
	}
	if config.Pipeline.ErrorThreshold != 0.1 {
		t.Errorf("Expected ErrorThreshold 0.1, got %f", config.Pipeline.ErrorThreshold)
	}

	// Test monitoring defaults
	if !config.Monitoring.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
	if config.Monitoring.MetricsInterval != 10*time.Second {
		t.Errorf("Expected MetricsInterval 10s, got %v", config.Monitoring.MetricsInterval)
	}
}

func TestNewStreamProcessor(t *testing.T) {
	tests := []struct {
		name        string
		config      *model.StreamProcessorConfig
		expectError bool
	}{
		{
			name:        "with nil config (uses defaults)",
			config:      nil,
			expectError: false,
		},
		{
			name:        "with default config",
			config:      model.DefaultStreamProcessorConfig(),
			expectError: false,
		},
		{
			name: "with custom config",
			config: func() *model.StreamProcessorConfig {
				config := model.DefaultStreamProcessorConfig()
				config.WorkerPool.MaxWorkers = 5
				config.Stream.BufferSize = 50
				return config
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor, err := model.NewStreamProcessor[StreamTestStruct](tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && processor == nil {
				t.Error("Expected processor instance but got nil")
			}

			// Clean up
			if processor != nil {
				_ = processor.Stop()
			}
		})
	}
}

func TestStreamProcessor_StartStop(t *testing.T) {
	processor, err := model.NewStreamProcessor[StreamTestStruct](nil)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	// Test initial state
	if processor.IsRunning() {
		t.Error("Processor should not be running initially")
	}

	// Test start
	if err := processor.Start(); err != nil {
		t.Errorf("Failed to start processor: %v", err)
	}

	if !processor.IsRunning() {
		t.Error("Processor should be running after start")
	}

	// Test double start (should fail)
	if err := processor.Start(); err == nil {
		t.Error("Expected error when starting already running processor")
	}

	// Test stop
	if err := processor.Stop(); err != nil {
		t.Errorf("Failed to stop processor: %v", err)
	}

	if processor.IsRunning() {
		t.Error("Processor should not be running after stop")
	}

	// Test double stop (should be safe)
	if err := processor.Stop(); err != nil {
		t.Errorf("Double stop should be safe: %v", err)
	}
}

func TestStreamProcessor_ProcessStream(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 3
	config.Stream.BufferSize = 20

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create test data
	testItems := []StreamTestStruct{
		{ID: 1, Name: "Alice", Email: "alice@example.com", Value: 100},
		{ID: 2, Name: "Bob", Email: "bob@example.com", Value: 200},
		{ID: 3, Name: "Charlie", Email: "charlie@example.com", Value: 300},
	}

	// Convert to JSON bytes
	jsonItems := make([][]byte, 0, len(testItems))
	for _, item := range testItems {
		jsonData, err := json.Marshal(item)
		if err != nil {
			t.Fatalf("Failed to marshal test data: %v", err)
		}
		jsonItems = append(jsonItems, jsonData)
	}

	// Process stream
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, jsonItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect results
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Verify results
	if len(results) != len(testItems) {
		t.Errorf("Expected %d results, got %d", len(testItems), len(results))
	}

	// Check that all results are successful
	for i, result := range results {
		if !result.Success {
			t.Errorf("Result %d failed: %v", i, result.Error)
		}
		if result.Result == nil {
			t.Errorf("Result %d has nil result", i)
		}
		if result.Duration <= 0 {
			t.Errorf("Result %d has invalid duration", i)
		}
	}
}

func TestStreamProcessor_ProcessChannel(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.Stream.BufferSize = 10

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create input channel
	inputChan := make(chan []byte, 5)

	// Send test data
	testItems := []StreamTestStruct{
		{ID: 1, Name: "Test1", Email: "test1@example.com", Value: 1},
		{ID: 2, Name: "Test2", Email: "test2@example.com", Value: 2},
	}

	go func() {
		defer close(inputChan)
		for _, item := range testItems {
			jsonData, _ := json.Marshal(item)
			inputChan <- jsonData
		}
	}()

	// Process channel
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessChannel(ctx, inputChan)
	if err != nil {
		t.Fatalf("Failed to process channel: %v", err)
	}
	defer resultStream.Close()

	// Collect results
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Verify results
	if len(results) != len(testItems) {
		t.Errorf("Expected %d results, got %d", len(testItems), len(results))
	}

	for i, result := range results {
		if !result.Success {
			t.Errorf("Result %d failed: %v", i, result.Error)
		}
	}
}

func TestStreamProcessor_InvalidData(t *testing.T) {
	processor, err := model.NewStreamProcessor[StreamTestStruct](nil)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create invalid test data
	invalidItems := [][]byte{
		[]byte(`{"id": "invalid", "name": "", "email": "not-an-email"}`),
		[]byte(`{"id": -1, "name": "Test", "email": "test@example.com"}`),
		[]byte(`{malformed json`),
	}

	// Process stream
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, invalidItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect results
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Verify that invalid items fail
	if len(results) != len(invalidItems) {
		t.Errorf("Expected %d results, got %d", len(invalidItems), len(results))
	}

	successCount := 0
	for _, result := range results {
		if result.Success {
			successCount++
		} else if result.Error == nil {
			t.Error("Expected error for failed result")
		}
	}

	// Most or all should fail due to validation errors
	if successCount > 1 {
		t.Errorf("Expected most results to fail, but %d succeeded", successCount)
	}
}

func TestStreamProcessor_Metrics(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.Monitoring.EnableMetrics = true
	config.Monitoring.MetricsInterval = 100 * time.Millisecond

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Initial metrics should be zero
	metrics := processor.GetMetrics()
	if metrics.TotalItems != 0 {
		t.Errorf("Expected 0 total items, got %d", metrics.TotalItems)
	}

	// Process some test items
	testItems := []StreamTestStruct{
		{ID: 1, Name: "Test1", Email: "test1@example.com", Value: 1},
		{ID: 2, Name: "Test2", Email: "test2@example.com", Value: 2},
		{ID: 3, Name: "Test3", Email: "test3@example.com", Value: 3},
	}

	jsonItems := make([][]byte, 0, len(testItems))
	for _, item := range testItems {
		jsonData, _ := json.Marshal(item)
		jsonItems = append(jsonItems, jsonData)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, jsonItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Consume all results
	var resultCount int
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		resultCount++
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Wait for metrics to update
	time.Sleep(200 * time.Millisecond)
	metrics = processor.GetMetrics()

	// Check metrics after processing
	if metrics.TotalItems != int64(len(testItems)) {
		t.Errorf("Expected %d total items, got %d", len(testItems), metrics.TotalItems)
	}

	if metrics.AverageProcessingTime <= 0 {
		t.Error("Expected positive average processing time")
	}

	if metrics.Throughput <= 0 {
		t.Error("Expected positive throughput")
	}

	// Check uptime
	uptime := processor.GetUptime()
	if uptime <= 0 {
		t.Error("Expected positive uptime")
	}
}

func TestStreamProcessor_BackpressureHandling(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.Stream.BackpressureSize = 3 // Very small buffer to test backpressure
	config.WorkerPool.MaxWorkers = 2

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create many items to test backpressure
	jsonItems := make([][]byte, 0, 20)
	for i := 0; i < 20; i++ {
		item := StreamTestStruct{
			ID:    i + 1,
			Name:  "Test",
			Email: "test@example.com",
			Value: i,
		}
		jsonData, _ := json.Marshal(item)
		jsonItems = append(jsonItems, jsonData)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, jsonItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Consume results with delay to test backpressure
	var resultCount int
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		resultCount++
		// Small delay to simulate slow consumer
		time.Sleep(10 * time.Millisecond)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Should process all items despite backpressure
	if resultCount != len(jsonItems) {
		t.Errorf("Expected %d results, got %d", len(jsonItems), resultCount)
	}
}

func TestStreamProcessor_CircuitBreaker(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.Pipeline.ErrorThreshold = 0.5 // 50% error rate threshold
	config.WorkerPool.MaxWorkers = 1

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create many invalid items to trigger circuit breaker
	invalidItems := make([][]byte, 0, 15)
	for i := 0; i < 15; i++ {
		// Invalid JSON to cause parsing errors
		invalidItems = append(invalidItems, []byte(`{invalid json`))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, invalidItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect results
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Check that we get results (some may be circuit breaker errors)
	if len(results) == 0 {
		t.Error("Expected some results even with circuit breaker")
	}

	// Wait for circuit breaker to potentially open
	time.Sleep(100 * time.Millisecond)
	metrics := processor.GetMetrics()

	// Check error rate
	if metrics.ErrorRate == 0 {
		t.Error("Expected some error rate due to invalid data")
	}
}

func TestStreamProcessor_RetryMechanism(t *testing.T) {
	config := model.DefaultStreamProcessorConfig()
	config.Pipeline.RetryAttempts = 2
	config.Pipeline.RetryBackoff = 50 * time.Millisecond

	processor, err := model.NewStreamProcessor[StreamTestStruct](config)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create mix of valid and invalid items
	items := [][]byte{
		[]byte(`{"id": 1, "name": "Valid", "email": "valid@example.com", "value": 1}`),
		[]byte(`{invalid json`), // This will cause retry attempts
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, items)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Collect results
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
	})

	if err != nil {
		t.Errorf("Failed to collect results: %v", err)
	}

	// Verify results
	if len(results) != len(items) {
		t.Errorf("Expected %d results, got %d", len(items), len(results))
	}

	// Check that failed item shows retry attempts
	for _, result := range results {
		if !result.Success {
			if result.Attempts <= 1 {
				t.Errorf("Expected multiple attempts for failed result, got %d", result.Attempts)
			}
		}
	}
}

func TestStreamProcessor_ContextCancellation(t *testing.T) {
	processor, err := model.NewStreamProcessor[StreamTestStruct](nil)
	if err != nil {
		t.Fatalf("Failed to create processor: %v", err)
	}

	if err := processor.Start(); err != nil {
		t.Fatalf("Failed to start processor: %v", err)
	}
	defer func() { _ = processor.Stop() }()

	// Create test items
	jsonItems := make([][]byte, 0, 10)
	for i := 0; i < 10; i++ {
		item := StreamTestStruct{
			ID:    i + 1,
			Name:  "Test",
			Email: "test@example.com",
			Value: i,
		}
		jsonData, _ := json.Marshal(item)
		jsonItems = append(jsonItems, jsonData)
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	resultStream, err := processor.ProcessStream(ctx, jsonItems)
	if err != nil {
		t.Fatalf("Failed to process stream: %v", err)
	}
	defer resultStream.Close()

	// Try to collect results (should be interrupted by context)
	var results []*model.StreamResult
	err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
		results = append(results, result)
		// Add small delay to ensure context timeout
		time.Sleep(20 * time.Millisecond)
	})

	// Should get context deadline exceeded or partial results
	if err != context.DeadlineExceeded && len(results) == len(jsonItems) {
		t.Error("Expected context cancellation to interrupt processing")
	}
}
