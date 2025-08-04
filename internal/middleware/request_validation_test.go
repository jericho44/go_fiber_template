package middleware

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"strings"
	"testing"

	"go-fiber-template/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestValidationMiddleware_RequestSizeLimit(t *testing.T) {
	// Create test config
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        1024, // 1KB
			MaxFileSize:        2048, // 2KB
			MaxMultipartSize:   4096, // 4KB
			AllowedMimeTypes:   []string{"text/plain", "image/jpeg"},
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)
	app := fiber.New()
	app.Use(middleware.RequestSizeLimit())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	tests := []struct {
		name           string
		bodySize       int
		expectedStatus int
	}{
		{
			name:           "Small request should pass",
			bodySize:       512,
			expectedStatus: 200,
		},
		{
			name:           "Large request should be rejected",
			bodySize:       2048,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.Repeat("a", tt.bodySize)
			req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
			req.Header.Set("Content-Type", "text/plain")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequestValidationMiddleware_FileUploadSizeLimit(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        10240, // 10KB
			MaxFileSize:        1024,  // 1KB
			MaxMultipartSize:   5120,  // 5KB
			AllowedMimeTypes:   []string{"text/plain", "image/jpeg"},
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)
	app := fiber.New()
	app.Use(middleware.FileUploadSizeLimit())
	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	tests := []struct {
		name           string
		fileSize       int
		expectedStatus int
	}{
		{
			name:           "Small file should pass",
			fileSize:       512,
			expectedStatus: 200,
		},
		{
			name:           "Large file should be rejected",
			fileSize:       2048,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create multipart form
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Create file field
			fileContent := strings.Repeat("a", tt.fileSize)
			part, err := writer.CreateFormFile("file", "test.txt")
			require.NoError(t, err)
			_, err = part.Write([]byte(fileContent))
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/upload", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequestValidationMiddleware_ContentTypeValidation(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        10240,
			MaxFileSize:        5120,
			MaxMultipartSize:   10240,
			AllowedMimeTypes:   []string{"text/plain", "image/jpeg"},
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)
	app := fiber.New()
	app.Use(middleware.ContentTypeValidation())
	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	tests := []struct {
		name           string
		fileContent    string
		filename       string
		expectedStatus int
	}{
		{
			name:           "Text file should pass",
			fileContent:    "Hello, World!",
			filename:       "test.txt",
			expectedStatus: 200,
		},
		{
			name:           "Binary file should be rejected",
			fileContent:    "\x89PNG\r\n\x1a\n", // PNG header but not in allowed types
			filename:       "test.png",
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			part, err := writer.CreateFormFile("file", tt.filename)
			require.NoError(t, err)
			_, err = part.Write([]byte(tt.fileContent))
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/upload", &buf)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
		})
	}
}

func TestRequestValidationMiddleware_RequestSanitization(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        10240,
			MaxFileSize:        5120,
			MaxMultipartSize:   10240,
			AllowedMimeTypes:   []string{"text/plain"},
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)
	app := fiber.New()
	app.Use(middleware.RequestSanitization())
	app.Get("/test", func(c *fiber.Ctx) error {
		// Return the sanitized query parameter
		return c.JSON(fiber.Map{
			"param": c.Query("test"),
		})
	})

	tests := []struct {
		name           string
		queryParam     string
		expectedStatus int
	}{
		{
			name:           "Clean parameter should pass",
			queryParam:     "hello",
			expectedStatus: 200,
		},
		{
			name:           "Parameter with script should be sanitized",
			queryParam:     "<script>alert('xss')</script>hello",
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?test="+tt.queryParam, nil)

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// For successful requests, verify sanitization occurred
			if tt.expectedStatus == 200 {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				// The response should not contain the original script tag
				assert.NotContains(t, string(body), "<script>")
			}
		})
	}
}

func TestRequestValidationMiddleware_CombinedRequestValidation(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			MaxBodySize:        1024,
			MaxFileSize:        512,
			MaxMultipartSize:   2048,
			AllowedMimeTypes:   []string{"text/plain"},
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)
	app := fiber.New()
	app.Use(middleware.CombinedRequestValidation())
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test with oversized request
	body := strings.Repeat("a", 2048)
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "text/plain")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLegacyMiddleware_RequestSizeLimitMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(RequestSizeLimitMiddleware(1024))
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Test with small request
	req := httptest.NewRequest("POST", "/test", strings.NewReader("small"))
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Test with large request
	body := strings.Repeat("a", 2048)
	req = httptest.NewRequest("POST", "/test", strings.NewReader(body))
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLegacyMiddleware_FileUploadLimitMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(FileUploadLimitMiddleware(1024))
	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Create multipart form with large file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	fileContent := strings.Repeat("a", 2048)
	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte(fileContent))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

func TestLegacyMiddleware_ContentTypeValidationMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(ContentTypeValidationMiddleware([]string{"text/plain"}))
	app.Post("/upload", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	// Create multipart form with text file
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", "test.txt")
	require.NoError(t, err)
	_, err = part.Write([]byte("Hello, World!"))
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/upload", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestRequestValidationMiddleware_isMimeTypeAllowed(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			AllowedMimeTypes: []string{"text/plain", "image/*", "application/json"},
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)

	tests := []struct {
		name     string
		mimeType string
		expected bool
	}{
		{
			name:     "Exact match should be allowed",
			mimeType: "text/plain",
			expected: true,
		},
		{
			name:     "Wildcard match should be allowed",
			mimeType: "image/jpeg",
			expected: true,
		},
		{
			name:     "Non-matching type should be rejected",
			mimeType: "video/mp4",
			expected: false,
		},
		{
			name:     "MIME type with parameters should work",
			mimeType: "text/plain; charset=utf-8",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.isMimeTypeAllowed(tt.mimeType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRequestValidationMiddleware_sanitizeString(t *testing.T) {
	cfg := &config.Config{
		Request: config.RequestConfig{
			EnableSanitization: true,
		},
	}

	middleware := NewRequestValidationMiddleware(cfg)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Clean string should remain unchanged",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "Script tags should be removed",
			input:    "<script>alert('xss')</script>hello",
			expected: "hello",
		},
		{
			name:     "HTML should be escaped",
			input:    "<div>content</div>",
			expected: "&lt;div&gt;content&lt;/div&gt;",
		},
		{
			name:     "Null bytes should be removed",
			input:    "hello\x00world",
			expected: "helloworld",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
