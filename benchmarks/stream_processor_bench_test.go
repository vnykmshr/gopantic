package benchmarks

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// BenchmarkUser represents a typical user structure for streaming benchmarks
type BenchmarkUser struct {
	ID       int                    `json:"id" validate:"required,min=1"`
	Name     string                 `json:"name" validate:"required,min=2,max=100"`
	Email    string                 `json:"email" validate:"required,email"`
	Age      int                    `json:"age" validate:"min=0,max=150"`
	Company  string                 `json:"company" validate:"required,min=2"`
	Active   bool                   `json:"active"`
	Tags     []string               `json:"tags"`
	JoinDate string                 `json:"join_date" validate:"required"`
	Metadata map[string]interface{} `json:"metadata"`
}

// generateBenchmarkData creates test data for streaming benchmarks
func generateBenchmarkData(count int) [][]byte {
	data := make([][]byte, count)
	for i := 0; i < count; i++ {
		user := BenchmarkUser{
			ID:       i + 1,
			Name:     fmt.Sprintf("User-%d", i+1),
			Email:    fmt.Sprintf("user%d@company%d.com", i+1, (i%10)+1),
			Age:      25 + (i % 40),
			Company:  fmt.Sprintf("Company-%d", (i%20)+1),
			Active:   i%2 == 0,
			Tags:     []string{"tag1", "tag2", fmt.Sprintf("tag%d", i%5)},
			JoinDate: time.Now().AddDate(0, -(i % 12), 0).Format("2006-01-02"),
			Metadata: map[string]interface{}{
				"source":    "benchmark",
				"batch":     i / 100,
				"generated": time.Now().Unix(),
			},
		}
		jsonData, _ := json.Marshal(user)
		data[i] = jsonData
	}
	return data
}

// Benchmark: Vanilla JSON parsing (standard library)
func BenchmarkVanillaJSON_Small(b *testing.B) {
	data := generateBenchmarkData(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			var user BenchmarkUser
			if err := json.Unmarshal(jsonData, &user); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

func BenchmarkVanillaJSON_Medium(b *testing.B) {
	data := generateBenchmarkData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			var user BenchmarkUser
			if err := json.Unmarshal(jsonData, &user); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

func BenchmarkVanillaJSON_Large(b *testing.B) {
	data := generateBenchmarkData(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			var user BenchmarkUser
			if err := json.Unmarshal(jsonData, &user); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

// Benchmark: Gopantic ParseInto (single-threaded)
func BenchmarkGopanticParseInto_Small(b *testing.B) {
	data := generateBenchmarkData(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			if user, err := model.ParseInto[BenchmarkUser](jsonData); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

func BenchmarkGopanticParseInto_Medium(b *testing.B) {
	data := generateBenchmarkData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			if user, err := model.ParseInto[BenchmarkUser](jsonData); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

func BenchmarkGopanticParseInto_Large(b *testing.B) {
	data := generateBenchmarkData(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			if user, err := model.ParseInto[BenchmarkUser](jsonData); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

// Benchmark: StreamProcessor (concurrent, low workers)
func BenchmarkStreamProcessor_Small_2Workers(b *testing.B) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 2
	config.Monitoring.EnableMetrics = false // Reduce overhead

	processor, err := model.NewStreamProcessor[BenchmarkUser](config)
	if err != nil {
		b.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Stop()

	if err := processor.Start(); err != nil {
		b.Fatalf("Failed to start processor: %v", err)
	}

	data := generateBenchmarkData(100)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		
		resultStream, err := processor.ProcessStream(ctx, data)
		if err != nil {
			b.Fatalf("Failed to process stream: %v", err)
		}

		var results []BenchmarkUser
		err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
			if result.Success {
				if user, ok := result.Result.(BenchmarkUser); ok {
					results = append(results, user)
				}
			}
		})
		resultStream.Close()
		cancel()

		if err != nil {
			b.Fatalf("Failed to collect results: %v", err)
		}
		_ = results
	}
}

func BenchmarkStreamProcessor_Medium_5Workers(b *testing.B) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 5
	config.Monitoring.EnableMetrics = false

	processor, err := model.NewStreamProcessor[BenchmarkUser](config)
	if err != nil {
		b.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Stop()

	if err := processor.Start(); err != nil {
		b.Fatalf("Failed to start processor: %v", err)
	}

	data := generateBenchmarkData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		resultStream, err := processor.ProcessStream(ctx, data)
		if err != nil {
			b.Fatalf("Failed to process stream: %v", err)
		}

		var results []BenchmarkUser
		err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
			if result.Success {
				if user, ok := result.Result.(BenchmarkUser); ok {
					results = append(results, user)
				}
			}
		})
		resultStream.Close()
		cancel()

		if err != nil {
			b.Fatalf("Failed to collect results: %v", err)
		}
		_ = results
	}
}

func BenchmarkStreamProcessor_Large_10Workers(b *testing.B) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 10
	config.Monitoring.EnableMetrics = false

	processor, err := model.NewStreamProcessor[BenchmarkUser](config)
	if err != nil {
		b.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Stop()

	if err := processor.Start(); err != nil {
		b.Fatalf("Failed to start processor: %v", err)
	}

	data := generateBenchmarkData(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		
		resultStream, err := processor.ProcessStream(ctx, data)
		if err != nil {
			b.Fatalf("Failed to process stream: %v", err)
		}

		var results []BenchmarkUser
		err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
			if result.Success {
				if user, ok := result.Result.(BenchmarkUser); ok {
					results = append(results, user)
				}
			}
		})
		resultStream.Close()
		cancel()

		if err != nil {
			b.Fatalf("Failed to collect results: %v", err)
		}
		_ = results
	}
}

// Benchmark: StreamProcessor with high concurrency
func BenchmarkStreamProcessor_Large_20Workers(b *testing.B) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 20
	config.Monitoring.EnableMetrics = false

	processor, err := model.NewStreamProcessor[BenchmarkUser](config)
	if err != nil {
		b.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Stop()

	if err := processor.Start(); err != nil {
		b.Fatalf("Failed to start processor: %v", err)
	}

	data := generateBenchmarkData(10000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		
		resultStream, err := processor.ProcessStream(ctx, data)
		if err != nil {
			b.Fatalf("Failed to process stream: %v", err)
		}

		var results []BenchmarkUser
		err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
			if result.Success {
				if user, ok := result.Result.(BenchmarkUser); ok {
					results = append(results, user)
				}
			}
		})
		resultStream.Close()
		cancel()

		if err != nil {
			b.Fatalf("Failed to collect results: %v", err)
		}
		_ = results
	}
}

