package benchmarks

import (
	"testing"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Optimization comparison benchmarks
func BenchmarkGopantic_SimpleUser_Original(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[SimpleUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_SimpleUser_Optimized(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.OptimizedParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_ComplexUser_Original(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ComplexUser](complexUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_ComplexUser_Optimized(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test with frequent struct parsing to show caching benefits
func BenchmarkStructInfo_Caching_Benefits(b *testing.B) {
	// This benchmark shows the benefit of struct info caching
	// when parsing the same struct type multiple times

	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := model.ParseInto[ComplexUser](complexUserJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, err := model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Memory allocation comparison
func BenchmarkMemoryOptimization_SimpleUser(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.ParseInto[SimpleUser](simpleUserJSON)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.OptimizedParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
		}
	})
}

// Test struct info cache performance
func BenchmarkStructInfoCache_Performance(b *testing.B) {
	// Clear cache to ensure clean start
	model.ClearStructInfoCache()

	b.Run("FirstAccess", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.ClearStructInfoCache() // Force cache miss each time
			model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
		}
	})

	b.Run("CachedAccess", func(b *testing.B) {
		// Warm up cache
		model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
		}
	})
}

// Concurrent access to cached struct info
func BenchmarkStructInfoCache_Concurrent(b *testing.B) {
	// Warm up cache
	model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)

	b.ReportAllocs()
	b.SetParallelism(4)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := model.OptimizedParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Object pooling benchmarks
func BenchmarkGopantic_SimpleUser_Pooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.PooledParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGopantic_ComplexUser_Pooled(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.PooledParseIntoWithFormat[ComplexUser](complexUserJSON, model.FormatJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Full optimization comparison
func BenchmarkOptimizationComparison_SimpleUser(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.ParseInto[SimpleUser](simpleUserJSON)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.OptimizedParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
		}
	})

	b.Run("Pooled", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			model.PooledParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
		}
	})
}

// Memory pressure test
func BenchmarkMemoryPressure_ManyAllocations(b *testing.B) {
	// This test simulates high allocation pressure to show pooling benefits
	b.Run("WithoutPools", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 10; j++ {
				model.OptimizedParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
			}
		}
	})

	b.Run("WithPools", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for j := 0; j < 10; j++ {
				model.PooledParseIntoWithFormat[SimpleUser](simpleUserJSON, model.FormatJSON)
			}
		}
	})
}
