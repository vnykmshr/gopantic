package model

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/vnykmshr/goflow/pkg/scheduling/workerpool"
	"github.com/vnykmshr/goflow/pkg/streaming/stream"
)

// StreamProcessorConfig configures the streaming processor behavior
type StreamProcessorConfig struct {
	// Worker pool configuration
	WorkerPool struct {
		MaxWorkers  int           // Maximum number of workers
		QueueSize   int           // Size of the work queue
		IdleTimeout time.Duration // How long workers stay idle before termination
	}

	// Stream processing configuration
	Stream struct {
		BufferSize       int           // Size of internal stream buffers
		BatchSize        int           // Number of items to process in each batch
		FlushInterval    time.Duration // Maximum time to wait before flushing incomplete batches
		MaxConcurrency   int           // Maximum concurrent stream operations
		BackpressureSize int           // Buffer size for backpressure management
	}

	// Pipeline configuration
	Pipeline struct {
		RetryAttempts  int           // Number of retry attempts for failed processing
		RetryBackoff   time.Duration // Base backoff duration for retries
		TimeoutPerItem time.Duration // Timeout for processing individual items
		ErrorThreshold float64       // Error rate threshold (0.0-1.0) for circuit breaking
	}

	// Monitoring and metrics
	Monitoring struct {
		EnableMetrics     bool          // Enable metrics collection
		MetricsInterval   time.Duration // Interval for metrics collection
		LogSlowOperations time.Duration // Log operations slower than this threshold
	}
}

// DefaultStreamProcessorConfig returns a default configuration for stream processing
func DefaultStreamProcessorConfig() *StreamProcessorConfig {
	config := &StreamProcessorConfig{}

	// Worker pool defaults
	config.WorkerPool.MaxWorkers = 10
	config.WorkerPool.QueueSize = 1000
	config.WorkerPool.IdleTimeout = 30 * time.Second

	// Stream processing defaults
	config.Stream.BufferSize = 100
	config.Stream.BatchSize = 50
	config.Stream.FlushInterval = 1 * time.Second
	config.Stream.MaxConcurrency = 5
	config.Stream.BackpressureSize = 200

	// Pipeline defaults
	config.Pipeline.RetryAttempts = 3
	config.Pipeline.RetryBackoff = 100 * time.Millisecond
	config.Pipeline.TimeoutPerItem = 30 * time.Second
	config.Pipeline.ErrorThreshold = 0.1 // 10% error rate

	// Monitoring defaults
	config.Monitoring.EnableMetrics = true
	config.Monitoring.MetricsInterval = 10 * time.Second
	config.Monitoring.LogSlowOperations = 1 * time.Second

	return config
}

// StreamItem represents a single item to be processed in the stream
type StreamItem struct {
	ID       string                 // Unique identifier for the item
	Data     []byte                 // Raw data to process
	Target   reflect.Type           // Target type for parsing
	Context  context.Context        // Context for the processing
	Metadata map[string]interface{} // Additional metadata
}

// StreamResult represents the result of processing a stream item
type StreamResult struct {
	ID        string        // Matches StreamItem.ID
	Success   bool          // Whether processing succeeded
	Result    interface{}   // Parsed result (if successful)
	Error     error         // Error (if failed)
	Duration  time.Duration // Time taken for processing
	Attempts  int           // Number of attempts made
	Timestamp time.Time     // When processing completed
	BatchID   string        // ID of the batch this item was processed in
}

// StreamMetrics tracks stream processing performance and statistics
type StreamMetrics struct {
	mu sync.RWMutex

	// Processing counters
	TotalItems       int64 // Total items processed
	SuccessfulItems  int64 // Successfully processed items
	FailedItems      int64 // Failed processing items
	RetryCount       int64 // Total number of retries
	BatchesProcessed int64 // Total batches processed

	// Timing metrics
	AverageProcessingTime time.Duration // Average time per item
	TotalProcessingTime   time.Duration // Total processing time
	AverageBatchTime      time.Duration // Average time per batch
	TotalBatchTime        time.Duration // Total batch processing time

	// Stream metrics
	ItemsInBuffer int       // Current items in buffer
	ActiveWorkers int       // Currently active workers
	QueuedItems   int       // Items queued for processing
	Throughput    float64   // Items per second
	ErrorRate     float64   // Current error rate (0.0-1.0)
	LastFlushTime time.Time // Last time buffers were flushed

	// Error tracking
	ErrorsByType       map[string]int64 // Errors grouped by type
	CircuitBreakerOpen bool             // Whether circuit breaker is open
}

