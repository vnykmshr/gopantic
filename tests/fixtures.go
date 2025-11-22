// Package tests provides comprehensive test coverage for gopantic's parsing, validation, and caching functionality.
package tests

import (
	"time"
)

// Basic test structures (shared across multiple test files)

// User represents a basic user for simple parsing tests
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

// Product represents a product with pricing for parsing tests
type Product struct {
	ID      uint64  `json:"id"`
	Name    string  `json:"name"`
	Price   float64 `json:"price"`
	InStock bool    `json:"in_stock"`
}

// Event represents an event with timestamps for time parsing tests
type Event struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	StartTime time.Time `json:"start_time"`
	CreatedAt time.Time `json:"created_at"`
}

// Config represents configuration with nested structures
type Config struct {
	Port     int      `json:"port"`
	Enabled  bool     `json:"enabled"`
	Features []string `json:"features"`
	DB       Database `json:"database"`
}

// Database represents database connection settings
type Database struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Comprehensive E2E test structures (from integration tests)

// E2EUser represents a complete user with validation for end-to-end testing
type E2EUser struct {
	ID        int       `json:"id" validate:"required,min=1"`
	Username  string    `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email     string    `json:"email" validate:"required,email"`
	FirstName string    `json:"first_name" validate:"required,alpha"`
	LastName  string    `json:"last_name" validate:"required,alpha"`
	Age       int       `json:"age" validate:"min=13,max=120"`
	IsActive  bool      `json:"is_active"`
	Profile   Profile   `json:"profile" validate:"required"`
	Settings  *Settings `json:"settings"`
	CreatedAt time.Time `json:"created_at"`
}

// Profile represents a user profile with bio and skills
type Profile struct {
	Bio       string   `json:"bio" validate:"max=500"`
	Website   *string  `json:"website"`
	Location  string   `json:"location"`
	Skills    []string `json:"skills"`
	Languages []string `json:"languages" validate:"required"`
}

// Settings represents user preferences and privacy settings
type Settings struct {
	Theme         string                 `json:"theme" validate:"required"`
	Notifications map[string]interface{} `json:"notifications"`
	Privacy       PrivacySettings        `json:"privacy" validate:"required"`
}

// PrivacySettings represents privacy preferences for a user
type PrivacySettings struct {
	ProfileVisible bool `json:"profile_visible"`
	EmailVisible   bool `json:"email_visible"`
	ShowOnline     bool `json:"show_online"`
}

// APIResponse represents a generic API response wrapper with metadata
type APIResponse[T any] struct {
	Success   bool      `json:"success"`
	Data      *T        `json:"data"`
	Error     *APIError `json:"error"`
	Meta      Meta      `json:"meta"`
	Timestamp time.Time `json:"timestamp"`
}

// APIError represents an error response from an API
type APIError struct {
	Code    string                 `json:"code" validate:"required"`
	Message string                 `json:"message" validate:"required"`
	Details map[string]interface{} `json:"details"`
}

// Meta represents metadata for API responses
type Meta struct {
	RequestID   string `json:"request_id" validate:"required"`
	Version     string `json:"version" validate:"required"`
	ProcessTime int    `json:"process_time_ms" validate:"min=0"`
}

// DatabaseConfig represents database connection configuration
type DatabaseConfig struct {
	Host     string `json:"host" validate:"required"`
	Port     int    `json:"port" validate:"required,min=1,max=65535"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=8"`
	Database string `json:"database" validate:"required"`
	SSL      bool   `json:"ssl"`
	Timeout  int    `json:"timeout" validate:"min=1000"`
}

// AppConfig represents application configuration with database settings
type AppConfig struct {
	Name        string         `json:"name" validate:"required"`
	Version     string         `json:"version" validate:"required"`
	Environment string         `json:"environment" validate:"required"`
	Database    DatabaseConfig `json:"database" validate:"required"`
	Debug       bool           `json:"debug"`
	Features    []string       `json:"features"`
}
