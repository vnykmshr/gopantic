package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vnykmshr/gopantic/pkg/model"
)

// Real-world application configuration structures
type ApplicationConfig struct {
	// Application metadata
	Name        string `json:"name" yaml:"name" validate:"required"`
	Version     string `json:"version" yaml:"version" validate:"required"`
	Environment string `json:"environment" yaml:"environment" validate:"required"`
	Debug       bool   `json:"debug" yaml:"debug"`

	// Server configuration
	Server ServerConfig `json:"server" yaml:"server" validate:"required"`

	// Database configuration
	Database DatabaseConfig `json:"database" yaml:"database" validate:"required"`

	// Redis cache configuration
	Redis RedisConfig `json:"redis" yaml:"redis" validate:"required"`

	// Logging configuration
	Logging LoggingConfig `json:"logging" yaml:"logging" validate:"required"`

	// Security configuration
	Security SecurityConfig `json:"security" yaml:"security" validate:"required"`

	// Feature flags
	Features FeatureFlags `json:"features" yaml:"features"`

	// External services
	Services ExternalServices `json:"services" yaml:"services"`

	// Monitoring and observability
	Monitoring MonitoringConfig `json:"monitoring" yaml:"monitoring"`
}

type ServerConfig struct {
	Host            string     `json:"host" yaml:"host" validate:"required"`
	Port            int        `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	ReadTimeout     string     `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout    string     `json:"write_timeout" yaml:"write_timeout"`
	ShutdownTimeout string     `json:"shutdown_timeout" yaml:"shutdown_timeout"`
	MaxHeaderBytes  int        `json:"max_header_bytes" yaml:"max_header_bytes" validate:"min=1024"`
	TLS             *TLSConfig `json:"tls" yaml:"tls"`
}

type TLSConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	CertFile   string `json:"cert_file" yaml:"cert_file"`
	KeyFile    string `json:"key_file" yaml:"key_file"`
	MinVersion string `json:"min_version" yaml:"min_version"`
}

type DatabaseConfig struct {
	Driver          string          `json:"driver" yaml:"driver" validate:"required"`
	Host            string          `json:"host" yaml:"host" validate:"required"`
	Port            int             `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	Username        string          `json:"username" yaml:"username" validate:"required"`
	Password        string          `json:"password" yaml:"password" validate:"required,min=8"`
	Database        string          `json:"database" yaml:"database" validate:"required"`
	MaxOpenConns    int             `json:"max_open_conns" yaml:"max_open_conns" validate:"min=1,max=1000"`
	MaxIdleConns    int             `json:"max_idle_conns" yaml:"max_idle_conns" validate:"min=1"`
	ConnMaxLifetime string          `json:"conn_max_lifetime" yaml:"conn_max_lifetime"`
	SSLMode         string          `json:"ssl_mode" yaml:"ssl_mode" validate:"required"`
	Migrations      MigrationConfig `json:"migrations" yaml:"migrations"`
}

type MigrationConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Directory string `json:"directory" yaml:"directory"`
	Table     string `json:"table" yaml:"table"`
}

type RedisConfig struct {
	Host        string         `json:"host" yaml:"host" validate:"required"`
	Port        int            `json:"port" yaml:"port" validate:"required,min=1,max=65535"`
	Password    string         `json:"password" yaml:"password"`
	DB          int            `json:"db" yaml:"db" validate:"min=0,max=15"`
	PoolSize    int            `json:"pool_size" yaml:"pool_size" validate:"min=1,max=1000"`
	DialTimeout string         `json:"dial_timeout" yaml:"dial_timeout"`
	ReadTimeout string         `json:"read_timeout" yaml:"read_timeout"`
	Cluster     *ClusterConfig `json:"cluster" yaml:"cluster"`
}

type ClusterConfig struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Nodes   []string `json:"nodes" yaml:"nodes"`
}

type LoggingConfig struct {
	Level      string                 `json:"level" yaml:"level" validate:"required"`
	Format     string                 `json:"format" yaml:"format" validate:"required"`
	Output     string                 `json:"output" yaml:"output" validate:"required"`
	File       *FileConfig            `json:"file" yaml:"file"`
	Structured bool                   `json:"structured" yaml:"structured"`
	Fields     map[string]interface{} `json:"fields" yaml:"fields"`
}

type FileConfig struct {
	Path       string `json:"path" yaml:"path" validate:"required"`
	MaxSize    int    `json:"max_size" yaml:"max_size" validate:"min=1"`
	MaxBackups int    `json:"max_backups" yaml:"max_backups" validate:"min=0"`
	MaxAge     int    `json:"max_age" yaml:"max_age" validate:"min=0"`
	Compress   bool   `json:"compress" yaml:"compress"`
}

type SecurityConfig struct {
	JWT        JWTConfig        `json:"jwt" yaml:"jwt" validate:"required"`
	RateLimit  RateLimitConfig  `json:"rate_limit" yaml:"rate_limit"`
	CORS       CORSConfig       `json:"cors" yaml:"cors"`
	CSP        string           `json:"csp" yaml:"csp"`
	Encryption EncryptionConfig `json:"encryption" yaml:"encryption"`
}

type JWTConfig struct {
	Secret         string `json:"secret" yaml:"secret" validate:"required,min=32"`
	Algorithm      string `json:"algorithm" yaml:"algorithm" validate:"required"`
	ExpirationTime string `json:"expiration_time" yaml:"expiration_time"`
	RefreshTime    string `json:"refresh_time" yaml:"refresh_time"`
	Issuer         string `json:"issuer" yaml:"issuer" validate:"required"`
	Audience       string `json:"audience" yaml:"audience" validate:"required"`
}

type RateLimitConfig struct {
	Enabled   bool     `json:"enabled" yaml:"enabled"`
	Requests  int      `json:"requests" yaml:"requests" validate:"min=1"`
	Window    string   `json:"window" yaml:"window"`
	SkipPaths []string `json:"skip_paths" yaml:"skip_paths"`
}

type CORSConfig struct {
	Enabled        bool     `json:"enabled" yaml:"enabled"`
	AllowedOrigins []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`
	MaxAge         int      `json:"max_age" yaml:"max_age"`
}

