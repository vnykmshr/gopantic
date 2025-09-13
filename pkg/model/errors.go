// Package model provides core parsing and validation functionality for gopantic.
// It includes type coercion, error handling, and the main ParseInto function
// for converting JSON data into typed Go structs.
package model

import (
	"encoding/json"
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
	Field     string
	FieldPath string // Full field path for nested structures (e.g., "User.Address.Street")
	Value     interface{}
	Rule      string
	Message   string
	Details   map[string]interface{} // Additional structured information
}

func (e ValidationError) Error() string {
	fieldName := e.Field
	if e.FieldPath != "" {
		fieldName = e.FieldPath
	}

	if fieldName != "" {
		return fmt.Sprintf("validation error on field %q: %s", fieldName, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new ValidationError
func NewValidationError(field string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Field:     field,
		FieldPath: field, // For backward compatibility
		Value:     value,
		Rule:      rule,
		Message:   message,
		Details:   make(map[string]interface{}),
	}
}

// NewValidationErrorWithPath creates a new ValidationError with explicit field path
func NewValidationErrorWithPath(field, fieldPath string, value interface{}, rule, message string) *ValidationError {
	return &ValidationError{
		Field:     field,
		FieldPath: fieldPath,
		Value:     value,
		Rule:      rule,
		Message:   message,
		Details:   make(map[string]interface{}),
	}
}

// NewValidationErrorWithDetails creates a new ValidationError with additional structured details
func NewValidationErrorWithDetails(field, fieldPath string, value interface{}, rule, message string, details map[string]interface{}) *ValidationError {
	if details == nil {
		details = make(map[string]interface{})
	}

	return &ValidationError{
		Field:     field,
		FieldPath: fieldPath,
		Value:     value,
		Rule:      rule,
		Message:   message,
		Details:   details,
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
// If the error is itself an ErrorList, it flattens the errors to avoid nesting
func (el *ErrorList) Add(err error) {
	if err != nil {
		// Check if the error is another ErrorList and flatten it
		if nestedErrorList, ok := err.(ErrorList); ok {
			*el = append(*el, nestedErrorList...)
		} else {
			*el = append(*el, err)
		}
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

// ValidationErrors returns only the ValidationError instances from the ErrorList
func (el ErrorList) ValidationErrors() []*ValidationError {
	var validationErrors []*ValidationError
	for _, err := range el {
		if validationErr, ok := err.(*ValidationError); ok {
			validationErrors = append(validationErrors, validationErr)
		}
	}
	return validationErrors
}

// GroupByField groups validation errors by field path
func (el ErrorList) GroupByField() map[string][]*ValidationError {
	groups := make(map[string][]*ValidationError)
	for _, err := range el {
		if validationErr, ok := err.(*ValidationError); ok {
			fieldPath := validationErr.FieldPath
			if fieldPath == "" {
				fieldPath = validationErr.Field
			}
			groups[fieldPath] = append(groups[fieldPath], validationErr)
		}
	}
	return groups
}

// StructuredErrorReport represents a structured validation error report for JSON serialization
type StructuredErrorReport struct {
	Errors []FieldError `json:"errors"`
	Count  int          `json:"count"`
}

// FieldError represents a single field's validation errors
type FieldError struct {
	Field     string                `json:"field"`
	FieldPath string                `json:"field_path"`
	Value     interface{}           `json:"value,omitempty"`
	Errors    []ValidationErrorInfo `json:"validation_errors"`
}

// ValidationErrorInfo represents detailed information about a validation error
type ValidationErrorInfo struct {
	Rule    string                 `json:"rule"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToStructuredReport converts an ErrorList to a structured error report for JSON serialization
func (el ErrorList) ToStructuredReport() *StructuredErrorReport {
	fieldGroups := el.GroupByField()
	fieldErrors := make([]FieldError, 0, len(fieldGroups))

	for fieldPath, validationErrors := range fieldGroups {
		var errorInfos []ValidationErrorInfo
		var field string
		var value interface{}

		for _, validationErr := range validationErrors {
			errorInfos = append(errorInfos, ValidationErrorInfo{
				Rule:    validationErr.Rule,
				Message: validationErr.Message,
				Details: validationErr.Details,
			})

			// Use the first error's field and value info
			if field == "" {
				field = validationErr.Field
				value = validationErr.Value
			}
		}

		fieldErrors = append(fieldErrors, FieldError{
			Field:     field,
			FieldPath: fieldPath,
			Value:     value,
			Errors:    errorInfos,
		})
	}

	return &StructuredErrorReport{
		Errors: fieldErrors,
		Count:  len(fieldErrors),
	}
}

// ToJSON converts an ErrorList to JSON for API responses
func (el ErrorList) ToJSON() ([]byte, error) {
	report := el.ToStructuredReport()
	return json.Marshal(report)
}
