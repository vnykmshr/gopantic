package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/vnykmshr/goflow/pkg/scheduling/workerpool"
)

// ValidationPipelineConfig configures the validation pipeline behavior
type ValidationPipelineConfig struct {
	// Worker pool configuration
	WorkerPool struct {
		MinWorkers    int           // Minimum number of workers
		MaxWorkers    int           // Maximum number of workers
		IdleTimeout   time.Duration // How long workers stay idle before termination
		QueueSize     int           // Size of the work queue
		ScalingFactor float64       // Factor for worker scaling (0-1)
	}

	// Batch processing configuration
	BatchProcessing struct {
		Enabled       bool          // Enable batch processing
		BatchSize     int           // Number of items per batch
		BatchTimeout  time.Duration // Maximum time to wait for batch completion
		FlushInterval time.Duration // Interval to flush incomplete batches
	}

	// Pipeline configuration
	Pipeline struct {
		MaxConcurrentBatches int           // Maximum concurrent batches
		TimeoutPerItem       time.Duration // Timeout per individual validation
		RetryAttempts        int           // Number of retry attempts for failed validations
		BackoffMultiplier    float64       // Backoff multiplier for retries
	}

	// Monitoring and metrics
	Monitoring struct {
		EnableMetrics   bool          // Enable metrics collection
		MetricsInterval time.Duration // Interval for metrics collection
		EnableProfiling bool          // Enable performance profiling
		LogSlowRequests time.Duration // Log requests slower than this threshold
	}
}

// DefaultValidationPipelineConfig returns a default pipeline configuration
func DefaultValidationPipelineConfig() *ValidationPipelineConfig {
	config := &ValidationPipelineConfig{}

	// Worker pool defaults
	config.WorkerPool.MinWorkers = 2
	config.WorkerPool.MaxWorkers = 10
	config.WorkerPool.IdleTimeout = 30 * time.Second
	config.WorkerPool.QueueSize = 100
	config.WorkerPool.ScalingFactor = 0.8

	// Batch processing defaults
	config.BatchProcessing.Enabled = true
	config.BatchProcessing.BatchSize = 10
	config.BatchProcessing.BatchTimeout = 5 * time.Second
	config.BatchProcessing.FlushInterval = 1 * time.Second

	// Pipeline defaults
	config.Pipeline.MaxConcurrentBatches = 5
	config.Pipeline.TimeoutPerItem = 10 * time.Second
	config.Pipeline.RetryAttempts = 3
	config.Pipeline.BackoffMultiplier = 2.0

	// Monitoring defaults
	config.Monitoring.EnableMetrics = true
	config.Monitoring.MetricsInterval = 10 * time.Second
	config.Monitoring.EnableProfiling = false
	config.Monitoring.LogSlowRequests = 1 * time.Second

	return config
}

// ValidationItem represents a single item to be validated
type ValidationItem struct {
	ID       string                 // Unique identifier for the item
	Data     []byte                 // Raw data to validate
	Target   reflect.Type           // Target type for validation
	Context  context.Context        // Context for the validation
	Metadata map[string]interface{} // Additional metadata
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	ID        string        // Matches ValidationItem.ID
	Success   bool          // Whether validation succeeded
	Result    interface{}   // Parsed result (if successful)
	Error     error         // Error (if failed)
	Duration  time.Duration // Time taken for validation
	Attempts  int           // Number of attempts made
	Timestamp time.Time     // When validation completed
}

// ValidationPipeline provides parallel validation processing with worker pools
type ValidationPipeline struct {
	config  *ValidationPipelineConfig
	workers workerpool.Pool

	// Channels for pipeline communication
	inputChan  chan *ValidationItem
	outputChan chan *ValidationResult

	// Pipeline state
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex

	// Metrics and monitoring
	metrics   *PipelineMetrics
	startTime time.Time
}

// PipelineMetrics tracks pipeline performance and statistics
type PipelineMetrics struct {
	mu sync.RWMutex

	// Counters
	TotalProcessed  int64
	TotalSuccessful int64
	TotalFailed     int64
	TotalRetries    int64

	// Timing
	AverageProcessingTime time.Duration
	TotalProcessingTime   time.Duration

	// Worker pool stats
	ActiveWorkers    int
	QueuedItems      int
	ProcessedBatches int64

	// Error tracking
	ErrorsByType map[string]int64
}

