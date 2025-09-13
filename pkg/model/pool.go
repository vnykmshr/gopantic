package model

import (
	"reflect"
	"sync"
)

// ObjectPool manages reusable objects to reduce memory allocations.
// This generic pool can be used for any type T and automatically handles
// object creation, reset, and reuse for improved performance.
type ObjectPool[T any] struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool for type T.
// The pool automatically creates new instances of T when needed and
// reuses existing instances to reduce garbage collection pressure.
func NewObjectPool[T any]() *ObjectPool[T] {
	return &ObjectPool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				var zero T
				return &zero
			},
		},
	}
}

// Get retrieves an object from the pool.
// If no objects are available, a new one is created automatically.
// The returned object should be returned to the pool using Put() when done.
func (p *ObjectPool[T]) Get() *T {
	return p.pool.Get().(*T)
}

// Put returns an object to the pool after resetting it to zero value.
// This ensures that pooled objects are clean and ready for reuse.
// Always call Put() when you're done with an object from Get().
func (p *ObjectPool[T]) Put(obj *T) {
	// Reset the object to zero value before returning to pool
	var zero T
	*obj = zero
	p.pool.Put(obj)
}

// Global pools for common types
var (
	errorListPool = sync.Pool{
		New: func() interface{} {
			return make(ErrorList, 0, 4) // Pre-allocate capacity for 4 errors
		},
	}

	mapPool = sync.Pool{
		New: func() interface{} {
			return make(map[string]interface{}, 16) // Pre-allocate for 16 keys
		},
	}

	stringSlicePool = sync.Pool{
		New: func() interface{} {
			return make([]string, 0, 8) // Pre-allocate for 8 strings
		},
	}
)

// GetErrorList gets an ErrorList from the global pool.
// This is optimized for gopantic's internal error handling and reduces
// allocations during parsing operations.
func GetErrorList() ErrorList {
	return errorListPool.Get().(ErrorList)
}

// PutErrorList returns an ErrorList to the global pool after clearing it.
// The slice is reset to zero length but capacity is preserved for reuse.
func PutErrorList(el ErrorList) {
	// Clear the slice but keep capacity
	el = el[:0]
	errorListPool.Put(&el)
}

// GetMap gets a map[string]interface{} from the global pool.
// Pre-allocated with capacity for 16 keys to optimize common parsing scenarios.
func GetMap() map[string]interface{} {
	return mapPool.Get().(map[string]interface{})
}

