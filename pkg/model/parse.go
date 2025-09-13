package model

import (
	"fmt"
	"reflect"
	"time"
)

// ParseInto parses raw data (JSON by default) into a struct of type T with coercion and validation
func ParseInto[T any](raw []byte) (T, error) {
	// Auto-detect format and use appropriate parser
	format := DetectFormat(raw)
	return ParseIntoWithFormat[T](raw, format)
}

// ParseIntoWithFormat parses raw data of a specific format into a struct of type T
func ParseIntoWithFormat[T any](raw []byte, format Format) (T, error) {
	var zero T
	var errors ErrorList

	// Get the appropriate parser for the format
	parser := GetParser(format)

	// Parse into a generic map structure
	data, err := parser.Parse(raw)
	if err != nil {
		errors.Add(err)
		return zero, errors.AsError()
	}

	// Create new instance of T
	resultValue := reflect.New(reflect.TypeOf(zero)).Elem()
	resultType := resultValue.Type()

	// Parse validation rules for this struct type
	validation := ParseValidationTags(resultType)

	// Process each field in the struct (parsing and coercion pass)
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		fieldValue := resultValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field key from appropriate tag (json or yaml), fallback to field name
		fieldKey := getFieldKey(field, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Get value from data map
		rawValue, exists := data[fieldKey]
		if !exists {
			// Field not present in data, leave as zero value
			rawValue = nil
		}

		// Coerce and set the value
		if err := setFieldValue(fieldValue, rawValue, field.Name, format); err != nil {
			errors.Add(err)
		}
	}

	// Validation pass - now that all fields are parsed, we can do cross-field validation
	for i := 0; i < resultType.NumField(); i++ {
		field := resultType.Field(i)
		fieldValue := resultValue.Field(i)

		// Skip unexported fields
		if !fieldValue.CanSet() {
			continue
		}

		// Get field key from appropriate tag (json or yaml), fallback to field name
		fieldKey := getFieldKey(field, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Apply validation rules (including cross-field validators)
		if err := validateFieldValueWithStruct(field.Name, fieldKey, fieldValue.Interface(), validation, resultValue); err != nil {
			errors.Add(err)
		}
	}

	if errors.HasErrors() {
		return zero, errors.AsError()
	}

	return resultValue.Interface().(T), nil
}

// setFieldValue coerces and sets a value on a struct field
func setFieldValue(fieldValue reflect.Value, rawValue interface{}, fieldName string, format Format) error {
	fieldType := fieldValue.Type()
	fieldKind := fieldType.Kind()

	// Handle direct assignment for matching types first
	if rawValue != nil && reflect.TypeOf(rawValue).AssignableTo(fieldType) {
		fieldValue.Set(reflect.ValueOf(rawValue))
		return nil
	}

	// Handle specific types that need special treatment
	if fieldType == reflect.TypeOf(time.Time{}) {
		coercedValue, err := CoerceValueWithFormat(rawValue, fieldType, fieldName, format)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(coercedValue))
		return nil
	}

	// Use coercion for basic type conversion
	coercedValue, err := CoerceValueWithFormat(rawValue, fieldType, fieldName, format)
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
	case reflect.Struct:
		fieldValue.Set(reflect.ValueOf(coercedValue))
	case reflect.Ptr:
		fieldValue.Set(reflect.ValueOf(coercedValue))
	default:
		return NewParseError(fieldName, rawValue, fieldType.String(),
			fmt.Sprintf("unsupported field type: %s", fieldType))
	}

	return nil
}

// getFieldKey extracts the appropriate field key based on the data format
func getFieldKey(field reflect.StructField, format Format) string {
	var tagName string

	// Determine which tag to use based on format
	switch format {
	case FormatYAML:
		tagName = "yaml"
	default:
		tagName = "json"
	}

	tag := field.Tag.Get(tagName)
	if tag == "" {
		// Fallback to json tag if yaml tag is not present
		if tagName == "yaml" {
			tag = field.Tag.Get("json")
		}

		// If still empty, use field name
		if tag == "" {
			return field.Name
		}
	}

	// Handle tag options like "name,omitempty"
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

func validateFieldValueWithStruct(fieldName, jsonKey string, value interface{}, validation *StructValidation, structValue reflect.Value) error {
	// Find validation rules for this field
	for _, fieldValidation := range validation.Fields {
		if fieldValidation.FieldName == fieldName || fieldValidation.JSONKey == jsonKey {
			// Apply all validation rules for this field (including cross-field validators)
			return ValidateValueWithStruct(fieldName, value, fieldValidation.Rules, structValue)
		}
	}

	// No validation rules found for this field
	return nil
}
