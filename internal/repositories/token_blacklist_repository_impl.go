package repositories

import (
	"context"
	"time"

	"go-fiber-template/internal/models"

	"gorm.io/gorm"
)

// TokenBlacklistRepositoryImpl implements the TokenBlacklistRepository interface
type TokenBlacklistRepositoryImpl struct {
	db *gorm.DB
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db *gorm.DB) TokenBlacklistRepository {
	return &TokenBlacklistRepositoryImpl{db: db}
}

// Create creates a new token blacklist entry
func (r *TokenBlacklistRepositoryImpl) Create(entity *models.TokenBlacklist) error {
	return r.db.Create(entity).Error
}

// CreateTx creates a new token blacklist entry within a transaction
func (r *TokenBlacklistRepositoryImpl) CreateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	return tx.Create(entity).Error
}

// CreateWithContext creates a new token blacklist entry with context
func (r *TokenBlacklistRepositoryImpl) CreateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	return r.db.WithContext(ctx).Create(entity).Error
}

// GetByID retrieves a token blacklist entry by ID
func (r *TokenBlacklistRepositoryImpl) GetByID(id uint) (models.TokenBlacklist, error) {
	var entity models.TokenBlacklist
	err := r.db.First(&entity, id).Error
	return entity, err
}

// GetByIDWithContext retrieves a token blacklist entry by ID with context
func (r *TokenBlacklistRepositoryImpl) GetByIDWithContext(ctx context.Context, id uint) (models.TokenBlacklist, error) {
	var entity models.TokenBlacklist
	err := r.db.WithContext(ctx).First(&entity, id).Error
	return entity, err
}

// Update updates a token blacklist entry
func (r *TokenBlacklistRepositoryImpl) Update(entity *models.TokenBlacklist) error {
	return r.db.Save(entity).Error
}

// UpdateTx updates a token blacklist entry within a transaction
func (r *TokenBlacklistRepositoryImpl) UpdateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	return tx.Save(entity).Error
}

// UpdateWithContext updates a token blacklist entry with context
func (r *TokenBlacklistRepositoryImpl) UpdateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	return r.db.WithContext(ctx).Save(entity).Error
}

// Delete deletes a token blacklist entry by ID
func (r *TokenBlacklistRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.TokenBlacklist{}, id).Error
}

// DeleteTx deletes a token blacklist entry by ID within a transaction
func (r *TokenBlacklistRepositoryImpl) DeleteTx(tx *gorm.DB, id uint) error {
	return tx.Delete(&models.TokenBlacklist{}, id).Error
}

// DeleteWithContext deletes a token blacklist entry by ID with context
func (r *TokenBlacklistRepositoryImpl) DeleteWithContext(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&models.TokenBlacklist{}, id).Error
}

// List retrieves token blacklist entries with pagination
func (r *TokenBlacklistRepositoryImpl) List(limit, offset int) ([]models.TokenBlacklist, error) {
	var entities []models.TokenBlacklist
	err := r.db.Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

// ListWithContext retrieves token blacklist entries with pagination and context
func (r *TokenBlacklistRepositoryImpl) ListWithContext(ctx context.Context, limit, offset int) ([]models.TokenBlacklist, error) {
	var entities []models.TokenBlacklist
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&entities).Error
	return entities, err
}

// Count counts the total number of token blacklist entries
func (r *TokenBlacklistRepositoryImpl) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.TokenBlacklist{}).Count(&count).Error
	return count, err
}

// CountWithContext counts the total number of token blacklist entries with context
func (r *TokenBlacklistRepositoryImpl) CountWithContext(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TokenBlacklist{}).Count(&count).Error
	return count, err
}

// IsTokenBlacklisted checks if a token hash is blacklisted
func (r *TokenBlacklistRepositoryImpl) IsTokenBlacklisted(tokenHash string) (bool, error) {
	var count int64
	err := r.db.Model(&models.TokenBlacklist{}).Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).Count(&count).Error
	return count > 0, err
}