// StreamProcessor provides high-throughput stream processing with concurrent parsing workflows
type StreamProcessor[T any] struct {
	config  *StreamProcessorConfig
	workers workerpool.Pool

	// Stream components
	inputStream  stream.Stream[*StreamItem]
	outputStream chan *StreamResult

	// Processing state
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	running bool
	mu      sync.RWMutex

	// Metrics and monitoring
	metrics   *StreamMetrics
	startTime time.Time
	lastFlush time.Time

	// Circuit breaker
	errorWindow []time.Time
	cbMu        sync.Mutex
}

// NewStreamProcessor creates a new stream processor with concurrent parsing workflows
func NewStreamProcessor[T any](config *StreamProcessorConfig) (*StreamProcessor[T], error) {
	if config == nil {
		config = DefaultStreamProcessorConfig()
	}

	// Create context for processor lifecycle
	ctx, cancel := context.WithCancel(context.Background())

	// Create worker pool
	workers := workerpool.New(config.WorkerPool.MaxWorkers, config.WorkerPool.QueueSize)

	processor := &StreamProcessor[T]{
		config:       config,
		workers:      workers,
		outputStream: make(chan *StreamResult, config.Stream.BufferSize),
		ctx:          ctx,
		cancel:       cancel,
		metrics: &StreamMetrics{
			ErrorsByType: make(map[string]int64),
		},
		startTime:   time.Now(),
		lastFlush:   time.Now(),
		errorWindow: make([]time.Time, 0),
	}

	return processor, nil
}

// Start begins the stream processing
func (sp *StreamProcessor[T]) Start() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if sp.running {
		return fmt.Errorf("stream processor is already running")
	}

	// Start processing goroutines
	sp.wg.Add(2)
	go sp.metricsCollector()
	go sp.circuitBreakerMonitor()

	sp.running = true
	sp.startTime = time.Now()

	return nil
}

// Stop gracefully shuts down the stream processor
func (sp *StreamProcessor[T]) Stop() error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	if !sp.running {
		return nil
	}

	// Signal shutdown
	sp.cancel()

	// Close input stream if it exists
	if sp.inputStream != nil {
		_ = sp.inputStream.Close()
	}

	// Wait for all workers to complete
	sp.wg.Wait()

	// Shutdown worker pool
	<-sp.workers.Shutdown()

	close(sp.outputStream)
	sp.running = false

	return nil
}

// ProcessStream processes a stream of raw data items concurrently
func (sp *StreamProcessor[T]) ProcessStream(ctx context.Context, items [][]byte) (stream.Stream[*StreamResult], error) {
	sp.mu.RLock()
	if !sp.running {
		sp.mu.RUnlock()
		return nil, fmt.Errorf("stream processor is not running")
	}
	sp.mu.RUnlock()

	// Convert raw data to stream items
	streamItems := make([]*StreamItem, len(items))
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	for i, data := range items {
		streamItems[i] = &StreamItem{
			ID:      fmt.Sprintf("item-%d-%d", time.Now().Unix(), i),
			Data:    data,
			Target:  targetType,
			Context: ctx,
			Metadata: map[string]interface{}{
				"index": i,
				"batch": time.Now().Unix(),
			},
		}
	}

	// Create input stream
	inputStream := stream.FromSlice(streamItems)

	// Process items through the pipeline
	resultStream := sp.processItemsWithBackpressure(ctx, inputStream)

	return resultStream, nil
}

