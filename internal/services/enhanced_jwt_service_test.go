package services

import (
	"context"
	"testing"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Mock implementations for testing
type mockTokenSessionRepository struct {
	mock.Mock
}

func (m *mockTokenSessionRepository) CreateSession(ctx context.Context, session *models.TokenSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) CreateSessionTx(tx *gorm.DB, session *models.TokenSession) error {
	args := m.Called(tx, session)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) GetSessionByID(ctx context.Context, sessionID string) (*models.TokenSession, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*models.TokenSession), args.Error(1)
}

func (m *mockTokenSessionRepository) GetSessionByRefreshToken(ctx context.Context, refreshTokenHash string) (*models.TokenSession, error) {
	args := m.Called(ctx, refreshTokenHash)
	return args.Get(0).(*models.TokenSession), args.Error(1)
}

func (m *mockTokenSessionRepository) UpdateSession(ctx context.Context, session *models.TokenSession) error {
	args := m.Called(ctx, session)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) UpdateSessionTx(tx *gorm.DB, session *models.TokenSession) error {
	args := m.Called(tx, session)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) DeactivateSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) DeactivateSessionTx(tx *gorm.DB, sessionID string) error {
	args := m.Called(tx, sessionID)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) DeactivateUserSessions(ctx context.Context, userID uint) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) GetActiveSessions(ctx context.Context, userID uint) ([]*models.TokenSession, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.TokenSession), args.Error(1)
}

func (m *mockTokenSessionRepository) GetExpiredSessions(ctx context.Context) ([]*models.TokenSession, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*models.TokenSession), args.Error(1)
}

func (m *mockTokenSessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenSessionRepository) RecordTokenUsage(ctx context.Context, usage *models.TokenUsage) error {
	args := m.Called(ctx, usage)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) RecordTokenUsageTx(tx *gorm.DB, usage *models.TokenUsage) error {
	args := m.Called(tx, usage)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) GetTokenUsages(ctx context.Context, sessionID string, limit int) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, sessionID, limit)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *mockTokenSessionRepository) GetSuspiciousUsages(ctx context.Context, sessionID string, since time.Time) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, sessionID, since)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *mockTokenSessionRepository) RecordTokenRotation(ctx context.Context, rotation *models.TokenRotation) error {
	args := m.Called(ctx, rotation)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) RecordTokenRotationTx(tx *gorm.DB, rotation *models.TokenRotation) error {
	args := m.Called(tx, rotation)
	return args.Error(0)
}

func (m *mockTokenSessionRepository) GetTokenRotations(ctx context.Context, sessionID string, limit int) ([]*models.TokenRotation, error) {
	args := m.Called(ctx, sessionID, limit)
	return args.Get(0).([]*models.TokenRotation), args.Error(1)
}

func (m *mockTokenSessionRepository) GetRecentUsagesByIP(ctx context.Context, userID uint, ipAddress string, since time.Time) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, userID, ipAddress, since)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *mockTokenSessionRepository) GetRecentUsagesByUserAgent(ctx context.Context, userID uint, userAgent string, since time.Time) ([]*models.TokenUsage, error) {
	args := m.Called(ctx, userID, userAgent, since)
	return args.Get(0).([]*models.TokenUsage), args.Error(1)
}

func (m *mockTokenSessionRepository) CountSessionsByFingerprint(ctx context.Context, userID uint, fingerprint string) (int64, error) {
	args := m.Called(ctx, userID, fingerprint)
	return args.Get(0).(int64), args.Error(1)
}

type mockRedisService struct {
	mock.Mock
}

func (m *mockRedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockRedisService) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockRedisService) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *mockRedisService) Exists(ctx context.Context, keys ...string) (int64, error) {
	args := m.Called(ctx, keys)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisService) BlacklistToken(ctx context.Context, tokenHash string, expiration time.Duration) error {
	args := m.Called(ctx, tokenHash, expiration)
	return args.Error(0)
}

func (m *mockRedisService) IsTokenBlacklisted(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *mockRedisService) BlacklistTokenWithMetadata(ctx context.Context, tokenHash string, metadata map[string]interface{}, expiration time.Duration) error {
	args := m.Called(ctx, tokenHash, metadata, expiration)
	return args.Error(0)
}

func (m *mockRedisService) SetSession(ctx context.Context, sessionID string, data interface{}, expiration time.Duration) error {
	args := m.Called(ctx, sessionID, data, expiration)
	return args.Error(0)
}

