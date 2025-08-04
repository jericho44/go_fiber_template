package utils

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

// SanitizationOptions holds configuration for sanitization
type SanitizationOptions struct {
	HTMLEscape         bool     // Whether to HTML escape the input
	RemoveNullBytes    bool     // Whether to remove null bytes
	RemoveScriptTags   bool     // Whether to remove script tags
	RemoveSQLPatterns  bool     // Whether to remove SQL injection patterns
	RemoveControlChars bool     // Whether to remove control characters
	TrimWhitespace     bool     // Whether to trim leading/trailing whitespace
	MaxLength          int      // Maximum allowed length (0 = no limit)
	AllowedCharsets    []string // Allowed character sets (empty = all allowed)
	ForbiddenPatterns  []string // Custom forbidden regex patterns
}

// DefaultSanitizationOptions returns default sanitization options
func DefaultSanitizationOptions() *SanitizationOptions {
	return &SanitizationOptions{
		HTMLEscape:         true,
		RemoveNullBytes:    true,
		RemoveScriptTags:   true,
		RemoveSQLPatterns:  true,
		RemoveControlChars: true,
		TrimWhitespace:     true,
		MaxLength:          0,
		AllowedCharsets:    []string{},
		ForbiddenPatterns:  []string{},
	}
}

// StrictSanitizationOptions returns strict sanitization options
func StrictSanitizationOptions() *SanitizationOptions {
	return &SanitizationOptions{
		HTMLEscape:         true,
		RemoveNullBytes:    true,
		RemoveScriptTags:   true,
		RemoveSQLPatterns:  true,
		RemoveControlChars: true,
		TrimWhitespace:     true,
		MaxLength:          1000,
		AllowedCharsets:    []string{"latin", "numeric", "space", "punctuation"},
		ForbiddenPatterns: []string{
			`(?i)<[^>]*>`,         // HTML tags
			`(?i)javascript:`,     // JavaScript URLs
			`(?i)vbscript:`,       // VBScript URLs
			`(?i)data:`,           // Data URLs
			`(?i)on\w+\s*=`,       // Event handlers
			`(?i)expression\s*\(`, // CSS expressions
			`(?i)@import`,         // CSS imports
			`(?i)\\x[0-9a-f]{2}`,  // Hex encoded characters
			`(?i)\\u[0-9a-f]{4}`,  // Unicode encoded characters
		},
	}
}

// LenientSanitizationOptions returns lenient sanitization options
func LenientSanitizationOptions() *SanitizationOptions {
	return &SanitizationOptions{
		HTMLEscape:         false,
		RemoveNullBytes:    true,
		RemoveScriptTags:   false,
		RemoveSQLPatterns:  false,
		RemoveControlChars: true,
		TrimWhitespace:     true,
		MaxLength:          0,
		AllowedCharsets:    []string{},
		ForbiddenPatterns:  []string{},
	}
}

// SanitizeString sanitizes a string according to the provided options
func SanitizeString(input string, options *SanitizationOptions) string {
	if options == nil {
		options = DefaultSanitizationOptions()
	}

	result := input

	// Apply length limit first
	if options.MaxLength > 0 && len(result) > options.MaxLength {
		result = result[:options.MaxLength]
	}

	// Remove null bytes
	if options.RemoveNullBytes {
		result = strings.ReplaceAll(result, "\x00", "")
	}

	// Remove control characters
	if options.RemoveControlChars {
		result = removeControlCharacters(result)
	}

	// Remove script tags
	if options.RemoveScriptTags {
		result = removeScriptTags(result)
	}

	// Remove SQL injection patterns
	if options.RemoveSQLPatterns {
		result = removeSQLInjectionPatterns(result)
	}

	// Apply custom forbidden patterns
	for _, pattern := range options.ForbiddenPatterns {
		regex := regexp.MustCompile(pattern)
		result = regex.ReplaceAllString(result, "")
	}

	// Filter by allowed character sets
	if len(options.AllowedCharsets) > 0 {
		result = filterByCharsets(result, options.AllowedCharsets)
	}

	// HTML escape
	if options.HTMLEscape {
		result = html.EscapeString(result)
	}

	// Trim whitespace
	if options.TrimWhitespace {
		result = strings.TrimSpace(result)
	}

	return result
}

