// Package model provides core parsing and validation functionality for gopantic.
// It includes type coercion, error handling, and the main ParseInto function
// for converting JSON data into typed Go structs.
package model

import (
	"fmt"
	"strings"
)

// ParseError represents an error that occurred during parsing
type ParseError struct {
	Field   string
	Value   interface{}
	Type    string
	Message string
}

func (e ParseError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("parse error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// NewParseError creates a new ParseError
func NewParseError(field string, value interface{}, targetType, message string) *ParseError {
	return &ParseError{
		Field:   field,
		Value:   value,
		Type:    targetType,
		Message: message,
	}
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Value   interface{}
	Rule    string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
	}
}

// ErrorList represents a collection of errors that can occur during parsing/validation
type ErrorList []error

func (el ErrorList) Error() string {
	if len(el) == 0 {
		return ""
	}
	if len(el) == 1 {
		return el[0].Error()
	}

	messages := make([]string, 0, len(el))
	for _, err := range el {
		messages = append(messages, err.Error())
	}
	return fmt.Sprintf("multiple errors: %s", strings.Join(messages, "; "))
}

// Add appends an error to the ErrorList
func (el *ErrorList) Add(err error) {
	if err != nil {
		*el = append(*el, err)
	}
}

// HasErrors returns true if the ErrorList contains any errors
func (el ErrorList) HasErrors() bool {
	return len(el) > 0
}

// AsError returns the ErrorList as an error if it contains any errors, nil otherwise
func (el ErrorList) AsError() error {
	if el.HasErrors() {
		return el
	}
	return nil
}