func (m *mockRedisService) GetSession(ctx context.Context, sessionID string, dest interface{}) error {
	args := m.Called(ctx, sessionID, dest)
	return args.Error(0)
}

func (m *mockRedisService) DeleteSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *mockRedisService) ExtendSessionExpiry(ctx context.Context, sessionID string, expiration time.Duration) error {
	args := m.Called(ctx, sessionID, expiration)
	return args.Error(0)
}

func (m *mockRedisService) TrackTokenUsage(ctx context.Context, tokenHash string, metadata map[string]interface{}) error {
	args := m.Called(ctx, tokenHash, metadata)
	return args.Error(0)
}

func (m *mockRedisService) GetTokenUsageStats(ctx context.Context, tokenHash string) (map[string]interface{}, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *mockRedisService) IncrementCounter(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	args := m.Called(ctx, key, expiration)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisService) GetCounter(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRedisService) SetCounterWithExpiry(ctx context.Context, key string, value int64, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockRedisService) RecordSecurityEvent(ctx context.Context, eventType string, metadata map[string]interface{}, expiration time.Duration) error {
	args := m.Called(ctx, eventType, metadata, expiration)
	return args.Error(0)
}

func (m *mockRedisService) GetSecurityEvents(ctx context.Context, eventType string, limit int) ([]map[string]interface{}, error) {
	args := m.Called(ctx, eventType, limit)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *mockRedisService) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	args := m.Called(ctx, key, limit, window)
	return args.Bool(0), args.Get(1).(int64), args.Error(2)
}

func (m *mockRedisService) ResetRateLimit(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockRedisService) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRedisService) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockAnomalyDetector struct {
	mock.Mock
}

func (m *mockAnomalyDetector) AnalyzeTokenUsage(ctx context.Context, usage *models.TokenUsage) (*AnomalyResult, error) {
	args := m.Called(ctx, usage)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckIPAnomalies(ctx context.Context, userID uint, ipAddress string) (*AnomalyResult, error) {
	args := m.Called(ctx, userID, ipAddress)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckUserAgentAnomalies(ctx context.Context, userID uint, userAgent string) (*AnomalyResult, error) {
	args := m.Called(ctx, userID, userAgent)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckLocationAnomalies(ctx context.Context, userID uint, location string) (*AnomalyResult, error) {
	args := m.Called(ctx, userID, location)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckFrequencyAnomalies(ctx context.Context, sessionID string) (*AnomalyResult, error) {
	args := m.Called(ctx, sessionID)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckDeviceFingerprintAnomalies(ctx context.Context, userID uint, fingerprint string) (*AnomalyResult, error) {
	args := m.Called(ctx, userID, fingerprint)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

func (m *mockAnomalyDetector) CheckTimeBasedAnomalies(ctx context.Context, userID uint, timestamp time.Time) (*AnomalyResult, error) {
	args := m.Called(ctx, userID, timestamp)
	return args.Get(0).(*AnomalyResult), args.Error(1)
}

type mockTransactionManager struct {
	mock.Mock
}

func (m *mockTransactionManager) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	// Execute the function with a mock transaction
	err := fn(&gorm.DB{})
	if err != nil {
		return err
	}
	return args.Error(0)
}

func (m *mockTransactionManager) WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, fn)
	// Execute the function with a mock transaction
	err := fn(ctx, &gorm.DB{})
	if err != nil {
		return err
	}
	return args.Error(0)
}

func (m *mockTransactionManager) BeginTransaction() (*gorm.DB, error) {
	args := m.Called()
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *mockTransactionManager) BeginTransactionWithContext(ctx context.Context) (*gorm.DB, error) {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *mockTransactionManager) CommitTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *mockTransactionManager) RollbackTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *mockTransactionManager) WithSavepoint(tx *gorm.DB, name string, fn func(*gorm.DB) error) error {
	args := m.Called(tx, name, fn)
	return args.Error(0)
}

func (m *mockTransactionManager) WithRetryableTransaction(fn func(*gorm.DB) error, maxRetries int) error {
	args := m.Called(fn, maxRetries)
	return args.Error(0)
}

func (m *mockTransactionManager) WithIsolationLevel(isolationLevel string, fn func(*gorm.DB) error) error {
	args := m.Called(isolationLevel, fn)
	return args.Error(0)
}

func (m *mockTransactionManager) WithIsolationLevelContext(ctx context.Context, isolationLevel string, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, isolationLevel, fn)
	return args.Error(0)
}

func (m *mockTransactionManager) WithTimeout(timeout time.Duration, fn func(*gorm.DB) error) error {
	args := m.Called(timeout, fn)
	return args.Error(0)
}

