package benchmarks

import (
	"encoding/json"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Real-world API response structures for benchmarking
type APIResponse struct {
	Status    string      `json:"status" validate:"required"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id" validate:"required"`
}

type UserProfile struct {
	UserID      int       `json:"user_id" validate:"required,min=1"`
	Username    string    `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email       string    `json:"email" validate:"required,email"`
	FirstName   string    `json:"first_name" validate:"required,alpha,min=1,max=50"`
	LastName    string    `json:"last_name" validate:"required,alpha,min=1,max=50"`
	DateOfBirth string    `json:"date_of_birth"`
	PhoneNumber *string   `json:"phone_number"`
	Address     *Address  `json:"address"`
	Preferences Settings  `json:"preferences"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
}

type Settings struct {
	Theme         string            `json:"theme" validate:"required"`
	Language      string            `json:"language" validate:"required,min=2,max=5"`
	Timezone      string            `json:"timezone" validate:"required"`
	Notifications map[string]bool   `json:"notifications"`
	CustomFields  map[string]string `json:"custom_fields"`
}

// Configuration parsing structures
type DatabaseConfig struct {
	Host        string        `json:"host" validate:"required"`
	Port        int           `json:"port" validate:"required,min=1,max=65535"`
	Username    string        `json:"username" validate:"required"`
	Password    string        `json:"password" validate:"required,min=8"`
	Database    string        `json:"database" validate:"required,alphanum"`
	MaxConns    int           `json:"max_conns" validate:"min=1,max=1000"`
	Timeout     time.Duration `json:"timeout"`
	SSLMode     string        `json:"ssl_mode" validate:"required"`
	SSLCert     *string       `json:"ssl_cert"`
	SSLKey      *string       `json:"ssl_key"`
	SSLRootCert *string       `json:"ssl_root_cert"`
}

type ApplicationConfig struct {
	AppName     string         `json:"app_name" validate:"required"`
	Version     string         `json:"version" validate:"required"`
	Environment string         `json:"environment" validate:"required"`
	Debug       bool           `json:"debug"`
	LogLevel    string         `json:"log_level" validate:"required"`
	Database    DatabaseConfig `json:"database" validate:"required"`
	Redis       RedisConfig    `json:"redis" validate:"required"`
	API         APIConfig      `json:"api" validate:"required"`
	Features    FeatureFlags   `json:"features"`
}

type RedisConfig struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Password string `json:"password"`
	DB       int    `json:"db" validate:"min=0,max=15"`
	Timeout  int    `json:"timeout" validate:"min=100"`
}

type APIConfig struct {
	Host         string   `json:"host" validate:"required"`
	Port         int      `json:"port" validate:"required,min=1,max=65535"`
	ReadTimeout  int      `json:"read_timeout" validate:"min=1"`
	WriteTimeout int      `json:"write_timeout" validate:"min=1"`
	CORS         bool     `json:"cors"`
	AllowedHosts []string `json:"allowed_hosts"`
}

type FeatureFlags struct {
	EnableMetrics     bool `json:"enable_metrics"`
	EnableTracing     bool `json:"enable_tracing"`
	EnableProfiling   bool `json:"enable_profiling"`
	EnableCompression bool `json:"enable_compression"`
}

// Generate benchmark data
var (
	userProfileJSON = []byte(`{
		"user_id": 12345,
		"username": "johndoe123",
		"email": "john.doe@example.com",
		"first_name": "John",
		"last_name": "Doe", 
		"date_of_birth": "1990-01-15",
		"phone_number": "+1-555-0123",
		"address": {
			"street": "123 Main Street",
			"city": "San Francisco", 
			"zip": "94102",
			"country": "USA"
		},
		"preferences": {
			"theme": "dark",
			"language": "en_US",
			"timezone": "America/Los_Angeles",
			"notifications": {
				"email": true,
				"push": false,
				"sms": true
			},
			"custom_fields": {
				"newsletter": "weekly",
				"marketing": "opt_in"
			}
		},
		"created_at": "2023-01-15T10:30:00Z",
		"updated_at": "2023-12-01T15:45:30Z",
		"is_active": true,
		"roles": ["user", "premium", "beta_tester"],
		"permissions": ["read_profile", "update_profile", "delete_account"]
	}`)

	appConfigJSON = []byte(`{
		"app_name": "MyApplication",
		"version": "1.2.3",
		"environment": "production",
		"debug": false,
		"log_level": "info",
		"database": {
			"host": "db.example.com",
			"port": 5432,
			"username": "app_user",
			"password": "secure_password123",
			"database": "myapp_prod",
			"max_conns": 100,
			"timeout": "30s",
			"ssl_mode": "require",
			"ssl_cert": "/etc/ssl/client-cert.pem",
			"ssl_key": "/etc/ssl/client-key.pem",
			"ssl_root_cert": "/etc/ssl/ca-cert.pem"
		},
		"redis": {
			"host": "redis.example.com",
			"port": 6379,
			"password": "redis_password",
			"db": 0,
			"timeout": 5000
		},
		"api": {
			"host": "0.0.0.0",
			"port": 8080,
			"read_timeout": 30,
			"write_timeout": 30,
			"cors": true,
			"allowed_hosts": ["example.com", "api.example.com", "app.example.com"]
		},
		"features": {
			"enable_metrics": true,
			"enable_tracing": true,
			"enable_profiling": false,
			"enable_compression": true
		}
	}`)

	// Create large batch of user profiles for stress testing
	largeBatchJSON = func() []byte {
		profiles := make([]map[string]interface{}, 500)
		for i := 0; i < 500; i++ {
			profiles[i] = map[string]interface{}{
				"user_id":     i + 1000,
				"username":    fmt.Sprintf("user%d", i),
				"email":       fmt.Sprintf("user%d@example.com", i),
				"first_name":  fmt.Sprintf("User%d", i),
				"last_name":   "TestUser",
				"is_active":   i%3 != 0,
				"roles":       []string{"user"},
				"permissions": []string{"read_profile"},
				"created_at":  "2023-01-15T10:30:00Z",
				"updated_at":  "2023-12-01T15:45:30Z",
				"preferences": map[string]interface{}{
					"theme":    "light",
					"language": "en_US",
					"timezone": "UTC",
					"notifications": map[string]bool{
						"email": true,
						"push":  false,
					},
				},
			}
		}
		data, _ := json.Marshal(profiles)
		return data
	}()
)

