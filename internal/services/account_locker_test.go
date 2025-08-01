package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
)

func setupAccountLockerTest(t *testing.T) (*gorm.DB, AccountLocker, func()) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the User model
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	// Create transaction manager
	txManager := repositories.NewTransactionManager(db)

	// Create account locker with test configuration
	config := &LockoutConfig{
		MaxFailedAttempts: 3,
		BaseLockDuration:  1 * time.Minute,
		MaxLockDuration:   10 * time.Minute,
		ProgressiveMode:   true,
	}
	accountLocker := NewAccountLocker(db, txManager, config)

	// Create a test user
	user := &models.User{
		Email:     "test@example.com",
		Password:  "hashedpassword",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}
	err = db.Create(user).Error
	require.NoError(t, err)

	cleanup := func() {
		// Clean up database
		db.Exec("DELETE FROM users")
	}

	return db, accountLocker, cleanup
}

func TestAccountLocker_RecordFailedAttempt(t *testing.T) {
	db, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	// Test recording failed attempts
	for i := 1; i <= 3; i++ {
		err := accountLocker.RecordFailedAttempt(ctx, email)
		assert.NoError(t, err)

		// Check failed attempt count
		var user models.User
		err = db.Where("email = ?", email).First(&user).Error
		assert.NoError(t, err)
		assert.Equal(t, i, user.GetFailedLoginCount())

		// Check if account is locked after max attempts
		if i >= 3 {
			assert.True(t, user.IsLocked())
			assert.NotNil(t, user.LockedUntil)
		} else {
			assert.False(t, user.IsLocked())
		}
	}
}

func TestAccountLocker_RecordFailedAttempt_NonExistentUser(t *testing.T) {
	_, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "nonexistent@example.com"

	// Should not return error for non-existent user (security measure)
	err := accountLocker.RecordFailedAttempt(ctx, email)
	assert.NoError(t, err)
}

func TestAccountLocker_IsAccountLocked(t *testing.T) {
	db, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	// Initially not locked
	isLocked, err := accountLocker.IsAccountLocked(ctx, email)
	assert.NoError(t, err)
	assert.False(t, isLocked)

	// Lock the account manually
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)

	lockUntil := time.Now().Add(5 * time.Minute)
	user.LockAccount(lockUntil)
	err = db.Save(&user).Error
	assert.NoError(t, err)

	// Should be locked now
	isLocked, err = accountLocker.IsAccountLocked(ctx, email)
	assert.NoError(t, err)
	assert.True(t, isLocked)

	// Test with expired lock
	user.LockedUntil = &time.Time{} // Set to past time
	err = db.Save(&user).Error
	assert.NoError(t, err)

	isLocked, err = accountLocker.IsAccountLocked(ctx, email)
	assert.NoError(t, err)
	assert.False(t, isLocked)
}

func TestAccountLocker_IsAccountLocked_NonExistentUser(t *testing.T) {
	_, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "nonexistent@example.com"

	// Should return false for non-existent user (security measure)
	isLocked, err := accountLocker.IsAccountLocked(ctx, email)
	assert.NoError(t, err)
	assert.False(t, isLocked)
}

func TestAccountLocker_UnlockAccount(t *testing.T) {
	db, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	// Lock the account first
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)

	user.FailedLoginCount = 5
	lockUntil := time.Now().Add(5 * time.Minute)
	user.LockAccount(lockUntil)
	err = db.Save(&user).Error
	assert.NoError(t, err)

	// Verify account is locked
	assert.True(t, user.IsLocked())
	assert.Equal(t, 5, user.GetFailedLoginCount())

	// Unlock the account
	err = accountLocker.UnlockAccount(ctx, email)
	assert.NoError(t, err)

	// Verify account is unlocked and failed attempts reset
	err = db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)
	assert.False(t, user.IsLocked())
	assert.Equal(t, 0, user.GetFailedLoginCount())
	assert.Nil(t, user.LockedUntil)
}

func TestAccountLocker_UnlockAccount_NonExistentUser(t *testing.T) {
	_, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "nonexistent@example.com"

	// Should return error for non-existent user
	err := accountLocker.UnlockAccount(ctx, email)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get user")
}