// IsTokenBlacklistedWithContext checks if a token hash is blacklisted with context
func (r *TokenBlacklistRepositoryImpl) IsTokenBlacklistedWithContext(ctx context.Context, tokenHash string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TokenBlacklist{}).Where("token_hash = ? AND expires_at > ?", tokenHash, time.Now()).Count(&count).Error
	return count > 0, err
}

// BlacklistToken adds a token to the blacklist
func (r *TokenBlacklistRepositoryImpl) BlacklistToken(tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	blacklistEntry := &models.TokenBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	return r.Create(blacklistEntry)
}

// BlacklistTokenWithContext adds a token to the blacklist with context
func (r *TokenBlacklistRepositoryImpl) BlacklistTokenWithContext(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	blacklistEntry := &models.TokenBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	return r.CreateWithContext(ctx, blacklistEntry)
}

// BlacklistTokenTx adds a token to the blacklist within a transaction
func (r *TokenBlacklistRepositoryImpl) BlacklistTokenTx(tx *gorm.DB, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	blacklistEntry := &models.TokenBlacklist{
		TokenHash: tokenHash,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	return r.CreateTx(tx, blacklistEntry)
}

// GetBlacklistedToken retrieves a blacklisted token by hash
func (r *TokenBlacklistRepositoryImpl) GetBlacklistedToken(tokenHash string) (*models.TokenBlacklist, error) {
	var entity models.TokenBlacklist
	err := r.db.Where("token_hash = ?", tokenHash).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// GetBlacklistedTokenWithContext retrieves a blacklisted token by hash with context
func (r *TokenBlacklistRepositoryImpl) GetBlacklistedTokenWithContext(ctx context.Context, tokenHash string) (*models.TokenBlacklist, error) {
	var entity models.TokenBlacklist
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (r *TokenBlacklistRepositoryImpl) CleanupExpiredTokens() (int64, error) {
	result := r.db.Where("expires_at <= ?", time.Now()).Delete(&models.TokenBlacklist{})
	return result.RowsAffected, result.Error
}

// CleanupExpiredTokensWithContext removes expired tokens from the blacklist with context
func (r *TokenBlacklistRepositoryImpl) CleanupExpiredTokensWithContext(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&models.TokenBlacklist{})
	return result.RowsAffected, result.Error
}

// CleanupExpiredTokensTx removes expired tokens from the blacklist within a transaction
func (r *TokenBlacklistRepositoryImpl) CleanupExpiredTokensTx(tx *gorm.DB) (int64, error) {
	result := tx.Where("expires_at <= ?", time.Now()).Delete(&models.TokenBlacklist{})
	return result.RowsAffected, result.Error
}

// GetUserBlacklistedTokens retrieves all blacklisted tokens for a user
func (r *TokenBlacklistRepositoryImpl) GetUserBlacklistedTokens(userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	var entities []models.TokenBlacklist
	err := r.db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("blacklisted_at DESC").Find(&entities).Error
	return entities, err
}

// GetUserBlacklistedTokensWithContext retrieves all blacklisted tokens for a user with context
func (r *TokenBlacklistRepositoryImpl) GetUserBlacklistedTokensWithContext(ctx context.Context, userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	var entities []models.TokenBlacklist
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("blacklisted_at DESC").Find(&entities).Error
	return entities, err
}

// CountUserBlacklistedTokens counts blacklisted tokens for a user
func (r *TokenBlacklistRepositoryImpl) CountUserBlacklistedTokens(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&models.TokenBlacklist{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}

// CountUserBlacklistedTokensWithContext counts blacklisted tokens for a user with context
func (r *TokenBlacklistRepositoryImpl) CountUserBlacklistedTokensWithContext(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TokenBlacklist{}).Where("user_id = ?", userID).Count(&count).Error
	return count, err
}
