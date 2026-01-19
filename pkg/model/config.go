package model

import (
	"sync"
)

// Thread-safe configuration accessors
//
// The package-level variables (MaxInputSize, MaxCacheSize, MaxValidationDepth)
// are exported for backwards compatibility but are NOT thread-safe for concurrent
// modification. For concurrent access, use the Get/Set functions below.
//
// Typical usage patterns:
//   - Startup configuration (before goroutines): Direct variable access is fine
//   - Runtime modification (with goroutines): Use Get/Set functions
//
// Note: The Get/Set functions maintain their own copy of the values. After using
// SetMaxInputSize(), internal code will use the new value on subsequent calls.

var configMu sync.RWMutex

// configValues holds thread-safe copies of configuration values.
// These are initialized from the exported variables on first access.
var configValues struct {
	maxInputSize       int
	maxCacheSize       int
	maxValidationDepth int
	initialized        bool
}

// initConfigOnce ensures configuration is initialized from exported variables
func initConfigOnce() {
	if !configValues.initialized {
		configMu.Lock()
		defer configMu.Unlock()
		if !configValues.initialized {
			configValues.maxInputSize = MaxInputSize
			configValues.maxCacheSize = MaxCacheSize
			configValues.maxValidationDepth = MaxValidationDepth
			configValues.initialized = true
		}
	}
}

// GetMaxInputSize returns the maximum input size in a thread-safe manner.
// Default: 10MB (10 * 1024 * 1024 bytes). Set to 0 to disable size checking.
func GetMaxInputSize() int {
	initConfigOnce()
	configMu.RLock()
	defer configMu.RUnlock()
	return configValues.maxInputSize
}

// SetMaxInputSize sets the maximum input size in a thread-safe manner.
// Set to 0 to disable size checking.
//
// Note: This also updates the exported MaxInputSize variable for compatibility,
// but that update is not atomic with respect to direct variable reads.
func SetMaxInputSize(size int) {
	initConfigOnce()
	configMu.Lock()
	defer configMu.Unlock()
	configValues.maxInputSize = size
	MaxInputSize = size
}

// GetMaxCacheSize returns the maximum cache size in a thread-safe manner.
// Default: 1000 types.
func GetMaxCacheSize() int {
	initConfigOnce()
	configMu.RLock()
	defer configMu.RUnlock()
	return configValues.maxCacheSize
}

// SetMaxCacheSize sets the maximum cache size in a thread-safe manner.
// Set to 0 for unlimited caching (not recommended for long-running services).
//
// Note: This also updates the exported MaxCacheSize variable for compatibility,
// but that update is not atomic with respect to direct variable reads.
func SetMaxCacheSize(size int) {
	initConfigOnce()
	configMu.Lock()
	defer configMu.Unlock()
	configValues.maxCacheSize = size
	MaxCacheSize = size
}

// GetMaxValidationDepth returns the maximum validation depth in a thread-safe manner.
// Default: 32 levels.
func GetMaxValidationDepth() int {
	initConfigOnce()
	configMu.RLock()
	defer configMu.RUnlock()
	return configValues.maxValidationDepth
}

// SetMaxValidationDepth sets the maximum validation depth in a thread-safe manner.
// This prevents stack overflow and DoS attacks from deeply nested structures.
//
// Note: This also updates the exported MaxValidationDepth variable for compatibility,
// but that update is not atomic with respect to direct variable reads.
func SetMaxValidationDepth(depth int) {
	initConfigOnce()
	configMu.Lock()
	defer configMu.Unlock()
	configValues.maxValidationDepth = depth
	MaxValidationDepth = depth
}
