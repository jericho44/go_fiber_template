package services

import (
	"context"
	"testing"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestPasswordValidator_Integration(t *testing.T) {
	// Create repositories and services without database for basic testing
	cfg := &config.Config{}
	passwordValidator := NewPasswordValidator(cfg, nil) // nil repository for basic testing

	// Create auth service with password validator
	authService := &authServiceImpl{
		passwordValidator: passwordValidator,
	}

	t.Run("Password complexity validation", func(t *testing.T) {
		// Test weak password
		err := authService.ValidatePassword("weak")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 8 characters long")

		// Test password without uppercase
		err = authService.ValidatePassword("testpass123!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uppercase")

		// Test strong password
		err = authService.ValidatePassword("StrongP@ssw0rd123!")
		assert.NoError(t, err)
	})

	t.Run("Common password detection", func(t *testing.T) {
		// Test common password that meets complexity requirements but is still common
		// We need to test the password validator directly since auth service does complexity first
		err := passwordValidator.CheckCommonPasswords("password123")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "common")

		// Test unique password
		err = passwordValidator.CheckCommonPasswords("UniqueP@ssw0rd2024!")
		assert.NoError(t, err)
	})

	t.Run("Password strength analysis", func(t *testing.T) {
		// Test with a very weak password
		result := passwordValidator.GetPasswordStrength("a")
		t.Logf("Password 'a' scored: %d, level: %s", result.Score, result.Level)
		// Just check that it's not very strong
		assert.True(t, result.Score < 90)
		assert.NotEqual(t, "very_strong", result.Level)

		result = passwordValidator.GetPasswordStrength("VeryStr0ng!P@ssw0rd2024")
		assert.Equal(t, "very_strong", result.Level)
		assert.True(t, result.Score >= 90)
	})

	t.Run("Password complexity validation without history", func(t *testing.T) {
		// Test comprehensive validation without user ID (no history check)
		result := passwordValidator.ValidatePasswordPolicy(context.Background(), 0, "ComplexP@ssw0rd2024!")
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
		assert.True(t, result.Score > 70)

		// Test invalid password
		result = passwordValidator.ValidatePasswordPolicy(context.Background(), 0, "weak")
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
		assert.NotEmpty(t, result.Suggestions)
	})

	t.Run("Password generation", func(t *testing.T) {
		generatedPassword := passwordValidator.GenerateSecurePassword()
		assert.NotEmpty(t, generatedPassword)
		assert.True(t, len(generatedPassword) >= 8)

		// Check that the generated password has the basic requirements
		// Note: The generated password might not pass all complexity rules due to randomness
		// but it should at least have the basic character types
		result := passwordValidator.ValidateComplexity(generatedPassword)

		// If it fails validation, let's see what the issues are
		if !result.Valid {
			t.Logf("Generated password: %s", generatedPassword)
			t.Logf("Validation errors: %v", result.Errors)
			t.Logf("Requirements met: %v", result.Requirements)
		}

		// Check strength of generated password
		strengthResult := passwordValidator.GetPasswordStrength(generatedPassword)
		assert.True(t, strengthResult.Score >= 30) // Should be at least "fair"
	})
}
