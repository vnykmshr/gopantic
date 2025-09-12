package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

type User struct {
	ID    int    `json:"id" yaml:"id" validate:"required,min=1"`
	Name  string `json:"name" yaml:"name" validate:"required,min=2"`
	Email string `json:"email" yaml:"email" validate:"required,email"`
	Age   int    `json:"age" yaml:"age" validate:"min=18,max=120"`
}

type Config struct {
	Database struct {
		Host string `yaml:"host" json:"host" validate:"required"`
		Port int    `yaml:"port" json:"port" validate:"min=1,max=65535"`
	} `yaml:"database" json:"database" validate:"required"`
	Server struct {
		Port int `yaml:"port" json:"port" validate:"min=1000,max=65535"`
	} `yaml:"server" json:"server" validate:"required"`
}

func main() {
	fmt.Println("=== Gopantic Caching Demo ===")

	// Example 1: Using CachedParser instance
	fmt.Println("1. Using CachedParser Instance:")

	// Create a custom cached parser with specific configuration
	cacheConfig := &model.CacheConfig{
		TTL:        5 * time.Minute,
		MaxEntries: 500,
		Namespace:  "demo:users",
	}

	userParser, err := model.NewCachedParser[User](cacheConfig)
	if err != nil {
		log.Fatal("Failed to create cached parser:", err)
	}
	defer userParser.Close()

	userData := []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30
	}`)

	// First parse - will be cached
	fmt.Println("First parse (cache miss):")
	start := time.Now()
	user1, err := userParser.Parse(userData)
	duration1 := time.Since(start)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	fmt.Printf("Result: %+v\n", user1)
	fmt.Printf("Duration: %v\n", duration1)

	// Second parse - will use cache
	fmt.Println("\nSecond parse (cache hit):")
	start = time.Now()
	user2, err := userParser.Parse(userData)
	duration2 := time.Since(start)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	fmt.Printf("Result: %+v\n", user2)
	fmt.Printf("Duration: %v\n", duration2)

	// Show cache statistics
	stats := userParser.Stats()
	fmt.Printf("\nCache Statistics:\n")
	fmt.Printf("- Hits: %d\n", stats.Hits())
	fmt.Printf("- Misses: %d\n", stats.Misses())
	fmt.Printf("- Hit Rate: %.2f%%\n", stats.HitRate())
	fmt.Printf("- Total Requests: %d\n", stats.Total())

	// Example 2: Global cached functions
	fmt.Println("\n2. Using Global Cached Functions:")

	configData := []byte(`
database:
  host: localhost
  port: 5432
server:
  port: 8080
`)

	// First call - cache miss
	fmt.Println("First parse (global cache miss):")
	start = time.Now()
	config1, err := model.ParseIntoCached[Config](configData)
	duration1 = time.Since(start)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	fmt.Printf("Result: %+v\n", config1)
	fmt.Printf("Duration: %v\n", duration1)

	// Second call - cache hit
	fmt.Println("\nSecond parse (global cache hit):")
	start = time.Now()
	config2, err := model.ParseIntoCached[Config](configData)
	duration2 = time.Since(start)
	if err != nil {
		log.Fatal("Parse error:", err)
	}
	fmt.Printf("Result: %+v\n", config2)
	fmt.Printf("Duration: %v\n", duration2)

	// Show global cache statistics
	globalStats := model.GetGlobalCacheStats()
	fmt.Printf("\nGlobal Cache Statistics:\n")
	for typeKey, stats := range globalStats {
		fmt.Printf("Type: %s\n", typeKey)
		fmt.Printf("- Hits: %d\n", stats.Hits())
		fmt.Printf("- Misses: %d\n", stats.Misses())
		fmt.Printf("- Hit Rate: %.2f%%\n", stats.HitRate())
		fmt.Printf("- Total Requests: %d\n", stats.Total())
	}

	// Example 3: Format-specific caching
	fmt.Println("\n3. Format-specific Caching:")

	yamlUserData := []byte(`
id: 456
name: "Jane Smith"
email: "jane@test.com" 
age: 28
`)

	jsonUserData := []byte(`{
		"id": 456,
		"name": "Jane Smith",
		"email": "jane@test.com",
		"age": 28
	}`)

	// Parse YAML version
	yamlUser, err := model.ParseIntoWithFormatCached[User](yamlUserData, model.FormatYAML)
	if err != nil {
		log.Fatal("YAML parse error:", err)
	}
	fmt.Printf("YAML Result: %+v\n", yamlUser)

	// Parse JSON version - should be cached separately due to different format
	jsonUser, err := model.ParseIntoWithFormatCached[User](jsonUserData, model.FormatJSON)
	if err != nil {
		log.Fatal("JSON parse error:", err)
	}
	fmt.Printf("JSON Result: %+v\n", jsonUser)

	// Example 4: Performance comparison
	fmt.Println("\n4. Performance Comparison:")

	// Clear existing cache for fair comparison
	model.ClearAllCaches()

	largeConfigData := []byte(`
database:
  host: localhost
  port: 5432
  username: admin
  password: secret123
  ssl: true
  timeout: "30s"
  
server:
  port: 8080
  workers: 10
  hosts:
    - api.example.com
    - cdn.example.com
    - static.example.com
    - admin.example.com
    - metrics.example.com