// NewValidationPipeline creates a new validation pipeline with worker pools
func NewValidationPipeline(config *ValidationPipelineConfig) (*ValidationPipeline, error) {
	if config == nil {
		config = DefaultValidationPipelineConfig()
	}

	// Create context for pipeline lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker pool
	workers := workerpool.New(config.WorkerPool.MaxWorkers, config.WorkerPool.QueueSize)

	pipeline := &ValidationPipeline{
		config:     config,
		workers:    workers,
		inputChan:  make(chan *ValidationItem, config.WorkerPool.QueueSize),
		outputChan: make(chan *ValidationResult, config.WorkerPool.QueueSize),
		ctx:        ctx,
		cancel:     cancel,
		metrics: &PipelineMetrics{
			ErrorsByType: make(map[string]int64),
		},
		startTime: time.Now(),
	}

	return pipeline, nil
}

// Start begins the validation pipeline processing
func (vp *ValidationPipeline) Start() error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if vp.running {
		return fmt.Errorf("pipeline is already running")
	}

	// Start pipeline workers
	vp.wg.Add(3)
	go vp.inputProcessor()
	go vp.batchProcessor()
	go vp.metricsCollector()

	// Start processing results from worker pool
	go vp.resultProcessor()

	vp.running = true
	vp.startTime = time.Now()

	return nil
}

// Stop gracefully shuts down the validation pipeline
func (vp *ValidationPipeline) Stop() error {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if !vp.running {
		return nil
	}

	// Signal shutdown
	vp.cancel()

	// Close input channel to stop accepting new work
	close(vp.inputChan)

	// Wait for all workers to complete
	vp.wg.Wait()

	// Shutdown worker pool
	<-vp.workers.Shutdown()

	close(vp.outputChan)
	vp.running = false

	return nil
}

// Submit adds a validation item to the pipeline for processing
func (vp *ValidationPipeline) Submit(item *ValidationItem) error {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	if !vp.running {
		return fmt.Errorf("pipeline is not running")
	}

	if item.Context == nil {
		item.Context = context.Background()
	}

	select {
	case vp.inputChan <- item:
		return nil
	case <-vp.ctx.Done():
		return fmt.Errorf("pipeline is shutting down")
	default:
		return fmt.Errorf("pipeline queue is full")
	}
}

// Results returns a channel for receiving validation results
func (vp *ValidationPipeline) Results() <-chan *ValidationResult {
	return vp.outputChan
}

// GetMetrics returns current pipeline metrics
func (vp *ValidationPipeline) GetMetrics() *PipelineMetrics {
	vp.metrics.mu.RLock()
	defer vp.metrics.mu.RUnlock()

	// Create a copy of metrics to avoid race conditions
	metrics := &PipelineMetrics{
		TotalProcessed:        vp.metrics.TotalProcessed,
		TotalSuccessful:       vp.metrics.TotalSuccessful,
		TotalFailed:           vp.metrics.TotalFailed,
		TotalRetries:          vp.metrics.TotalRetries,
		AverageProcessingTime: vp.metrics.AverageProcessingTime,
		TotalProcessingTime:   vp.metrics.TotalProcessingTime,
		ActiveWorkers:         vp.metrics.ActiveWorkers,
		QueuedItems:           vp.metrics.QueuedItems,
		ProcessedBatches:      vp.metrics.ProcessedBatches,
		ErrorsByType:          make(map[string]int64),
	}

	for k, v := range vp.metrics.ErrorsByType {
		metrics.ErrorsByType[k] = v
	}

	return metrics
}

// inputProcessor handles incoming validation items
func (vp *ValidationPipeline) inputProcessor() {
	defer vp.wg.Done()

	for {
		select {
		case item, ok := <-vp.inputChan:
			if !ok {
				return // Channel closed, shutdown
			}

			// Queue the item for batch processing
			vp.queueForBatchProcessing(item)

		case <-vp.ctx.Done():
			return
		}
	}
}

// queueForBatchProcessing adds an item to the batch processing queue
func (vp *ValidationPipeline) queueForBatchProcessing(item *ValidationItem) {
	// Create a task for the worker pool
	task := workerpool.TaskFunc(func(ctx context.Context) error {
		start := time.Now()

		// Perform the actual validation
		result, err := vp.performValidation(item)

		duration := time.Since(start)

		// Update metrics
		vp.updateMetrics(err == nil, duration, err)

		// Create validation result
		validationResult := &ValidationResult{
			ID:        item.ID,
			Success:   err == nil,
			Result:    result,
			Error:     err,
			Duration:  duration,
			Attempts:  1, // Will be updated by retry logic
			Timestamp: time.Now(),
		}

		// Send result to output channel
		select {
		case vp.outputChan <- validationResult:
		case <-vp.ctx.Done():
		}

		return err
	})

	// Submit to worker pool
	_ = vp.workers.Submit(task) // Error handling is done within the task
}

