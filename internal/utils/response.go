package utils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// APIResponse represents the standard response structure for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// APIError represents error information in API responses
type APIError struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// Meta represents pagination and additional metadata
type Meta struct {
	Page       int   `json:"page,omitempty"`
	Limit      int   `json:"limit,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"total_pages,omitempty"`
}

// SuccessResponse creates a successful API response
func SuccessResponse(c *fiber.Ctx, message string, data interface{}) error {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	return c.Status(http.StatusOK).JSON(response)
}

// SuccessResponseWithMeta creates a successful API response with pagination metadata
func SuccessResponseWithMeta(c *fiber.Ctx, message string, data interface{}, meta *Meta) error {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	}
	return c.Status(http.StatusOK).JSON(response)
}

// CreatedResponse creates a response for successful resource creation
func CreatedResponse(c *fiber.Ctx, message string, data interface{}) error {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

// ErrorResponse creates an error API response
func ErrorResponse(c *fiber.Ctx, statusCode int, code string, message string, details map[string]string) error {
	response := APIResponse{
		Success: false,
		Message: "Request failed",
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	return c.Status(statusCode).JSON(response)
}

// BadRequestResponse creates a 400 Bad Request response
func BadRequestResponse(c *fiber.Ctx, message string, details map[string]string) error {
	return ErrorResponse(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

// UnauthorizedResponse creates a 401 Unauthorized response
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

// ForbiddenResponse creates a 403 Forbidden response
func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

// NotFoundResponse creates a 404 Not Found response
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

// InternalServerErrorResponse creates a 500 Internal Server Error response
func InternalServerErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message, nil)
}

// ValidationErrorResponse creates a response for validation errors
func ValidationErrorResponse(c *fiber.Ctx, details map[string]string) error {
	return ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Validation failed", details)
}

// TooManyRequestsResponse creates a 429 Too Many Requests response
func TooManyRequestsResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", message, nil)
}
