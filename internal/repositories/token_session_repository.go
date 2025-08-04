package repositories

import (
	"context"
	"time"

	"go-fiber-template/internal/models"

	"gorm.io/gorm"
)

// TokenSessionRepository defines the interface for token session operations
type TokenSessionRepository interface {
	// Session management
	CreateSession(ctx context.Context, session *models.TokenSession) error
	CreateSessionTx(tx *gorm.DB, session *models.TokenSession) error
	GetSessionByID(ctx context.Context, sessionID string) (*models.TokenSession, error)
	GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.TokenSession, error)
	UpdateSession(ctx context.Context, session *models.TokenSession) error
	UpdateSessionTx(tx *gorm.DB, session *models.TokenSession) error
	DeactivateSession(ctx context.Context, sessionID string) error
	DeactivateSessionTx(tx *gorm.DB, sessionID string) error
	DeactivateUserSessions(ctx context.Context, userID uint) error

	// Session queries
	GetActiveSessions(ctx context.Context, userID uint) ([]*models.TokenSession, error)
	GetExpiredSessions(ctx context.Context) ([]*models.TokenSession, error)
	CleanupExpiredSessions(ctx context.Context) (int64, error)

	// Token usage tracking
	RecordTokenUsage(ctx context.Context, usage *models.TokenUsage) error
	RecordTokenUsageTx(tx *gorm.DB, usage *models.TokenUsage) error
	GetTokenUsages(ctx context.Context, sessionID string, limit int) ([]*models.TokenUsage, error)
	GetSuspiciousUsages(ctx context.Context, sessionID string, since time.Time) ([]*models.TokenUsage, error)

	// Token rotation tracking
	RecordTokenRotation(ctx context.Context, rotation *models.TokenRotation) error
	RecordTokenRotationTx(tx *gorm.DB, rotation *models.TokenRotation) error
	GetTokenRotations(ctx context.Context, sessionID string, limit int) ([]*models.TokenRotation, error)

	// Anomaly detection
	GetRecentUsagesByIP(ctx context.Context, userID uint, ipAddress string, since time.Time) ([]*models.TokenUsage, error)
	GetRecentUsagesByUserAgent(ctx context.Context, userID uint, userAgent string, since time.Time) ([]*models.TokenUsage, error)
	CountSessionsByFingerprint(ctx context.Context, userID uint, fingerprint string) (int64, error)
}

// tokenSessionRepositoryImpl implements TokenSessionRepository
type tokenSessionRepositoryImpl struct {
	db *gorm.DB
}

// NewTokenSessionRepository creates a new token session repository
func NewTokenSessionRepository(db *gorm.DB) TokenSessionRepository {
	return &tokenSessionRepositoryImpl{
		db: db,
	}
}

// CreateSession creates a new token session
func (r *tokenSessionRepositoryImpl) CreateSession(ctx context.Context, session *models.TokenSession) error {
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return err
	}
	return nil
}

// CreateSessionTx creates a new token session within a transaction
func (r *tokenSessionRepositoryImpl) CreateSessionTx(tx *gorm.DB, session *models.TokenSession) error {
	if err := tx.Create(session).Error; err != nil {
		return err
	}
	return nil
}

