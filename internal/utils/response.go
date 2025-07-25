package utils

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Response codes
const (
	// Success codes
	CodeSuccess = 200
	CodeCreated = 201

	// Client error codes
	CodeBadRequest      = 400
	CodeUnauthorized    = 401
	CodeForbidden       = 403
	CodeNotFound        = 404
	CodeValidationError = 422
	CodeTooManyRequests = 429

	// Server error codes
	CodeInternalServerError = 500
)

// APIResponse represents the standard response structure for all API endpoints
type APIResponse struct {
	Success bool        `json:"success"`
	Code    int         `json:"code"`
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
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	}
	return c.Status(http.StatusOK).JSON(response)
}

// SuccessResponseWithMeta creates a successful API response with pagination metadata
func SuccessResponseWithMeta(c *fiber.Ctx, message string, data interface{}, meta *Meta) error {
	response := APIResponse{
		Success: true,
		Code:    CodeSuccess,
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
		Code:    CodeCreated,
		Message: message,
		Data:    data,
	}
	return c.Status(http.StatusCreated).JSON(response)
}

// ErrorResponse creates an error API response
func ErrorResponse(c *fiber.Ctx, statusCode int, code int, message string, details map[string]string) error {
	response := APIResponse{
		Success: false,
		Code:    code,
		Message: "Request failed",
		Error: &APIError{
			Code:    getStringCode(code),
			Message: message,
			Details: details,
		},
	}
	return c.Status(statusCode).JSON(response)
}

// getStringCode converts integer code to string code for backward compatibility in error details
func getStringCode(code int) string {
	switch code {
	case CodeBadRequest:
		return "BAD_REQUEST"
	case CodeUnauthorized:
		return "UNAUTHORIZED"
	case CodeForbidden:
		return "FORBIDDEN"
	case CodeNotFound:
		return "NOT_FOUND"
	case CodeValidationError:
		return "VALIDATION_ERROR"
	case CodeTooManyRequests:
		return "TOO_MANY_REQUESTS"
	case CodeInternalServerError:
		return "INTERNAL_SERVER_ERROR"
	default:
		return "UNKNOWN_ERROR"
	}
}

// BadRequestResponse creates a 400 Bad Request response
func BadRequestResponse(c *fiber.Ctx, message string, details map[string]string) error {
	return ErrorResponse(c, http.StatusBadRequest, CodeBadRequest, message, details)
}

// UnauthorizedResponse creates a 401 Unauthorized response
func UnauthorizedResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

// ForbiddenResponse creates a 403 Forbidden response
func ForbiddenResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusForbidden, CodeForbidden, message, nil)
}

// NotFoundResponse creates a 404 Not Found response
func NotFoundResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusNotFound, CodeNotFound, message, nil)
}

// InternalServerErrorResponse creates a 500 Internal Server Error response
func InternalServerErrorResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusInternalServerError, CodeInternalServerError, message, nil)
}

// ValidationErrorResponse creates a response for validation errors
func ValidationErrorResponse(c *fiber.Ctx, details map[string]string) error {
	return ErrorResponse(c, http.StatusBadRequest, CodeValidationError, "Validation failed", details)
}

// TooManyRequestsResponse creates a 429 Too Many Requests response
func TooManyRequestsResponse(c *fiber.Ctx, message string) error {
	return ErrorResponse(c, http.StatusTooManyRequests, CodeTooManyRequests, message, nil)
}