// ProcessChannel processes items from a channel with streaming
func (sp *StreamProcessor[T]) ProcessChannel(ctx context.Context, input <-chan []byte) (stream.Stream[*StreamResult], error) {
	sp.mu.RLock()
	if !sp.running {
		sp.mu.RUnlock()
		return nil, fmt.Errorf("stream processor is not running")
	}
	sp.mu.RUnlock()

	// Convert channel to stream items
	itemChan := make(chan *StreamItem, sp.config.Stream.BufferSize)
	targetType := reflect.TypeOf((*T)(nil)).Elem()

	// Start goroutine to convert input to stream items
	go func() {
		defer close(itemChan)
		itemCounter := 0

		for {
			select {
			case data, ok := <-input:
				if !ok {
					return
				}

				item := &StreamItem{
					ID:      fmt.Sprintf("channel-item-%d-%d", time.Now().Unix(), itemCounter),
					Data:    data,
					Target:  targetType,
					Context: ctx,
					Metadata: map[string]interface{}{
						"index":  itemCounter,
						"source": "channel",
						"batch":  time.Now().Unix(),
					},
				}

				select {
				case itemChan <- item:
					itemCounter++
				case <-ctx.Done():
					return
				case <-sp.ctx.Done():
					return
				}

			case <-ctx.Done():
				return
			case <-sp.ctx.Done():
				return
			}
		}
	}()

	// Create input stream from channel
	inputStream := stream.FromChannel(itemChan)

	// Process items through the pipeline
	resultStream := sp.processItemsWithBackpressure(ctx, inputStream)

	return resultStream, nil
}

// processItemsWithBackpressure processes stream items with backpressure management
func (sp *StreamProcessor[T]) processItemsWithBackpressure(ctx context.Context, inputStream stream.Stream[*StreamItem]) stream.Stream[*StreamResult] {
	// Create output channel for results with backpressure
	resultChan := make(chan *StreamResult, sp.config.Stream.BackpressureSize)

	// Start processing pipeline
	go func() {
		defer close(resultChan)
		defer func() { _ = inputStream.Close() }()

		// Use a wait group to track all processing tasks
		var wg sync.WaitGroup

		// Process items with concurrent processing
		err := inputStream.
			Peek(func(item *StreamItem) {
				sp.updateMetrics(func(m *StreamMetrics) {
					m.ItemsInBuffer++
				})
			}).
			ForEach(ctx, func(item *StreamItem) {
				wg.Add(1)
				go func(item *StreamItem) {
					defer wg.Done()
					sp.processItemDirectly(ctx, item, resultChan)
				}(item)
			})

		// Wait for all items to be processed
		wg.Wait()

		if err != nil && err != context.Canceled {
			// Log error but don't stop processing
			fmt.Printf("Stream processing error: %v\n", err)
		}
	}()

	// Return stream from result channel
	return stream.FromChannel(resultChan)
}

// processItemDirectly processes a single item directly without using the worker pool
func (sp *StreamProcessor[T]) processItemDirectly(ctx context.Context, item *StreamItem, resultChan chan<- *StreamResult) {
	// Check circuit breaker
	if sp.isCircuitBreakerOpen() {
		result := &StreamResult{
			ID:        item.ID,
			Success:   false,
			Error:     fmt.Errorf("circuit breaker is open"),
			Duration:  0,
			Attempts:  0,
			Timestamp: time.Now(),
		}

		select {
		case resultChan <- result:
		case <-ctx.Done():
		case <-sp.ctx.Done():
		}
		return
	}

	start := time.Now()
	attempts := 0

	for attempts <= sp.config.Pipeline.RetryAttempts {
		attempts++

		// Create timeout context for this attempt
		attemptCtx, cancel := context.WithTimeout(ctx, sp.config.Pipeline.TimeoutPerItem)

		// Perform the actual parsing/processing
		result, err := sp.performProcessing(attemptCtx, item)
		cancel()

		duration := time.Since(start)

		// Create result
		streamResult := &StreamResult{
			ID:        item.ID,
			Success:   err == nil,
			Result:    result,
			Error:     err,
			Duration:  duration,
			Attempts:  attempts,
			Timestamp: time.Now(),
			BatchID:   fmt.Sprintf("batch-%d", time.Now().Unix()),
		}

		// Update metrics
		sp.updateProcessingMetrics(err == nil, duration, err)

		// If successful or max attempts reached, stop retrying
		if err == nil || attempts > sp.config.Pipeline.RetryAttempts {
			// Send final result
			select {
			case resultChan <- streamResult:
			case <-ctx.Done():
				return
			case <-sp.ctx.Done():
				return
			}
			break
		}

		// Wait before retry with exponential backoff
		backoff := time.Duration(attempts) * sp.config.Pipeline.RetryBackoff
		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return
		case <-sp.ctx.Done():
			return
		}
	}
}