// SanitizeMap sanitizes all string values in a map
func SanitizeMap(input map[string]interface{}, options *SanitizationOptions) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range input {
		sanitizedKey := SanitizeString(key, options)

		switch v := value.(type) {
		case string:
			result[sanitizedKey] = SanitizeString(v, options)
		case map[string]interface{}:
			result[sanitizedKey] = SanitizeMap(v, options)
		case []interface{}:
			result[sanitizedKey] = sanitizeSlice(v, options)
		default:
			result[sanitizedKey] = v
		}
	}

	return result
}

// SanitizeSlice sanitizes all string values in a slice
func SanitizeSlice(input []string, options *SanitizationOptions) []string {
	result := make([]string, len(input))
	for i, str := range input {
		result[i] = SanitizeString(str, options)
	}
	return result
}

// sanitizeSlice sanitizes all values in an interface slice
func sanitizeSlice(input []interface{}, options *SanitizationOptions) []interface{} {
	result := make([]interface{}, len(input))

	for i, value := range input {
		switch v := value.(type) {
		case string:
			result[i] = SanitizeString(v, options)
		case map[string]interface{}:
			result[i] = SanitizeMap(v, options)
		case []interface{}:
			result[i] = sanitizeSlice(v, options)
		default:
			result[i] = v
		}
	}

	return result
}

// removeControlCharacters removes control characters except for common whitespace
func removeControlCharacters(input string) string {
	var result strings.Builder
	for _, r := range input {
		// Keep common whitespace characters (space, tab, newline, carriage return)
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			result.WriteRune(r)
		} else if !unicode.IsControl(r) {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// removeScriptTags removes script tags and their content
func removeScriptTags(input string) string {
	// Remove script tags (case insensitive)
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	result := scriptRegex.ReplaceAllString(input, "")

	// Remove noscript tags
	noscriptRegex := regexp.MustCompile(`(?i)<noscript[^>]*>.*?</noscript>`)
	result = noscriptRegex.ReplaceAllString(result, "")

	return result
}

// removeSQLInjectionPatterns removes common SQL injection patterns
func removeSQLInjectionPatterns(input string) string {
	patterns := []string{
		`(?i)\bunion\s+select\b`,
		`(?i)\bdrop\s+table\b`,
		`(?i)\bdelete\s+from\b`,
		`(?i)\binsert\s+into\b`,
		`(?i)\bupdate\s+.*\bset\b`,
		`(?i)\bselect\s+.*\bfrom\b`,
		`(?i)\bcreate\s+table\b`,
		`(?i)\balter\s+table\b`,
		`(?i)\btruncate\s+table\b`,
		`(?i)--`,
		`(?i)/\*.*?\*/`,
		`(?i)\bor\s+1\s*=\s*1\b`,
		`(?i)\band\s+1\s*=\s*1\b`,
		`(?i)\bor\s+true\b`,
		`(?i)\band\s+true\b`,
		`(?i)'\s*or\s*'`,
		`(?i)"\s*or\s*"`,
		`(?i);\s*drop\b`,
		`(?i);\s*delete\b`,
		`(?i);\s*update\b`,
		`(?i);\s*insert\b`,
	}

	result := input
	for _, pattern := range patterns {
		regex := regexp.MustCompile(pattern)
		result = regex.ReplaceAllString(result, "")
	}

	return result
}

// filterByCharsets filters input to only allow specified character sets
func filterByCharsets(input string, allowedCharsets []string) string {
	var result strings.Builder

	for _, r := range input {
		allowed := false

		for _, charset := range allowedCharsets {
			switch charset {
			case "latin":
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
					allowed = true
				}
			case "numeric":
				if r >= '0' && r <= '9' {
					allowed = true
				}
			case "space":
				if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
					allowed = true
				}
			case "punctuation":
				if unicode.IsPunct(r) {
					allowed = true
				}
			case "symbol":
				if unicode.IsSymbol(r) {
					allowed = true
				}
			case "unicode":
				if r > 127 {
					allowed = true
				}
			}

			if allowed {
				break
			}
		}

		if allowed {
			result.WriteRune(r)
		}
	}

	return result.String()
}