func (m *mockTransactionManager) WithTimeoutContext(ctx context.Context, timeout time.Duration, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, timeout, fn)
	return args.Error(0)
}

func (m *mockTransactionManager) WithTimeoutAndRetry(timeout time.Duration, fn func(*gorm.DB) error, maxRetries int) error {
	args := m.Called(timeout, fn, maxRetries)
	return args.Error(0)
}

func (m *mockTransactionManager) GetIsolationLevel(tx *gorm.DB) (string, error) {
	args := m.Called(tx)
	return args.String(0), args.Error(1)
}

func (m *mockTransactionManager) SetIsolationLevel(tx *gorm.DB, isolationLevel string) error {
	args := m.Called(tx, isolationLevel)
	return args.Error(0)
}

type mockTokenBlacklistRepository struct {
	mock.Mock
}

// BaseRepository methods
func (m *mockTokenBlacklistRepository) Create(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) GetByID(id uint) (models.TokenBlacklist, error) {
	args := m.Called(id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) Update(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) List(limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CreateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) UpdateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) DeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) CreateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) GetByIDWithContext(ctx context.Context, id uint) (models.TokenBlacklist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) UpdateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) DeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) ListWithContext(ctx context.Context, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CountWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// TokenBlacklistRepository specific methods
func (m *mockTokenBlacklistRepository) IsTokenBlacklisted(tokenHash string) (bool, error) {
	args := m.Called(tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *mockTokenBlacklistRepository) IsTokenBlacklistedWithContext(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *mockTokenBlacklistRepository) BlacklistToken(tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) BlacklistTokenWithContext(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(ctx, tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) BlacklistTokenTx(tx *gorm.DB, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(tx, tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *mockTokenBlacklistRepository) GetBlacklistedToken(tokenHash string) (*models.TokenBlacklist, error) {
	args := m.Called(tokenHash)
	return args.Get(0).(*models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) GetBlacklistedTokenWithContext(ctx context.Context, tokenHash string) (*models.TokenBlacklist, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CleanupExpiredTokens() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CleanupExpiredTokensWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CleanupExpiredTokensTx(tx *gorm.DB) (int64, error) {
	args := m.Called(tx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenBlacklistRepository) GetUserBlacklistedTokens(userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) GetUserBlacklistedTokensWithContext(ctx context.Context, userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CountUserBlacklistedTokens(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockTokenBlacklistRepository) CountUserBlacklistedTokensWithContext(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// Test cases
func TestEnhancedJWTService_GenerateTokenPairWithContext(t *testing.T) {
	// Setup mocks
	mockTokenSessionRepo := &mockTokenSessionRepository{}
	mockRedisService := &mockRedisService{}
	mockAnomalyDetector := &mockAnomalyDetector{}
	mockTransactionManager := &mockTransactionManager{}
	mockTokenBlacklistRepo := &mockTokenBlacklistRepository{}

	// Create JWT manager
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	// Create service
	service := NewEnhancedJWTService(
		jwtManager,
		mockTokenBlacklistRepo,
		mockTransactionManager,
		mockTokenSessionRepo,
		mockRedisService,
		mockAnomalyDetector,
	)

	// Test data
	tokenCtx := utils.TokenContext{
		UserID:            1,
		Email:             "test@example.com",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
	}

	// Setup expectations
	mockTokenSessionRepo.On("CreateSession", mock.Anything, mock.AnythingOfType("*models.TokenSession")).Return(nil)
	mockRedisService.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	result, err := service.GenerateTokenPairWithContext(context.Background(), tokenCtx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEmpty(t, result.SessionID)
	assert.Equal(t, "test-fingerprint", result.DeviceFingerprint)
	assert.Equal(t, "standard", result.SecurityLevel)
	assert.False(t, result.IsRotated)

	// Verify mocks
	mockTokenSessionRepo.AssertExpectations(t)
	mockRedisService.AssertExpectations(t)
}

func TestEnhancedJWTService_RotateTokens(t *testing.T) {
	// Setup mocks
	mockTokenSessionRepo := &mockTokenSessionRepository{}
	mockRedisService := &mockRedisService{}
	mockAnomalyDetector := &mockAnomalyDetector{}
	mockTransactionManager := &mockTransactionManager{}
	mockTokenBlacklistRepo := &mockTokenBlacklistRepository{}

	// Create JWT manager
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	// Create service
	service := NewEnhancedJWTService(
		jwtManager,
		mockTokenBlacklistRepo,
		mockTransactionManager,
		mockTokenSessionRepo,
		mockRedisService,
		mockAnomalyDetector,
	)

	// Generate initial token pair for testing
	tokenCtx := utils.TokenContext{
		UserID:            1,
		Email:             "test@example.com",
		SessionID:         "test-session-id",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
	}

	refreshToken, err := jwtManager.GenerateTokenWithContext(tokenCtx, utils.RefreshToken)
	assert.NoError(t, err)

	// Create mock session
	session := &models.TokenSession{
		ID:                1,
		UserID:            1,
		SessionID:         "test-session-id",
		RefreshTokenHash:  jwtManager.GetTokenHash(refreshToken),
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
		UserAgent:         "test-agent",
		IsActive:          true,
		LastUsedAt:        time.Now(),
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	requestCtx := RequestContext{
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
		Endpoint:  "POST /auth/refresh",
		Timestamp: time.Now(),
	}

	// Setup expectations
	mockRedisService.On("IsTokenBlacklisted", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockTokenBlacklistRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
	mockTokenSessionRepo.On("GetSessionByRefreshToken", mock.Anything, mock.AnythingOfType("string")).Return(session, nil)
	mockTokenSessionRepo.On("GetTokenRotations", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int")).Return([]*models.TokenRotation{}, nil)
	mockTokenSessionRepo.On("UpdateSessionTx", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.TokenSession")).Return(nil)
	mockTokenSessionRepo.On("RecordTokenRotationTx", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("*models.TokenRotation")).Return(nil)
	mockTokenBlacklistRepo.On("BlacklistTokenTx", mock.AnythingOfType("*gorm.DB"), mock.AnythingOfType("string"), mock.AnythingOfType("uint"), mock.AnythingOfType("time.Time"), mock.AnythingOfType("string")).Return(nil)
	mockTransactionManager.On("WithTransactionContext", mock.Anything, mock.AnythingOfType("func(context.Context, *gorm.DB) error")).Return(nil)
	mockRedisService.On("Set", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)
	mockRedisService.On("Del", mock.Anything, mock.AnythingOfType("[]string")).Return(nil)
	mockRedisService.On("BlacklistToken", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)
	mockRedisService.On("IncrementCounter", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(int64(1), nil)

	// Execute
	result, err := service.RotateTokens(context.Background(), refreshToken, requestCtx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.AccessToken)
	assert.NotEmpty(t, result.RefreshToken)
	assert.NotEmpty(t, result.SessionID)
	assert.NotEqual(t, "test-session-id", result.SessionID) // Should be a new session ID
	assert.Equal(t, "test-fingerprint", result.DeviceFingerprint)
	assert.True(t, result.IsRotated)
	assert.Equal(t, "standard", result.SecurityLevel)

	// Verify mocks
	mockTokenSessionRepo.AssertExpectations(t)
	mockRedisService.AssertExpectations(t)
	mockTransactionManager.AssertExpectations(t)
	mockTokenBlacklistRepo.AssertExpectations(t)
}

func TestEnhancedJWTService_ValidateAccessTokenWithContext(t *testing.T) {
	// Setup mocks
	mockTokenSessionRepo := &mockTokenSessionRepository{}
	mockRedisService := &mockRedisService{}
	mockAnomalyDetector := &mockAnomalyDetector{}
	mockTransactionManager := &mockTransactionManager{}
	mockTokenBlacklistRepo := &mockTokenBlacklistRepository{}

	// Create JWT manager
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	// Create service
	service := NewEnhancedJWTService(
		jwtManager,
		mockTokenBlacklistRepo,
		mockTransactionManager,
		mockTokenSessionRepo,
		mockRedisService,
		mockAnomalyDetector,
	)

	// Generate test token
	tokenCtx := utils.TokenContext{
		UserID:            1,
		Email:             "test@example.com",
		SessionID:         "test-session-id",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
	}

	accessToken, err := jwtManager.GenerateTokenWithContext(tokenCtx, utils.AccessToken)
	assert.NoError(t, err)

	// Create mock session
	session := &models.TokenSession{
		ID:                1,
		UserID:            1,
		SessionID:         "test-session-id",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
		UserAgent:         "test-agent",
		IsActive:          true,
		LastUsedAt:        time.Now(),
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	requestCtx := RequestContext{
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
		Endpoint:  "GET /api/protected",
		Timestamp: time.Now(),
	}

	// Setup expectations
	mockRedisService.On("IsTokenBlacklisted", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockTokenBlacklistRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
	mockRedisService.On("TrackTokenUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockTokenSessionRepo.On("GetSessionByID", mock.Anything, mock.AnythingOfType("string")).Return(session, nil)
	mockTokenSessionRepo.On("UpdateSession", mock.Anything, mock.AnythingOfType("*models.TokenSession")).Return(nil)
	mockTokenSessionRepo.On("RecordTokenUsage", mock.Anything, mock.AnythingOfType("*models.TokenUsage")).Return(nil)
	mockAnomalyDetector.On("AnalyzeTokenUsage", mock.Anything, mock.AnythingOfType("*models.TokenUsage")).Return(&AnomalyResult{IsAnomaly: false}, nil)
	mockRedisService.On("RecordSecurityEvent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	claims, err := service.ValidateAccessTokenWithContext(context.Background(), accessToken, requestCtx)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, uint(1), claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "test-session-id", claims.SessionID)
	assert.Equal(t, utils.AccessToken, claims.TokenType)

	// Verify mocks
	mockTokenSessionRepo.AssertExpectations(t)
	mockRedisService.AssertExpectations(t)
	mockAnomalyDetector.AssertExpectations(t)
	mockTokenBlacklistRepo.AssertExpectations(t)
}

func TestEnhancedJWTService_ValidateAccessTokenWithContext_AnomalyDetected(t *testing.T) {
	// Setup mocks
	mockTokenSessionRepo := &mockTokenSessionRepository{}
	mockRedisService := &mockRedisService{}
	mockAnomalyDetector := &mockAnomalyDetector{}
	mockTransactionManager := &mockTransactionManager{}
	mockTokenBlacklistRepo := &mockTokenBlacklistRepository{}

	// Create JWT manager
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	// Create service
	service := NewEnhancedJWTService(
		jwtManager,
		mockTokenBlacklistRepo,
		mockTransactionManager,
		mockTokenSessionRepo,
		mockRedisService,
		mockAnomalyDetector,
	)

	// Generate test token
	tokenCtx := utils.TokenContext{
		UserID:            1,
		Email:             "test@example.com",
		SessionID:         "test-session-id",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
	}

	accessToken, err := jwtManager.GenerateTokenWithContext(tokenCtx, utils.AccessToken)
	assert.NoError(t, err)

	// Create mock session
	session := &models.TokenSession{
		ID:                1,
		UserID:            1,
		SessionID:         "test-session-id",
		DeviceFingerprint: "test-fingerprint",
		IPAddress:         "192.168.1.1",
		UserAgent:         "test-agent",
		IsActive:          true,
		LastUsedAt:        time.Now(),
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	requestCtx := RequestContext{
		IPAddress: "192.168.1.1",
		UserAgent: "test-agent",
		Endpoint:  "GET /api/protected",
		Timestamp: time.Now(),
	}

	// Setup expectations for critical anomaly
	mockRedisService.On("IsTokenBlacklisted", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockTokenBlacklistRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
	mockRedisService.On("TrackTokenUsage", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}")).Return(nil)
	mockTokenSessionRepo.On("GetSessionByID", mock.Anything, mock.AnythingOfType("string")).Return(session, nil)
	mockTokenSessionRepo.On("UpdateSession", mock.Anything, mock.AnythingOfType("*models.TokenSession")).Return(nil)
	mockTokenSessionRepo.On("RecordTokenUsage", mock.Anything, mock.AnythingOfType("*models.TokenUsage")).Return(nil)

	// Mock critical anomaly detection
	criticalAnomaly := &AnomalyResult{
		IsAnomaly:   true,
		Confidence:  0.9,
		AnomalyType: "ip",
		Description: "Suspicious IP change detected",
		Severity:    SeverityCritical,
		Timestamp:   time.Now(),
	}
	mockAnomalyDetector.On("AnalyzeTokenUsage", mock.Anything, mock.AnythingOfType("*models.TokenUsage")).Return(criticalAnomaly, nil)
	mockRedisService.On("RecordSecurityEvent", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("map[string]interface {}"), mock.AnythingOfType("time.Duration")).Return(nil)

	// Execute
	claims, err := service.ValidateAccessTokenWithContext(context.Background(), accessToken, requestCtx)

	// Assert - should fail due to critical anomaly
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "suspicious activity detected")

	// Verify mocks
	mockTokenSessionRepo.AssertExpectations(t)
	mockRedisService.AssertExpectations(t)
	mockAnomalyDetector.AssertExpectations(t)
	mockTokenBlacklistRepo.AssertExpectations(t)
}
