package repositories

import (
	"context"
	"go-fiber-template/internal/models"
	"time"

	"gorm.io/gorm"
)

// BaseRepository defines common repository operations that all repositories should implement
type BaseRepository[T any] interface {
	// Standard CRUD operations
	Create(entity *T) error
	GetByID(id uint) (T, error)
	Update(entity *T) error
	Delete(id uint) error
	List(limit, offset int) ([]T, error)
	Count() (int64, error)

	// Transactional CRUD operations
	CreateTx(tx *gorm.DB, entity *T) error
	UpdateTx(tx *gorm.DB, entity *T) error
	DeleteTx(tx *gorm.DB, id uint) error

	// Context-aware operations
	CreateWithContext(ctx context.Context, entity *T) error
	GetByIDWithContext(ctx context.Context, id uint) (T, error)
	UpdateWithContext(ctx context.Context, entity *T) error
	DeleteWithContext(ctx context.Context, id uint) error
	ListWithContext(ctx context.Context, limit, offset int) ([]T, error)
	CountWithContext(ctx context.Context) (int64, error)
}

// TokenBlacklistRepository defines the interface for token blacklist operations
type TokenBlacklistRepository interface {
	BaseRepository[models.TokenBlacklist]

	// IsTokenBlacklisted checks if a token hash is blacklisted
	IsTokenBlacklisted(tokenHash string) (bool, error)
	IsTokenBlacklistedWithContext(ctx context.Context, tokenHash string) (bool, error)

	// BlacklistToken adds a token to the blacklist
	BlacklistToken(tokenHash string, userID uint, expiresAt time.Time, reason string) error
	BlacklistTokenWithContext(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time, reason string) error
	BlacklistTokenTx(tx *gorm.DB, tokenHash string, userID uint, expiresAt time.Time, reason string) error

	// GetBlacklistedToken retrieves a blacklisted token by hash
	GetBlacklistedToken(tokenHash string) (*models.TokenBlacklist, error)
	GetBlacklistedTokenWithContext(ctx context.Context, tokenHash string) (*models.TokenBlacklist, error)

	// CleanupExpiredTokens removes expired tokens from the blacklist
	CleanupExpiredTokens() (int64, error)
	CleanupExpiredTokensWithContext(ctx context.Context) (int64, error)
	CleanupExpiredTokensTx(tx *gorm.DB) (int64, error)

	// GetUserBlacklistedTokens retrieves all blacklisted tokens for a user
	GetUserBlacklistedTokens(userID uint, limit, offset int) ([]models.TokenBlacklist, error)
	GetUserBlacklistedTokensWithContext(ctx context.Context, userID uint, limit, offset int) ([]models.TokenBlacklist, error)

	// CountUserBlacklistedTokens counts blacklisted tokens for a user
	CountUserBlacklistedTokens(userID uint) (int64, error)
	CountUserBlacklistedTokensWithContext(ctx context.Context, userID uint) (int64, error)
}

// TransactionManager defines the interface for database transaction management
type TransactionManager interface {
	// WithTransaction executes a function within a database transaction
	// If the function returns an error, the transaction is rolled back
	// Otherwise, the transaction is committed
	WithTransaction(fn func(*gorm.DB) error) error

	// WithTransactionContext executes a function within a database transaction with context
	WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error

	// BeginTransaction starts a new database transaction
	BeginTransaction() (*gorm.DB, error)

	// BeginTransactionWithContext starts a new database transaction with context
	BeginTransactionWithContext(ctx context.Context) (*gorm.DB, error)

	// CommitTransaction commits the given transaction
	CommitTransaction(tx *gorm.DB) error

	// RollbackTransaction rolls back the given transaction
	RollbackTransaction(tx *gorm.DB) error

	// WithSavepoint executes a function within a savepoint
	// If the function returns an error, the savepoint is rolled back
	WithSavepoint(tx *gorm.DB, name string, fn func(*gorm.DB) error) error

	// WithRetryableTransaction executes a function within a retryable transaction
	// Useful for handling deadlocks and temporary failures
	WithRetryableTransaction(fn func(*gorm.DB) error, maxRetries int) error

	// WithIsolationLevel executes a function within a transaction with specific isolation level
	WithIsolationLevel(isolationLevel string, fn func(*gorm.DB) error) error

	// WithIsolationLevelContext executes a function within a transaction with specific isolation level and context
	WithIsolationLevelContext(ctx context.Context, isolationLevel string, fn func(context.Context, *gorm.DB) error) error

	// WithTimeout executes a function within a transaction with timeout
	WithTimeout(timeout time.Duration, fn func(*gorm.DB) error) error

	// WithTimeoutContext executes a function within a transaction with timeout and context
	WithTimeoutContext(ctx context.Context, timeout time.Duration, fn func(context.Context, *gorm.DB) error) error

	// WithTimeoutAndRetry executes a function within a retryable transaction with timeout
	WithTimeoutAndRetry(timeout time.Duration, fn func(*gorm.DB) error, maxRetries int) error

	// GetIsolationLevel returns the current transaction isolation level
	GetIsolationLevel(tx *gorm.DB) (string, error)

	// SetIsolationLevel sets the transaction isolation level
	SetIsolationLevel(tx *gorm.DB, isolationLevel string) error
}

// PasswordHistoryRepository defines the interface for password history operations
type PasswordHistoryRepository interface {
	BaseRepository[models.PasswordHistory]

	// Create adds a new password history entry
	CreatePasswordHistory(ctx context.Context, entry *models.PasswordHistory) error
	CreatePasswordHistoryTx(tx *gorm.DB, entry *models.PasswordHistory) error

	// GetRecentPasswords retrieves recent passwords for a user
	GetRecentPasswords(ctx context.Context, userID uint, limit int) ([]*models.PasswordHistory, error)
	GetRecentPasswordsTx(tx *gorm.DB, userID uint, limit int) ([]*models.PasswordHistory, error)

	// CleanupOldPasswords removes old password history entries
	CleanupOldPasswords(ctx context.Context, userID uint, keepCount int) error
	CleanupOldPasswordsTx(tx *gorm.DB, userID uint, keepCount int) error

	// DeleteByUserID removes all password history for a user
	DeleteByUserID(ctx context.Context, userID uint) error
	DeleteByUserIDTx(tx *gorm.DB, userID uint) error

	// CountUserPasswords counts password history entries for a user
	CountUserPasswords(ctx context.Context, userID uint) (int64, error)
	CountUserPasswordsTx(tx *gorm.DB, userID uint) (int64, error)
}