// GetSessionByID retrieves a session by its ID
func (r *tokenSessionRepositoryImpl) GetSessionByID(ctx context.Context, sessionID string) (*models.TokenSession, error) {
	var session models.TokenSession
	if err := r.db.WithContext(ctx).Where("session_id = ? AND is_active = ?", sessionID, true).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token hash
func (r *tokenSessionRepositoryImpl) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.TokenSession, error) {
	var session models.TokenSession
	if err := r.db.WithContext(ctx).Where("refresh_token_hash = ? AND is_active = ?", refreshTokenHash, true).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateSession updates a token session
func (r *tokenSessionRepositoryImpl) UpdateSession(ctx context.Context, session *models.TokenSession) error {
	if err := r.db.WithContext(ctx).Save(session).Error; err != nil {
		return err
	}
	return nil
}

// UpdateSessionTx updates a token session within a transaction
func (r *tokenSessionRepositoryImpl) UpdateSessionTx(tx *gorm.DB, session *models.TokenSession) error {
	if err := tx.Save(session).Error; err != nil {
		return err
	}
	return nil
}

// DeactivateSession deactivates a session
func (r *tokenSessionRepositoryImpl) DeactivateSession(ctx context.Context, sessionID string) error {
	if err := r.db.WithContext(ctx).Model(&models.TokenSession{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error; err != nil {
		return err
	}
	return nil
}

// DeactivateSessionTx deactivates a session within a transaction
func (r *tokenSessionRepositoryImpl) DeactivateSessionTx(tx *gorm.DB, sessionID string) error {
	if err := tx.Model(&models.TokenSession{}).
		Where("session_id = ?", sessionID).
		Update("is_active", false).Error; err != nil {
		return err
	}
	return nil
}

// DeactivateUserSessions deactivates all sessions for a user
func (r *tokenSessionRepositoryImpl) DeactivateUserSessions(ctx context.Context, userID uint) error {
	if err := r.db.WithContext(ctx).Model(&models.TokenSession{}).
		Where("user_id = ?", userID).
		Update("is_active", false).Error; err != nil {
		return err
	}
	return nil
}

// GetActiveSessions retrieves all active sessions for a user
func (r *tokenSessionRepositoryImpl) GetActiveSessions(ctx context.Context, userID uint) ([]*models.TokenSession, error) {
	var sessions []*models.TokenSession
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_active = ? AND expires_at > ?", userID, true, time.Now()).
		Order("last_used_at DESC").
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// GetExpiredSessions retrieves expired sessions
func (r *tokenSessionRepositoryImpl) GetExpiredSessions(ctx context.Context) ([]*models.TokenSession, error) {
	var sessions []*models.TokenSession
	if err := r.db.WithContext(ctx).
		Where("expires_at <= ? OR is_active = ?", time.Now(), false).
		Find(&sessions).Error; err != nil {
		return nil, err
	}
	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions
func (r *tokenSessionRepositoryImpl) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("expires_at <= ? OR is_active = ?", time.Now(), false).
		Delete(&models.TokenSession{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// RecordTokenUsage records token usage
func (r *tokenSessionRepositoryImpl) RecordTokenUsage(ctx context.Context, usage *models.TokenUsage) error {
	if err := r.db.WithContext(ctx).Create(usage).Error; err != nil {
		return err
	}
	return nil
}

// RecordTokenUsageTx records token usage within a transaction
func (r *tokenSessionRepositoryImpl) RecordTokenUsageTx(tx *gorm.DB, usage *models.TokenUsage) error {
	if err := tx.Create(usage).Error; err != nil {
		return err
	}
	return nil
}

// GetTokenUsages retrieves token usages for a session
func (r *tokenSessionRepositoryImpl) GetTokenUsages(ctx context.Context, sessionID string, limit int) ([]*models.TokenUsage, error) {
	var usages []*models.TokenUsage
	query := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("used_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// GetSuspiciousUsages retrieves suspicious token usages
func (r *tokenSessionRepositoryImpl) GetSuspiciousUsages(ctx context.Context, sessionID string, since time.Time) ([]*models.TokenUsage, error) {
	var usages []*models.TokenUsage
	if err := r.db.WithContext(ctx).
		Where("session_id = ? AND is_anomaly = ? AND used_at >= ?", sessionID, true, since).
		Order("used_at DESC").
		Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// RecordTokenRotation records token rotation
func (r *tokenSessionRepositoryImpl) RecordTokenRotation(ctx context.Context, rotation *models.TokenRotation) error {
	if err := r.db.WithContext(ctx).Create(rotation).Error; err != nil {
		return err
	}
	return nil
}

// RecordTokenRotationTx records token rotation within a transaction
func (r *tokenSessionRepositoryImpl) RecordTokenRotationTx(tx *gorm.DB, rotation *models.TokenRotation) error {
	if err := tx.Create(rotation).Error; err != nil {
		return err
	}
	return nil
}

// GetTokenRotations retrieves token rotations for a session
func (r *tokenSessionRepositoryImpl) GetTokenRotations(ctx context.Context, sessionID string, limit int) ([]*models.TokenRotation, error) {
	var rotations []*models.TokenRotation
	query := r.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("rotated_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&rotations).Error; err != nil {
		return nil, err
	}
	return rotations, nil
}

// GetRecentUsagesByIP retrieves recent usages by IP address
func (r *tokenSessionRepositoryImpl) GetRecentUsagesByIP(ctx context.Context, userID uint, ipAddress string, since time.Time) ([]*models.TokenUsage, error) {
	var usages []*models.TokenUsage
	if err := r.db.WithContext(ctx).
		Joins("JOIN token_sessions ON token_usages.session_id = token_sessions.session_id").
		Where("token_sessions.user_id = ? AND token_usages.ip_address = ? AND token_usages.used_at >= ?", userID, ipAddress, since).
		Order("token_usages.used_at DESC").
		Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// GetRecentUsagesByUserAgent retrieves recent usages by user agent
func (r *tokenSessionRepositoryImpl) GetRecentUsagesByUserAgent(ctx context.Context, userID uint, userAgent string, since time.Time) ([]*models.TokenUsage, error) {
	var usages []*models.TokenUsage
	if err := r.db.WithContext(ctx).
		Joins("JOIN token_sessions ON token_usages.session_id = token_sessions.session_id").
		Where("token_sessions.user_id = ? AND token_usages.user_agent = ? AND token_usages.used_at >= ?", userID, userAgent, since).
		Order("token_usages.used_at DESC").
		Find(&usages).Error; err != nil {
		return nil, err
	}
	return usages, nil
}

// CountSessionsByFingerprint counts sessions by device fingerprint
func (r *tokenSessionRepositoryImpl) CountSessionsByFingerprint(ctx context.Context, userID uint, fingerprint string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.TokenSession{}).
		Where("user_id = ? AND device_fingerprint = ? AND is_active = ?", userID, fingerprint, true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
