package model

import (
	"fmt"
	"reflect"
	"time"
)

// MaxInputSize is the default maximum size for input data (10MB).
// Set to 0 to disable size checking. This prevents resource exhaustion
// from maliciously large inputs.
var MaxInputSize int = 10 * 1024 * 1024 // 10MB

// ParseInto parses raw data into a struct of type T with automatic format detection, type coercion, and validation.
// The format is automatically detected (JSON or YAML) based on the content structure.
// This is the main entry point for parsing operations in gopantic.
//
// The function checks input size against MaxInputSize (default 10MB) to prevent resource exhaustion.
// Set MaxInputSize to 0 to disable size checking.
//
// Example:
//
//	type User struct {
//	    ID   int    `json:"id" validate:"required,min=1"`
//	    Name string `json:"name" validate:"required,min=2"`
//	}
//
//	user, err := model.ParseInto[User](jsonData)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseInto[T any](raw []byte) (T, error) {
	// Check input size
	var zero T
	if MaxInputSize > 0 && len(raw) > MaxInputSize {
		return zero, fmt.Errorf("input size %d bytes exceeds maximum allowed size %d bytes", len(raw), MaxInputSize)
	}

	// Auto-detect format and use appropriate parser
	format := DetectFormat(raw)
	return ParseIntoWithFormat[T](raw, format)
}

// ParseIntoWithFormat parses raw data of a specific format into a struct of type T with type coercion and validation.
// Use this when you know the exact format or want to enforce a specific format.
// Supports JSON and YAML formats.
//
// The function checks input size against MaxInputSize (default 10MB) to prevent resource exhaustion.
// Set MaxInputSize to 0 to disable size checking.
//
// Example:
//
//	user, err := model.ParseIntoWithFormat[User](yamlData, model.FormatYAML)
//	if err != nil {
//	    log.Fatal(err)
//	}
func ParseIntoWithFormat[T any](raw []byte, format Format) (T, error) {
	var zero T

	// Check input size
	if MaxInputSize > 0 && len(raw) > MaxInputSize {
		return zero, fmt.Errorf("input size %d bytes exceeds maximum allowed size %d bytes", len(raw), MaxInputSize)
	}

	var errors ErrorList

	// Get the appropriate parser for the format
	parser := GetParser(format)

	// Parse into a generic interface{} structure
	data, err := parser.Parse(raw)
	if err != nil {
		errors.Add(err)
		return zero, errors.AsError()
	}

	// Create new instance of T
	resultValue := reflect.New(reflect.TypeOf(zero)).Elem()
	resultType := resultValue.Type()

	// Handle different target types
	if resultType.Kind() == reflect.Slice || resultType.Kind() == reflect.Array {
		// Handle array/slice parsing
		return parseIntoSlice[T](data, resultType, format)
	}

	// Ensure data is a map for struct parsing
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		errors.Add(fmt.Errorf("cannot parse non-object data into struct"))
		return zero, errors.AsError()
	}

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
		rawValue, exists := dataMap[fieldKey]
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

// parseIntoSlice handles parsing of array/slice data into slice/array types
func parseIntoSlice[T any](data interface{}, resultType reflect.Type, format Format) (T, error) {
	var zero T
	var errors ErrorList

	// Ensure data is an array
	dataSlice, ok := data.([]interface{})
	if !ok {
		errors.Add(fmt.Errorf("cannot parse non-array data into slice/array"))
		return zero, errors.AsError()
	}

	if resultType.Kind() == reflect.Slice {
		// Handle slice parsing
		slice := reflect.MakeSlice(resultType, len(dataSlice), len(dataSlice))

		for i, item := range dataSlice {
			elemValue := slice.Index(i)
			if err := setFieldValue(elemValue, item, fmt.Sprintf("[%d]", i), format); err != nil {
				errors.Add(err)
			}
		}

		if len(errors) > 0 {
			return zero, errors.AsError()
		}

		return slice.Interface().(T), nil
	} else if resultType.Kind() == reflect.Array {
		// Handle array parsing
		arrayLen := resultType.Len()
		if len(dataSlice) != arrayLen {
			errors.Add(fmt.Errorf("array length mismatch: expected %d elements, got %d", arrayLen, len(dataSlice)))
			return zero, errors.AsError()
		}

		array := reflect.New(resultType).Elem()

		for i, item := range dataSlice {
			elemValue := array.Index(i)
			if err := setFieldValue(elemValue, item, fmt.Sprintf("[%d]", i), format); err != nil {
				errors.Add(err)
			}
		}

		if len(errors) > 0 {
			return zero, errors.AsError()
		}

		return array.Interface().(T), nil
	}

	errors.Add(fmt.Errorf("unsupported type: %s", resultType.Kind()))
	return zero, errors.AsError()
}