// Benchmark: Manual concurrent processing (for comparison)
func BenchmarkManualConcurrent_Medium_5Goroutines(b *testing.B) {
	data := generateBenchmarkData(1000)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		const workers = 5
		resultsChan := make(chan BenchmarkUser, len(data))
		var wg sync.WaitGroup
		dataChan := make(chan []byte, len(data))

		// Send data to channel
		for _, jsonData := range data {
			dataChan <- jsonData
		}
		close(dataChan)

		// Start workers
		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for jsonData := range dataChan {
					if user, err := model.ParseInto[BenchmarkUser](jsonData); err == nil {
						resultsChan <- user
					}
				}
			}()
		}

		// Wait and collect results
		go func() {
			wg.Wait()
			close(resultsChan)
		}()

		var results []BenchmarkUser
		for user := range resultsChan {
			results = append(results, user)
		}
		_ = results
	}
}

// Memory benchmarks
func BenchmarkMemory_VanillaJSON_1000Items(b *testing.B) {
	data := generateBenchmarkData(1000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var results []BenchmarkUser
		for _, jsonData := range data {
			var user BenchmarkUser
			if err := json.Unmarshal(jsonData, &user); err == nil {
				results = append(results, user)
			}
		}
		_ = results
	}
}

func BenchmarkMemory_StreamProcessor_1000Items(b *testing.B) {
	config := model.DefaultStreamProcessorConfig()
	config.WorkerPool.MaxWorkers = 5
	config.Monitoring.EnableMetrics = false

	processor, err := model.NewStreamProcessor[BenchmarkUser](config)
	if err != nil {
		b.Fatalf("Failed to create processor: %v", err)
	}
	defer processor.Stop()

	if err := processor.Start(); err != nil {
		b.Fatalf("Failed to start processor: %v", err)
	}

	data := generateBenchmarkData(1000)
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		
		resultStream, err := processor.ProcessStream(ctx, data)
		if err != nil {
			b.Fatalf("Failed to process stream: %v", err)
		}

		var results []BenchmarkUser
		err = resultStream.ForEach(ctx, func(result *model.StreamResult) {
			if result.Success {
				if user, ok := result.Result.(BenchmarkUser); ok {
					results = append(results, user)
				}
			}
		})
		resultStream.Close()
		cancel()

		if err != nil {
			b.Fatalf("Failed to collect results: %v", err)
		}
		_ = results
	}
}