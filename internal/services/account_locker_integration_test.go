package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"
	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
)

// TestAccountLockerIntegration tests the account locker with real database
func TestAccountLockerIntegration(t *testing.T) {
	// Load test configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Initialize test database
	err = database.Initialize(&cfg.Database)
	require.NoError(t, err)
	defer database.Close()

	db := database.GetDB()

	// Clean up test data
	db.Where("email LIKE ?", "%test-lockout%").Delete(&models.User{})

	// Create transaction manager and account locker
	txManager := repositories.NewTransactionManager(db)
	config := &LockoutConfig{
		MaxFailedAttempts: 3,
		BaseLockDuration:  1 * time.Second, // Short duration for testing
		MaxLockDuration:   10 * time.Second,
		ProgressiveMode:   true,
	}
	accountLocker := NewAccountLocker(db, txManager, config)

	// Create a test user
	user := &models.User{
		Email:     "test-lockout@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	ctx := context.Background()
	email := "test-lockout@example.com"

	t.Run("Account Lockout Flow", func(t *testing.T) {
		// Initially not locked
		isLocked, err := accountLocker.IsAccountLocked(ctx, email)
		assert.NoError(t, err)
		assert.False(t, isLocked)

		// Record failed attempts
		for i := 1; i <= 3; i++ {
			err := accountLocker.RecordFailedAttempt(ctx, email)
			assert.NoError(t, err)

			// Check if account is locked after max attempts
			isLocked, err := accountLocker.IsAccountLocked(ctx, email)
			assert.NoError(t, err)

			if i >= 3 {
				assert.True(t, isLocked, "Account should be locked after %d failed attempts", i)
			} else {
				assert.False(t, isLocked, "Account should not be locked after %d failed attempts", i)
			}
		}

		// Get lockout info
		info, err := accountLocker.GetLockoutInfo(ctx, email)
		assert.NoError(t, err)
		assert.True(t, info.IsLocked)
		assert.Equal(t, 3, info.FailedAttempts)
		assert.NotNil(t, info.LockedUntil)

		// Wait for lock to expire
		time.Sleep(2 * time.Second)

		// Should be unlocked now
		isLocked, err = accountLocker.IsAccountLocked(ctx, email)
		assert.NoError(t, err)
		assert.False(t, isLocked)

		// Reset failed attempts (simulate successful login)
		err = accountLocker.ResetFailedAttempts(ctx, email)
		assert.NoError(t, err)

		// Verify reset
		info, err = accountLocker.GetLockoutInfo(ctx, email)
		assert.NoError(t, err)
		assert.False(t, info.IsLocked)
		assert.Equal(t, 0, info.FailedAttempts)
	})

	t.Run("Admin Unlock", func(t *testing.T) {
		// Lock the account again
		for i := 0; i < 3; i++ {
			err := accountLocker.RecordFailedAttempt(ctx, email)
			assert.NoError(t, err)
		}

		// Verify locked
		isLocked, err := accountLocker.IsAccountLocked(ctx, email)
		assert.NoError(t, err)
		assert.True(t, isLocked)

		// Admin unlock
		err = accountLocker.UnlockAccount(ctx, email)
		assert.NoError(t, err)

		// Verify unlocked
		isLocked, err = accountLocker.IsAccountLocked(ctx, email)
		assert.NoError(t, err)
		assert.False(t, isLocked)

		// Verify failed attempts reset
		info, err := accountLocker.GetLockoutInfo(ctx, email)
		assert.NoError(t, err)
		assert.Equal(t, 0, info.FailedAttempts)
	})

	// Clean up test data
	db.Where("email = ?", email).Delete(&models.User{})
}

// TestAccountLockerWithAuthService tests integration with auth service
func TestAccountLockerWithAuthService(t *testing.T) {
	// Load test configuration
	cfg, err := config.Load()
	require.NoError(t, err)

	// Initialize test database
	err = database.Initialize(&cfg.Database)
	require.NoError(t, err)
	defer database.Close()

	db := database.GetDB()

	// Clean up test data
	db.Where("email LIKE ?", "%auth-lockout%").Delete(&models.User{})

	// Create services
	txManager := repositories.NewTransactionManager(db)
	passwordHistoryRepo := repositories.NewPasswordHistoryRepository(db)
	passwordValidator := NewPasswordValidator(cfg, passwordHistoryRepo)

	lockoutConfig := &LockoutConfig{
		MaxFailedAttempts: 2, // Lower threshold for testing
		BaseLockDuration:  1 * time.Second,
		MaxLockDuration:   10 * time.Second,
		ProgressiveMode:   true,
	}
	accountLocker := NewAccountLocker(db, txManager, lockoutConfig)

	authService := NewAuthService(db, txManager, passwordValidator, accountLocker)

	ctx := context.Background()

	// Register a user
	registerReq := RegisterRequest{
		Email:     "auth-lockout@example.com",
		Password:  "Password123!",
		FirstName: "Auth",
		LastName:  "Test",
	}

	_, err = authService.Register(ctx, registerReq)
	require.NoError(t, err)

	t.Run("Login Attempts with Account Lockout", func(t *testing.T) {
		loginReq := LoginRequest{
			Email:    "auth-lockout@example.com",
			Password: "WrongPassword123!",
		}

		// First failed attempt
		_, err := authService.Login(ctx, loginReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email or password")

		// Check not locked yet
		isLocked, err := accountLocker.IsAccountLocked(ctx, loginReq.Email)
		assert.NoError(t, err)
		assert.False(t, isLocked)

		// Second failed attempt - should lock the account
		_, err = authService.Login(ctx, loginReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email or password")

		// Check account is locked
		isLocked, err = accountLocker.IsAccountLocked(ctx, loginReq.Email)
		assert.NoError(t, err)
		assert.True(t, isLocked)

		// Third attempt should be blocked due to lockout
		_, err = authService.Login(ctx, loginReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Account is temporarily locked")

		// Wait for lock to expire
		time.Sleep(2 * time.Second)

		// Should be able to attempt login again (but still fail with wrong password)
		_, err = authService.Login(ctx, loginReq)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Invalid email or password")

		// Successful login should reset failed attempts
		correctLoginReq := LoginRequest{
			Email:    "auth-lockout@example.com",
			Password: "Password123!",
		}

		_, err = authService.Login(ctx, correctLoginReq)
		assert.NoError(t, err)

		// Verify failed attempts are reset
		info, err := accountLocker.GetLockoutInfo(ctx, loginReq.Email)
		assert.NoError(t, err)
		assert.Equal(t, 0, info.FailedAttempts)
		assert.False(t, info.IsLocked)
	})

	// Clean up test data
	db.Where("email = ?", "auth-lockout@example.com").Delete(&models.User{})
}
