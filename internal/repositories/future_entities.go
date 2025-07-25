package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// TokenBlacklist represents the token blacklist model interface
// This interface is implemented by models.TokenBlacklist
type TokenBlacklist interface {
	GetID() uint
	GetTokenHash() string
	GetUserID() uint
	GetExpiresAt() time.Time
	GetReason() string
}

// ProfileRepository defines the interface for user profile operations
// This is a placeholder for future user profile functionality
type ProfileRepository interface {
	BaseRepository[Profile]

	// Profile-specific operations
	GetByUserID(userID uint) (Profile, error)
	GetByUserIDTx(tx *gorm.DB, userID uint) (Profile, error)
	GetByUserIDWithContext(ctx context.Context, userID uint) (Profile, error)

	UpdateByUserID(userID uint, profile Profile) error
	UpdateByUserIDTx(tx *gorm.DB, userID uint, profile Profile) error
	UpdateByUserIDWithContext(ctx context.Context, userID uint, profile Profile) error

	DeleteByUserID(userID uint) error
	DeleteByUserIDTx(tx *gorm.DB, userID uint) error
	DeleteByUserIDWithContext(ctx context.Context, userID uint) error
}

// Profile represents the user profile model interface
type Profile interface {
	GetID() uint
	GetUserID() uint
	GetBio() string
	GetAvatarURL() string
	GetPhoneNumber() string
	GetDateOfBirth() *time.Time
}

// AuditLogRepository defines the interface for audit log operations
// This is for tracking user actions and system events
type AuditLogRepository interface {
	BaseRepository[AuditLog]

	// Audit-specific operations
	GetByUserID(userID uint, limit, offset int) ([]AuditLog, error)
	GetByUserIDTx(tx *gorm.DB, userID uint, limit, offset int) ([]AuditLog, error)
	GetByUserIDWithContext(ctx context.Context, userID uint, limit, offset int) ([]AuditLog, error)

	GetByAction(action string, limit, offset int) ([]AuditLog, error)
	GetByActionTx(tx *gorm.DB, action string, limit, offset int) ([]AuditLog, error)
	GetByActionWithContext(ctx context.Context, action string, limit, offset int) ([]AuditLog, error)

	GetByDateRange(startDate, endDate time.Time, limit, offset int) ([]AuditLog, error)
	GetByDateRangeTx(tx *gorm.DB, startDate, endDate time.Time, limit, offset int) ([]AuditLog, error)
	GetByDateRangeWithContext(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]AuditLog, error)

	// Cleanup old audit logs
	CleanupOldLogs(olderThan time.Time) error
	CleanupOldLogsTx(tx *gorm.DB, olderThan time.Time) error
	CleanupOldLogsWithContext(ctx context.Context, olderThan time.Time) error
}

// AuditLog represents the audit log model interface
type AuditLog interface {
	GetID() uint
	GetUserID() uint
	GetAction() string
	GetResource() string
	GetResourceID() uint
	GetOldValues() string
	GetNewValues() string
	GetIPAddress() string
	GetUserAgent() string
	GetCreatedAt() time.Time
}