// performProcessing performs the actual parsing/processing of an item
func (sp *StreamProcessor[T]) performProcessing(ctx context.Context, item *StreamItem) (interface{}, error) {
	// Create a new instance of the target type
	targetValue := reflect.New(item.Target)
	targetPtr := targetValue.Interface()

	// Use the existing parsing infrastructure
	format := DetectFormat(item.Data)
	parser := GetParser(format)

	// Parse into generic structure
	parsedData, err := parser.Parse(item.Data)
	if err != nil {
		return nil, fmt.Errorf("parsing failed for item %s: %w", item.ID, err)
	}

	// Use existing parseIntoTarget function from validation pipeline
	err = parseIntoTarget(parsedData, targetPtr, format)
	if err != nil {
		return nil, fmt.Errorf("processing failed for item %s: %w", item.ID, err)
	}

	// Return the parsed value
	return targetValue.Elem().Interface(), nil
}

// Results returns a channel for receiving processing results
func (sp *StreamProcessor[T]) Results() <-chan *StreamResult {
	return sp.outputStream
}

// GetMetrics returns current processor metrics
func (sp *StreamProcessor[T]) GetMetrics() *StreamMetrics {
	sp.metrics.mu.RLock()
	defer sp.metrics.mu.RUnlock()

	// Create a copy of metrics to avoid race conditions
	metrics := &StreamMetrics{
		TotalItems:            sp.metrics.TotalItems,
		SuccessfulItems:       sp.metrics.SuccessfulItems,
		FailedItems:           sp.metrics.FailedItems,
		RetryCount:            sp.metrics.RetryCount,
		BatchesProcessed:      sp.metrics.BatchesProcessed,
		AverageProcessingTime: sp.metrics.AverageProcessingTime,
		TotalProcessingTime:   sp.metrics.TotalProcessingTime,
		AverageBatchTime:      sp.metrics.AverageBatchTime,
		TotalBatchTime:        sp.metrics.TotalBatchTime,
		ItemsInBuffer:         sp.metrics.ItemsInBuffer,
		ActiveWorkers:         sp.metrics.ActiveWorkers,
		QueuedItems:           sp.metrics.QueuedItems,
		Throughput:            sp.metrics.Throughput,
		ErrorRate:             sp.metrics.ErrorRate,
		LastFlushTime:         sp.metrics.LastFlushTime,
		ErrorsByType:          make(map[string]int64),
		CircuitBreakerOpen:    sp.metrics.CircuitBreakerOpen,
	}

	for k, v := range sp.metrics.ErrorsByType {
		metrics.ErrorsByType[k] = v
	}

	return metrics
}

// IsRunning returns whether the processor is currently running
func (sp *StreamProcessor[T]) IsRunning() bool {
	sp.mu.RLock()
	defer sp.mu.RUnlock()
	return sp.running
}

// GetUptime returns how long the processor has been running
func (sp *StreamProcessor[T]) GetUptime() time.Duration {
	return time.Since(sp.startTime)
}

// updateMetrics updates processor metrics with a function
func (sp *StreamProcessor[T]) updateMetrics(updateFunc func(*StreamMetrics)) {
	sp.metrics.mu.Lock()
	defer sp.metrics.mu.Unlock()
	updateFunc(sp.metrics)
}

