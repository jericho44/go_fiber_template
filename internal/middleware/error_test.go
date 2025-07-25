package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAppWithErrorHandler() *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(),
	})
	return app
}

func TestErrorHandler_AppError(t *testing.T) {
	app := setupTestAppWithErrorHandler()

	app.Get("/test", func(c *fiber.Ctx) error {
		return utils.NewNotFoundError("User")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response utils.APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Request failed", response.Message)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "User not found", response.Error.Message)
}

func TestErrorHandler_FiberError(t *testing.T) {
	app := setupTestAppWithErrorHandler()

	app.Get("/test", func(c *fiber.Ctx) error {
		return fiber.NewError(http.StatusBadRequest, "Invalid request")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response utils.APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Request failed", response.Message)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, "Invalid request", response.Error.Message)
}

func TestErrorHandler_UnknownError(t *testing.T) {
	app := setupTestAppWithErrorHandler()

	app.Get("/test", func(c *fiber.Ctx) error {
		return errors.New("unknown error")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response utils.APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "Request failed", response.Message)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	assert.Equal(t, "An unexpected error occurred", response.Error.Message)
}

func TestGetErrorCodeFromStatus(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		expected   string
	}{
		{
			name:       "bad request",
			statusCode: http.StatusBadRequest,
			expected:   "BAD_REQUEST",
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			expected:   "UNAUTHORIZED",
		},
		{
			name:       "forbidden",
			statusCode: http.StatusForbidden,
			expected:   "FORBIDDEN",
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			expected:   "NOT_FOUND",
		},
		{
			name:       "method not allowed",
			statusCode: http.StatusMethodNotAllowed,
			expected:   "METHOD_NOT_ALLOWED",
		},
		{
			name:       "conflict",
			statusCode: http.StatusConflict,
			expected:   "CONFLICT",
		},
		{
			name:       "unprocessable entity",
			statusCode: http.StatusUnprocessableEntity,
			expected:   "UNPROCESSABLE_ENTITY",
		},
		{
			name:       "too many requests",
			statusCode: http.StatusTooManyRequests,
			expected:   "TOO_MANY_REQUESTS",
		},
		{
			name:       "internal server error",
			statusCode: http.StatusInternalServerError,
			expected:   "INTERNAL_SERVER_ERROR",
		},
		{
			name:       "bad gateway",
			statusCode: http.StatusBadGateway,
			expected:   "BAD_GATEWAY",
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			expected:   "SERVICE_UNAVAILABLE",
		},
		{
			name:       "gateway timeout",
			statusCode: http.StatusGatewayTimeout,
			expected:   "GATEWAY_TIMEOUT",
		},
		{
			name:       "unknown status",
			statusCode: 999,
			expected:   "UNKNOWN_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorCodeFromStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRecoverMiddleware(t *testing.T) {
	app := fiber.New()
	app.Use(RecoverMiddleware())

	app.Get("/panic", func(c *fiber.Ctx) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	// The middleware should recover from panic and return a response
	// Note: The exact status code might vary depending on how the middleware handles the panic
	assert.True(t, resp.StatusCode >= 400)
}

func TestRecoverMiddleware_NoPanic(t *testing.T) {
	app := fiber.New()
	app.Use(RecoverMiddleware())

	app.Get("/normal", func(c *fiber.Ctx) error {
		return c.JSON(map[string]string{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/normal", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response map[string]string
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["message"])
}
