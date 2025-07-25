package utils

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorFormatter formats validation errors into a map
type ValidationErrorFormatter struct {
	validator *validator.Validate
}

// NewValidationErrorFormatter creates a new validation error formatter
func NewValidationErrorFormatter() *ValidationErrorFormatter {
	return &ValidationErrorFormatter{
		validator: validator.New(),
	}
}

// FormatValidationErrors converts validator errors to a map of field errors
func (vef *ValidationErrorFormatter) FormatValidationErrors(err error) map[string]string {
	errors := make(map[string]string)

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			fieldName := getJSONFieldName(fieldError)
			errors[fieldName] = getErrorMessage(fieldError)
		}
	}

	return errors
}

// ValidateStruct validates a struct and returns formatted errors
func (vef *ValidationErrorFormatter) ValidateStruct(s interface{}) *AppError {
	if err := vef.validator.Struct(s); err != nil {
		details := vef.FormatValidationErrors(err)
		return NewValidationError(details)
	}
	return nil
}

// getJSONFieldName extracts the JSON field name from the struct field
func getJSONFieldName(fieldError validator.FieldError) string {
	field := fieldError.Field()

	// Try to get the JSON tag name
	if fieldError.StructNamespace() != "" {
		// Get the struct type
		structType := reflect.TypeOf(fieldError.StructNamespace())
		if structType != nil && structType.Kind() == reflect.Ptr {
			structType = structType.Elem()
		}

		// This is a simplified approach - in a real implementation,
		// you might want to parse the full namespace to get the exact field
		return strings.ToLower(field)
	}

	return strings.ToLower(field)
}

// getErrorMessage generates a human-readable error message based on the validation tag
func getErrorMessage(fieldError validator.FieldError) string {
	field := fieldError.Field()
	tag := fieldError.Tag()
	param := fieldError.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	case "numeric":
		return fmt.Sprintf("%s must be numeric", field)
	case "alpha":
		return fmt.Sprintf("%s must contain only letters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only letters and numbers", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, param)
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	case "eqfield":
		return fmt.Sprintf("%s must be equal to %s", field, param)
	case "nefield":
		return fmt.Sprintf("%s must not be equal to %s", field, param)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

// Common validation helpers

// ValidateAndRespond validates a struct and returns an error response if validation fails
func ValidateAndRespond(s interface{}) *AppError {
	formatter := NewValidationErrorFormatter()
	return formatter.ValidateStruct(s)
}

// IsValidationError checks if an error is a validation error
func IsValidationError(err error) bool {
	_, ok := err.(validator.ValidationErrors)
	return ok
}