func TestAccountLocker_GetLockoutInfo(t *testing.T) {
	db, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	// Test with no failed attempts
	info, err := accountLocker.GetLockoutInfo(ctx, email)
	assert.NoError(t, err)
	assert.False(t, info.IsLocked)
	assert.Equal(t, 0, info.FailedAttempts)
	assert.Nil(t, info.LockedUntil)

	// Add some failed attempts
	var user models.User
	err = db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)

	user.FailedLoginCount = 2
	err = db.Save(&user).Error
	assert.NoError(t, err)

	info, err = accountLocker.GetLockoutInfo(ctx, email)
	assert.NoError(t, err)
	assert.False(t, info.IsLocked)
	assert.Equal(t, 2, info.FailedAttempts)
	assert.NotEmpty(t, info.NextLockDuration)

	// Lock the account
	lockUntil := time.Now().Add(5 * time.Minute)
	user.LockAccount(lockUntil)
	err = db.Save(&user).Error
	assert.NoError(t, err)

	info, err = accountLocker.GetLockoutInfo(ctx, email)
	assert.NoError(t, err)
	assert.True(t, info.IsLocked)
	assert.Equal(t, 2, info.FailedAttempts)
	assert.NotNil(t, info.LockedUntil)
}

func TestAccountLocker_GetLockoutInfo_NonExistentUser(t *testing.T) {
	_, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "nonexistent@example.com"

	// Should return default info for non-existent user
	info, err := accountLocker.GetLockoutInfo(ctx, email)
	assert.NoError(t, err)
	assert.False(t, info.IsLocked)
	assert.Equal(t, 0, info.FailedAttempts)
}

func TestAccountLocker_ResetFailedAttempts(t *testing.T) {
	db, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	ctx := context.Background()
	email := "test@example.com"

	// Set some failed attempts
	var user models.User
	err := db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)

	user.FailedLoginCount = 3
	lockUntil := time.Now().Add(5 * time.Minute)
	user.LockAccount(lockUntil)
	err = db.Save(&user).Error
	assert.NoError(t, err)

	// Reset failed attempts
	err = accountLocker.ResetFailedAttempts(ctx, email)
	assert.NoError(t, err)

	// Verify reset
	err = db.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, 0, user.GetFailedLoginCount())
	assert.Nil(t, user.LockedUntil)
	assert.NotNil(t, user.LastLoginAt)
}

func TestAccountLocker_ProgressiveLockDuration(t *testing.T) {
	_, accountLocker, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	impl := accountLocker.(*accountLockerImpl)

	// Test progressive lock duration calculation
	testCases := []struct {
		failedAttempts   int
		expectedDuration time.Duration
	}{
		{3, 1 * time.Minute},  // First lock: base duration
		{4, 2 * time.Minute},  // Second lock: 2x base
		{5, 4 * time.Minute},  // Third lock: 4x base
		{6, 8 * time.Minute},  // Fourth lock: 8x base
		{7, 10 * time.Minute}, // Fifth lock: capped at max
		{8, 10 * time.Minute}, // Sixth lock: still capped at max
	}

	for _, tc := range testCases {
		duration := impl.calculateLockDuration(tc.failedAttempts)
		assert.Equal(t, tc.expectedDuration, duration,
			"Failed attempts: %d, expected: %v, got: %v",
			tc.failedAttempts, tc.expectedDuration, duration)
	}
}

func TestAccountLocker_NonProgressiveLockDuration(t *testing.T) {
	db, _, cleanup := setupAccountLockerTest(t)
	defer cleanup()

	// Create account locker with non-progressive mode
	txManager := repositories.NewTransactionManager(db)
	config := &LockoutConfig{
		MaxFailedAttempts: 3,
		BaseLockDuration:  5 * time.Minute,
		MaxLockDuration:   10 * time.Minute,
		ProgressiveMode:   false,
	}
	accountLocker := NewAccountLocker(db, txManager, config)
	impl := accountLocker.(*accountLockerImpl)

	// Test non-progressive lock duration (should always be base duration)
	testCases := []int{3, 4, 5, 6, 7, 8}
	expectedDuration := 5 * time.Minute

	for _, failedAttempts := range testCases {
		duration := impl.calculateLockDuration(failedAttempts)
		assert.Equal(t, expectedDuration, duration,
			"Failed attempts: %d, expected: %v, got: %v",
			failedAttempts, expectedDuration, duration)
	}
}

func TestDefaultLockoutConfig(t *testing.T) {
	config := DefaultLockoutConfig()

	assert.Equal(t, 5, config.MaxFailedAttempts)
	assert.Equal(t, 5*time.Minute, config.BaseLockDuration)
	assert.Equal(t, 24*time.Hour, config.MaxLockDuration)
	assert.True(t, config.ProgressiveMode)
}