// updateProcessingMetrics updates metrics after processing an item
func (sp *StreamProcessor[T]) updateProcessingMetrics(success bool, duration time.Duration, err error) {
	sp.metrics.mu.Lock()
	defer sp.metrics.mu.Unlock()

	sp.metrics.TotalItems++
	sp.metrics.TotalProcessingTime += duration

	if success {
		sp.metrics.SuccessfulItems++
	} else {
		sp.metrics.FailedItems++

		// Track error types
		if err != nil {
			errorType := reflect.TypeOf(err).String()
			sp.metrics.ErrorsByType[errorType]++
		}

		// Update circuit breaker error window
		sp.cbMu.Lock()
		sp.errorWindow = append(sp.errorWindow, time.Now())
		sp.cbMu.Unlock()
	}

	// Calculate averages
	if sp.metrics.TotalItems > 0 {
		sp.metrics.AverageProcessingTime = time.Duration(
			int64(sp.metrics.TotalProcessingTime) / sp.metrics.TotalItems,
		)

		sp.metrics.ErrorRate = float64(sp.metrics.FailedItems) / float64(sp.metrics.TotalItems)
	}

	// Log slow operations if monitoring is enabled
	if sp.config.Monitoring.LogSlowOperations > 0 && duration > sp.config.Monitoring.LogSlowOperations {
		fmt.Printf("Slow stream processing operation: duration=%v, success=%v\n", duration, success)
	}
}

// metricsCollector periodically collects and updates metrics
func (sp *StreamProcessor[T]) metricsCollector() {
	defer sp.wg.Done()

	if !sp.config.Monitoring.EnableMetrics {
		return
	}

	ticker := time.NewTicker(sp.config.Monitoring.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sp.collectMetrics()

		case <-sp.ctx.Done():
			return
		}
	}
}

// collectMetrics collects current processor metrics
func (sp *StreamProcessor[T]) collectMetrics() {
	sp.metrics.mu.Lock()
	defer sp.metrics.mu.Unlock()

	// Calculate throughput (items per second)
	uptime := time.Since(sp.startTime).Seconds()
	if uptime > 0 {
		sp.metrics.Throughput = float64(sp.metrics.TotalItems) / uptime
	}

	// Update queue metrics (simplified - would need access to worker pool internals)
	sp.metrics.QueuedItems = 0                                 // Would be implemented with worker pool stats
	sp.metrics.ActiveWorkers = sp.config.WorkerPool.MaxWorkers // Simplified

	sp.metrics.LastFlushTime = time.Now()
}

// isCircuitBreakerOpen checks if the circuit breaker should be open
func (sp *StreamProcessor[T]) isCircuitBreakerOpen() bool {
	sp.cbMu.Lock()
	defer sp.cbMu.Unlock()

	// Clean old errors (older than 1 minute)
	cutoff := time.Now().Add(-time.Minute)
	var recentErrors []time.Time
	for _, errorTime := range sp.errorWindow {
		if errorTime.After(cutoff) {
			recentErrors = append(recentErrors, errorTime)
		}
	}
	sp.errorWindow = recentErrors

	// Check if error rate exceeds threshold
	if len(sp.errorWindow) >= 10 { // Minimum 10 errors to consider
		errorRate := float64(len(sp.errorWindow)) / 60.0 // Errors per second over last minute
		threshold := sp.config.Pipeline.ErrorThreshold * float64(sp.config.WorkerPool.MaxWorkers)

		return errorRate > threshold
	}

	return false
}

// circuitBreakerMonitor monitors circuit breaker status
func (sp *StreamProcessor[T]) circuitBreakerMonitor() {
	defer sp.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			isOpen := sp.isCircuitBreakerOpen()
			sp.updateMetrics(func(m *StreamMetrics) {
				m.CircuitBreakerOpen = isOpen
			})

		case <-sp.ctx.Done():
			return
		}
	}
}
