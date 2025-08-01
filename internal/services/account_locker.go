package services

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
)

// AccountLocker defines the interface for account lockout operations
type AccountLocker interface {
	// RecordFailedAttempt records a failed login attempt and potentially locks the account
	RecordFailedAttempt(ctx context.Context, email string) error

	// IsAccountLocked checks if an account is currently locked
	IsAccountLocked(ctx context.Context, email string) (bool, error)

	// UnlockAccount unlocks an account (admin functionality)
	UnlockAccount(ctx context.Context, email string) error

	// GetLockoutInfo returns lockout information for an account
	GetLockoutInfo(ctx context.Context, email string) (*LockoutInfo, error)

	// ResetFailedAttempts resets failed login attempts for successful login
	ResetFailedAttempts(ctx context.Context, email string) error
}

// LockoutInfo contains information about account lockout status
type LockoutInfo struct {
	IsLocked         bool       `json:"is_locked"`
	FailedAttempts   int        `json:"failed_attempts"`
	LockedUntil      *time.Time `json:"locked_until,omitempty"`
	NextLockDuration string     `json:"next_lock_duration,omitempty"`
}

// LockoutConfig defines the configuration for account lockout behavior
type LockoutConfig struct {
	MaxFailedAttempts int           `json:"max_failed_attempts"`
	BaseLockDuration  time.Duration `json:"base_lock_duration"`
	MaxLockDuration   time.Duration `json:"max_lock_duration"`
	ProgressiveMode   bool          `json:"progressive_mode"`
}

// DefaultLockoutConfig returns the default lockout configuration
func DefaultLockoutConfig() *LockoutConfig {
	return &LockoutConfig{
		MaxFailedAttempts: 5,
		BaseLockDuration:  5 * time.Minute,
		MaxLockDuration:   24 * time.Hour,
		ProgressiveMode:   true,
	}
}

// accountLockerImpl is the concrete implementation of AccountLocker
type accountLockerImpl struct {
	db        *gorm.DB
	txManager repositories.TransactionManager
	config    *LockoutConfig
}

// NewAccountLocker creates a new instance of AccountLocker
func NewAccountLocker(db *gorm.DB, txManager repositories.TransactionManager, config *LockoutConfig) AccountLocker {
	if config == nil {
		config = DefaultLockoutConfig()
	}
	return &accountLockerImpl{
		db:        db,
		txManager: txManager,
		config:    config,
	}
}

// RecordFailedAttempt records a failed login attempt and potentially locks the account
func (a *accountLockerImpl) RecordFailedAttempt(ctx context.Context, email string) error {
	return a.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Get user by email
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Don't reveal that the user doesn't exist
				return nil
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Increment failed login count
		user.IncrementFailedLoginCount()

		// Check if account should be locked
		if user.GetFailedLoginCount() >= a.config.MaxFailedAttempts {
			lockDuration := a.calculateLockDuration(user.GetFailedLoginCount())
			lockUntil := time.Now().Add(lockDuration)
			user.LockAccount(lockUntil)
		}

		// Update user in database
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// IsAccountLocked checks if an account is currently locked
func (a *accountLockerImpl) IsAccountLocked(ctx context.Context, email string) (bool, error) {
	var user models.User
	if err := a.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal that the user doesn't exist
			return false, nil
		}
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.IsLocked(), nil
}

// UnlockAccount unlocks an account (admin functionality)
func (a *accountLockerImpl) UnlockAccount(ctx context.Context, email string) error {
	return a.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Unlock the account
		user.UnlockAccount()

		// Update user in database
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// GetLockoutInfo returns lockout information for an account
func (a *accountLockerImpl) GetLockoutInfo(ctx context.Context, email string) (*LockoutInfo, error) {
	var user models.User
	if err := a.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default info for non-existent users
			return &LockoutInfo{
				IsLocked:       false,
				FailedAttempts: 0,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	info := &LockoutInfo{
		IsLocked:       user.IsLocked(),
		FailedAttempts: user.GetFailedLoginCount(),
		LockedUntil:    user.LockedUntil,
	}

	// Calculate next lock duration if not currently locked
	if !info.IsLocked && user.GetFailedLoginCount() > 0 {
		nextLockDuration := a.calculateLockDuration(user.GetFailedLoginCount() + 1)
		info.NextLockDuration = nextLockDuration.String()
	}

	return info, nil
}

// ResetFailedAttempts resets failed login attempts for successful login
func (a *accountLockerImpl) ResetFailedAttempts(ctx context.Context, email string) error {
	return a.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		var user models.User
		if err := tx.Where("email = ?", email).First(&user).Error; err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Reset failed attempts and update last login
		user.ResetFailedLoginCount()
		user.UpdateLastLoginAt()

		// Clear any existing lock
		if user.LockedUntil != nil {
			user.LockedUntil = nil
		}

		// Update user in database
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}

		return nil
	})
}

// calculateLockDuration calculates the lock duration based on failed attempts
func (a *accountLockerImpl) calculateLockDuration(failedAttempts int) time.Duration {
	if !a.config.ProgressiveMode {
		return a.config.BaseLockDuration
	}

	// Progressive lockout: exponential backoff with maximum cap
	// Formula: base_duration * 2^(attempts - max_attempts)
	exponent := failedAttempts - a.config.MaxFailedAttempts
	if exponent < 0 {
		exponent = 0
	}

	duration := a.config.BaseLockDuration
	for i := 0; i < exponent; i++ {
		duration *= 2
		if duration > a.config.MaxLockDuration {
			return a.config.MaxLockDuration
		}
	}

	return duration
}