// Real-world parsing benchmarks
func BenchmarkRealWorld_UserProfile_StdJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var profile UserProfile
		if err := json.Unmarshal(userProfileJSON, &profile); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealWorld_UserProfile_Gopantic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[UserProfile](userProfileJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealWorld_UserProfile_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[UserProfile](nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(userProfileJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Configuration parsing benchmarks
func BenchmarkConfig_Parsing_StdJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var config ApplicationConfig
		if err := json.Unmarshal(appConfigJSON, &config); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfig_Parsing_Gopantic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[ApplicationConfig](appConfigJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConfig_Parsing_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[ApplicationConfig](nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(appConfigJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Large batch processing benchmarks
func BenchmarkLargeBatch_StdJSON(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var profiles []UserProfile
		if err := json.Unmarshal(largeBatchJSON, &profiles); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLargeBatch_Gopantic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[[]UserProfile](largeBatchJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLargeBatch_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[[]UserProfile](nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(largeBatchJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Memory allocation analysis benchmarks
func BenchmarkMemoryProfile_StdJSON(b *testing.B) {
	b.ReportAllocs()
	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < b.N; i++ {
		var profile UserProfile
		json.Unmarshal(userProfileJSON, &profile)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}

func BenchmarkMemoryProfile_Gopantic(b *testing.B) {
	b.ReportAllocs()
	var m1, m2 runtime.MemStats

	runtime.GC()
	runtime.ReadMemStats(&m1)

	for i := 0; i < b.N; i++ {
		model.ParseInto[UserProfile](userProfileJSON)
	}

	runtime.GC()
	runtime.ReadMemStats(&m2)

	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}

// Concurrent parsing benchmarks
func BenchmarkConcurrent_StdJSON(b *testing.B) {
	b.ReportAllocs()
	b.SetParallelism(runtime.NumCPU())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var profile UserProfile
			if err := json.Unmarshal(userProfileJSON, &profile); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkConcurrent_Gopantic(b *testing.B) {
	b.ReportAllocs()
	b.SetParallelism(runtime.NumCPU())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := model.ParseInto[UserProfile](userProfileJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkConcurrent_Cached(b *testing.B) {
	parser, _ := model.NewCachedParser[UserProfile](nil)
	b.ReportAllocs()
	b.SetParallelism(runtime.NumCPU())
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := parser.Parse(userProfileJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Error handling performance benchmarks
func BenchmarkErrorHandling_ValidationErrors(b *testing.B) {
	invalidJSON := []byte(`{
		"user_id": 0,
		"username": "ab",
		"email": "invalid_email",
		"first_name": "",
		"last_name": "",
		"is_active": true,
		"roles": [],
		"permissions": [],
		"created_at": "invalid_date",
		"updated_at": "invalid_date",
		"preferences": {
			"theme": "",
			"language": "x",
			"timezone": ""
		}
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[UserProfile](invalidJSON)
		if err == nil {
			b.Fatal("Expected validation errors")
		}
	}
}

// Type coercion performance benchmarks
func BenchmarkTypeCoercion_MixedTypes(b *testing.B) {
	mixedTypeJSON := []byte(`{
		"user_id": "12345",
		"username": "johndoe123",
		"email": "john.doe@example.com",
		"first_name": "John",
		"last_name": "Doe",
		"is_active": "true",
		"roles": "user,premium",
		"created_at": 1640995800,
		"updated_at": "1701434730",
		"preferences": {
			"theme": "dark",
			"language": "en_US",
			"timezone": "America/Los_Angeles"
		}
	}`)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[UserProfile](mixedTypeJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark runner helper for comparison analysis
func BenchmarkComparison_All(b *testing.B) {
	// This benchmark runs all comparison scenarios and can be used for analysis
	benchmarks := []struct {
		name string
		fn   func(*testing.B)
	}{
		{"StdJSON/SimpleUser", BenchmarkStdJSON_SimpleUser},
		{"Gopantic/SimpleUser", BenchmarkGopantic_SimpleUser},
		{"Gopantic/SimpleUser/Cached", BenchmarkGopantic_SimpleUser_Cached},
		{"StdJSON/ComplexUser", BenchmarkStdJSON_ComplexUser},
		{"Gopantic/ComplexUser", BenchmarkGopantic_ComplexUser},
		{"Gopantic/ComplexUser/Cached", BenchmarkGopantic_ComplexUser_Cached},
	}

	for _, benchmark := range benchmarks {
		b.Run(benchmark.name, benchmark.fn)
	}
}
