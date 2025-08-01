package repositories

import (
	"context"
	"fmt"

	"go-fiber-template/internal/models"

	"gorm.io/gorm"
)

// passwordHistoryRepositoryImpl implements PasswordHistoryRepository
type passwordHistoryRepositoryImpl struct {
	db *gorm.DB
}

// NewPasswordHistoryRepository creates a new password history repository
func NewPasswordHistoryRepository(db *gorm.DB) PasswordHistoryRepository {
	return &passwordHistoryRepositoryImpl{
		db: db,
	}
}

// BaseRepository implementation

// Create creates a new password history entry
func (r *passwordHistoryRepositoryImpl) Create(entity *models.PasswordHistory) error {
	return r.db.Create(entity).Error
}

// GetByID retrieves a password history entry by ID
func (r *passwordHistoryRepositoryImpl) GetByID(id uint) (models.PasswordHistory, error) {
	var entity models.PasswordHistory
	err := r.db.First(&entity, id).Error
	return entity, err
}

// Update updates a password history entry
func (r *passwordHistoryRepositoryImpl) Update(entity *models.PasswordHistory) error {
	return r.db.Save(entity).Error
}

// Delete deletes a password history entry by ID
func (r *passwordHistoryRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.PasswordHistory{}, id).Error
}

// List retrieves password history entries with pagination
func (r *passwordHistoryRepositoryImpl) List(limit, offset int) ([]models.PasswordHistory, error) {
	var entities []models.PasswordHistory
	err := r.db.Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

// Count counts all password history entries
func (r *passwordHistoryRepositoryImpl) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.PasswordHistory{}).Count(&count).Error
	return count, err
}

// Transactional operations

// CreateTx creates a new password history entry within a transaction
func (r *passwordHistoryRepositoryImpl) CreateTx(tx *gorm.DB, entity *models.PasswordHistory) error {
	return tx.Create(entity).Error
}

// UpdateTx updates a password history entry within a transaction
func (r *passwordHistoryRepositoryImpl) UpdateTx(tx *gorm.DB, entity *models.PasswordHistory) error {
	return tx.Save(entity).Error
}

// DeleteTx deletes a password history entry by ID within a transaction
func (r *passwordHistoryRepositoryImpl) DeleteTx(tx *gorm.DB, id uint) error {
	return tx.Delete(&models.PasswordHistory{}, id).Error
}

// Context-aware operations

// CreateWithContext creates a new password history entry with context
func (r *passwordHistoryRepositoryImpl) CreateWithContext(ctx context.Context, entity *models.PasswordHistory) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByIDWithContext retrieves a password history entry by ID with context
func (r *passwordHistoryRepositoryImpl) GetByIDWithContext(ctx context.Context, id uint) (models.PasswordHistory, error) {
	var entity models.PasswordHistory
	err := r.db.WithContext(ctx).First(&entity, id).Error
	return entity, err
}

// UpdateWithContext updates a password history entry with context
func (r *passwordHistoryRepositoryImpl) UpdateWithContext(ctx context.Context, entity *models.PasswordHistory) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// DeleteWithContext deletes a password history entry by ID with context
func (r *passwordHistoryRepositoryImpl) DeleteWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.PasswordHistory{}, id).Error
}

// ListWithContext retrieves password history entries with pagination and context
func (r *passwordHistoryRepositoryImpl) ListWithContext(ctx context.Context, limit, offset int) ([]models.PasswordHistory, error) {
	var entities []models.PasswordHistory
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

// CountWithContext counts all password history entries with context
func (r *passwordHistoryRepositoryImpl) CountWithContext(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.PasswordHistory{}).Count(&count).Error
	return count, err
}

// Password history specific operations

// CreatePasswordHistory adds a new password history entry
func (r *passwordHistoryRepositoryImpl) CreatePasswordHistory(ctx context.Context, entry *models.PasswordHistory) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

