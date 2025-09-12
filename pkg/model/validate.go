package model

import (
	"reflect"
	"strconv"
	"strings"
)

// Validator represents a validation rule that can be applied to a field
type Validator interface {
	// Validate checks if the value is valid according to this validator's rules
	Validate(fieldName string, value interface{}) error
	// Name returns the name of this validator (e.g., "required", "min", "email")
	Name() string
}

// ValidationRule represents a single validation rule parsed from struct tags
type ValidationRule struct {
	Name       string                 // Name of the validator (e.g., "min")
	Validator  Validator              // The validator instance
	Parameters map[string]interface{} // Parameters for the validator (e.g., {"value": 5})
}

// FieldValidation contains all validation rules for a single struct field
type FieldValidation struct {
	FieldName string           // Name of the struct field
	JSONKey   string           // JSON key for this field
	Rules     []ValidationRule // List of validation rules to apply
}

// StructValidation contains validation information for an entire struct
type StructValidation struct {
	Fields []FieldValidation // Validation rules for each field
}

// ValidatorRegistry manages the collection of available validators
type ValidatorRegistry struct {
	validators map[string]func(params map[string]interface{}) Validator
}

// NewValidatorRegistry creates a new validator registry with built-in validators
func NewValidatorRegistry() *ValidatorRegistry {
	registry := &ValidatorRegistry{
		validators: make(map[string]func(params map[string]interface{}) Validator),
	}

	// Register built-in validators
	registry.Register("required", func(params map[string]interface{}) Validator {
		return &RequiredValidator{}
	})

	registry.Register("min", func(params map[string]interface{}) Validator {
		if val, ok := params["value"]; ok {
			if minVal, err := toFloat64(val); err == nil {
				return &MinValidator{Min: minVal}
			}
		}
		return &MinValidator{Min: 0} // Default minimum
	})

	registry.Register("max", func(params map[string]interface{}) Validator {
		if val, ok := params["value"]; ok {
			if maxVal, err := toFloat64(val); err == nil {
				return &MaxValidator{Max: maxVal}
			}
		}
		return &MaxValidator{Max: 0} // Default maximum
	})

	registry.Register("email", func(params map[string]interface{}) Validator {
		return &EmailValidator{}
	})

	registry.Register("length", func(params map[string]interface{}) Validator {
		if val, ok := params["value"]; ok {
			if lengthVal, err := toInt(val); err == nil {
				return &LengthValidator{Length: lengthVal}
			}
		}
		return &LengthValidator{Length: 0} // Default length
	})

	registry.Register("alpha", func(params map[string]interface{}) Validator {
		return &AlphaValidator{}
	})

	registry.Register("alphanum", func(params map[string]interface{}) Validator {
		return &AlphanumValidator{}
	})

	return registry
}

// Register adds a new validator to the registry
func (r *ValidatorRegistry) Register(name string, factory func(params map[string]interface{}) Validator) {
	r.validators[name] = factory
}

// Create creates a validator instance from the registry
func (r *ValidatorRegistry) Create(name string, params map[string]interface{}) Validator {
	if factory, exists := r.validators[name]; exists {
		return factory(params)
	}
	return nil // Unknown validator
}

// Global validator registry instance
var defaultRegistry = NewValidatorRegistry()

// GetDefaultRegistry returns the default global validator registry
func GetDefaultRegistry() *ValidatorRegistry {
	return defaultRegistry
}

// ParseValidationTags parses validation tags from a struct and returns validation info
func ParseValidationTags(structType reflect.Type) *StructValidation {
	validation := &StructValidation{
		Fields: make([]FieldValidation, 0),
	}

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get validation tag
		validateTag := field.Tag.Get("validate")
		if validateTag == "" || validateTag == "-" {
			continue // No validation rules for this field
		}

		// Parse JSON key
		jsonKey := getJSONKeyFromField(field)
		if jsonKey == "-" {
			continue // Field is excluded from JSON
		}

		// Parse validation rules
		rules, err := parseValidationRules(validateTag)
		if err != nil {
			// Skip field with invalid validation syntax
			// TODO: Consider logging this error
			continue
		}

		if len(rules) > 0 {
			fieldValidation := FieldValidation{
				FieldName: field.Name,
				JSONKey:   jsonKey,
				Rules:     rules,
			}
			validation.Fields = append(validation.Fields, fieldValidation)
		}
	}

	return validation
}

// parseValidationRules parses a validation tag string into ValidationRule structs
// Example: "required,min=5,max=100,email" -> []ValidationRule
func parseValidationRules(tag string) ([]ValidationRule, error) {
	rules := make([]ValidationRule, 0)
	registry := GetDefaultRegistry()

	// Split by comma to get individual rules
	ruleParts := strings.Split(tag, ",")

	for _, part := range ruleParts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Parse rule name and parameters
		// Format: "min=5" or "required" or "range=1:10"
		var ruleName string
		params := make(map[string]interface{})

		if equalPos := strings.Index(part, "="); equalPos > 0 {
			// Rule with parameter: "min=5"
			ruleName = part[:equalPos]
			paramValue := part[equalPos+1:]

			// Try to parse parameter as number, fallback to string
			if numVal, err := strconv.ParseFloat(paramValue, 64); err == nil {
				params["value"] = numVal
			} else if intVal, err := strconv.ParseInt(paramValue, 10, 64); err == nil {
				params["value"] = intVal
			} else {
				params["value"] = paramValue
			}
		} else {
			// Simple rule without parameters: "required"
			ruleName = part
		}

		// Create validator instance
		validator := registry.Create(ruleName, params)
		if validator != nil {
			rule := ValidationRule{
				Name:       ruleName,
				Validator:  validator,
				Parameters: params,
			}
			rules = append(rules, rule)
		}
		// TODO: Consider returning error for unknown validators
	}

	return rules, nil
}

// ValidateValue applies validation rules to a single value
func ValidateValue(fieldName string, value interface{}, rules []ValidationRule) error {
	var errors ErrorList

	for _, rule := range rules {
		if err := rule.Validator.Validate(fieldName, value); err != nil {
			errors.Add(err)
		}
	}

	return errors.AsError()
}

// getJSONKeyFromField extracts JSON key from struct field (reused from parse.go logic)
func getJSONKeyFromField(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	if tag == "-" {
		return "-"
	}

	// Split on comma and take first part (the name)
	for i, char := range tag {
		if char == ',' {
			return tag[:i]
		}
	}

	return tag
}

// toFloat64 converts various numeric types to float64 for validation purposes
func toFloat64(value interface{}) (float64, error) {
	switch v := value.(type) {
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case string:
		return strconv.ParseFloat(v, 64)
	default:
		return 0, NewParseError("", value, "float64", "cannot convert to numeric value for validation")
	}
}

// toInt converts various numeric types to int for validation purposes
func toInt(value interface{}) (int, error) {
	switch v := value.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		if v > 9223372036854775807 { // max int64
			return 0, NewParseError("", value, "int", "value too large for int")
		}
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		if v > 9223372036854775807 { // max int64
			return 0, NewParseError("", value, "int", "value too large for int")
		}
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, NewParseError("", value, "int", "cannot convert to integer value for validation")
	}
}