// ValidateFileExtension validates file extension against allowed extensions
func ValidateFileExtension(filename string, allowedExtensions []string) bool {
	if len(allowedExtensions) == 0 {
		return true // No restrictions
	}

	// Extract file extension
	parts := strings.Split(filename, ".")
	if len(parts) < 2 {
		return false // No extension
	}

	extension := strings.ToLower(parts[len(parts)-1])

	for _, allowed := range allowedExtensions {
		if strings.ToLower(allowed) == extension {
			return true
		}
	}

	return false
}

// ValidateFileName validates filename for security issues
func ValidateFileName(filename string) error {
	if filename == "" {
		return NewAppError("INVALID_FILENAME", "Filename cannot be empty", 400, nil)
	}

	// Check for path traversal attempts
	if strings.Contains(filename, "..") {
		return NewAppError("INVALID_FILENAME", "Filename cannot contain path traversal sequences", 400, nil)
	}

	// Check for absolute paths
	if strings.HasPrefix(filename, "/") || strings.HasPrefix(filename, "\\") {
		return NewAppError("INVALID_FILENAME", "Filename cannot be an absolute path", 400, nil)
	}

	// Check for Windows drive letters
	if len(filename) >= 2 && filename[1] == ':' {
		return NewAppError("INVALID_FILENAME", "Filename cannot contain drive letters", 400, nil)
	}

	// Check for reserved Windows filenames
	reservedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	}

	baseFilename := strings.ToUpper(filename)
	if dotIndex := strings.Index(baseFilename, "."); dotIndex != -1 {
		baseFilename = baseFilename[:dotIndex]
	}

	for _, reserved := range reservedNames {
		if baseFilename == reserved {
			return NewAppError("INVALID_FILENAME", "Filename cannot be a reserved system name", 400, nil)
		}
	}

	// Check for invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return NewAppError("INVALID_FILENAME", "Filename contains invalid characters", 400, nil)
		}
	}

	// Check length
	if len(filename) > 255 {
		return NewAppError("INVALID_FILENAME", "Filename is too long", 400, nil)
	}

	return nil
}

// SanitizeFileName sanitizes a filename for safe storage
func SanitizeFileName(filename string) string {
	// Remove path components
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")

	// Remove invalid characters
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		filename = strings.ReplaceAll(filename, char, "_")
	}

	// Remove control characters
	filename = removeControlCharacters(filename)

	// Trim whitespace and dots
	filename = strings.Trim(filename, " .")

	// Ensure it's not empty
	if filename == "" {
		filename = "unnamed_file"
	}

	// Limit length
	if len(filename) > 255 {
		// Try to preserve extension
		if dotIndex := strings.LastIndex(filename, "."); dotIndex != -1 && dotIndex > 200 {
			extension := filename[dotIndex:]
			filename = filename[:255-len(extension)] + extension
		} else {
			filename = filename[:255]
		}
	}

	return filename
}

// IsValidMimeType checks if a MIME type is valid
func IsValidMimeType(mimeType string) bool {
	// Basic MIME type format validation
	parts := strings.Split(mimeType, "/")
	if len(parts) != 2 {
		return false
	}

	// Check for empty parts
	if parts[0] == "" || parts[1] == "" {
		return false
	}

	// Check for valid characters (basic validation)
	validPattern := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9!#$&\-\^_]*$`)
	return validPattern.MatchString(parts[0]) && validPattern.MatchString(parts[1])
}

// GetSafeMimeType returns a safe MIME type, defaulting to application/octet-stream
func GetSafeMimeType(mimeType string) string {
	if IsValidMimeType(mimeType) {
		return mimeType
	}
	return "application/octet-stream"
}
