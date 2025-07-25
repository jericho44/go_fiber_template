package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateStruct(t *testing.T) {
	// Test with valid struct
	validUser := User{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}

	err := ValidateStruct(validUser)
	assert.NoError(t, err)

	// Test with invalid struct
	invalidUser := User{
		Email:     "invalid-email",
		Password:  "",
		FirstName: "",
		LastName:  "",
	}

	err = ValidateStruct(invalidUser)
	require.Error(t, err)
}

func TestGetValidator(t *testing.T) {
	validator := GetValidator()
	assert.NotNil(t, validator)

	// Test that it's the same instance
	validator2 := GetValidator()
	assert.Equal(t, validator, validator2)
}