// resultProcessor processes results from the worker pool
func (vp *ValidationPipeline) resultProcessor() {
	// Process results from worker pool if needed
	// For now, results are sent directly from tasks to output channel
	// This method can be used for additional result processing
	for {
		select {
		case result := <-vp.workers.Results():
			// Log or process worker pool results if needed
			if result.Error != nil && vp.config.Monitoring.LogSlowRequests > 0 {
				fmt.Printf("Worker pool task error: %v\n", result.Error)
			}
		case <-vp.ctx.Done():
			return
		}
	}
}

// performValidation performs the actual validation of an item
func (vp *ValidationPipeline) performValidation(item *ValidationItem) (interface{}, error) {
	// Create a new instance of the target type
	targetValue := reflect.New(item.Target)
	targetPtr := targetValue.Interface()

	// Use reflection to call ParseInto with the correct type
	// This is a simplified implementation - in practice you'd use the generic ParseInto
	if err := parseIntoReflection(item.Data, targetPtr); err != nil {
		return nil, fmt.Errorf("validation failed for item %s: %w", item.ID, err)
	}

	// Return the parsed value
	return targetValue.Elem().Interface(), nil
}

// parseIntoReflection performs parsing using reflection (simplified implementation)
func parseIntoReflection(data []byte, target interface{}) error {
	// Detect format and parse
	format := DetectFormat(data)

	switch format {
	case FormatJSON:
		return parseJSONIntoReflection(data, target)
	case FormatYAML:
		return parseYAMLIntoReflection(data, target)
	default:
		return fmt.Errorf("unsupported format: %v", format)
	}
}

// batchProcessor handles batch processing of validation items
func (vp *ValidationPipeline) batchProcessor() {
	defer vp.wg.Done()

	if !vp.config.BatchProcessing.Enabled {
		return
	}

	ticker := time.NewTicker(vp.config.BatchProcessing.FlushInterval)
	defer ticker.Stop()

	var batch []*ValidationItem
	var batchTimer *time.Timer

	for {
		select {
		case <-ticker.C:
			if len(batch) > 0 {
				vp.processBatch(batch)
				batch = nil
				if batchTimer != nil {
					batchTimer.Stop()
					batchTimer = nil
				}
			}

		case <-vp.ctx.Done():
			// Process any remaining items in the batch
			if len(batch) > 0 {
				vp.processBatch(batch)
			}
			return
		}
	}
}

// processBatch processes a batch of validation items
func (vp *ValidationPipeline) processBatch(batch []*ValidationItem) {
	vp.metrics.mu.Lock()
	vp.metrics.ProcessedBatches++
	vp.metrics.mu.Unlock()

	// Process batch items concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, vp.config.Pipeline.MaxConcurrentBatches)

	for _, item := range batch {
		wg.Add(1)
		go func(item *ValidationItem) {
			defer wg.Done()

			semaphore <- struct{}{}        // Acquire
			defer func() { <-semaphore }() // Release

			vp.queueForBatchProcessing(item)
		}(item)
	}

	wg.Wait()
}

// metricsCollector periodically collects and updates metrics
func (vp *ValidationPipeline) metricsCollector() {
	defer vp.wg.Done()

	if !vp.config.Monitoring.EnableMetrics {
		return
	}

	ticker := time.NewTicker(vp.config.Monitoring.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			vp.collectMetrics()

		case <-vp.ctx.Done():
			return
		}
	}
}

// collectMetrics collects current pipeline metrics
func (vp *ValidationPipeline) collectMetrics() {
	vp.metrics.mu.Lock()
	defer vp.metrics.mu.Unlock()

	// Update queue metrics
	vp.metrics.QueuedItems = len(vp.inputChan)

	// Calculate average processing time
	if vp.metrics.TotalProcessed > 0 {
		vp.metrics.AverageProcessingTime = time.Duration(
			int64(vp.metrics.TotalProcessingTime) / vp.metrics.TotalProcessed,
		)
	}
}

