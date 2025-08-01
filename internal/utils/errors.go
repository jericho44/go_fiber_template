package utils

import (
	"errors"
	"fmt"
	"net/http"
)

// AppError represents a custom application error
type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	StatusCode int               `json:"-"`
	Details    map[string]string `json:"details,omitempty"`
	Err        error             `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// NewAppError creates a new application error
func NewAppError(code, message string, statusCode int, details map[string]string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}
}

// NewAppErrorWithCause creates a new application error with an underlying cause
func NewAppErrorWithCause(code, message string, statusCode int, cause error, details map[string]string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
		Err:        cause,
	}
}

// Predefined error constructors

// NewValidationError creates a validation error
func NewValidationError(details map[string]string) *AppError {
	return NewAppError("VALIDATION_ERROR", "Validation failed", http.StatusBadRequest, details)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s not found", resource), http.StatusNotFound, nil)
}

// NewUnauthorizedError creates an unauthorized error
func NewUnauthorizedError(message string) *AppError {
	if message == "" {
		message = "Authentication required"
	}
	return NewAppError("UNAUTHORIZED", message, http.StatusUnauthorized, nil)
}

// NewForbiddenError creates a forbidden error
func NewForbiddenError(message string) *AppError {
	if message == "" {
		message = "Access denied"
	}
	return NewAppError("FORBIDDEN", message, http.StatusForbidden, nil)
}

// NewConflictError creates a conflict error
func NewConflictError(message string) *AppError {
	if message == "" {
		message = "Resource conflict"
	}
	return NewAppError("CONFLICT", message, http.StatusConflict, nil)
}

// NewInternalServerError creates an internal server error
func NewInternalServerError(message string, cause error) *AppError {
	if message == "" {
		message = "Internal server error"
	}
	return NewAppErrorWithCause("INTERNAL_SERVER_ERROR", message, http.StatusInternalServerError, cause, nil)
}

// NewBadRequestError creates a bad request error
func NewBadRequestError(message string, details map[string]string) *AppError {
	if message == "" {
		message = "Bad request"
	}
	return NewAppError("BAD_REQUEST", message, http.StatusBadRequest, details)
}

// NewDatabaseError creates a database-related error
func NewDatabaseError(message string, cause error) *AppError {
	if message == "" {
		message = "Database operation failed"
	}
	return NewAppErrorWithCause("DATABASE_ERROR", message, http.StatusInternalServerError, cause, nil)
}

// NewAuthenticationError creates an authentication error
func NewAuthenticationError(message string) *AppError {
	if message == "" {
		message = "Invalid credentials"
	}
	return NewAppError("AUTHENTICATION_ERROR", message, http.StatusUnauthorized, nil)
}

// NewTokenError creates a token-related error
func NewTokenError(message string) *AppError {
	if message == "" {
		message = "Invalid or expired token"
	}
	return NewAppError("TOKEN_ERROR", message, http.StatusUnauthorized, nil)
}

// Error type checking helpers

// IsAppError checks if an error is an AppError
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to AppError if possible
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetStatusCode extracts status code from error, defaults to 500
func GetStatusCode(err error) int {
	if appErr, ok := AsAppError(err); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}

// GetErrorCode extracts error code from error, defaults to INTERNAL_SERVER_ERROR
func GetErrorCode(err error) string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return "INTERNAL_SERVER_ERROR"
}

// GetErrorDetails extracts error details from error
func GetErrorDetails(err error) map[string]string {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Details
	}
	return nil
}

// GetIntCodeFromString converts string error codes to integer codes for ErrorResponse
func GetIntCodeFromString(code string) int {
	switch code {
	case "BAD_REQUEST":
		return CodeBadRequest
	case "UNAUTHORIZED":
		return CodeUnauthorized
	case "FORBIDDEN":
		return CodeForbidden
	case "NOT_FOUND":
		return CodeNotFound
	case "VALIDATION_ERROR":
		return CodeValidationError
	case "CONFLICT":
		return CodeBadRequest // Using 400 for conflict as it's a client error
	case "TOO_MANY_REQUESTS":
		return CodeTooManyRequests
	case "INTERNAL_SERVER_ERROR":
		return CodeInternalServerError
	default:
		return CodeInternalServerError
	}
}
