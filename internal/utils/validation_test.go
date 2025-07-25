package utils

import (
	"reflect"
	"testing"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for validation
type TestUser struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Age      int    `json:"age" validate:"gte=18,lte=100"`
	Username string `json:"username" validate:"required,alphanum,min=3,max=20"`
	Website  string `json:"website" validate:"omitempty,url"`
	Role     string `json:"role" validate:"oneof=admin user guest"`
}

type TestProfile struct {
	FirstName string `json:"first_name" validate:"required,alpha"`
	LastName  string `json:"last_name" validate:"required,alpha"`
	Phone     string `json:"phone" validate:"omitempty,numeric"`
}

func TestNewValidationErrorFormatter(t *testing.T) {
	formatter := NewValidationErrorFormatter()
	assert.NotNil(t, formatter)
	assert.NotNil(t, formatter.validator)
}

func TestValidationErrorFormatter_FormatValidationErrors(t *testing.T) {
	formatter := NewValidationErrorFormatter()

	// Create a struct with validation errors
	user := TestUser{
		Email:    "invalid-email",
		Password: "short",
		Age:      15,
		Username: "ab",
		Website:  "not-a-url",
		Role:     "invalid-role",
	}

	err := formatter.validator.Struct(user)
	require.Error(t, err)

	errors := formatter.FormatValidationErrors(err)

	// Check that we have errors for the expected fields
	assert.Contains(t, errors, "email")
	assert.Contains(t, errors, "password")
	assert.Contains(t, errors, "age")
	assert.Contains(t, errors, "username")
	assert.Contains(t, errors, "website")
	assert.Contains(t, errors, "role")

	// Check some specific error messages
	assert.Contains(t, errors["email"], "valid email")
	assert.Contains(t, errors["password"], "at least 8")
	assert.Contains(t, errors["age"], "greater than or equal to 18")
	assert.Contains(t, errors["username"], "at least 3")
}

func TestValidationErrorFormatter_FormatValidationErrors_NonValidationError(t *testing.T) {
	formatter := NewValidationErrorFormatter()

	// Test with a non-validation error
	errors := formatter.FormatValidationErrors(assert.AnError)

	assert.Empty(t, errors)
}

func TestValidationErrorFormatter_ValidateStruct(t *testing.T) {
	formatter := NewValidationErrorFormatter()

	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{
			name: "valid struct",
			input: TestUser{
				Email:    "test@example.com",
				Password: "password123",
				Age:      25,
				Username: "testuser",
				Website:  "https://example.com",
				Role:     "user",
			},
			expectErr: false,
		},
		{
			name: "invalid struct",
			input: TestUser{
				Email:    "invalid-email",
				Password: "short",
				Age:      15,
				Username: "ab",
				Role:     "invalid",
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := formatter.ValidateStruct(tt.input)

			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Equal(t, "VALIDATION_ERROR", err.Code)
				assert.NotEmpty(t, err.Details)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		field    string
		param    string
		expected string
	}{
		{
			name:     "required field",
			tag:      "required",
			field:    "Email",
			param:    "",
			expected: "Email is required",
		},
		{
			name:     "email validation",
			tag:      "email",
			field:    "Email",
			param:    "",
			expected: "Email must be a valid email address",
		},
		{
			name:     "min length",
			tag:      "min",
			field:    "Password",
			param:    "8",
			expected: "Password must be at least 8 characters long",
		},
		{
			name:     "max length",
			tag:      "max",
			field:    "Username",
			param:    "20",
			expected: "Username must be at most 20 characters long",
		},
		{
			name:     "numeric validation",
			tag:      "numeric",
			field:    "Phone",
			param:    "",
			expected: "Phone must be numeric",
		},
		{
			name:     "alpha validation",
			tag:      "alpha",
			field:    "Name",
			param:    "",
			expected: "Name must contain only letters",
		},
		{
			name:     "alphanum validation",
			tag:      "alphanum",
			field:    "Username",
			param:    "",
			expected: "Username must contain only letters and numbers",
		},
		{
			name:     "url validation",
			tag:      "url",
			field:    "Website",
			param:    "",
			expected: "Website must be a valid URL",
		},
		{
			name:     "oneof validation",
			tag:      "oneof",
			field:    "Role",
			param:    "admin user guest",
			expected: "Role must be one of: admin user guest",
		},
		{
			name:     "greater than",
			tag:      "gt",
			field:    "Age",
			param:    "0",
			expected: "Age must be greater than 0",
		},
		{
			name:     "greater than or equal",
			tag:      "gte",
			field:    "Age",
			param:    "18",
			expected: "Age must be greater than or equal to 18",
		},
		{
			name:     "less than",
			tag:      "lt",
			field:    "Age",
			param:    "100",
			expected: "Age must be less than 100",
		},
		{
			name:     "less than or equal",
			tag:      "lte",
			field:    "Age",
			param:    "100",
			expected: "Age must be less than or equal to 100",
		},
		{
			name:     "unknown tag",
			tag:      "unknown",
			field:    "Field",
			param:    "",
			expected: "Field is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock field error
			mockFieldError := &mockFieldError{
				field: tt.field,
				tag:   tt.tag,
				param: tt.param,
			}

			result := getErrorMessage(mockFieldError)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAndRespond(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		expectErr bool
	}{
		{
			name: "valid struct",
			input: TestProfile{
				FirstName: "John",
				LastName:  "Doe",
				Phone:     "1234567890",
			},
			expectErr: false,
		},
		{
			name: "invalid struct",
			input: TestProfile{
				FirstName: "John123", // Should be alpha only
				LastName:  "",        // Required
				Phone:     "abc",     // Should be numeric
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAndRespond(tt.input)

			if tt.expectErr {
				assert.NotNil(t, err)
				assert.Equal(t, "VALIDATION_ERROR", err.Code)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestIsValidationError(t *testing.T) {
	validator := validator.New()

	// Create validation error
	user := TestUser{Email: "invalid"}
	validationErr := validator.Struct(user)

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation error",
			err:      validationErr,
			expected: true,
		},
		{
			name:     "regular error",
			err:      assert.AnError,
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
			result := IsValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Mock field error for testing
type mockFieldError struct {
	field string
	tag   string
	param string
}

func (m *mockFieldError) Tag() string                       { return m.tag }
func (m *mockFieldError) ActualTag() string                 { return m.tag }
func (m *mockFieldError) Namespace() string                 { return "" }
func (m *mockFieldError) StructNamespace() string           { return "" }
func (m *mockFieldError) Field() string                     { return m.field }
func (m *mockFieldError) StructField() string               { return m.field }
func (m *mockFieldError) Value() interface{}                { return nil }
func (m *mockFieldError) Param() string                     { return m.param }
func (m *mockFieldError) Kind() reflect.Kind                { return reflect.String }
func (m *mockFieldError) Type() reflect.Type                { return nil }
func (m *mockFieldError) Translate(ut ut.Translator) string { return "" }
func (m *mockFieldError) Error() string                     { return "" }
