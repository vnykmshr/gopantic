package tests

import (
	"testing"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Benchmark structures
type BenchUser struct {
	ID        int    `json:"id" yaml:"id" validate:"required,min=1"`
	Name      string `json:"name" yaml:"name" validate:"required,min=2"`
	Email     string `json:"email" yaml:"email" validate:"required,email"`
	Age       int    `json:"age" yaml:"age" validate:"min=18,max=120"`
	IsActive  bool   `json:"is_active" yaml:"is_active"`
	CreatedAt string `json:"created_at" yaml:"created_at"`
}

type BenchConfig struct {
	Database struct {
		Host     string `json:"host" yaml:"host" validate:"required"`
		Port     int    `json:"port" yaml:"port" validate:"min=1,max=65535"`
		Username string `json:"username" yaml:"username" validate:"required"`
		Password string `json:"password" yaml:"password" validate:"required"`
		SSL      bool   `json:"ssl" yaml:"ssl"`
		Timeout  string `json:"timeout" yaml:"timeout" validate:"required"`
	} `json:"database" yaml:"database" validate:"required"`
	Server struct {
		Port    int      `json:"port" yaml:"port" validate:"min=1000,max=65535"`
		Workers int      `json:"workers" yaml:"workers" validate:"min=1,max=100"`
		Hosts   []string `json:"hosts" yaml:"hosts"`
		TLS     struct {
			Enabled  bool   `json:"enabled" yaml:"enabled"`
			CertFile string `json:"cert_file" yaml:"cert_file"`
			KeyFile  string `json:"key_file" yaml:"key_file"`
		} `json:"tls" yaml:"tls"`
	} `json:"server" yaml:"server" validate:"required"`
	Logging struct {
		Level    string `json:"level" yaml:"level" validate:"required,oneof=debug info warn error"`
		File     string `json:"file" yaml:"file"`
		MaxSize  int    `json:"max_size" yaml:"max_size" validate:"min=1"`
		Compress bool   `json:"compress" yaml:"compress"`
	} `json:"logging" yaml:"logging" validate:"required"`
}

// Test data
var (
	simpleUserJSON = []byte(`{
		"id": 123,
		"name": "John Doe",
		"email": "john@example.com",
		"age": 30,
		"is_active": true,
		"created_at": "2023-12-25T10:30:00Z"
	}`)

	simpleUserYAML = []byte(`
id: 123
name: "John Doe"
email: "john@example.com"
age: 30
is_active: true
created_at: "2023-12-25T10:30:00Z"
`)

	complexConfigJSON = []byte(`{
		"database": {
			"host": "localhost",
			"port": 5432,
			"username": "admin",
			"password": "secret123",
			"ssl": true,
			"timeout": "30s"
		},
		"server": {
			"port": 8080,
			"workers": 10,
			"hosts": [
				"api.example.com",
				"cdn.example.com",
				"static.example.com",
				"admin.example.com",
				"metrics.example.com"
			],
			"tls": {
				"enabled": true,
				"cert_file": "/etc/ssl/certs/server.crt",
				"key_file": "/etc/ssl/private/server.key"
			}
		},
		"logging": {
			"level": "info",
			"file": "/var/log/app.log",
			"max_size": 100,
			"compress": true
		}
	}`)

	complexConfigYAML = []byte(`
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
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"

logging:
  level: info
  file: "/var/log/app.log"
  max_size: 100
  compress: true
`)
)

// Simple parsing benchmarks
func BenchmarkParseInto_SimpleUser_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntoCached_SimpleUser_JSON(b *testing.B) {
	// Clear cache before benchmark
	model.ClearAllCaches()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseIntoCached[BenchUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_SimpleUser_YAML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](simpleUserYAML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntoCached_SimpleUser_YAML(b *testing.B) {
	// Clear cache before benchmark
	model.ClearAllCaches()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseIntoCached[BenchUser](simpleUserYAML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Complex parsing benchmarks
func BenchmarkParseInto_ComplexConfig_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchConfig](complexConfigJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntoCached_ComplexConfig_JSON(b *testing.B) {
	// Clear cache before benchmark
	model.ClearAllCaches()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseIntoCached[BenchConfig](complexConfigJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseInto_ComplexConfig_YAML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchConfig](complexConfigYAML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntoCached_ComplexConfig_YAML(b *testing.B) {
	// Clear cache before benchmark
	model.ClearAllCaches()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseIntoCached[BenchConfig](complexConfigYAML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// CachedParser instance benchmarks
func BenchmarkCachedParser_SimpleUser_JSON(b *testing.B) {
	parser, err := model.NewCachedParser[BenchUser](&model.CacheConfig{
		TTL:        time.Hour,
		MaxEntries: 1000,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer parser.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCachedParser_ComplexConfig_YAML(b *testing.B) {
	parser, err := model.NewCachedParser[BenchConfig](&model.CacheConfig{
		TTL:        time.Hour,
		MaxEntries: 1000,
	})
	if err != nil {
		b.Fatal(err)
	}
	defer parser.Close()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := parser.ParseWithFormat(complexConfigYAML, model.FormatYAML)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Format detection benchmarks
func BenchmarkDetectFormat_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = model.DetectFormat(simpleUserJSON)
	}
}

func BenchmarkDetectFormat_YAML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = model.DetectFormat(simpleUserYAML)
	}
}

func BenchmarkDetectFormat_ComplexJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = model.DetectFormat(complexConfigJSON)
	}
}

func BenchmarkDetectFormat_ComplexYAML(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = model.DetectFormat(complexConfigYAML)
	}
}

// Memory allocation benchmarks
func BenchmarkParseInto_SimpleUser_Allocs(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseInto[BenchUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParseIntoCached_SimpleUser_Allocs(b *testing.B) {
	b.ReportAllocs()
	model.ClearAllCaches()

	for i := 0; i < b.N; i++ {
		_, err := model.ParseIntoCached[BenchUser](simpleUserJSON)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Parallel benchmarks
func BenchmarkParseInto_SimpleUser_Parallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := model.ParseInto[BenchUser](simpleUserJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkParseIntoCached_SimpleUser_Parallel(b *testing.B) {
	model.ClearAllCaches()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := model.ParseIntoCached[BenchUser](simpleUserJSON)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
