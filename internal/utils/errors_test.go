package utils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		expected string
	}{
		{
			name: "error without cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
			},
			expected: "Test error message",
		},
		{
			name: "error with cause",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
				Err:     errors.New("underlying error"),
			},
			expected: "Test error message: underlying error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.appError.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	appErr := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Err:     underlyingErr,
	}

	result := appErr.Unwrap()
	assert.Equal(t, underlyingErr, result)
}

func TestNewAppError(t *testing.T) {
	details := map[string]string{"field": "error"}
	appErr := NewAppError("TEST_CODE", "Test message", http.StatusBadRequest, details)

	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "Test message", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, details, appErr.Details)
	assert.Nil(t, appErr.Err)
}

func TestNewAppErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	details := map[string]string{"field": "error"}
	appErr := NewAppErrorWithCause("TEST_CODE", "Test message", http.StatusInternalServerError, cause, details)

	assert.Equal(t, "TEST_CODE", appErr.Code)
	assert.Equal(t, "Test message", appErr.Message)
	assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
	assert.Equal(t, details, appErr.Details)
	assert.Equal(t, cause, appErr.Err)
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{
		"email":    "Email is required",
		"password": "Password is too short",
	}

	appErr := NewValidationError(details)

	assert.Equal(t, "VALIDATION_ERROR", appErr.Code)
	assert.Equal(t, "Validation failed", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, details, appErr.Details)
}

func TestNewNotFoundError(t *testing.T) {
	appErr := NewNotFoundError("User")

	assert.Equal(t, "NOT_FOUND", appErr.Code)
	assert.Equal(t, "User not found", appErr.Message)
	assert.Equal(t, http.StatusNotFound, appErr.StatusCode)
	assert.Nil(t, appErr.Details)
}

func TestNewUnauthorizedError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "with custom message",
			message:  "Custom auth message",
			expected: "Custom auth message",
		},
		{
			name:     "with empty message",
			message:  "",
			expected: "Authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewUnauthorizedError(tt.message)

			assert.Equal(t, "UNAUTHORIZED", appErr.Code)
			assert.Equal(t, tt.expected, appErr.Message)
			assert.Equal(t, http.StatusUnauthorized, appErr.StatusCode)
		})
	}
}

func TestNewForbiddenError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "with custom message",
			message:  "Custom forbidden message",
			expected: "Custom forbidden message",
		},
		{
			name:     "with empty message",
			message:  "",
			expected: "Access denied",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			appErr := NewForbiddenError(tt.message)

			assert.Equal(t, "FORBIDDEN", appErr.Code)
			assert.Equal(t, tt.expected, appErr.Message)
			assert.Equal(t, http.StatusForbidden, appErr.StatusCode)
		})
	}
}

func TestNewConflictError(t *testing.T) {
	appErr := NewConflictError("Email already exists")

	assert.Equal(t, "CONFLICT", appErr.Code)
	assert.Equal(t, "Email already exists", appErr.Message)
	assert.Equal(t, http.StatusConflict, appErr.StatusCode)
}

func TestNewInternalServerError(t *testing.T) {
	cause := errors.New("database connection failed")
	appErr := NewInternalServerError("Database error", cause)

	assert.Equal(t, "INTERNAL_SERVER_ERROR", appErr.Code)
	assert.Equal(t, "Database error", appErr.Message)
	assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
	assert.Equal(t, cause, appErr.Err)
}

func TestNewBadRequestError(t *testing.T) {
	details := map[string]string{"param": "invalid"}
	appErr := NewBadRequestError("Invalid parameters", details)

	assert.Equal(t, "BAD_REQUEST", appErr.Code)
	assert.Equal(t, "Invalid parameters", appErr.Message)
	assert.Equal(t, http.StatusBadRequest, appErr.StatusCode)
	assert.Equal(t, details, appErr.Details)
}

func TestNewDatabaseError(t *testing.T) {
	cause := errors.New("connection timeout")
	appErr := NewDatabaseError("Database operation failed", cause)

	assert.Equal(t, "DATABASE_ERROR", appErr.Code)
	assert.Equal(t, "Database operation failed", appErr.Message)
	assert.Equal(t, http.StatusInternalServerError, appErr.StatusCode)
	assert.Equal(t, cause, appErr.Err)
}

func TestNewAuthenticationError(t *testing.T) {
	appErr := NewAuthenticationError("Invalid password")

	assert.Equal(t, "AUTHENTICATION_ERROR", appErr.Code)
	assert.Equal(t, "Invalid password", appErr.Message)
	assert.Equal(t, http.StatusUnauthorized, appErr.StatusCode)
}

func TestNewTokenError(t *testing.T) {
	appErr := NewTokenError("Token expired")

	assert.Equal(t, "TOKEN_ERROR", appErr.Code)
	assert.Equal(t, "Token expired", appErr.Message)
	assert.Equal(t, http.StatusUnauthorized, appErr.StatusCode)
}

func TestIsAppError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is app error",
			err:      NewNotFoundError("User"),
			expected: true,
		},
		{
			name:     "is not app error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAppError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAsAppError(t *testing.T) {
	appErr := NewNotFoundError("User")
	regularErr := errors.New("regular error")

	// Test with AppError
	result, ok := AsAppError(appErr)
	assert.True(t, ok)
	assert.Equal(t, appErr, result)

	// Test with regular error
	result, ok = AsAppError(regularErr)
	assert.False(t, ok)
	assert.Nil(t, result)

	// Test with nil
	result, ok = AsAppError(nil)
	assert.False(t, ok)
	assert.Nil(t, result)
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "app error",
			err:      NewNotFoundError("User"),
			expected: http.StatusNotFound,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: http.StatusInternalServerError,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetStatusCode(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "app error",
			err:      NewNotFoundError("User"),
			expected: "NOT_FOUND",
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: "INTERNAL_SERVER_ERROR",
		},
		{
			name:     "nil error",
			err:      nil,
			expected: "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorCode(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetErrorDetails(t *testing.T) {
	details := map[string]string{"field": "error"}

	tests := []struct {
		name     string
		err      error
		expected map[string]string
	}{
		{
			name:     "app error with details",
			err:      NewValidationError(details),
			expected: details,
		},
		{
			name:     "app error without details",
			err:      NewNotFoundError("User"),
			expected: nil,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetErrorDetails(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
