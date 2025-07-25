package utils

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp() *fiber.App {
	return fiber.New()
}

func TestSuccessResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessResponse(c, "Operation successful", map[string]string{"key": "value"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "Operation successful", response.Message)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
	assert.Nil(t, response.Meta)
}

func TestSuccessResponseWithMeta(t *testing.T) {
	app := setupTestApp()

	meta := &Meta{
		Page:       1,
		Limit:      10,
		Total:      100,
		TotalPages: 10,
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return SuccessResponseWithMeta(c, "Data retrieved", []string{"item1", "item2"}, meta)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, CodeSuccess, response.Code)
	assert.Equal(t, "Data retrieved", response.Message)
	assert.NotNil(t, response.Data)
	assert.Nil(t, response.Error)
	assert.NotNil(t, response.Meta)
	assert.Equal(t, 1, response.Meta.Page)
	assert.Equal(t, 10, response.Meta.Limit)
	assert.Equal(t, int64(100), response.Meta.Total)
	assert.Equal(t, 10, response.Meta.TotalPages)
}

func TestCreatedResponse(t *testing.T) {
	app := setupTestApp()

	app.Post("/test", func(c *fiber.Ctx) error {
		return CreatedResponse(c, "Resource created", map[string]interface{}{"id": 1})
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.True(t, response.Success)
	assert.Equal(t, CodeCreated, response.Code)
	assert.Equal(t, "Resource created", response.Message)
	assert.NotNil(t, response.Data)
}

func TestErrorResponse(t *testing.T) {
	app := setupTestApp()

	details := map[string]string{
		"field1": "error message 1",
		"field2": "error message 2",
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		return ErrorResponse(c, http.StatusBadRequest, CodeBadRequest, "Test error message", details)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeBadRequest, response.Code)
	assert.Equal(t, "Request failed", response.Message)
	assert.Nil(t, response.Data)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, "Test error message", response.Error.Message)
	assert.Equal(t, details, response.Error.Details)
}

func TestBadRequestResponse(t *testing.T) {
	app := setupTestApp()

	details := map[string]string{"email": "Invalid email format"}

	app.Get("/test", func(c *fiber.Ctx) error {
		return BadRequestResponse(c, "Invalid input", details)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeBadRequest, response.Code)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "BAD_REQUEST", response.Error.Code)
	assert.Equal(t, "Invalid input", response.Error.Message)
	assert.Equal(t, details, response.Error.Details)
}

func TestUnauthorizedResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return UnauthorizedResponse(c, "Authentication required")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeUnauthorized, response.Code)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
	assert.Equal(t, "Authentication required", response.Error.Message)
	assert.Nil(t, response.Error.Details)
}

func TestForbiddenResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return ForbiddenResponse(c, "Access denied")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeForbidden, response.Code)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "FORBIDDEN", response.Error.Code)
	assert.Equal(t, "Access denied", response.Error.Message)
}

func TestNotFoundResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return NotFoundResponse(c, "Resource not found")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "NOT_FOUND", response.Error.Code)
	assert.Equal(t, "Resource not found", response.Error.Message)
}

func TestInternalServerErrorResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return InternalServerErrorResponse(c, "Something went wrong")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "INTERNAL_SERVER_ERROR", response.Error.Code)
	assert.Equal(t, "Something went wrong", response.Error.Message)
}

func TestValidationErrorResponse(t *testing.T) {
	app := setupTestApp()

	details := map[string]string{
		"email":    "Email is required",
		"password": "Password must be at least 8 characters",
	}

	app.Post("/test", func(c *fiber.Ctx) error {
		return ValidationErrorResponse(c, details)
	})

	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeValidationError, response.Code)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	assert.Equal(t, "Validation failed", response.Error.Message)
	assert.Equal(t, details, response.Error.Details)
}

func TestTooManyRequestsResponse(t *testing.T) {
	app := setupTestApp()

	app.Get("/test", func(c *fiber.Ctx) error {
		return TooManyRequestsResponse(c, "Rate limit exceeded")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var response APIResponse
	err = json.Unmarshal(body, &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, CodeTooManyRequests, response.Code)
	assert.NotNil(t, response.Error)
	assert.Equal(t, "TOO_MANY_REQUESTS", response.Error.Code)
	assert.Equal(t, "Rate limit exceeded", response.Error.Message)
}

func TestGetStringCode(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{"Bad Request", CodeBadRequest, "BAD_REQUEST"},
		{"Unauthorized", CodeUnauthorized, "UNAUTHORIZED"},
		{"Forbidden", CodeForbidden, "FORBIDDEN"},
		{"Not Found", CodeNotFound, "NOT_FOUND"},
		{"Validation Error", CodeValidationError, "VALIDATION_ERROR"},
		{"Too Many Requests", CodeTooManyRequests, "TOO_MANY_REQUESTS"},
		{"Internal Server Error", CodeInternalServerError, "INTERNAL_SERVER_ERROR"},
		{"Unknown Code", 999, "UNKNOWN_ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStringCode(tt.code)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponseCodeConstants(t *testing.T) {
	// Test that our constants match expected HTTP status codes
	assert.Equal(t, 200, CodeSuccess)
	assert.Equal(t, 201, CodeCreated)
	assert.Equal(t, 400, CodeBadRequest)
	assert.Equal(t, 401, CodeUnauthorized)
	assert.Equal(t, 403, CodeForbidden)
	assert.Equal(t, 404, CodeNotFound)
	assert.Equal(t, 422, CodeValidationError)
	assert.Equal(t, 429, CodeTooManyRequests)
	assert.Equal(t, 500, CodeInternalServerError)
}
