package middleware

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLoggingMiddleware_Development(t *testing.T) {
	// Capture stdout since Fiber logger writes to stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{
		Env: config.Development,
	}

	app := fiber.New()
	app.Use(RequestLoggingMiddleware(cfg))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	logOutput := buf.String()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output (development format should be human-readable)
	assert.Contains(t, logOutput, "200")
	assert.Contains(t, logOutput, "GET")
	assert.Contains(t, logOutput, "/test")
}

func TestRequestLoggingMiddleware_Production(t *testing.T) {
	// Capture stdout since Fiber logger writes to stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cfg := &config.Config{
		Env: config.Production,
	}

	app := fiber.New()
	app.Use(RequestLoggingMiddleware(cfg))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "test-agent")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Restore stdout and read captured output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	logOutput := buf.String()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output (production format should be JSON-like)
	assert.Contains(t, logOutput, `"status":"200"`)
	assert.Contains(t, logOutput, `"method":"GET"`)
	assert.Contains(t, logOutput, `"path":"/test"`)
	assert.Contains(t, logOutput, `"user_agent":"test-agent"`)
}

func TestCustomRequestLoggingMiddleware(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New()
	app.Use(CustomRequestLoggingMiddleware())

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "custom-test-agent")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, "GET /test")
	assert.Contains(t, logOutput, "Status: 200")
	assert.Contains(t, logOutput, "User: anonymous")
	assert.Contains(t, logOutput, "UA: custom-test-agent")
}

func TestCustomRequestLoggingMiddleware_WithAuthenticatedUser(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New()

	// Add middleware that sets user context
	app.Use(func(c *fiber.Ctx) error {
		// Mock authenticated user
		user := &models.User{ID: 123}
		c.Locals(UserContextKey, user)
		return c.Next()
	})

	app.Use(CustomRequestLoggingMiddleware())

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, "User: user_123")
}

func TestCustomRequestLoggingMiddleware_WithError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Custom error handler to ensure status codes are set properly
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(CustomRequestLoggingMiddleware())

	// Add test route that returns an error
	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusInternalServerError, "test error")
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Status: 500")
	assert.Contains(t, logOutput, "Request error: test error")
}

func TestStructuredLoggingMiddleware(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New()
	app.Use(StructuredLoggingMiddleware())

	// Add test route
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request with body
	body := strings.NewReader(`{"test": "data"}`)
	req, err := http.NewRequest("POST", "/test", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "structured-test-agent")

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, "REQUEST_LOG:")
	assert.Contains(t, logOutput, `"method":"POST"`)
	assert.Contains(t, logOutput, `"path":"/test"`)
	assert.Contains(t, logOutput, `"status_code":200`)
	assert.Contains(t, logOutput, `"user_agent":"structured-test-agent"`)
	assert.Contains(t, logOutput, `"latency_ms":`)
	assert.Contains(t, logOutput, `"response_size":`)
}

func TestStructuredLoggingMiddleware_WithAuthenticatedUser(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New()

	// Add middleware that sets user context
	app.Use(func(c *fiber.Ctx) error {
		// Mock authenticated user
		user := &models.User{ID: 456}
		c.Locals(UserContextKey, user)
		return c.Next()
	})

	app.Use(StructuredLoggingMiddleware())

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"user_id":456`)
}

func TestStructuredLoggingMiddleware_WithError(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Custom error handler to ensure status codes are set properly
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})
	app.Use(StructuredLoggingMiddleware())

	// Add test route that returns an error
	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(fiber.StatusBadRequest, "structured test error")
	})

	// Create request
	req, err := http.NewRequest("GET", "/test", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Check log output
	logOutput := buf.String()
	assert.Contains(t, logOutput, `"status_code":400`)
	assert.Contains(t, logOutput, `"error":"structured test error"`)
}

func TestLoggingMiddleware_LatencyMeasurement(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	app := fiber.New()
	app.Use(CustomRequestLoggingMiddleware())

	// Add test route with artificial delay
	app.Get("/slow", func(c *fiber.Ctx) error {
		time.Sleep(10 * time.Millisecond)
		return c.JSON(fiber.Map{"message": "slow response"})
	})

	// Create request
	req, err := http.NewRequest("GET", "/slow", nil)
	require.NoError(t, err)

	// Execute request
	resp, err := app.Test(req, 1000) // 1 second timeout
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check response
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check log output contains latency information
	logOutput := buf.String()
	assert.Contains(t, logOutput, "Latency:")
	// Should contain some measurable latency (at least a few milliseconds)
	assert.Contains(t, logOutput, "ms")
}

// Helper function to capture and restore log output
func captureLogOutput(fn func()) string {
	var buf bytes.Buffer
	originalOutput := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(originalOutput)

	fn()

	return buf.String()
}

func TestLoggingMiddleware_Integration(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		userAgent      string
	}{
		{
			name:           "GET request",
			method:         "GET",
			path:           "/api/test",
			expectedStatus: 200,
			userAgent:      "test-client/1.0",
		},
		{
			name:           "POST request",
			method:         "POST",
			path:           "/api/create",
			expectedStatus: 201,
			userAgent:      "api-client/2.0",
		},
		{
			name:           "PUT request",
			method:         "PUT",
			path:           "/api/update/123",
			expectedStatus: 200,
			userAgent:      "update-client/1.5",
		},
		{
			name:           "DELETE request",
			method:         "DELETE",
			path:           "/api/delete/456",
			expectedStatus: 204,
			userAgent:      "delete-client/1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logOutput := captureLogOutput(func() {
				app := fiber.New()
				app.Use(CustomRequestLoggingMiddleware())

				// Add routes for different methods
				app.Get("/api/test", func(c *fiber.Ctx) error {
					return c.JSON(fiber.Map{"message": "test"})
				})
				app.Post("/api/create", func(c *fiber.Ctx) error {
					return c.Status(201).JSON(fiber.Map{"message": "created"})
				})
				app.Put("/api/update/:id", func(c *fiber.Ctx) error {
					return c.JSON(fiber.Map{"message": "updated"})
				})
				app.Delete("/api/delete/:id", func(c *fiber.Ctx) error {
					return c.SendStatus(204)
				})

				// Create request
				req, err := http.NewRequest(tt.method, tt.path, nil)
				require.NoError(t, err)
				req.Header.Set("User-Agent", tt.userAgent)

				// Execute request
				resp, err := app.Test(req)
				require.NoError(t, err)
				defer resp.Body.Close()

				// Check response
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			})

			// Check log output
			assert.Contains(t, logOutput, tt.method)
			assert.Contains(t, logOutput, tt.path)
			assert.Contains(t, logOutput, tt.userAgent)
		})
	}
}