type EncryptionConfig struct {
	Key       string `json:"key" yaml:"key" validate:"required,min=32"`
	Algorithm string `json:"algorithm" yaml:"algorithm" validate:"required"`
}

type FeatureFlags struct {
	EnableMetrics     bool `json:"enable_metrics" yaml:"enable_metrics"`
	EnableTracing     bool `json:"enable_tracing" yaml:"enable_tracing"`
	EnableProfiling   bool `json:"enable_profiling" yaml:"enable_profiling"`
	EnableSwagger     bool `json:"enable_swagger" yaml:"enable_swagger"`
	EnableHealthCheck bool `json:"enable_health_check" yaml:"enable_health_check"`
	MaintenanceMode   bool `json:"maintenance_mode" yaml:"maintenance_mode"`
}

type ExternalServices struct {
	EmailService        ServiceConfig `json:"email_service" yaml:"email_service"`
	PaymentService      ServiceConfig `json:"payment_service" yaml:"payment_service"`
	StorageService      ServiceConfig `json:"storage_service" yaml:"storage_service"`
	NotificationService ServiceConfig `json:"notification_service" yaml:"notification_service"`
}

type ServiceConfig struct {
	Enabled        bool                 `json:"enabled" yaml:"enabled"`
	URL            string               `json:"url" yaml:"url"`
	APIKey         string               `json:"api_key" yaml:"api_key"`
	Timeout        string               `json:"timeout" yaml:"timeout"`
	Retries        int                  `json:"retries" yaml:"retries" validate:"min=0,max=10"`
	CircuitBreaker CircuitBreakerConfig `json:"circuit_breaker" yaml:"circuit_breaker"`
}

type CircuitBreakerConfig struct {
	Enabled          bool   `json:"enabled" yaml:"enabled"`
	FailureThreshold int    `json:"failure_threshold" yaml:"failure_threshold"`
	RecoveryTimeout  string `json:"recovery_timeout" yaml:"recovery_timeout"`
	HalfOpenRequests int    `json:"half_open_requests" yaml:"half_open_requests"`
}

type MonitoringConfig struct {
	Metrics MetricsConfig `json:"metrics" yaml:"metrics"`
	Tracing TracingConfig `json:"tracing" yaml:"tracing"`
	Health  HealthConfig  `json:"health" yaml:"health"`
}

type MetricsConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Endpoint  string `json:"endpoint" yaml:"endpoint"`
	Namespace string `json:"namespace" yaml:"namespace"`
	Interval  string `json:"interval" yaml:"interval"`
}

type TracingConfig struct {
	Enabled     bool    `json:"enabled" yaml:"enabled"`
	ServiceName string  `json:"service_name" yaml:"service_name"`
	Endpoint    string  `json:"endpoint" yaml:"endpoint"`
	SampleRate  float64 `json:"sample_rate" yaml:"sample_rate" validate:"min=0,max=1"`
}

type HealthConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Endpoint string `json:"endpoint" yaml:"endpoint"`
}

func main() {
	fmt.Println("🔧 Configuration Parsing Example")
	fmt.Println("This demonstrates real-world configuration parsing and validation using gopantic")
	fmt.Println()

	// Get the current directory
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get current directory:", err)
	}

	// Look for config files in the examples directory
	examplesDir := filepath.Join(currentDir, "examples", "config_parsing")

	// Try to parse different configuration formats
	configs := []struct {
		name   string
		file   string
		format string
	}{
		{"JSON Configuration", "config.json", "JSON"},
		{"YAML Configuration", "config.yaml", "YAML"},
		{"Development Environment", "config.dev.yaml", "YAML"},
		{"Production Environment", "config.prod.json", "JSON"},
	}

	for _, configInfo := range configs {
		fmt.Printf("📄 Parsing %s (%s)\n", configInfo.name, configInfo.format)
		fmt.Printf("   File: %s\n", configInfo.file)

		filePath := filepath.Join(examplesDir, configInfo.file)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("   ❌ File not found: %s\n", filePath)

			// Create sample file for demonstration
			if err := createSampleConfig(filePath, configInfo.format); err != nil {
				fmt.Printf("   ❌ Failed to create sample file: %v\n", err)
			} else {
				fmt.Printf("   ✅ Created sample file: %s\n", filePath)
			}
		}

		// Parse the configuration
		config, err := parseConfigFile(filePath)
		if err != nil {
			fmt.Printf("   ❌ Parsing failed: %v\n", err)

			// If it's a validation error, show structured details
			if errorList, ok := err.(model.ErrorList); ok {
				if jsonData, jsonErr := errorList.ToJSON(); jsonErr == nil {
					fmt.Printf("   📋 Validation Errors:\n")
					fmt.Printf("   %s\n", string(jsonData))
				}
			}
		} else {
			fmt.Printf("   ✅ Configuration parsed successfully\n")
			fmt.Printf("   📊 App: %s v%s (%s)\n", config.Name, config.Version, config.Environment)
			fmt.Printf("   🚀 Server: %s:%d\n", config.Server.Host, config.Server.Port)
			fmt.Printf("   🗃️  Database: %s@%s:%d/%s\n", config.Database.Username, config.Database.Host, config.Database.Port, config.Database.Database)
			fmt.Printf("   🔄 Redis: %s:%d (DB %d)\n", config.Redis.Host, config.Redis.Port, config.Redis.DB)
			fmt.Printf("   📝 Logging: %s level, %s format\n", config.Logging.Level, config.Logging.Format)
		}
		fmt.Println()
	}

	// Demonstrate configuration validation scenarios
	fmt.Println("🔍 Configuration Validation Scenarios")
	demonstrateValidationScenarios()

	// Show environment variable override example
	fmt.Println("\n🌍 Environment Variable Override Example")
	demonstrateEnvironmentOverrides()
}

