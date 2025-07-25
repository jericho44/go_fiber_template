package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// UserModel represents the concrete user model that implements UserEntity interface
// In actual usage, this would be imported from the models package
type UserModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"unique;not null;size:255"`
	Password  string         `json:"-" gorm:"not null;size:255"`
	FirstName string         `json:"first_name" gorm:"size:100"`
	LastName  string         `json:"last_name" gorm:"size:100"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the User model
func (UserModel) TableName() string {
	return "users"
}

// BeforeCreate is a GORM hook that runs before creating a user
func (u *UserModel) BeforeCreate(tx *gorm.DB) error {
	u.Email = strings.ToLower(u.Email)
	return nil
}

// BeforeUpdate is a GORM hook that runs before updating a user
func (u *UserModel) BeforeUpdate(tx *gorm.DB) error {
	u.Email = strings.ToLower(u.Email)
	return nil
}

// Implement UserEntity interface methods
func (u *UserModel) GetID() uint             { return u.ID }
func (u *UserModel) GetEmail() string        { return u.Email }
func (u *UserModel) SetEmail(email string)   { u.Email = strings.ToLower(email) }
func (u *UserModel) GetFirstName() string    { return u.FirstName }
func (u *UserModel) GetLastName() string     { return u.LastName }
func (u *UserModel) GetIsActive() bool       { return u.IsActive }
func (u *UserModel) SetIsActive(active bool) { u.IsActive = active }

// userRepositoryImpl is the GORM-based implementation of UserRepository
type userRepositoryImpl struct {
	db *gorm.DB
}

// NewUserRepository creates a new instance of UserRepository
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepositoryImpl{
		db: db,
	}
}

// Helper function to convert UserModel to UserEntity interface
func (u *UserModel) toEntity() UserEntity {
	return u
}

// Helper function to convert UserEntity to UserModel
func fromEntity(entity UserEntity) *UserModel {
	return &UserModel{
		ID:        entity.GetID(),
		Email:     entity.GetEmail(),
		FirstName: entity.GetFirstName(),
		LastName:  entity.GetLastName(),
		IsActive:  entity.GetIsActive(),
	}
}

// BaseRepository implementation

// Create creates a new user
func (r *userRepositoryImpl) Create(entity *UserEntity) error {
	user := fromEntity(*entity)
	err := r.db.Create(user).Error
	if err != nil {
		return err
	}
	*entity = user.toEntity()
	return nil
}

// CreateTx creates a new user within a transaction
func (r *userRepositoryImpl) CreateTx(tx *gorm.DB, entity *UserEntity) error {
	user := fromEntity(*entity)
	err := tx.Create(user).Error
	if err != nil {
		return err
	}
	*entity = user.toEntity()
	return nil
}

// CreateWithContext creates a new user with context
func (r *userRepositoryImpl) CreateWithContext(ctx context.Context, entity *UserEntity) error {
	user := fromEntity(*entity)
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return err
	}
	*entity = user.toEntity()
	return nil
}

// GetByID retrieves a user by ID
func (r *userRepositoryImpl) GetByID(id uint) (UserEntity, error) {
	var user UserModel
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return user.toEntity(), nil
}

// GetByIDWithContext retrieves a user by ID with context
func (r *userRepositoryImpl) GetByIDWithContext(ctx context.Context, id uint) (UserEntity, error) {
	var user UserModel
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with ID %d not found", id)
		}
		return nil, err
	}
	return user.toEntity(), nil
}

// Update updates an existing user
func (r *userRepositoryImpl) Update(entity *UserEntity) error {
	user := fromEntity(*entity)
	return r.db.Save(user).Error
}

// UpdateTx updates an existing user within a transaction
func (r *userRepositoryImpl) UpdateTx(tx *gorm.DB, entity *UserEntity) error {
	user := fromEntity(*entity)
	return tx.Save(user).Error
}

// UpdateWithContext updates an existing user with context
func (r *userRepositoryImpl) UpdateWithContext(ctx context.Context, entity *UserEntity) error {
	user := fromEntity(*entity)
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete soft deletes a user
func (r *userRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&UserModel{}, id).Error
}

// DeleteTx soft deletes a user within a transaction
func (r *userRepositoryImpl) DeleteTx(tx *gorm.DB, id uint) error {
	return tx.Delete(&UserModel{}, id).Error
}

// DeleteWithContext soft deletes a user with context
func (r *userRepositoryImpl) DeleteWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&UserModel{}, id).Error
}

// List retrieves users with pagination
func (r *userRepositoryImpl) List(limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// ListWithContext retrieves users with pagination and context
func (r *userRepositoryImpl) ListWithContext(ctx context.Context, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// Count returns the total number of users
func (r *userRepositoryImpl) Count() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Count(&count).Error
	return count, err
}

// CountWithContext returns the total number of users with context
func (r *userRepositoryImpl) CountWithContext(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Count(&count).Error
	return count, err
}

// UserRepository specific methods

// GetByEmail retrieves a user by email
func (r *userRepositoryImpl) GetByEmail(email string) (UserEntity, error) {
	var user UserModel
	err := r.db.Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}
	return user.toEntity(), nil
}