// PutMap returns a map to the global pool after clearing all entries.
// The underlying capacity is preserved for efficient reuse.
func PutMap(m map[string]interface{}) {
	// Clear the map but keep capacity
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

// GetStringSlice gets a []string from the global pool.
// Pre-allocated with capacity for 8 strings to optimize common parsing scenarios.
func GetStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the global pool after clearing it.
// The slice is reset to zero length but capacity is preserved for reuse.
func PutStringSlice(s []string) {
	// Clear the slice but keep capacity
	s = s[:0]
	stringSlicePool.Put(&s)
}

// PooledParseIntoWithFormat parses data into type T using object pools to reduce memory allocations.
// This function combines the performance benefits of OptimizedParseIntoWithFormat with object pooling
// for maximum efficiency in high-throughput scenarios. Use this when memory allocation pressure
// is a concern.
//
// Example:
//
//	user, err := model.PooledParseIntoWithFormat[User](jsonData, model.FormatJSON)
//	if err != nil {
//	    log.Fatal(err)
//	}
func PooledParseIntoWithFormat[T any](raw []byte, format Format) (T, error) {
	var zero T
	errors := GetErrorList()
	defer PutErrorList(errors)

	// Get the appropriate parser for the format
	parser := GetParser(format)

	// Parse into a generic map structure
	data, err := parser.Parse(raw)
	if err != nil {
		errors.Add(err)
		return zero, errors.AsError()
	}

	// Get struct info (cached)
	resultType := reflect.TypeOf(zero)
	structInfo := GetStructInfo(resultType)

	// Create new instance of T
	resultValue := reflect.New(resultType).Elem()

	// Parse validation rules for this struct type (cached)
	validation := ParseValidationTags(resultType)

	// Process each field using cached info (parsing and coercion pass)
	for i := range structInfo.Fields {
		fieldInfo := &structInfo.Fields[i]
		fieldValue := resultValue.Field(fieldInfo.Index)

		// Get field key for the format
		fieldKey := GetFieldKeyForFormat(fieldInfo, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Get value from data map
		rawValue, exists := data[fieldKey]
		if !exists {
			rawValue = nil
		}

		// Coerce and set the value
		if err := setFieldValue(fieldValue, rawValue, fieldInfo.Name, format); err != nil {
			errors.Add(err)
		}
	}

	// Validation pass - now that all fields are parsed, we can do cross-field validation
	for i := range structInfo.Fields {
		fieldInfo := &structInfo.Fields[i]
		fieldValue := resultValue.Field(fieldInfo.Index)

		// Get field key for the format
		fieldKey := GetFieldKeyForFormat(fieldInfo, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Apply validation rules (including cross-field validators)
		if err := validateFieldValueWithStruct(fieldInfo.Name, fieldKey, fieldValue.Interface(), validation, resultValue); err != nil {
			errors.Add(err)
		}
	}

	if errors.HasErrors() {
		// Create a new ErrorList since we're returning the pooled one
		finalErrors := make(ErrorList, len(errors))
		copy(finalErrors, errors)
		return zero, finalErrors.AsError()
	}

	return resultValue.Interface().(T), nil
}

// PerformanceMetrics tracks parsing performance statistics across parsing operations.
// This provides detailed insights into parsing performance, cache effectiveness,
// and can help identify optimization opportunities.
type PerformanceMetrics struct {
	ParseCount  int64
	TotalTime   int64 // nanoseconds
	TotalBytes  int64
	ErrorCount  int64
	CacheHits   int64
	CacheMisses int64

	mutex sync.RWMutex
}

// GlobalMetrics is the global instance for tracking parsing performance statistics.
// Use this to monitor performance across your entire application.
var GlobalMetrics = &PerformanceMetrics{}

// RecordParse records a successful parse operation with timing and byte count.
// This method is thread-safe and can be called concurrently.
func (m *PerformanceMetrics) RecordParse(durationNs, bytes int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ParseCount++
	m.TotalTime += durationNs
	m.TotalBytes += bytes
}

// RecordError records a parse error occurrence.
// This method is thread-safe and can be called concurrently.
func (m *PerformanceMetrics) RecordError() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ErrorCount++
}

// RecordCacheHit records a cache hit occurrence.
// This method is thread-safe and can be called concurrently.
func (m *PerformanceMetrics) RecordCacheHit() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CacheHits++
}

// RecordCacheMiss records a cache miss occurrence.
// This method is thread-safe and can be called concurrently.
func (m *PerformanceMetrics) RecordCacheMiss() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CacheMisses++
}

// Stats holds performance statistics returned by GetStats().
// All timing values are in nanoseconds, and averages are calculated across all operations.
type Stats struct {
	ParseCount       int64
	ErrorCount       int64
	CacheHits        int64
	CacheMisses      int64
	AvgTimeNs        float64
	AvgBytesPerParse float64
}

// GetStats returns current performance statistics as a Stats struct.
// This method is thread-safe and provides a snapshot of current metrics.
func (m *PerformanceMetrics) GetStats() Stats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := Stats{
		ParseCount:  m.ParseCount,
		ErrorCount:  m.ErrorCount,
		CacheHits:   m.CacheHits,
		CacheMisses: m.CacheMisses,
	}

	if m.ParseCount > 0 {
		stats.AvgTimeNs = float64(m.TotalTime) / float64(m.ParseCount)
		stats.AvgBytesPerParse = float64(m.TotalBytes) / float64(m.ParseCount)
	}

	return stats
}

// Reset resets all metrics to zero.
// This method is thread-safe and useful for resetting measurements.
func (m *PerformanceMetrics) Reset() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ParseCount = 0
	m.TotalTime = 0
	m.TotalBytes = 0
	m.ErrorCount = 0
	m.CacheHits = 0
	m.CacheMisses = 0
}