func parseConfigFile(filePath string) (*ApplicationConfig, error) {
	// Read file content
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse using gopantic (auto-detects format)
	config, err := model.ParseInto[ApplicationConfig](data)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func createSampleConfig(filePath, format string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content string

	if format == "JSON" {
		content = getSampleJSONConfig()
	} else {
		content = getSampleYAMLConfig()
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

func getSampleJSONConfig() string {
	return `{
  "name": "MyApplication",
  "version": "1.0.0",
  "environment": "production",
  "debug": false,
  "server": {
    "host": "0.0.0.0",
    "port": 8080,
    "read_timeout": "30s",
    "write_timeout": "30s",
    "shutdown_timeout": "10s",
    "max_header_bytes": 1048576,
    "tls": {
      "enabled": true,
      "cert_file": "/etc/ssl/certs/server.crt",
      "key_file": "/etc/ssl/private/server.key",
      "min_version": "1.2"
    }
  },
  "database": {
    "driver": "postgres",
    "host": "db.example.com",
    "port": 5432,
    "username": "app_user",
    "password": "secure_password123",
    "database": "myapp_production",
    "max_open_conns": 25,
    "max_idle_conns": 10,
    "conn_max_lifetime": "300s",
    "ssl_mode": "require",
    "migrations": {
      "enabled": true,
      "directory": "/migrations",
      "table": "schema_migrations"
    }
  },
  "redis": {
    "host": "redis.example.com",
    "port": 6379,
    "password": "redis_password",
    "db": 0,
    "pool_size": 10,
    "dial_timeout": "5s",
    "read_timeout": "3s",
    "cluster": {
      "enabled": false,
      "nodes": []
    }
  },
  "logging": {
    "level": "info",
    "format": "json",
    "output": "file",
    "file": {
      "path": "/var/log/myapp.log",
      "max_size": 100,
      "max_backups": 3,
      "max_age": 28,
      "compress": true
    },
    "structured": true,
    "fields": {
      "service": "myapp",
      "version": "1.0.0"
    }
  },
  "security": {
    "jwt": {
      "secret": "your-super-secret-jwt-key-here-32-chars-min",
      "algorithm": "HS256",
      "expiration_time": "24h",
      "refresh_time": "168h",
      "issuer": "myapp",
      "audience": "myapp-users"
    },
    "rate_limit": {
      "enabled": true,
      "requests": 100,
      "window": "1m",
      "skip_paths": ["/health", "/metrics"]
    },
    "cors": {
      "enabled": true,
      "allowed_origins": ["https://myapp.com"],
      "allowed_methods": ["GET", "POST", "PUT", "DELETE"],
      "allowed_headers": ["Content-Type", "Authorization"],
      "max_age": 86400
    },
    "csp": "default-src 'self'",
    "encryption": {
      "key": "your-encryption-key-here-32-chars-minimum",
      "algorithm": "AES-256-GCM"
    }
  },
  "features": {
    "enable_metrics": true,
    "enable_tracing": true,
    "enable_profiling": false,
    "enable_swagger": true,
    "enable_health_check": true,
    "maintenance_mode": false
  },
  "services": {
    "email_service": {
      "enabled": true,
      "url": "https://api.emailservice.com",
      "api_key": "your-email-api-key",
      "timeout": "10s",
      "retries": 3,
      "circuit_breaker": {
        "enabled": true,
        "failure_threshold": 5,
        "recovery_timeout": "30s",
        "half_open_requests": 3
      }
    },
    "payment_service": {
      "enabled": true,
      "url": "https://api.paymentservice.com",
      "api_key": "your-payment-api-key",
      "timeout": "15s",
      "retries": 2,
      "circuit_breaker": {
        "enabled": true,
        "failure_threshold": 3,
        "recovery_timeout": "60s",
        "half_open_requests": 2
      }
    },
    "storage_service": {
      "enabled": false,
      "url": "",
      "api_key": "",
      "timeout": "30s",
      "retries": 1,
      "circuit_breaker": {
        "enabled": false,
        "failure_threshold": 0,
        "recovery_timeout": "0s",
        "half_open_requests": 0
      }
    },
    "notification_service": {
      "enabled": true,
      "url": "https://api.notificationservice.com",
      "api_key": "your-notification-api-key",
      "timeout": "5s",
      "retries": 1,
      "circuit_breaker": {
        "enabled": false,
        "failure_threshold": 0,
        "recovery_timeout": "0s",
        "half_open_requests": 0
      }
    }
  },
  "monitoring": {
    "metrics": {
      "enabled": true,
      "endpoint": "/metrics",
      "namespace": "myapp",
      "interval": "15s"
    },
    "tracing": {
      "enabled": true,
      "service_name": "myapp",
      "endpoint": "http://jaeger:14268/api/traces",
      "sample_rate": 0.1
    },
    "health": {
      "enabled": true,
      "endpoint": "/health"
    }
  }
}`
}

func getSampleYAMLConfig() string {
	return `name: MyApplication
version: 1.0.0
environment: development
debug: true

server:
  host: localhost
  port: 3000
  read_timeout: 30s
  write_timeout: 30s
  shutdown_timeout: 5s
  max_header_bytes: 1048576
  tls:
    enabled: false
    cert_file: ""
    key_file: ""
    min_version: "1.2"

database:
  driver: postgres
  host: localhost
  port: 5432
  username: dev_user
  password: dev_password123
  database: myapp_development
  max_open_conns: 10
  max_idle_conns: 5
  conn_max_lifetime: 300s
  ssl_mode: disable
  migrations:
    enabled: true
    directory: ./migrations
    table: schema_migrations

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 5
  dial_timeout: 5s
  read_timeout: 3s
  cluster:
    enabled: false
    nodes: []

logging:
  level: debug
  format: text
  output: stdout
  file: null
  structured: false
  fields:
    service: myapp
    version: 1.0.0
    environment: development

security:
  jwt:
    secret: dev-jwt-secret-key-for-development-only-32-chars
    algorithm: HS256
    expiration_time: 24h
    refresh_time: 168h
    issuer: myapp-dev
    audience: myapp-dev-users
  rate_limit:
    enabled: false
    requests: 1000
    window: 1m
    skip_paths: ["/health", "/metrics", "/debug"]
  cors:
    enabled: true
    allowed_origins: ["http://localhost:3000", "http://localhost:8080"]
    allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
    allowed_headers: ["*"]
    max_age: 86400
  csp: ""
  encryption:
    key: dev-encryption-key-for-development-only-32-chars
    algorithm: AES-256-GCM

features:
  enable_metrics: true
  enable_tracing: false
  enable_profiling: true
  enable_swagger: true
  enable_health_check: true
  maintenance_mode: false

services:
  email_service:
    enabled: false
    url: ""
    api_key: ""
    timeout: 10s
    retries: 1
    circuit_breaker:
      enabled: false
      failure_threshold: 0
      recovery_timeout: 0s
      half_open_requests: 0
  payment_service:
    enabled: false
    url: ""
    api_key: ""
    timeout: 15s
    retries: 1
    circuit_breaker:
      enabled: false
      failure_threshold: 0
      recovery_timeout: 0s
      half_open_requests: 0
  storage_service:
    enabled: false
    url: ""
    api_key: ""
    timeout: 30s
    retries: 1
    circuit_breaker:
      enabled: false
      failure_threshold: 0
      recovery_timeout: 0s
      half_open_requests: 0
  notification_service:
    enabled: false
    url: ""
    api_key: ""
    timeout: 5s
    retries: 1
    circuit_breaker:
      enabled: false
      failure_threshold: 0
      recovery_timeout: 0s
      half_open_requests: 0

monitoring:
  metrics:
    enabled: true
    endpoint: /metrics
    namespace: myapp_dev
    interval: 30s
  tracing:
    enabled: false
    service_name: myapp-dev
    endpoint: ""
    sample_rate: 1.0
  health:
    enabled: true
    endpoint: /health`
}

func demonstrateValidationScenarios() {
	scenarios := []struct {
		name   string
		config string
		format string
	}{
		{
			name:   "Missing Required Fields",
			format: "JSON",
			config: `{
				"name": "",
				"version": "1.0.0"
			}`,
		},
		{
			name:   "Invalid Port Range",
			format: "JSON",
			config: `{
				"name": "TestApp",
				"version": "1.0.0",
				"environment": "test",
				"server": {
					"host": "localhost",
					"port": 99999
				}
			}`,
		},
		{
			name:   "Weak Password",
			format: "YAML",
			config: `
name: TestApp
version: 1.0.0
environment: test
server:
  host: localhost
  port: 8080
database:
  driver: postgres
  host: localhost
  port: 5432
  username: user
  password: "123"
  database: test
  ssl_mode: disable`,
		},
	}

	for _, scenario := range scenarios {
		fmt.Printf("   🧪 %s (%s)\n", scenario.name, scenario.format)

		_, err := model.ParseInto[ApplicationConfig]([]byte(scenario.config))
		if err != nil {
			fmt.Printf("      ❌ Validation failed (expected): %v\n", err)
		} else {
			fmt.Printf("      ✅ Validation passed (unexpected)\n")
		}
	}
}

func demonstrateEnvironmentOverrides() {
	// Show how environment variables could override config values
	fmt.Println("Environment variables can override configuration values:")
	fmt.Println("   DATABASE_HOST=prod-db.example.com")
	fmt.Println("   DATABASE_PASSWORD=prod-password")
	fmt.Println("   SERVER_PORT=9000")
	fmt.Println("   DEBUG=false")

	// In a real application, you would:
	// 1. Parse the base configuration from file
	// 2. Apply environment variable overrides
	// 3. Re-validate the final configuration

	baseConfig := `{
		"name": "MyApp",
		"version": "1.0.0", 
		"environment": "production",
		"debug": true,
		"server": {
			"host": "localhost",
			"port": 8080
		}
	}`

	fmt.Printf("\n📄 Base configuration:\n")

	// Parse as raw JSON first (no struct validation)
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(baseConfig), &config); err != nil {
		fmt.Printf("   ❌ Parse error: %v\n", err)
		return
	}

	// Simulate environment overrides
	if serverConfig, ok := config["server"].(map[string]interface{}); ok {
		if port := os.Getenv("SERVER_PORT"); port != "" {
			serverConfig["port"] = port
			fmt.Printf("   🔄 SERVER_PORT override: %s\n", port)
		}
	}

	if debug := os.Getenv("DEBUG"); debug != "" {
		config["debug"] = strings.ToLower(debug) == "true"
		fmt.Printf("   🔄 DEBUG override: %s\n", debug)
	}

	fmt.Printf("   ✅ Configuration with overrides ready for final validation\n")
}
