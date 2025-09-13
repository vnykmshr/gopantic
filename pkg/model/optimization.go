package model

import (
	"reflect"
	"sync"
)

// FieldInfo caches parsed struct field information to avoid repeated reflection operations
type FieldInfo struct {
	Name        string // Field name
	JSONKey     string // JSON tag key
	YAMLKey     string // YAML tag key
	ValidateTag string // Validation tag
	CanSet      bool   // Whether the field can be set
	Type        reflect.Type
	Index       int // Field index in struct
}

// StructInfo caches parsed struct information
type StructInfo struct {
	Type         reflect.Type
	Fields       []FieldInfo
	FieldsByJSON map[string]*FieldInfo // Map from JSON key to field info
	FieldsByYAML map[string]*FieldInfo // Map from YAML key to field info
	FieldsByName map[string]*FieldInfo // Map from field name to field info
}

// structInfoCache is a global cache for parsed struct information
var (
	structInfoCache = make(map[reflect.Type]*StructInfo)
	structInfoMutex sync.RWMutex
)

// GetStructInfo returns cached struct information, parsing it if necessary
func GetStructInfo(t reflect.Type) *StructInfo {
	structInfoMutex.RLock()
	if info, exists := structInfoCache[t]; exists {
		structInfoMutex.RUnlock()
		return info
	}
	structInfoMutex.RUnlock()

	// Need to parse struct info
	structInfoMutex.Lock()
	defer structInfoMutex.Unlock()

	// Double-check after acquiring write lock
	if info, exists := structInfoCache[t]; exists {
		return info
	}

	info := parseStructInfo(t)
	structInfoCache[t] = info
	return info
}

// parseStructInfo parses struct field information once and caches it
func parseStructInfo(t reflect.Type) *StructInfo {
	numField := t.NumField()

	info := &StructInfo{
		Type:         t,
		Fields:       make([]FieldInfo, 0, numField),
		FieldsByJSON: make(map[string]*FieldInfo),
		FieldsByYAML: make(map[string]*FieldInfo),
		FieldsByName: make(map[string]*FieldInfo),
	}

	for i := 0; i < numField; i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		fieldInfo := FieldInfo{
			Name:        field.Name,
			JSONKey:     parseJSONTag(field),
			YAMLKey:     parseYAMLTag(field),
			ValidateTag: field.Tag.Get("validate"),
			CanSet:      true, // All exported fields can be set
			Type:        field.Type,
			Index:       i,
		}

		// Skip fields marked with "-"
		if fieldInfo.JSONKey == "-" && fieldInfo.YAMLKey == "-" {
			continue
		}

		info.Fields = append(info.Fields, fieldInfo)

		// Index by different keys
		fieldPtr := &info.Fields[len(info.Fields)-1]
		info.FieldsByName[fieldInfo.Name] = fieldPtr

		if fieldInfo.JSONKey != "-" {
			info.FieldsByJSON[fieldInfo.JSONKey] = fieldPtr
		}

		if fieldInfo.YAMLKey != "-" {
			info.FieldsByYAML[fieldInfo.YAMLKey] = fieldPtr
		}
	}

	return info
}

// parseJSONTag parses the JSON struct tag and returns the field key
func parseJSONTag(field reflect.StructField) string {
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

// parseYAMLTag parses the YAML struct tag and returns the field key
func parseYAMLTag(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" {
		// Fallback to JSON tag, then field name
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" && jsonTag != "-" {
			// Parse json tag the same way
			for i, char := range jsonTag {
				if char == ',' {
					return jsonTag[:i]
				}
			}
			return jsonTag
		}
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

// ClearStructInfoCache clears the struct info cache - useful for testing
func ClearStructInfoCache() {
	structInfoMutex.Lock()
	defer structInfoMutex.Unlock()

	for k := range structInfoCache {
		delete(structInfoCache, k)
	}
}

// GetFieldKeyForFormat returns the appropriate field key for the given format using cached info
func GetFieldKeyForFormat(fieldInfo *FieldInfo, format Format) string {
	switch format {
	case FormatYAML:
		return fieldInfo.YAMLKey
	default:
		return fieldInfo.JSONKey
	}
}

// OptimizedParseIntoWithFormat is an optimized version that uses cached struct info
func OptimizedParseIntoWithFormat[T any](raw []byte, format Format) (T, error) {
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

	// Get struct info (cached)
	resultType := reflect.TypeOf(zero)
	structInfo := GetStructInfo(resultType)

	// Create new instance of T
	resultValue := reflect.New(resultType).Elem()

	// Parse validation rules for this struct type (cached)
	validation := ParseValidationTags(resultType)

	// Process each field using cached info (parsing and coercion pass)
	for i := range structInfo.Fields {
		fieldInfo := &structInfo.Fields[i]
		fieldValue := resultValue.Field(fieldInfo.Index)

		// Get field key for the format
		fieldKey := GetFieldKeyForFormat(fieldInfo, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Get value from data map
		rawValue, exists := data[fieldKey]
		if !exists {
			rawValue = nil
		}

		// Coerce and set the value
		if err := setFieldValue(fieldValue, rawValue, fieldInfo.Name, format); err != nil {
			errors.Add(err)
		}
	}

	// Validation pass - now that all fields are parsed, we can do cross-field validation
	for i := range structInfo.Fields {
		fieldInfo := &structInfo.Fields[i]
		fieldValue := resultValue.Field(fieldInfo.Index)

		// Get field key for the format
		fieldKey := GetFieldKeyForFormat(fieldInfo, format)
		if fieldKey == "-" {
			continue // Skip fields with tag:"-"
		}

		// Apply validation rules (including cross-field validators)
		if err := validateFieldValueWithStruct(fieldInfo.Name, fieldKey, fieldValue.Interface(), validation, resultValue); err != nil {
			errors.Add(err)
		}
	}

	if errors.HasErrors() {
		return zero, errors.AsError()
	}

	return resultValue.Interface().(T), nil
}
