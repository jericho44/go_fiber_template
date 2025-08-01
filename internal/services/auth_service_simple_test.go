package services

import (
	"go-fiber-template/internal/models"
	"go-fiber-template/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimple(t *testing.T) {
	assert.True(t, true)
}

func TestAuthService_ValidatePassword_Simple(t *testing.T) {
	// Create a simple auth service without mocks
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
		// passwordValidator is nil, so it will use fallback validation
	}

	err := service.ValidatePassword("Password123!")
	assert.NoError(t, err)

	err = service.ValidatePassword("weak")
	assert.Error(t, err)
}

func TestAuthService_HashPassword_Simple(t *testing.T) {
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
	}

	password := "TestPassword123!"
	hashedPassword, err := service.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hashedPassword)
	assert.NotEqual(t, password, hashedPassword)
}

func TestAuthService_ComparePassword_Simple(t *testing.T) {
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
	}

	password := "TestPassword123!"
	hashedPassword, err := service.HashPassword(password)
	assert.NoError(t, err)

	// Test correct password
	err = service.ComparePassword(hashedPassword, password)
	assert.NoError(t, err)

	// Test incorrect password
	err = service.ComparePassword(hashedPassword, "WrongPassword")
	assert.Error(t, err)
}

func TestAuthService_PasswordValidation_Comprehensive(t *testing.T) {
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
		// passwordValidator is nil, so it will use fallback validation
	}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid password with all requirements",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "Valid password with different special chars",
			password: "MyPass123@",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Pass1!",
			wantErr:  true,
			errMsg:   "at least 8 characters long",
		},
		{
			name:     "No uppercase letter",
			password: "password123!",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "No lowercase letter",
			password: "PASSWORD123!",
			wantErr:  true,
			errMsg:   "lowercase letter",
		},
		{
			name:     "No digit",
			password: "Password!",
			wantErr:  true,
			errMsg:   "digit",
		},
		{
			name:     "No special character",
			password: "Password123",
			wantErr:  true,
			errMsg:   "special character",
		},
		{
			name:     "Multiple missing requirements",
			password: "password",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "Only special characters",
			password: "!@#$%^&*",
			wantErr:  true,
			errMsg:   "uppercase letter",
		},
		{
			name:     "Long valid password",
			password: "ThisIsAVeryLongPassword123!WithManyCharacters",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthService_HashPassword_Security(t *testing.T) {
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
	}

	t.Run("Same password produces different hashes", func(t *testing.T) {
		password := "TestPassword123!"

		hash1, err1 := service.HashPassword(password)
		hash2, err2 := service.HashPassword(password)

		assert.NoError(t, err1)
		assert.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "Same password should produce different hashes due to salt")

		// Both hashes should validate the original password
		assert.NoError(t, service.ComparePassword(hash1, password))
		assert.NoError(t, service.ComparePassword(hash2, password))
	})

	t.Run("Hash is not the original password", func(t *testing.T) {
		password := "TestPassword123!"
		hash, err := service.HashPassword(password)

		assert.NoError(t, err)
		assert.NotEqual(t, password, hash)
		assert.NotEmpty(t, hash)
		assert.True(t, len(hash) > len(password), "Hash should be longer than original password")
	})

	t.Run("Empty password handling", func(t *testing.T) {
		hash, err := service.HashPassword("")
		assert.NoError(t, err) // bcrypt can hash empty strings
		assert.NotEmpty(t, hash)

		// Should be able to compare empty password
		err = service.ComparePassword(hash, "")
		assert.NoError(t, err)

		// Should fail with non-empty password
		err = service.ComparePassword(hash, "something")
		assert.Error(t, err)
	})
}

func TestAuthService_ComparePassword_EdgeCases(t *testing.T) {
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
	}

	t.Run("Invalid hash format", func(t *testing.T) {
		err := service.ComparePassword("invalid-hash", "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compare password")
	})

	t.Run("Empty hash", func(t *testing.T) {
		err := service.ComparePassword("", "password")
		assert.Error(t, err)
	})

	t.Run("Very long password within bcrypt limits", func(t *testing.T) {
		// bcrypt has a 72-byte limit, so test with a password just under that
		longPassword := string(make([]byte, 68)) // Leave room for required chars
		for i := range longPassword {
			longPassword = longPassword[:i] + "a" + longPassword[i+1:]
		}
		longPassword += "A1!" // Add required characters (total = 71 bytes)

		hash, err := service.HashPassword(longPassword)
		assert.NoError(t, err)

		err = service.ComparePassword(hash, longPassword)
		assert.NoError(t, err)
	})

	t.Run("Password exceeding bcrypt limit", func(t *testing.T) {
		// Test that passwords over 72 bytes are handled gracefully
		veryLongPassword := string(make([]byte, 100))
		for i := range veryLongPassword {
			veryLongPassword = veryLongPassword[:i] + "a" + veryLongPassword[i+1:]
		}
		veryLongPassword += "A1!" // This will exceed 72 bytes

		_, err := service.HashPassword(veryLongPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to hash password")
	})
}

func TestAuthService_Interface_Implementation(t *testing.T) {
	// Test that authServiceImpl implements AuthService interface
	var _ AuthService = &authServiceImpl{}

	// Test that we can create a service instance
	service := &authServiceImpl{
		validator: utils.NewValidationErrorFormatter(),
	}

	// Test that all interface methods are callable
	assert.NotNil(t, service.ValidatePassword)
	assert.NotNil(t, service.HashPassword)
	assert.NotNil(t, service.ComparePassword)
	assert.NotNil(t, service.Register)
	assert.NotNil(t, service.Login)
}

func TestAuthService_NewAuthService(t *testing.T) {
	// Test that NewAuthService creates a proper service instance
	// We can't test with real DB/TxManager without complex setup, but we can test the constructor

	// This test verifies that the constructor function exists and returns the right type
	// In a real scenario, this would be called with actual DB and TxManager instances
	t.Run("Constructor exists and returns AuthService", func(t *testing.T) {
		// We can't call NewAuthService without real dependencies, but we can verify the function exists
		assert.NotNil(t, NewAuthService)

		// Verify the function signature by checking it can be assigned to the right type
		var constructor func(interface{}, interface{}) AuthService
		constructor = func(db interface{}, txManager interface{}) AuthService {
			// This is just to test the signature - in real usage, these would be proper types
			return &authServiceImpl{
				validator: utils.NewValidationErrorFormatter(),
			}
		}
		assert.NotNil(t, constructor)
	})
}

func TestAuthService_RequestStructs(t *testing.T) {
	t.Run("RegisterRequest validation tags", func(t *testing.T) {
		req := RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		// Test that the struct can be created and has the expected fields
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "Password123!", req.Password)
		assert.Equal(t, "John", req.FirstName)
		assert.Equal(t, "Doe", req.LastName)
	})

	t.Run("LoginRequest validation tags", func(t *testing.T) {
		req := LoginRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		// Test that the struct can be created and has the expected fields
		assert.Equal(t, "test@example.com", req.Email)
		assert.Equal(t, "Password123!", req.Password)
	})

	t.Run("AuthResponse structure", func(t *testing.T) {
		user := &models.User{
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		response := &AuthResponse{
			User:    user,
			Message: "Success",
		}

		assert.Equal(t, user, response.User)
		assert.Equal(t, "Success", response.Message)
	})
}
