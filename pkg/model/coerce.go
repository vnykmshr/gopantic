package model

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// CoerceValue attempts to coerce a value to the target type
func CoerceValue(value interface{}, targetType reflect.Type, fieldName string) (interface{}, error) {
	if value == nil {
		return getZeroValueForType(targetType), nil
	}

	// Handle specific struct types first
	if targetType == reflect.TypeOf(time.Time{}) {
		return coerceToTime(value, fieldName)
	}

	// Fall back to kind-based coercion
	targetKind := targetType.Kind()
	switch targetKind {
	case reflect.String:
		return coerceToString(value, fieldName)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return coerceToInt(value, fieldName)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return coerceToUint(value, fieldName)
	case reflect.Float32, reflect.Float64:
		return coerceToFloat(value, targetKind, fieldName)
	case reflect.Bool:
		return coerceToBool(value, fieldName)
	default:
		return nil, NewParseError(fieldName, value, targetType.String(),
			fmt.Sprintf("coercion to %s not supported", targetType))
	}
}

// coerceToString converts various types to string
func coerceToString(value interface{}, _ string) (string, error) {
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%g", v), nil
	case bool:
		if v {
			return "true", nil
		}
		return "false", nil
	default:
		return fmt.Sprintf("%v", v), nil
	}
}

// coerceToInt converts various types to int64
func coerceToInt(value interface{}, fieldName string) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		if v > 9223372036854775807 { // max int64
			return 0, NewParseError(fieldName, value, "int64", "value too large for int64")
		}
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		if v > 9223372036854775807 { // max int64
			return 0, NewParseError(fieldName, value, "int64", "value too large for int64")
		}
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, NewParseError(fieldName, value, "int64",
				fmt.Sprintf("cannot parse string %q as integer: %v", v, err))
		}
		return parsed, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, NewParseError(fieldName, value, "int64",
			fmt.Sprintf("cannot coerce %T to int64", value))
	}
}

// coerceToUint converts various types to uint64
func coerceToUint(value interface{}, fieldName string) (uint64, error) {
	switch v := value.(type) {
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case int:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case float32:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, NewParseError(fieldName, value, "uint64", "negative value cannot be coerced to uint64")
		}
		return uint64(v), nil
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, NewParseError(fieldName, value, "uint64",
				fmt.Sprintf("cannot parse string %q as unsigned integer: %v", v, err))
		}
		return parsed, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, NewParseError(fieldName, value, "uint64",
			fmt.Sprintf("cannot coerce %T to uint64", value))
	}
}

// coerceToFloat converts various types to float32/float64
func coerceToFloat(value interface{}, targetKind reflect.Kind, fieldName string) (float64, error) {
	switch v := value.(type) {
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
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
		bitSize := 64
		if targetKind == reflect.Float32 {
			bitSize = 32
		}
		parsed, err := strconv.ParseFloat(v, bitSize)
		if err != nil {
			return 0, NewParseError(fieldName, value, "float64",
				fmt.Sprintf("cannot parse string %q as float: %v", v, err))
		}
		return parsed, nil
	case bool:
		if v {
			return 1.0, nil
		}
		return 0.0, nil
	default:
		return 0, NewParseError(fieldName, value, "float64",
			fmt.Sprintf("cannot coerce %T to float64", value))
	}
}

// coerceToBool converts various types to bool
func coerceToBool(value interface{}, fieldName string) (bool, error) {
	switch v := value.(type) {
	case bool:
		return v, nil
	case string:
		switch v {
		case "true", "True", "TRUE", "1", "yes", "Yes", "YES", "on", "On", "ON":
			return true, nil
		case "false", "False", "FALSE", "0", "no", "No", "NO", "off", "Off", "OFF", "":
			return false, nil
		default:
			return false, NewParseError(fieldName, value, "bool",
				fmt.Sprintf("cannot parse string %q as boolean", v))
		}
	case int, int8, int16, int32, int64:
		return v != 0, nil
	case uint, uint8, uint16, uint32, uint64:
		return v != 0, nil
	case float32, float64:
		return v != 0.0, nil
	default:
		return false, NewParseError(fieldName, value, "bool",
			fmt.Sprintf("cannot coerce %T to bool", value))
	}
}

// coerceToTime converts various types to time.Time
func coerceToTime(value interface{}, fieldName string) (time.Time, error) {
	switch v := value.(type) {
	case time.Time:
		return v, nil
	case string:
		return parseTimeFromString(v, fieldName)
	case int64:
		// Unix timestamp (seconds)
		return time.Unix(v, 0), nil
	case float64:
		// Unix timestamp (seconds, may have fractional part)
		sec := int64(v)
		nsec := int64((v - float64(sec)) * 1e9)
		return time.Unix(sec, nsec), nil
	case int:
		// Unix timestamp (seconds)
		return time.Unix(int64(v), 0), nil
	default:
		return time.Time{}, NewParseError(fieldName, value, "time.Time",
			fmt.Sprintf("cannot coerce %T to time.Time", value))
	}
}

// parseTimeFromString attempts to parse time from string using multiple formats
func parseTimeFromString(s, fieldName string) (time.Time, error) {
	formats := []string{
		time.RFC3339,           // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano,       // "2006-01-02T15:04:05.999999999Z07:00"
		"2006-01-02T15:04:05Z", // ISO 8601 UTC
		"2006-01-02T15:04:05",  // ISO 8601 without timezone
		"2006-01-02 15:04:05",  // Common format
		"2006-01-02",           // Date only
		"15:04:05",             // Time only (today's date)
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, NewParseError(fieldName, s, "time.Time",
		fmt.Sprintf("cannot parse string %q as time.Time using standard formats", s))
}

// getZeroValueForType returns the zero value for the given type
func getZeroValueForType(t reflect.Type) interface{} {
	if t == reflect.TypeOf(time.Time{}) {
		return time.Time{}
	}

	// Fall back to kind-based zero values
	return getZeroValue(t.Kind())
}

// getZeroValue returns the zero value for the given kind
func getZeroValue(kind reflect.Kind) interface{} {
	switch kind {
	case reflect.String:
		return ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return int64(0)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uint64(0)
	case reflect.Float32, reflect.Float64:
		return 0.0
	case reflect.Bool:
		return false
	default:
		return nil
	}
}
