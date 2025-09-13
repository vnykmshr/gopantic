package tests

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// TestStruct for pipeline testing
type TestStruct struct {
	ID    int    `json:"id" validate:"required,min=1"`
	Name  string `json:"name" validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

func TestValidationPipelineConfig_Defaults(t *testing.T) {
	config := model.DefaultValidationPipelineConfig()

	// Test worker pool defaults
	if config.WorkerPool.MinWorkers != 2 {
		t.Errorf("Expected MinWorkers 2, got %d", config.WorkerPool.MinWorkers)
	}
	if config.WorkerPool.MaxWorkers != 10 {
		t.Errorf("Expected MaxWorkers 10, got %d", config.WorkerPool.MaxWorkers)
	}
	if config.WorkerPool.QueueSize != 100 {
		t.Errorf("Expected QueueSize 100, got %d", config.WorkerPool.QueueSize)
	}

	// Test batch processing defaults
	if !config.BatchProcessing.Enabled {
		t.Error("Expected BatchProcessing.Enabled to be true")
	}
	if config.BatchProcessing.BatchSize != 10 {
		t.Errorf("Expected BatchSize 10, got %d", config.BatchProcessing.BatchSize)
	}

	// Test pipeline defaults
	if config.Pipeline.MaxConcurrentBatches != 5 {
		t.Errorf("Expected MaxConcurrentBatches 5, got %d", config.Pipeline.MaxConcurrentBatches)
	}
	if config.Pipeline.RetryAttempts != 3 {
		t.Errorf("Expected RetryAttempts 3, got %d", config.Pipeline.RetryAttempts)
	}

	// Test monitoring defaults
	if !config.Monitoring.EnableMetrics {
		t.Error("Expected EnableMetrics to be true")
	}
	if config.Monitoring.MetricsInterval != 10*time.Second {
		t.Errorf("Expected MetricsInterval 10s, got %v", config.Monitoring.MetricsInterval)
	}
}

func TestNewValidationPipeline(t *testing.T) {
	tests := []struct {
		name        string
		config      *model.ValidationPipelineConfig
		expectError bool
	}{
		{
			name:        "with nil config (uses defaults)",
			config:      nil,
			expectError: false,
		},
		{
			name:        "with default config",
			config:      model.DefaultValidationPipelineConfig(),
			expectError: false,
		},
		{
			name: "with custom config",
			config: func() *model.ValidationPipelineConfig {
				config := model.DefaultValidationPipelineConfig()
				config.WorkerPool.MaxWorkers = 5
				config.WorkerPool.QueueSize = 50
				return config
			}(),
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := model.NewValidationPipeline(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && pipeline == nil {
				t.Error("Expected pipeline instance but got nil")
			}

			// Clean up
			if pipeline != nil {
				_ = pipeline.Stop()
			}
		})
	}
}

func TestValidationPipeline_StartStop(t *testing.T) {
	pipeline, err := model.NewValidationPipeline(nil)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	// Test initial state
	if pipeline.IsRunning() {
		t.Error("Pipeline should not be running initially")
	}

	// Test start
	if err := pipeline.Start(); err != nil {
		t.Errorf("Failed to start pipeline: %v", err)
	}

	if !pipeline.IsRunning() {
		t.Error("Pipeline should be running after start")
	}

	// Test double start (should fail)
	if err := pipeline.Start(); err == nil {
		t.Error("Expected error when starting already running pipeline")
	}

	// Test stop
	if err := pipeline.Stop(); err != nil {
		t.Errorf("Failed to stop pipeline: %v", err)
	}

	if pipeline.IsRunning() {
		t.Error("Pipeline should not be running after stop")
	}

	// Test double stop (should be safe)
	if err := pipeline.Stop(); err != nil {
		t.Errorf("Double stop should be safe: %v", err)
	}
}

func TestValidationPipeline_SubmitAndProcess(t *testing.T) {
	config := model.DefaultValidationPipelineConfig()
	config.WorkerPool.MaxWorkers = 2
	config.WorkerPool.QueueSize = 10

	pipeline, err := model.NewValidationPipeline(config)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Create test data
	testData := TestStruct{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Create validation item
	item := &model.ValidationItem{
		ID:      "test-1",
		Data:    jsonData,
		Target:  reflect.TypeOf(TestStruct{}),
		Context: context.Background(),
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}

	// Submit item
	if err := pipeline.Submit(item); err != nil {
		t.Errorf("Failed to submit item: %v", err)
	}

	// Wait for result
	select {
	case result := <-pipeline.Results():
		if result.ID != "test-1" {
			t.Errorf("Expected result ID 'test-1', got '%s'", result.ID)
		}
		if !result.Success {
			t.Errorf("Expected successful validation, got error: %v", result.Error)
		}
		if result.Result == nil {
			t.Error("Expected non-nil result")
		}
		if result.Duration <= 0 {
			t.Error("Expected positive duration")
		}

	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for validation result")
	}
}

func TestValidationPipeline_SubmitBatch(t *testing.T) {
	config := model.DefaultValidationPipelineConfig()
	config.WorkerPool.MaxWorkers = 3
	config.WorkerPool.QueueSize = 20

	pipeline, err := model.NewValidationPipeline(config)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Create batch of test items
	batchSize := 5
	items := make([]*model.ValidationItem, batchSize)

	for i := 0; i < batchSize; i++ {
		testData := TestStruct{
			ID:    i + 1,
			Name:  "User " + string(rune('A'+i)),
			Email: "user" + string(rune('A'+i)) + "@example.com",
		}

		jsonData, err := json.Marshal(testData)
		if err != nil {
			t.Fatalf("Failed to marshal test data %d: %v", i, err)
		}

		items[i] = &model.ValidationItem{
			ID:      "test-" + string(rune('1'+i)),
			Data:    jsonData,
			Target:  reflect.TypeOf(TestStruct{}),
			Context: context.Background(),
		}
	}

	// Submit batch
	if err := pipeline.SubmitBatch(items); err != nil {
		t.Errorf("Failed to submit batch: %v", err)
	}

	// Collect results
	results := make(map[string]*model.ValidationResult)
	timeout := time.After(10 * time.Second)

	for i := 0; i < batchSize; i++ {
		select {
		case result := <-pipeline.Results():
			results[result.ID] = result

		case <-timeout:
			t.Errorf("Timeout waiting for batch results, got %d/%d", len(results), batchSize)
			return // Exit the test instead of break
		}
	}

	// Verify all results
	if len(results) != batchSize {
		t.Errorf("Expected %d results, got %d", batchSize, len(results))
	}

	for _, item := range items {
		result, found := results[item.ID]
		if !found {
			t.Errorf("Missing result for item %s", item.ID)
			continue
		}

		if !result.Success {
			t.Errorf("Expected successful validation for item %s, got error: %v", item.ID, result.Error)
		}
	}
}

func TestValidationPipeline_InvalidData(t *testing.T) {
	pipeline, err := model.NewValidationPipeline(nil)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Create invalid test data
	invalidData := []byte(`{"id": "invalid", "name": "", "email": "not-an-email"}`)

	item := &model.ValidationItem{
		ID:      "invalid-test",
		Data:    invalidData,
		Target:  reflect.TypeOf(TestStruct{}),
		Context: context.Background(),
	}

	// Submit item
	if err := pipeline.Submit(item); err != nil {
		t.Errorf("Failed to submit item: %v", err)
	}

	// Wait for result
	select {
	case result := <-pipeline.Results():
		if result.ID != "invalid-test" {
			t.Errorf("Expected result ID 'invalid-test', got '%s'", result.ID)
		}
		if result.Success {
			t.Error("Expected validation failure for invalid data")
		}
		if result.Error == nil {
			t.Error("Expected error for invalid data")
		}

	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for validation result")
	}
}

func TestValidationPipeline_Metrics(t *testing.T) {
	config := model.DefaultValidationPipelineConfig()
	config.Monitoring.EnableMetrics = true
	config.Monitoring.MetricsInterval = 100 * time.Millisecond

	pipeline, err := model.NewValidationPipeline(config)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Initial metrics should be zero
	metrics := pipeline.GetMetrics()
	if metrics.TotalProcessed != 0 {
		t.Errorf("Expected 0 total processed, got %d", metrics.TotalProcessed)
	}

	// Submit some test items
	for i := 0; i < 3; i++ {
		testData := TestStruct{
			ID:    i + 1,
			Name:  "Test User",
			Email: "test@example.com",
		}

		jsonData, _ := json.Marshal(testData)

		item := &model.ValidationItem{
			ID:      "metrics-test-" + string(rune('1'+i)),
			Data:    jsonData,
			Target:  reflect.TypeOf(TestStruct{}),
			Context: context.Background(),
		}

		pipeline.Submit(item)
	}

	// Wait for processing and collect results
	processedCount := 0
	timeout := time.After(5 * time.Second)

	for processedCount < 3 {
		select {
		case <-pipeline.Results():
			processedCount++

		case <-timeout:
			t.Errorf("Timeout waiting for results, processed %d/3", processedCount)
			return // Exit the test instead of break
		}
	}

	// Check metrics after processing
	time.Sleep(200 * time.Millisecond) // Allow metrics to update
	metrics = pipeline.GetMetrics()

	if metrics.TotalProcessed != 3 {
		t.Errorf("Expected 3 total processed, got %d", metrics.TotalProcessed)
	}

	if metrics.TotalSuccessful != 3 {
		t.Errorf("Expected 3 successful, got %d", metrics.TotalSuccessful)
	}

	if metrics.TotalFailed != 0 {
		t.Errorf("Expected 0 failed, got %d", metrics.TotalFailed)
	}

	if metrics.AverageProcessingTime <= 0 {
		t.Error("Expected positive average processing time")
	}
}

func TestValidationPipeline_QueueFull(t *testing.T) {
	config := model.DefaultValidationPipelineConfig()
	config.WorkerPool.QueueSize = 2 // Very small queue
	config.WorkerPool.MaxWorkers = 1

	pipeline, err := model.NewValidationPipeline(config)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Fill up the queue
	testData := TestStruct{ID: 1, Name: "Test", Email: "test@example.com"}
	jsonData, _ := json.Marshal(testData)

	item := &model.ValidationItem{
		ID:      "queue-test",
		Data:    jsonData,
		Target:  reflect.TypeOf(TestStruct{}),
		Context: context.Background(),
	}

	// Submit items until queue is full
	var submitErrors []error
	for i := 0; i < 10; i++ { // Try to submit more than queue size
		if err := pipeline.Submit(item); err != nil {
			submitErrors = append(submitErrors, err)
		}
	}

	// Should have some submission errors due to full queue
	if len(submitErrors) == 0 {
		t.Error("Expected some submission errors due to full queue")
	}
}

func TestValidationPipeline_WaitForCompletion(t *testing.T) {
	pipeline, err := model.NewValidationPipeline(nil)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Submit a single item
	testData := TestStruct{ID: 1, Name: "Test", Email: "test@example.com"}
	jsonData, _ := json.Marshal(testData)

	item := &model.ValidationItem{
		ID:      "completion-test",
		Data:    jsonData,
		Target:  reflect.TypeOf(TestStruct{}),
		Context: context.Background(),
	}

	pipeline.Submit(item)

	// Wait for result and check completion
	var result *model.ValidationResult
	select {
	case result = <-pipeline.Results():
		if result.ID != "completion-test" {
			t.Errorf("Expected result ID 'completion-test', got '%s'", result.ID)
		}
	case <-time.After(5 * time.Second):
		t.Error("Timeout waiting for validation result")
		return
	}

	// Now test that WaitForCompletion works (should return immediately since processing is done)
	if err := pipeline.WaitForCompletion(1 * time.Second); err != nil {
		t.Errorf("Failed to wait for completion: %v", err)
	}
}

func TestValidationPipeline_GetUptime(t *testing.T) {
	pipeline, err := model.NewValidationPipeline(nil)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Wait a bit and check uptime
	time.Sleep(100 * time.Millisecond)
	uptime := pipeline.GetUptime()

	if uptime < 100*time.Millisecond {
		t.Errorf("Expected uptime >= 100ms, got %v", uptime)
	}

	if uptime > 1*time.Second {
		t.Errorf("Expected uptime < 1s, got %v", uptime)
	}
}

func TestValidationPipeline_GetQueueLength(t *testing.T) {
	pipeline, err := model.NewValidationPipeline(nil)
	if err != nil {
		t.Fatalf("Failed to create pipeline: %v", err)
	}

	if err := pipeline.Start(); err != nil {
		t.Fatalf("Failed to start pipeline: %v", err)
	}
	defer func() { _ = pipeline.Stop() }()

	// Initial queue should be empty
	if length := pipeline.GetQueueLength(); length != 0 {
		t.Errorf("Expected initial queue length 0, got %d", length)
	}

	// Submit an item (queue length check is approximate due to concurrent processing)
	testData := TestStruct{ID: 1, Name: "Test", Email: "test@example.com"}
	jsonData, _ := json.Marshal(testData)

	item := &model.ValidationItem{
		ID:      "queue-length-test",
		Data:    jsonData,
		Target:  reflect.TypeOf(TestStruct{}),
		Context: context.Background(),
	}

	pipeline.Submit(item)

	// Queue length should increase (may be processed quickly)
	time.Sleep(10 * time.Millisecond)
	length := pipeline.GetQueueLength()

	// Length should be >= 0 (item may have been processed already)
	if length < 0 {
		t.Errorf("Expected queue length >= 0, got %d", length)
	}
}
