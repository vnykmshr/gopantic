// Package model provides JSON/YAML parsing with type coercion and validation.
//
// Features: automatic format detection, type coercion, validation, and caching.
//
// # Usage
//
//	type User struct {
//	    ID    int    `json:"id" validate:"required,min=1"`
//	    Name  string `json:"name" validate:"required"`
//	    Email string `json:"email" validate:"email"`
//	}
//
//	user, err := model.ParseInto[User](data)
//
// # Type Coercion
//
// Converts between compatible types: "42"→42, "true"→true, Unix/RFC3339→time.Time
//
// # Validation
//
// Supported validators: required, min, max, length, email, alpha, alphanum
//
// # Performance Caching
//
//	parser := model.NewCachedParser[User](nil)
//	user, _ := parser.Parse(data) // 5-27x faster for repeated inputs
//
// See examples/ directory and docs/ for complete documentation.
package model
