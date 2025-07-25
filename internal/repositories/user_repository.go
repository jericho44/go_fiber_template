package repositories

import (
	"context"

	"gorm.io/gorm"
)

// UserEntity represents the user model interface for repository operations
// This allows the repository to work with any type that implements these methods
type UserEntity interface {
	GetID() uint
	GetEmail() string
	SetEmail(string)
	GetFirstName() string
	GetLastName() string
	GetIsActive() bool
	SetIsActive(bool)
}

// UserRepository defines the interface for user-specific database operations
type UserRepository interface {
	// Embed the base repository interface for common operations
	BaseRepository[UserEntity]

	// User-specific query methods
	GetByEmail(email string) (UserEntity, error)
	GetByEmailTx(tx *gorm.DB, email string) (UserEntity, error)
	GetByEmailWithContext(ctx context.Context, email string) (UserEntity, error)

	// User filtering and search operations
	GetActiveUsers(limit, offset int) ([]UserEntity, error)
	GetActiveUsersTx(tx *gorm.DB, limit, offset int) ([]UserEntity, error)
	GetActiveUsersWithContext(ctx context.Context, limit, offset int) ([]UserEntity, error)

	// User existence checks
	ExistsByEmail(email string) (bool, error)
	ExistsByEmailTx(tx *gorm.DB, email string) (bool, error)
	ExistsByEmailWithContext(ctx context.Context, email string) (bool, error)

	// Bulk operations
	CreateBatch(users []UserEntity) error
	CreateBatchTx(tx *gorm.DB, users []UserEntity) error
	CreateBatchWithContext(ctx context.Context, users []UserEntity) error

	UpdateBatch(users []UserEntity) error
	UpdateBatchTx(tx *gorm.DB, users []UserEntity) error
	UpdateBatchWithContext(ctx context.Context, users []UserEntity) error

	// Soft delete operations (since User model has DeletedAt)
	SoftDelete(id uint) error
	SoftDeleteTx(tx *gorm.DB, id uint) error
	SoftDeleteWithContext(ctx context.Context, id uint) error

	// Restore soft deleted users
	Restore(id uint) error
	RestoreTx(tx *gorm.DB, id uint) error
	RestoreWithContext(ctx context.Context, id uint) error

	// Get soft deleted users
	GetDeletedUsers(limit, offset int) ([]UserEntity, error)
	GetDeletedUsersTx(tx *gorm.DB, limit, offset int) ([]UserEntity, error)
	GetDeletedUsersWithContext(ctx context.Context, limit, offset int) ([]UserEntity, error)

	// Hard delete (permanent deletion)
	HardDelete(id uint) error
	HardDeleteTx(tx *gorm.DB, id uint) error
	HardDeleteWithContext(ctx context.Context, id uint) error

	// Search operations
	SearchByName(firstName, lastName string, limit, offset int) ([]UserEntity, error)
	SearchByNameTx(tx *gorm.DB, firstName, lastName string, limit, offset int) ([]UserEntity, error)
	SearchByNameWithContext(ctx context.Context, firstName, lastName string, limit, offset int) ([]UserEntity, error)

	// Count operations
	CountActiveUsers() (int64, error)
	CountActiveUsersTx(tx *gorm.DB) (int64, error)
	CountActiveUsersWithContext(ctx context.Context) (int64, error)

	CountDeletedUsers() (int64, error)
	CountDeletedUsersTx(tx *gorm.DB) (int64, error)
	CountDeletedUsersWithContext(ctx context.Context) (int64, error)
}