`)

	const iterations = 1000

	// Test without caching
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := model.ParseInto[Config](largeConfigData)
		if err != nil {
			log.Fatal("Parse error:", err)
		}
	}
	uncachedDuration := time.Since(start)

	// Test with caching
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := model.ParseIntoCached[Config](largeConfigData)
		if err != nil {
			log.Fatal("Cached parse error:", err)
		}
	}
	cachedDuration := time.Since(start)

	fmt.Printf("Performance Results (%d iterations):\n", iterations)
	fmt.Printf("- Uncached: %v (avg: %v per parse)\n",
		uncachedDuration, uncachedDuration/iterations)
	fmt.Printf("- Cached: %v (avg: %v per parse)\n",
		cachedDuration, cachedDuration/iterations)

	speedup := float64(uncachedDuration) / float64(cachedDuration)
	fmt.Printf("- Speedup: %.2fx faster with caching\n", speedup)

	// Final cache statistics
	finalStats := model.GetGlobalCacheStats()
	fmt.Printf("\nFinal Cache Statistics:\n")
	for typeKey, stats := range finalStats {
		fmt.Printf("Type: %s\n", typeKey)
		fmt.Printf("- Hits: %d\n", stats.Hits())
		fmt.Printf("- Misses: %d\n", stats.Misses())
		fmt.Printf("- Hit Rate: %.2f%%\n", stats.HitRate())
		fmt.Printf("- Total Requests: %d\n", stats.Total())
		fmt.Println()
	}

	// Example 5: Redis distributed caching (if Redis is available)
	fmt.Println("\n5. Redis Distributed Caching:")

	// Try to create a Redis cache parser
	redisConfig := model.DefaultRedisCacheConfig("localhost:6379")
	redisConfig.TTL = 2 * time.Minute
	redisConfig.RedisConfig.DB = 0 // Use default database
	redisConfig.RedisConfig.KeyPrefix = "gopantic-demo:"

	redisParser, err := model.NewCachedParser[User](redisConfig)
	if err != nil {
		fmt.Printf("Redis not available, skipping Redis demo: %v\n", err)
	} else {
		defer redisParser.Close()

		fmt.Println("Redis cache configured successfully!")

		redisUserData := []byte(`{
			"id": 999,
			"name": "Redis User", 
			"email": "redis@example.com",
			"age": 35
		}`)

		// First parse with Redis - cache miss
		fmt.Println("First Redis parse (cache miss):")
		start := time.Now()
		redisUser1, err := redisParser.Parse(redisUserData)
		redisDuration1 := time.Since(start)
		if err != nil {
			log.Printf("Redis parse error: %v", err)
		} else {
			fmt.Printf("Result: %+v\n", redisUser1)
			fmt.Printf("Duration: %v\n", redisDuration1)
		}

		// Second parse with Redis - cache hit from distributed cache
		fmt.Println("\nSecond Redis parse (distributed cache hit):")
		start = time.Now()
		redisUser2, err := redisParser.Parse(redisUserData)
		redisDuration2 := time.Since(start)
		if err != nil {
			log.Printf("Redis parse error: %v", err)
		} else {
			fmt.Printf("Result: %+v\n", redisUser2)
			fmt.Printf("Duration: %v\n", redisDuration2)

			// Show performance improvement
			if redisDuration1 > redisDuration2 {
				improvement := float64(redisDuration1) / float64(redisDuration2)
				fmt.Printf("Redis cache speedup: %.2fx faster\n", improvement)
			}
		}

		// Show Redis cache statistics
		redisStats := redisParser.Stats()
		fmt.Printf("\nRedis Cache Statistics:\n")
		fmt.Printf("- Hits: %d\n", redisStats.Hits())
		fmt.Printf("- Misses: %d\n", redisStats.Misses())
		fmt.Printf("- Hit Rate: %.2f%%\n", redisStats.HitRate())
		fmt.Printf("- Total Requests: %d\n", redisStats.Total())
		fmt.Printf("- Key Count: %d\n", redisStats.KeyCount())
	}

	// Example 6: Advanced Redis configuration
	fmt.Println("\n6. Advanced Redis Configuration:")

	advancedRedisConfig := &model.CacheConfig{
		TTL:                3 * time.Minute,
		CompressionEnabled: true,
		Namespace:          "myapp:advanced",
		Backend:            model.CacheBackendRedis,
		RedisConfig: &model.RedisConfig{
			Addr:      "localhost:6379",
			Password:  "", // Set password if needed
			DB:        1,  // Use database 1
			KeyPrefix: "advanced-demo:",
		},
	}

	advancedParser, err := model.NewCachedParser[Config](advancedRedisConfig)
	if err != nil {
		fmt.Printf("Advanced Redis config not available: %v\n", err)
	} else {
		defer advancedParser.Close()

		fmt.Println("Advanced Redis configuration successful!")
		fmt.Printf("- Backend: %s\n", advancedRedisConfig.Backend)
		fmt.Printf("- Redis Address: %s\n", advancedRedisConfig.RedisConfig.Addr)
		fmt.Printf("- Redis Database: %d\n", advancedRedisConfig.RedisConfig.DB)
		fmt.Printf("- Key Prefix: %s\n", advancedRedisConfig.RedisConfig.KeyPrefix)
		fmt.Printf("- TTL: %v\n", advancedRedisConfig.TTL)

		// Test with configuration data
		configData := []byte(`{
			"database": {
				"host": "redis-demo.local",
				"port": 5432
			},
			"server": {
				"port": 8080
			}
		}`)

		config, err := advancedParser.Parse(configData)
		if err != nil {
			fmt.Printf("Advanced Redis parse error: %v\n", err)
		} else {
			fmt.Printf("Advanced config parsed: Database=%s:%d, Server=%d\n",
				config.Database.Host, config.Database.Port, config.Server.Port)
		}
	}

	fmt.Println("\n=== Demo Complete ===")
}
