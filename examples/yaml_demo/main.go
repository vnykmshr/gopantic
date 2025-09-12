package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vnykmshr/gopantic/pkg/model"
	"gopkg.in/yaml.v3"
)

// YAMLConfig demonstrates a comprehensive configuration structure
type YAMLConfig struct {
	// Database configuration with nested struct
	Database struct {
		Host     string `yaml:"host" validate:"required"`
		Port     int    `yaml:"port" validate:"min=1,max=65535"`
		Username string `yaml:"username" validate:"required"`
		Password string `yaml:"password" validate:"required"`
		SSL      bool   `yaml:"ssl"`
		Timeout  string `yaml:"timeout" validate:"required"`
	} `yaml:"database" validate:"required"`

	// Server configuration
	Server struct {
		Port    int      `yaml:"port" validate:"min=1000,max=65535"`
		Workers int      `yaml:"workers" validate:"min=1,max=100"`
		Hosts   []string `yaml:"hosts"`
		TLS     struct {
			Enabled  bool   `yaml:"enabled"`
			CertFile string `yaml:"cert_file"`
			KeyFile  string `yaml:"key_file"`
		} `yaml:"tls"`
	} `yaml:"server" validate:"required"`

	// Logging configuration
	Logging struct {
		Level    string `yaml:"level" validate:"required,oneof=debug info warn error"`
		File     string `yaml:"file"`
		MaxSize  int    `yaml:"max_size" validate:"min=1"`
		Compress bool   `yaml:"compress"`
	} `yaml:"logging" validate:"required"`

	// Features configuration with various data types
	Features struct {
		Metrics   bool     `yaml:"metrics"`
		Tracing   bool     `yaml:"tracing"`
		RateLimit float64  `yaml:"rate_limit" validate:"min=0"`
		Tags      []string `yaml:"tags"`
	} `yaml:"features"`

	// Metadata
	Version   string    `yaml:"version" validate:"required"`
	CreatedAt time.Time `yaml:"created_at"`
	Debug     bool      `yaml:"debug"`
}

// YAMLUser demonstrates YAML parsing with different tag patterns
type YAMLUser struct {
	ID          int       `json:"id" yaml:"id" validate:"required,min=1"`
	Name        string    `json:"name" yaml:"name" validate:"required,min=2"`
	Email       string    `json:"email" yaml:"email" validate:"required,email"`
	Age         int       `json:"age" yaml:"age" validate:"min=18,max=120"`
	IsActive    bool      `json:"is_active" yaml:"is_active"`
	Preferences []string  `json:"preferences" yaml:"preferences"`
	LastLogin   time.Time `json:"last_login" yaml:"last_login"`
}

func main() {
	fmt.Println("=== YAML Configuration Example ===")

	// Example 1: Complex configuration structure
	yamlConfig := []byte(`
# Application Configuration
version: "1.2.3"
created_at: "2023-12-25T10:30:00Z"
debug: true

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
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/server.crt"
    key_file: "/etc/ssl/private/server.key"

logging:
  level: info
  file: "/var/log/app.log"
  max_size: 100
  compress: true

features:
  metrics: true
  tracing: false
  rate_limit: 100.5
  tags:
    - "production"
    - "v1"
    - "stable"
`)

	fmt.Println("Raw YAML data:")
	var rawData map[string]interface{}
	if err := yaml.Unmarshal(yamlConfig, &rawData); err != nil {
		log.Fatal("YAML unmarshal error:", err)
	}
	fmt.Printf("%+v\n\n", rawData)

	// Test gopantic parsing
	fmt.Println("Gopantic parsed result:")
	config, err := model.ParseInto[YAMLConfig](yamlConfig)
	if err != nil {
		fmt.Printf("Parse error: %v\n\n", err)
	} else {
		fmt.Printf("Config: %+v\n\n", config)
	}

	// Example 2: User data with auto-detection
	fmt.Println("=== Auto-Detection Example ===")

	userData := []byte(`
id: 123
name: "John Doe"  
email: "john@example.com"
age: 30
is_active: true
preferences:
  - "email_notifications"
  - "dark_mode"
  - "auto_save"
last_login: "2023-12-20T15:45:00Z"
`)

	fmt.Printf("Detected format: %d (0=JSON, 1=YAML)\n", model.DetectFormat(userData))

	user, err := model.ParseInto[YAMLUser](userData)
	if err != nil {
		fmt.Printf("User parse error: %v\n", err)
	} else {
		fmt.Printf("User: %+v\n\n", user)
	}

	// Example 3: Compare JSON vs YAML for same data
	fmt.Println("=== Format Comparison ===")

	jsonData := []byte(`{
		"id": 456,
		"name": "Jane Smith",
		"email": "jane@test.com", 
		"age": 28,
		"is_active": false,
		"preferences": ["api_access", "webhooks"],
		"last_login": "2023-12-19T12:00:00Z"
	}`)

	fmt.Printf("JSON detected format: %d\n", model.DetectFormat(jsonData))
	userFromJSON, err := model.ParseInto[YAMLUser](jsonData)
	if err != nil {
		fmt.Printf("JSON parse error: %v\n", err)
	} else {
		fmt.Printf("User from JSON: %+v\n\n", userFromJSON)
	}

	// Example 4: Validation errors
	fmt.Println("=== Validation Error Example ===")

	invalidYAML := []byte(`
id: 0  # Invalid - should be >= 1
name: "A"  # Invalid - should be >= 2 chars
email: "not-an-email"  # Invalid email format
age: 15  # Invalid - should be >= 18
is_active: true
preferences: []
`)

	_, err = model.ParseInto[YAMLUser](invalidYAML)
	if err != nil {
		fmt.Printf("Expected validation errors: %v\n", err)
	}
}
