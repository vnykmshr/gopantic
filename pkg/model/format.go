package model

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Format represents the input data format
type Format int

const (
	// FormatJSON represents JSON format
	FormatJSON Format = iota
	// FormatYAML represents YAML format
	FormatYAML
)

// FormatParser defines the interface for parsing different data formats
type FormatParser interface {
	// Parse parses raw bytes into a generic map structure
	Parse(raw []byte) (map[string]interface{}, error)
	// Format returns the format type this parser handles
	Format() Format
}

// JSONParser implements FormatParser for JSON format
type JSONParser struct{}

// Parse parses JSON data into a generic map
func (jp *JSONParser) Parse(raw []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("json parse error: %w", err)
	}
	return data, nil
}

// Format returns the JSON format type
func (jp *JSONParser) Format() Format {
	return FormatJSON
}

// YAMLParser implements FormatParser for YAML format
type YAMLParser struct{}

// Parse parses YAML data into a generic map
func (yp *YAMLParser) Parse(raw []byte) (map[string]interface{}, error) {
	var data map[string]interface{}
	if err := yaml.Unmarshal(raw, &data); err != nil {
		return nil, fmt.Errorf("yaml parse error: %w", err)
	}
	return data, nil
}

// Format returns the YAML format type
func (yp *YAMLParser) Format() Format {
	return FormatYAML
}

// DetectFormat attempts to detect the data format from raw bytes
func DetectFormat(raw []byte) Format {
	// Try to detect based on content characteristics
	if len(raw) == 0 {
		return FormatJSON // Default to JSON for empty input
	}

	// Trim whitespace and look at first non-whitespace character
	for i := 0; i < len(raw); i++ {
		switch raw[i] {
		case ' ', '\t', '\n', '\r':
			continue
		case '{', '[':
			return FormatJSON
		default:
			// Check for common YAML indicators
			content := string(raw)
			// YAML typically has key: value pairs without quotes around keys
			// or starts with --- document separator
			if containsYAMLPatterns(content) {
				return FormatYAML
			}
			return FormatJSON // Default to JSON if unsure
		}
	}

	return FormatJSON // Default to JSON
}

// containsYAMLPatterns checks for common YAML patterns
func containsYAMLPatterns(content string) bool {
	return hasYAMLDocumentSeparator(content) ||
		hasYAMLKeyValuePatterns(content) ||
		hasYAMLListPatterns(content)
}

// hasYAMLDocumentSeparator checks for YAML document separator
func hasYAMLDocumentSeparator(content string) bool {
	return len(content) >= 3 && content[:3] == "---"
}

// hasYAMLKeyValuePatterns checks for unquoted key-value patterns
func hasYAMLKeyValuePatterns(content string) bool {
	lines := 0
	yamlLines := 0

	for _, line := range splitLines(content) {
		if line == "" || lines >= 5 {
			continue
		}
		lines++

		if hasUnquotedKeyValue(line) {
			yamlLines++
		}
	}

	return lines > 1 && yamlLines >= lines/2
}

// hasYAMLListPatterns checks for YAML list indicators
func hasYAMLListPatterns(content string) bool {
	for _, line := range splitLines(content) {
		trimmed := trimLeadingSpace(line)
		if len(trimmed) >= 2 && trimmed[0] == '-' && trimmed[1] == ' ' {
			return true
		}
	}
	return false
}

// hasUnquotedKeyValue checks if a line has unquoted key:value pattern
func hasUnquotedKeyValue(line string) bool {
	for i := 0; i < len(line)-1; i++ {
		if line[i] == ':' && (i == 0 || line[i-1] != '"') &&
			(i == len(line)-1 || line[i+1] == ' ' || line[i+1] == '\t') {
			return true
		}
	}
	return false
}

// splitLines splits content by newlines, limiting to first 5 lines
func splitLines(content string) []string {
	var lines []string
	start := 0

	for i, char := range content {
		if char == '\n' || len(lines) >= 5 {
			if start <= i {
				lines = append(lines, content[start:i])
			}
			start = i + 1
			if len(lines) >= 5 {
				break
			}
		}
	}

	// Add last line if exists
	if start < len(content) && len(lines) < 5 {
		lines = append(lines, content[start:])
	}

	return lines
}

// trimLeadingSpace removes leading spaces and tabs
func trimLeadingSpace(s string) string {
	i := 0
	for i < len(s) && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	return s[i:]
}

// GetParser returns the appropriate parser for the given format
func GetParser(format Format) FormatParser {
	switch format {
	case FormatYAML:
		return &YAMLParser{}
	default:
		return &JSONParser{}
	}
}