// GetByEmailTx retrieves a user by email within a transaction
func (r *userRepositoryImpl) GetByEmailTx(tx *gorm.DB, email string) (UserEntity, error) {
	var user UserModel
	err := tx.Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}
	return user.toEntity(), nil
}

// GetByEmailWithContext retrieves a user by email with context
func (r *userRepositoryImpl) GetByEmailWithContext(ctx context.Context, email string) (UserEntity, error) {
	var user UserModel
	err := r.db.WithContext(ctx).Where("email = ?", strings.ToLower(email)).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with email %s not found", email)
		}
		return nil, err
	}
	return user.toEntity(), nil
}

// GetActiveUsers retrieves active users with pagination
func (r *userRepositoryImpl) GetActiveUsers(limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// GetActiveUsersTx retrieves active users with pagination within a transaction
func (r *userRepositoryImpl) GetActiveUsersTx(tx *gorm.DB, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := tx.Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// GetActiveUsersWithContext retrieves active users with pagination and context
func (r *userRepositoryImpl) GetActiveUsersWithContext(ctx context.Context, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// ExistsByEmail checks if a user exists by email
func (r *userRepositoryImpl) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("email = ?", strings.ToLower(email)).Count(&count).Error
	return count > 0, err
}

// ExistsByEmailTx checks if a user exists by email within a transaction
func (r *userRepositoryImpl) ExistsByEmailTx(tx *gorm.DB, email string) (bool, error) {
	var count int64
	err := tx.Model(&UserModel{}).Where("email = ?", strings.ToLower(email)).Count(&count).Error
	return count > 0, err
}

// ExistsByEmailWithContext checks if a user exists by email with context
func (r *userRepositoryImpl) ExistsByEmailWithContext(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("email = ?", strings.ToLower(email)).Count(&count).Error
	return count > 0, err
}

// Bulk operations

// CreateBatch creates multiple users in a single operation
func (r *userRepositoryImpl) CreateBatch(users []UserEntity) error {
	userModels := make([]UserModel, len(users))
	for i, entity := range users {
		userModels[i] = *fromEntity(entity)
	}
	return r.db.Create(&userModels).Error
}

// CreateBatchTx creates multiple users in a single operation within a transaction
func (r *userRepositoryImpl) CreateBatchTx(tx *gorm.DB, users []UserEntity) error {
	userModels := make([]UserModel, len(users))
	for i, entity := range users {
		userModels[i] = *fromEntity(entity)
	}
	return tx.Create(&userModels).Error
}

// CreateBatchWithContext creates multiple users in a single operation with context
func (r *userRepositoryImpl) CreateBatchWithContext(ctx context.Context, users []UserEntity) error {
	userModels := make([]UserModel, len(users))
	for i, entity := range users {
		userModels[i] = *fromEntity(entity)
	}
	return r.db.WithContext(ctx).Create(&userModels).Error
}

// UpdateBatch updates multiple users in a single operation
func (r *userRepositoryImpl) UpdateBatch(users []UserEntity) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, entity := range users {
			user := fromEntity(entity)
			if err := tx.Save(user).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// UpdateBatchTx updates multiple users in a single operation within a transaction
func (r *userRepositoryImpl) UpdateBatchTx(tx *gorm.DB, users []UserEntity) error {
	for _, entity := range users {
		user := fromEntity(entity)
		if err := tx.Save(user).Error; err != nil {
			return err
		}
	}
	return nil
}

// UpdateBatchWithContext updates multiple users in a single operation with context
func (r *userRepositoryImpl) UpdateBatchWithContext(ctx context.Context, users []UserEntity) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, entity := range users {
			user := fromEntity(entity)
			if err := tx.Save(user).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// Soft delete operations

// SoftDelete soft deletes a user (same as Delete for GORM with DeletedAt)
func (r *userRepositoryImpl) SoftDelete(id uint) error {
	return r.Delete(id)
}

// SoftDeleteTx soft deletes a user within a transaction
func (r *userRepositoryImpl) SoftDeleteTx(tx *gorm.DB, id uint) error {
	return r.DeleteTx(tx, id)
}

// SoftDeleteWithContext soft deletes a user with context
func (r *userRepositoryImpl) SoftDeleteWithContext(ctx context.Context, id uint) error {
	return r.DeleteWithContext(ctx, id)
}

// Restore restores a soft-deleted user
func (r *userRepositoryImpl) Restore(id uint) error {
	return r.db.Unscoped().Model(&UserModel{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

// RestoreTx restores a soft-deleted user within a transaction
func (r *userRepositoryImpl) RestoreTx(tx *gorm.DB, id uint) error {
	return tx.Unscoped().Model(&UserModel{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

// RestoreWithContext restores a soft-deleted user with context
func (r *userRepositoryImpl) RestoreWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Model(&UserModel{}).Where("id = ?", id).Update("deleted_at", nil).Error
}

// GetDeletedUsers retrieves soft-deleted users
func (r *userRepositoryImpl) GetDeletedUsers(limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// GetDeletedUsersTx retrieves soft-deleted users within a transaction
func (r *userRepositoryImpl) GetDeletedUsersTx(tx *gorm.DB, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := tx.Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// GetDeletedUsersWithContext retrieves soft-deleted users with context
func (r *userRepositoryImpl) GetDeletedUsersWithContext(ctx context.Context, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// HardDelete permanently deletes a user
func (r *userRepositoryImpl) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&UserModel{}, id).Error
}

// HardDeleteTx permanently deletes a user within a transaction
func (r *userRepositoryImpl) HardDeleteTx(tx *gorm.DB, id uint) error {
	return tx.Unscoped().Delete(&UserModel{}, id).Error
}

// HardDeleteWithContext permanently deletes a user with context
func (r *userRepositoryImpl) HardDeleteWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Unscoped().Delete(&UserModel{}, id).Error
}

// Search operations

// SearchByName searches users by first name and/or last name
func (r *userRepositoryImpl) SearchByName(firstName, lastName string, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	query := r.db.Model(&UserModel{})

	if firstName != "" && lastName != "" {
		query = query.Where("first_name ILIKE ? AND last_name ILIKE ?", "%"+firstName+"%", "%"+lastName+"%")
	} else if firstName != "" {
		query = query.Where("first_name ILIKE ?", "%"+firstName+"%")
	} else if lastName != "" {
		query = query.Where("last_name ILIKE ?", "%"+lastName+"%")
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// SearchByNameTx searches users by first name and/or last name within a transaction
func (r *userRepositoryImpl) SearchByNameTx(tx *gorm.DB, firstName, lastName string, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	query := tx.Model(&UserModel{})

	if firstName != "" && lastName != "" {
		query = query.Where("first_name ILIKE ? AND last_name ILIKE ?", "%"+firstName+"%", "%"+lastName+"%")
	} else if firstName != "" {
		query = query.Where("first_name ILIKE ?", "%"+firstName+"%")
	} else if lastName != "" {
		query = query.Where("last_name ILIKE ?", "%"+lastName+"%")
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// SearchByNameWithContext searches users by first name and/or last name with context
func (r *userRepositoryImpl) SearchByNameWithContext(ctx context.Context, firstName, lastName string, limit, offset int) ([]UserEntity, error) {
	var users []UserModel
	query := r.db.WithContext(ctx).Model(&UserModel{})

	if firstName != "" && lastName != "" {
		query = query.Where("first_name ILIKE ? AND last_name ILIKE ?", "%"+firstName+"%", "%"+lastName+"%")
	} else if firstName != "" {
		query = query.Where("first_name ILIKE ?", "%"+firstName+"%")
	} else if lastName != "" {
		query = query.Where("last_name ILIKE ?", "%"+lastName+"%")
	}

	err := query.Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}

	entities := make([]UserEntity, len(users))
	for i, user := range users {
		entities[i] = user.toEntity()
	}
	return entities, nil
}

// Count operations

// CountActiveUsers returns the count of active users
func (r *userRepositoryImpl) CountActiveUsers() (int64, error) {
	var count int64
	err := r.db.Model(&UserModel{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CountActiveUsersTx returns the count of active users within a transaction
func (r *userRepositoryImpl) CountActiveUsersTx(tx *gorm.DB) (int64, error) {
	var count int64
	err := tx.Model(&UserModel{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CountActiveUsersWithContext returns the count of active users with context
func (r *userRepositoryImpl) CountActiveUsersWithContext(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&UserModel{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// CountDeletedUsers returns the count of soft-deleted users
func (r *userRepositoryImpl) CountDeletedUsers() (int64, error) {
	var count int64
	err := r.db.Unscoped().Model(&UserModel{}).Where("deleted_at IS NOT NULL").Count(&count).Error
	return count, err
}

// CountDeletedUsersTx returns the count of soft-deleted users within a transaction
func (r *userRepositoryImpl) CountDeletedUsersTx(tx *gorm.DB) (int64, error) {
	var count int64
	err := tx.Unscoped().Model(&UserModel{}).Where("deleted_at IS NOT NULL").Count(&count).Error
	return count, err
}

// CountDeletedUsersWithContext returns the count of soft-deleted users with context
func (r *userRepositoryImpl) CountDeletedUsersWithContext(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&UserModel{}).Where("deleted_at IS NOT NULL").Count(&count).Error
	return count, err
}
