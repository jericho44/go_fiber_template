package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		options  *SanitizationOptions
		expected string
	}{
		{
			name:     "Default sanitization",
			input:    "<script>alert('xss')</script>Hello World!",
			options:  DefaultSanitizationOptions(),
			expected: "Hello World!",
		},
		{
			name:     "Strict sanitization with length limit",
			input:    "This is a very long string that should be truncated",
			options:  &SanitizationOptions{MaxLength: 10, TrimWhitespace: true},
			expected: "This is a ",
		},
		{
			name:     "Lenient sanitization",
			input:    "<div>Hello World!</div>",
			options:  LenientSanitizationOptions(),
			expected: "<div>Hello World!</div>",
		},
		{
			name:     "Remove SQL injection patterns",
			input:    "Hello'; DROP TABLE users; --",
			options:  DefaultSanitizationOptions(),
			expected: "Hello&#39;",
		},
		{
			name:     "Remove null bytes",
			input:    "Hello\x00World",
			options:  DefaultSanitizationOptions(),
			expected: "HelloWorld",
		},
		{
			name:     "Filter by charset - latin only",
			input:    "Hello123!@#世界",
			options:  &SanitizationOptions{AllowedCharsets: []string{"latin"}},
			expected: "Hello",
		},
		{
			name:     "Filter by charset - latin and numeric",
			input:    "Hello123!@#世界",
			options:  &SanitizationOptions{AllowedCharsets: []string{"latin", "numeric"}},
			expected: "Hello123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input, tt.options)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSanitizeMap(t *testing.T) {
	input := map[string]interface{}{
		"<script>key</script>": "<script>alert('xss')</script>value",
		"nested": map[string]interface{}{
			"inner": "<div>content</div>",
		},
		"array": []interface{}{
			"<script>item1</script>",
			"item2",
		},
		"number": 123,
	}

	options := DefaultSanitizationOptions()
	result := SanitizeMap(input, options)

	// Check that script tags are removed from keys and values
	assert.Contains(t, result, "key")
	assert.NotContains(t, result, "<script>key</script>")

	// Check string value sanitization
	if val, ok := result["key"].(string); ok {
		assert.Equal(t, "value", val)
	}

	// Check nested map sanitization
	if nested, ok := result["nested"].(map[string]interface{}); ok {
		if inner, ok := nested["inner"].(string); ok {
			assert.Equal(t, "&lt;div&gt;content&lt;/div&gt;", inner)
		}
	}

	// Check array sanitization
	if arr, ok := result["array"].([]interface{}); ok {
		assert.Equal(t, "item1", arr[0])
		assert.Equal(t, "item2", arr[1])
	}

	// Check that non-string values are preserved
	assert.Equal(t, 123, result["number"])
}

func TestSanitizeSlice(t *testing.T) {
	input := []string{
		"<script>alert('xss')</script>hello",
		"normal string",
		"<div>content</div>",
	}

	options := DefaultSanitizationOptions()
	result := SanitizeSlice(input, options)

	expected := []string{
		"hello",
		"normal string",
		"&lt;div&gt;content&lt;/div&gt;",
	}

	assert.Equal(t, expected, result)
}

func TestRemoveControlCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "String with control characters",
			input:    "Hello\x01\x02World\x03",
			expected: "HelloWorld",
		},
		{
			name:     "String with allowed whitespace",
			input:    "Hello\t\n\r World",
			expected: "Hello\t\n\r World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeControlCharacters(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveScriptTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple script tag",
			input:    "<script>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "Script tag with attributes",
			input:    "<script type='text/javascript'>alert('xss')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "Case insensitive",
			input:    "<SCRIPT>alert('xss')</SCRIPT>Hello",
			expected: "Hello",
		},
		{
			name:     "Noscript tag",
			input:    "<noscript>No script content</noscript>Hello",
			expected: "Hello",
		},
		{
			name:     "No script tags",
			input:    "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeScriptTags(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveSQLInjectionPatterns(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "UNION SELECT attack",
			input:    "Hello' UNION SELECT * FROM users --",
			expected: "Hello'  * FROM users ",
		},
		{
			name:     "DROP TABLE attack",
			input:    "Hello'; DROP TABLE users; --",
			expected: "Hello'; ; ",
		},
		{
			name:     "OR 1=1 attack",
			input:    "Hello' OR 1=1 --",
			expected: "Hello'  ",
		},
		{
			name:     "Comment attack",
			input:    "Hello /* comment */ World",
			expected: "Hello  World",
		},
		{
			name:     "Clean string",
			input:    "Hello World",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeSQLInjectionPatterns(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFilterByCharsets(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		allowedCharsets []string
		expected        string
	}{
		{
			name:            "Latin only",
			input:           "Hello123!@#世界",
			allowedCharsets: []string{"latin"},
			expected:        "Hello",
		},
		{
			name:            "Latin and numeric",
			input:           "Hello123!@#世界",
			allowedCharsets: []string{"latin", "numeric"},
			expected:        "Hello123",
		},
		{
			name:            "Latin, numeric, and punctuation",
			input:           "Hello123!@#世界",
			allowedCharsets: []string{"latin", "numeric", "punctuation"},
			expected:        "Hello123!@#",
		},
		{
			name:            "All charsets",
			input:           "Hello 123!",
			allowedCharsets: []string{"latin", "numeric", "space", "punctuation"},
			expected:        "Hello 123!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterByCharsets(tt.input, tt.allowedCharsets)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFileExtension(t *testing.T) {
	tests := []struct {
		name              string
		filename          string
		allowedExtensions []string
		expected          bool
	}{
		{
			name:              "Allowed extension",
			filename:          "test.txt",
			allowedExtensions: []string{"txt", "pdf", "jpg"},
			expected:          true,
		},
		{
			name:              "Disallowed extension",
			filename:          "test.exe",
			allowedExtensions: []string{"txt", "pdf", "jpg"},
			expected:          false,
		},
		{
			name:              "No extension",
			filename:          "test",
			allowedExtensions: []string{"txt", "pdf", "jpg"},
			expected:          false,
		},
		{
			name:              "Case insensitive",
			filename:          "test.TXT",
			allowedExtensions: []string{"txt", "pdf", "jpg"},
			expected:          true,
		},
		{
			name:              "No restrictions",
			filename:          "test.anything",
			allowedExtensions: []string{},
			expected:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateFileExtension(tt.filename, tt.allowedExtensions)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name        string
		filename    string
		expectError bool
	}{
		{
			name:        "Valid filename",
			filename:    "test.txt",
			expectError: false,
		},
		{
			name:        "Empty filename",
			filename:    "",
			expectError: true,
		},
		{
			name:        "Path traversal",
			filename:    "../test.txt",
			expectError: true,
		},
		{
			name:        "Absolute path",
			filename:    "/etc/passwd",
			expectError: true,
		},
		{
			name:        "Windows drive letter",
			filename:    "C:\\test.txt",
			expectError: true,
		},
		{
			name:        "Reserved Windows name",
			filename:    "CON.txt",
			expectError: true,
		},
		{
			name:        "Invalid character",
			filename:    "test<file>.txt",
			expectError: true,
		},
		{
			name:        "Too long filename",
			filename:    string(make([]byte, 300)),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileName(tt.filename)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeFileName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid filename",
			input:    "test.txt",
			expected: "test.txt",
		},
		{
			name:     "Path traversal",
			input:    "../test.txt",
			expected: "__test.txt",
		},
		{
			name:     "Invalid characters",
			input:    "test<file>name.txt",
			expected: "test_file_name.txt",
		},
		{
			name:     "Empty filename",
			input:    "",
			expected: "unnamed_file",
		},
		{
			name:     "Whitespace and dots",
			input:    "  test.txt  ",
			expected: "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFileName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsValidMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{
			name:     "Valid MIME type",
			mimeType: "text/plain",
			expected: true,
		},
		{
			name:     "Valid MIME type with subtype",
			mimeType: "application/json",
			expected: true,
		},
		{
			name:     "Invalid MIME type - no slash",
			mimeType: "textplain",
			expected: false,
		},
		{
			name:     "Invalid MIME type - empty parts",
			mimeType: "/plain",
			expected: false,
		},
		{
			name:     "Invalid MIME type - multiple slashes",
			mimeType: "text/plain/extra",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSafeMimeType(t *testing.T) {
	tests := []struct {
		name     string
		mimeType string
		expected string
	}{
		{
			name:     "Valid MIME type",
			mimeType: "text/plain",
			expected: "text/plain",
		},
		{
			name:     "Invalid MIME type",
			mimeType: "invalid",
			expected: "application/octet-stream",
		},
		{
			name:     "Empty MIME type",
			mimeType: "",
			expected: "application/octet-stream",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSafeMimeType(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDefaultSanitizationOptions(t *testing.T) {
	options := DefaultSanitizationOptions()

	assert.True(t, options.HTMLEscape)
	assert.True(t, options.RemoveNullBytes)
	assert.True(t, options.RemoveScriptTags)
	assert.True(t, options.RemoveSQLPatterns)
	assert.True(t, options.RemoveControlChars)
	assert.True(t, options.TrimWhitespace)
	assert.Equal(t, 0, options.MaxLength)
	assert.Empty(t, options.AllowedCharsets)
	assert.Empty(t, options.ForbiddenPatterns)
}

func TestStrictSanitizationOptions(t *testing.T) {
	options := StrictSanitizationOptions()

	assert.True(t, options.HTMLEscape)
	assert.True(t, options.RemoveNullBytes)
	assert.True(t, options.RemoveScriptTags)
	assert.True(t, options.RemoveSQLPatterns)
	assert.True(t, options.RemoveControlChars)
	assert.True(t, options.TrimWhitespace)
	assert.Equal(t, 1000, options.MaxLength)
	assert.NotEmpty(t, options.AllowedCharsets)
	assert.NotEmpty(t, options.ForbiddenPatterns)
}

func TestLenientSanitizationOptions(t *testing.T) {
	options := LenientSanitizationOptions()

	assert.False(t, options.HTMLEscape)
	assert.True(t, options.RemoveNullBytes)
	assert.False(t, options.RemoveScriptTags)
	assert.False(t, options.RemoveSQLPatterns)
	assert.True(t, options.RemoveControlChars)
	assert.True(t, options.TrimWhitespace)
	assert.Equal(t, 0, options.MaxLength)
	assert.Empty(t, options.AllowedCharsets)
	assert.Empty(t, options.ForbiddenPatterns)
}