// updateMetrics updates pipeline metrics after processing an item
func (vp *ValidationPipeline) updateMetrics(success bool, duration time.Duration, err error) {
	vp.metrics.mu.Lock()
	defer vp.metrics.mu.Unlock()

	vp.metrics.TotalProcessed++
	vp.metrics.TotalProcessingTime += duration

	if success {
		vp.metrics.TotalSuccessful++
	} else {
		vp.metrics.TotalFailed++

		// Track error types
		if err != nil {
			errorType := reflect.TypeOf(err).String()
			vp.metrics.ErrorsByType[errorType]++
		}
	}

	// Log slow requests if monitoring is enabled
	if vp.config.Monitoring.LogSlowRequests > 0 && duration > vp.config.Monitoring.LogSlowRequests {
		// In practice, you'd use a proper logger here
		fmt.Printf("Slow validation request: duration=%v, success=%v\n", duration, success)
	}
}

// IsRunning returns whether the pipeline is currently running
func (vp *ValidationPipeline) IsRunning() bool {
	vp.mu.RLock()
	defer vp.mu.RUnlock()
	return vp.running
}

// GetUptime returns how long the pipeline has been running
func (vp *ValidationPipeline) GetUptime() time.Duration {
	return time.Since(vp.startTime)
}

// GetQueueLength returns the current length of the input queue
func (vp *ValidationPipeline) GetQueueLength() int {
	return len(vp.inputChan)
}

// SubmitBatch submits multiple validation items as a batch
func (vp *ValidationPipeline) SubmitBatch(items []*ValidationItem) error {
	for _, item := range items {
		if err := vp.Submit(item); err != nil {
			return fmt.Errorf("failed to submit item %s: %w", item.ID, err)
		}
	}
	return nil
}

// WaitForCompletion waits for all submitted items to be processed
func (vp *ValidationPipeline) WaitForCompletion(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(vp.ctx, timeout)
	defer cancel()

	// Wait until the input queue is empty and all workers are idle
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pipeline completion")
		default:
			if vp.GetQueueLength() == 0 && vp.areAllWorkersIdle() {
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// areAllWorkersIdle checks if all workers are currently idle
func (vp *ValidationPipeline) areAllWorkersIdle() bool {
	// Check if input queue is empty (simple heuristic)
	return len(vp.inputChan) == 0
}

// parseJSONIntoReflection parses JSON data into a target using reflection
func parseJSONIntoReflection(data []byte, target interface{}) error {
	// Use format detection and existing parser
	format := DetectFormat(data)
	parser := GetParser(format)

	// Parse into generic structure
	parsedData, err := parser.Parse(data)
	if err != nil {
		return err
	}

	// Use the same parsing logic as ParseIntoWithFormat but adapted for reflection
	return parseIntoTarget(parsedData, target, format)
}

// parseYAMLIntoReflection parses YAML data into a target using reflection
func parseYAMLIntoReflection(data []byte, target interface{}) error {
	// Same logic as JSON, format detection will handle the difference
	return parseJSONIntoReflection(data, target)
}

// parseIntoTarget fills a target struct from parsed data using the same logic as ParseIntoWithFormat
func parseIntoTarget(data map[string]interface{}, target interface{}, format Format) error {
	var errors ErrorList

	// Get target info
	targetValue := reflect.ValueOf(target).Elem()
	targetType := targetValue.Type()

	// Parse validation rules for this struct type
	validation := ParseValidationTags(targetType)

	// Process each field in the struct (parsing and coercion pass)
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field key from appropriate tag (json or yaml), fallback to field name
		fieldKey := getFieldKey(field, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Get value from data map
		rawValue, exists := data[fieldKey]
		if !exists {
			continue // Field not present in data
		}

		// Set field value with type coercion
		if err := setFieldValue(fieldValue, rawValue, fieldKey, format); err != nil {
			errors.Add(NewParseError(fieldKey, rawValue, fieldValue.Type().String(), err.Error()))
			continue
		}
	}

	// Return parsing errors if any occurred
	if errors.HasErrors() {
		return errors.AsError()
	}

	// Validation pass - validate all fields with their rules
	for i := 0; i < targetType.NumField(); i++ {
		field := targetType.Field(i)
		fieldValue := targetValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field key from appropriate tag (json or yaml), fallback to field name
		fieldKey := getFieldKey(field, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Apply validation rules (including cross-field validators)
		if err := validateFieldValueWithStruct(field.Name, fieldKey, fieldValue.Interface(), validation, targetValue); err != nil {
			errors.Add(err)
		}
	}

	// Return validation errors if any occurred
	if errors.HasErrors() {
		return errors.AsError()
	}

	return nil
}