// CreatePasswordHistoryTx adds a new password history entry within a transaction
func (r *passwordHistoryRepositoryImpl) CreatePasswordHistoryTx(tx *gorm.DB, entry *models.PasswordHistory) error {
	return tx.Create(entry).Error
}

// GetRecentPasswords retrieves recent passwords for a user
func (r *passwordHistoryRepositoryImpl) GetRecentPasswords(ctx context.Context, userID uint, limit int) ([]*models.PasswordHistory, error) {
	var history []*models.PasswordHistory

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent passwords: %w", err)
	}

	return history, nil
}

// GetRecentPasswordsTx retrieves recent passwords for a user within a transaction
func (r *passwordHistoryRepositoryImpl) GetRecentPasswordsTx(tx *gorm.DB, userID uint, limit int) ([]*models.PasswordHistory, error) {
	var history []*models.PasswordHistory

	err := tx.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent passwords: %w", err)
	}

	return history, nil
}

// CleanupOldPasswords removes old password history entries
func (r *passwordHistoryRepositoryImpl) CleanupOldPasswords(ctx context.Context, userID uint, keepCount int) error {
	// Get the IDs of entries to keep (most recent ones)
	var keepIDs []uint
	err := r.db.WithContext(ctx).
		Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(keepCount).
		Pluck("id", &keepIDs).Error

	if err != nil {
		return fmt.Errorf("failed to get IDs to keep: %w", err)
	}

	// Delete entries not in the keep list
	if len(keepIDs) > 0 {
		err = r.db.WithContext(ctx).
			Where("user_id = ? AND id NOT IN ?", userID, keepIDs).
			Delete(&models.PasswordHistory{}).Error
	} else {
		// If no entries to keep, delete all
		err = r.db.WithContext(ctx).
			Where("user_id = ?", userID).
			Delete(&models.PasswordHistory{}).Error
	}

	if err != nil {
		return fmt.Errorf("failed to cleanup old passwords: %w", err)
	}

	return nil
}

// CleanupOldPasswordsTx removes old password history entries within a transaction
func (r *passwordHistoryRepositoryImpl) CleanupOldPasswordsTx(tx *gorm.DB, userID uint, keepCount int) error {
	// Get the IDs of entries to keep (most recent ones)
	var keepIDs []uint
	err := tx.Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(keepCount).
		Pluck("id", &keepIDs).Error

	if err != nil {
		return fmt.Errorf("failed to get IDs to keep: %w", err)
	}

	// Delete entries not in the keep list
	if len(keepIDs) > 0 {
		err = tx.Where("user_id = ? AND id NOT IN ?", userID, keepIDs).
			Delete(&models.PasswordHistory{}).Error
	} else {
		// If no entries to keep, delete all
		err = tx.Where("user_id = ?", userID).
			Delete(&models.PasswordHistory{}).Error
	}

	if err != nil {
		return fmt.Errorf("failed to cleanup old passwords: %w", err)
	}

	return nil
}

// DeleteByUserID removes all password history for a user
func (r *passwordHistoryRepositoryImpl) DeleteByUserID(ctx context.Context, userID uint) error {
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Delete(&models.PasswordHistory{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete password history for user: %w", err)
	}

	return nil
}

// DeleteByUserIDTx removes all password history for a user within a transaction
func (r *passwordHistoryRepositoryImpl) DeleteByUserIDTx(tx *gorm.DB, userID uint) error {
	err := tx.Where("user_id = ?", userID).
		Delete(&models.PasswordHistory{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete password history for user: %w", err)
	}

	return nil
}

// CountUserPasswords counts password history entries for a user
func (r *passwordHistoryRepositoryImpl) CountUserPasswords(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count user passwords: %w", err)
	}

	return count, nil
}

// CountUserPasswordsTx counts password history entries for a user within a transaction
func (r *passwordHistoryRepositoryImpl) CountUserPasswordsTx(tx *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := tx.Model(&models.PasswordHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	if err != nil {
		return 0, fmt.Errorf("failed to count user passwords: %w", err)
	}

	return count, nil
}
