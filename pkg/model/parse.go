package model

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

// ParseInto parses raw JSON into a struct of type T with coercion and validation
func ParseInto[T any](raw []byte) (T, error) {
	var zero T
	var errors ErrorList

	// First, unmarshal into a generic map
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		errors.Add(fmt.Errorf("json parse error: %w", err))
		return zero, errors.AsError()
	}

	// Create new instance of T
	resultValue := reflect.New(reflect.TypeOf(zero)).Elem()
	resultType := resultValue.Type()

	// Parse validation rules for this struct type
	validation := ParseValidationTags(resultType)

	// Process each field in the struct
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		fieldValue := resultValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get JSON key from tag, fallback to field name
		jsonKey := getJSONKey(field)
		if jsonKey == "-" {
			continue // Skip fields with json:"-"
		}

		// Get value from data map
		rawValue, exists := data[jsonKey]
		if !exists {
			// Field not present in JSON, leave as zero value
			rawValue = nil
		}

		// Coerce and set the value
		if err := setFieldValue(fieldValue, rawValue, field.Name); err != nil {
			errors.Add(err)
			continue // Skip validation if coercion failed
		}

		// Apply validation rules
		if err := validateFieldValue(field.Name, jsonKey, fieldValue.Interface(), validation); err != nil {
			errors.Add(err)
		}
	}

	if errors.HasErrors() {
		return zero, errors.AsError()
	}

	return resultValue.Interface().(T), nil
}

// setFieldValue coerces and sets a value on a struct field
func setFieldValue(fieldValue reflect.Value, rawValue interface{}, fieldName string) error {
	fieldType := fieldValue.Type()
	fieldKind := fieldType.Kind()

	// Handle direct assignment for matching types first
	if rawValue != nil && reflect.TypeOf(rawValue).AssignableTo(fieldType) {
		fieldValue.Set(reflect.ValueOf(rawValue))
		return nil
	}

	// Handle specific types that need special treatment
	if fieldType == reflect.TypeOf(time.Time{}) {
		coercedValue, err := CoerceValue(rawValue, fieldType, fieldName)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(coercedValue))
		return nil
	}

	// Use coercion for basic type conversion
	coercedValue, err := CoerceValue(rawValue, fieldType, fieldName)
	if err != nil {
		return err
	}

	// Set the coerced value based on the field kind
	switch fieldKind {
	case reflect.String:
		fieldValue.SetString(coercedValue.(string))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fieldValue.SetInt(coercedValue.(int64))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fieldValue.SetUint(coercedValue.(uint64))
	case reflect.Float32, reflect.Float64:
		fieldValue.SetFloat(coercedValue.(float64))
	case reflect.Bool:
		fieldValue.SetBool(coercedValue.(bool))
	case reflect.Slice, reflect.Array:
		fieldValue.Set(reflect.ValueOf(coercedValue))
	default:
		return NewParseError(fieldName, rawValue, fieldType.String(),
			fmt.Sprintf("unsupported field type: %s", fieldType))
	}

	return nil
}

// getJSONKey extracts the JSON key from struct field tags
func getJSONKey(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" {
		return field.Name
	}

	// Handle json tag options like "name,omitempty"
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

// validateFieldValue applies validation rules to a field value
func validateFieldValue(fieldName, jsonKey string, value interface{}, validation *StructValidation) error {
	// Find validation rules for this field
	for _, fieldValidation := range validation.Fields {
		if fieldValidation.FieldName == fieldName || fieldValidation.JSONKey == jsonKey {
			// Apply all validation rules for this field
			return ValidateValue(fieldName, value, fieldValidation.Rules)
		}
	}

	// No validation rules found for this field
	return nil
}
