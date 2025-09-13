package model

import (
	"reflect"
	"sync"
)

// ObjectPool manages reusable objects to reduce memory allocations
type ObjectPool[T any] struct {
	pool sync.Pool
}

// NewObjectPool creates a new object pool for type T
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

// Get retrieves an object from the pool
func (p *ObjectPool[T]) Get() *T {
	return p.pool.Get().(*T)
}

// Put returns an object to the pool
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

// GetErrorList gets an ErrorList from the pool
func GetErrorList() ErrorList {
	return errorListPool.Get().(ErrorList)
}

// PutErrorList returns an ErrorList to the pool after clearing it
func PutErrorList(el ErrorList) {
	// Clear the slice but keep capacity
	el = el[:0]
	errorListPool.Put(&el)
}

// GetMap gets a map from the pool
func GetMap() map[string]interface{} {
	return mapPool.Get().(map[string]interface{})
}

// PutMap returns a map to the pool after clearing it
func PutMap(m map[string]interface{}) {
	// Clear the map but keep capacity
	for k := range m {
		delete(m, k)
	}
	mapPool.Put(m)
}

// GetStringSlice gets a string slice from the pool
func GetStringSlice() []string {
	return stringSlicePool.Get().([]string)
}

// PutStringSlice returns a string slice to the pool after clearing it
func PutStringSlice(s []string) {
	// Clear the slice but keep capacity
	s = s[:0]
	stringSlicePool.Put(&s)
}

// PooledParseIntoWithFormat is a version that uses object pools to reduce allocations
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

// PerformanceMetrics tracks parsing performance statistics
type PerformanceMetrics struct {
	ParseCount  int64
	TotalTime   int64 // nanoseconds
	TotalBytes  int64
	ErrorCount  int64
	CacheHits   int64
	CacheMisses int64

	mutex sync.RWMutex
}

// GlobalMetrics tracks global parsing performance statistics
var GlobalMetrics = &PerformanceMetrics{}

// RecordParse records a successful parse operation
func (m *PerformanceMetrics) RecordParse(durationNs, bytes int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ParseCount++
	m.TotalTime += durationNs
	m.TotalBytes += bytes
}

// RecordError records a parse error
func (m *PerformanceMetrics) RecordError() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.ErrorCount++
}

// RecordCacheHit records a cache hit
func (m *PerformanceMetrics) RecordCacheHit() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CacheHits++
}

// RecordCacheMiss records a cache miss
func (m *PerformanceMetrics) RecordCacheMiss() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.CacheMisses++
}

// Stats holds performance statistics
type Stats struct {
	ParseCount       int64
	ErrorCount       int64
	CacheHits        int64
	CacheMisses      int64
	AvgTimeNs        float64
	AvgBytesPerParse float64
}

// GetStats returns current performance statistics
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

// Reset resets all metrics to zero
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
