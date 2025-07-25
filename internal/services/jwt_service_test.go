package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockTokenBlacklistRepository is a mock implementation of TokenBlacklistRepository
type MockTokenBlacklistRepository struct {
	mock.Mock
}

func (m *MockTokenBlacklistRepository) Create(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) CreateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) CreateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) GetByID(id uint) (models.TokenBlacklist, error) {
	args := m.Called(id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) GetByIDWithContext(ctx context.Context, id uint) (models.TokenBlacklist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) Update(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) UpdateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) UpdateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) DeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) DeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) List(limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) ListWithContext(ctx context.Context, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CountWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) IsTokenBlacklisted(tokenHash string) (bool, error) {
	args := m.Called(tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockTokenBlacklistRepository) IsTokenBlacklistedWithContext(ctx context.Context, tokenHash string) (bool, error) {
	args := m.Called(ctx, tokenHash)
	return args.Bool(0), args.Error(1)
}

func (m *MockTokenBlacklistRepository) BlacklistToken(tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) BlacklistTokenWithContext(ctx context.Context, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(ctx, tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) BlacklistTokenTx(tx *gorm.DB, tokenHash string, userID uint, expiresAt time.Time, reason string) error {
	args := m.Called(tx, tokenHash, userID, expiresAt, reason)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) GetBlacklistedToken(tokenHash string) (*models.TokenBlacklist, error) {
	args := m.Called(tokenHash)
	return args.Get(0).(*models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) GetBlacklistedTokenWithContext(ctx context.Context, tokenHash string) (*models.TokenBlacklist, error) {
	args := m.Called(ctx, tokenHash)
	return args.Get(0).(*models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CleanupExpiredTokens() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CleanupExpiredTokensWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CleanupExpiredTokensTx(tx *gorm.DB) (int64, error) {
	args := m.Called(tx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) GetUserBlacklistedTokens(userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(userID, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) GetUserBlacklistedTokensWithContext(ctx context.Context, userID uint, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CountUserBlacklistedTokens(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CountUserBlacklistedTokensWithContext(ctx context.Context, userID uint) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

// MockTransactionManager is a mock implementation of TransactionManager
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithTransaction(fn func(*gorm.DB) error) error {
	args := m.Called(fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) BeginTransaction() (*gorm.DB, error) {
	args := m.Called()
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockTransactionManager) BeginTransactionWithContext(ctx context.Context) (*gorm.DB, error) {
	args := m.Called(ctx)
	return args.Get(0).(*gorm.DB), args.Error(1)
}

func (m *MockTransactionManager) CommitTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockTransactionManager) RollbackTransaction(tx *gorm.DB) error {
	args := m.Called(tx)
	return args.Error(0)
}

func (m *MockTransactionManager) WithSavepoint(tx *gorm.DB, name string, fn func(*gorm.DB) error) error {
	args := m.Called(tx, name, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithRetryableTransaction(fn func(*gorm.DB) error, maxRetries int) error {
	args := m.Called(fn, maxRetries)
	return args.Error(0)
}

func (m *MockTransactionManager) WithIsolationLevel(isolationLevel string, fn func(*gorm.DB) error) error {
	args := m.Called(isolationLevel, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithIsolationLevelContext(ctx context.Context, isolationLevel string, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, isolationLevel, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithTimeout(timeout time.Duration, fn func(*gorm.DB) error) error {
	args := m.Called(timeout, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithTimeoutContext(ctx context.Context, timeout time.Duration, fn func(context.Context, *gorm.DB) error) error {
	args := m.Called(ctx, timeout, fn)
	return args.Error(0)
}

func (m *MockTransactionManager) WithTimeoutAndRetry(timeout time.Duration, fn func(*gorm.DB) error, maxRetries int) error {
	args := m.Called(timeout, fn, maxRetries)
	return args.Error(0)
}

func (m *MockTransactionManager) GetIsolationLevel(tx *gorm.DB) (string, error) {
	args := m.Called(tx)
	return args.String(0), args.Error(1)
}

func (m *MockTransactionManager) SetIsolationLevel(tx *gorm.DB, isolationLevel string) error {
	args := m.Called(tx, isolationLevel)
	return args.Error(0)
}

func setupJWTService() (*JWTService, *utils.JWTManager, *MockTokenBlacklistRepository, *MockTransactionManager) {
	jwtManager := utils.NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)
	mockRepo := &MockTokenBlacklistRepository{}
	mockTxManager := &MockTransactionManager{}

	service := NewJWTService(jwtManager, mockRepo, mockTxManager)

	return service, jwtManager, mockRepo, mockTxManager
}

func TestNewJWTService(t *testing.T) {
	service, jwtManager, mockRepo, mockTxManager := setupJWTService()

	assert.NotNil(t, service)
	assert.Equal(t, jwtManager, service.jwtManager)
	assert.Equal(t, mockRepo, service.tokenBlacklistRepo)
	assert.Equal(t, mockTxManager, service.transactionManager)
}

func TestJWTService_GenerateTokenPair(t *testing.T) {
	service, _, _, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tokenPair, err := service.GenerateTokenPair(userID, email)

	require.NoError(t, err)
	require.NotNil(t, tokenPair)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.True(t, tokenPair.ExpiresAt > time.Now().Unix())
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name          string
		setupToken    func() string
		setupMocks    func()
		wantError     bool
		errorContains string
	}{
		{
			name: "Valid access token not blacklisted",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
			},
			wantError: false,
		},
		{
			name: "Valid access token but blacklisted",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(true, nil)
			},
			wantError:     true,
			errorContains: "token has been revoked",
		},
		{
			name: "Refresh token instead of access token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)
				return token
			},
			setupMocks: func() {
				// No mock setup needed as validation fails before blacklist check
			},
			wantError:     true,
			errorContains: "invalid token type",
		},
		{
			name: "Invalid token",
			setupToken: func() string {
				return "invalid-token"
			},
			setupMocks: func() {
				// No mock setup needed as validation fails before blacklist check
			},
			wantError:     true,
			errorContains: "token validation failed",
		},
		{
			name: "Blacklist check error",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, errors.New("database error"))
			},
			wantError:     true,
			errorContains: "failed to check token blacklist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			token := tt.setupToken()
			tt.setupMocks()

			claims, err := service.ValidateAccessToken(token)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, utils.AccessToken, claims.TokenType)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name          string
		setupToken    func() string
		setupMocks    func()
		wantError     bool
		errorContains string
	}{
		{
			name: "Valid refresh token not blacklisted",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
			},
			wantError: false,
		},
		{
			name: "Valid refresh token but blacklisted",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(true, nil)
			},
			wantError:     true,
			errorContains: "token has been revoked",
		},
		{
			name: "Access token instead of refresh token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				// No mock setup needed as validation fails before blacklist check
			},
			wantError:     true,
			errorContains: "invalid token type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			token := tt.setupToken()
			tt.setupMocks()

			claims, err := service.ValidateRefreshToken(token)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, utils.RefreshToken, claims.TokenType)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name          string
		setupToken    func() string
		setupMocks    func()
		wantError     bool
		errorContains string
	}{
		{
			name: "Valid refresh token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
			},
			wantError: false,
		},
		{
			name: "Blacklisted refresh token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(true, nil)
			},
			wantError:     true,
			errorContains: "invalid refresh token",
		},
		{
			name: "Access token instead of refresh token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				// No mock setup needed as validation fails before blacklist check
			},
			wantError:     true,
			errorContains: "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			refreshToken := tt.setupToken()
			tt.setupMocks()

			newAccessToken, err := service.RefreshAccessToken(refreshToken)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, newAccessToken)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newAccessToken)

				// Verify the new access token is valid
				claims, err := jwtManager.ValidateToken(newAccessToken)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, utils.AccessToken, claims.TokenType)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_BlacklistToken(t *testing.T) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"
	reason := "logout"

	tests := []struct {
		name          string
		setupToken    func() string
		setupMocks    func()
		wantError     bool
		errorContains string
	}{
		{
			name: "Successfully blacklist token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("BlacklistToken", mock.AnythingOfType("string"), userID, mock.AnythingOfType("time.Time"), reason).Return(nil)
			},
			wantError: false,
		},
		{
			name: "Repository error",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			setupMocks: func() {
				mockRepo.On("BlacklistToken", mock.AnythingOfType("string"), userID, mock.AnythingOfType("time.Time"), reason).Return(errors.New("database error"))
			},
			wantError:     true,
			errorContains: "failed to blacklist token",
		},
		{
			name: "Invalid token",
			setupToken: func() string {
				return "invalid-token"
			},
			setupMocks: func() {
				// No mock setup needed as token parsing fails first
			},
			wantError:     true,
			errorContains: "failed to get token expiry",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			token := tt.setupToken()
			tt.setupMocks()

			err := service.BlacklistToken(token, reason)

			if tt.wantError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_LogoutUser(t *testing.T) {
	service, jwtManager, mockRepo, mockTxManager := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	accessToken, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
	refreshToken, _ := jwtManager.GenerateToken(userID, email, utils.RefreshToken)

	tests := []struct {
		name         string
		accessToken  string
		refreshToken string
		setupMocks   func()
		wantError    bool
	}{
		{
			name:         "Successfully logout with both tokens",
			accessToken:  accessToken,
			refreshToken: refreshToken,
			setupMocks: func() {
				mockTxManager.On("WithTransaction", mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(0).(func(*gorm.DB) error)
					mockTx := &gorm.DB{}

					// Setup expectations for both token blacklisting
					mockRepo.On("BlacklistTokenTx", mockTx, mock.AnythingOfType("string"), userID, mock.AnythingOfType("time.Time"), "logout").Return(nil).Twice()

					fn(mockTx)
				})
			},
			wantError: false,
		},
		{
			name:         "Successfully logout with only access token",
			accessToken:  accessToken,
			refreshToken: "",
			setupMocks: func() {
				mockTxManager.On("WithTransaction", mock.AnythingOfType("func(*gorm.DB) error")).Return(nil).Run(func(args mock.Arguments) {
					fn := args.Get(0).(func(*gorm.DB) error)
					mockTx := &gorm.DB{}

					// Setup expectation for only access token blacklisting
					mockRepo.On("BlacklistTokenTx", mockTx, mock.AnythingOfType("string"), userID, mock.AnythingOfType("time.Time"), "logout").Return(nil).Once()

					fn(mockTx)
				})
			},
			wantError: false,
		},
		{
			name:         "Transaction error",
			accessToken:  accessToken,
			refreshToken: refreshToken,
			setupMocks: func() {
				mockTxManager.On("WithTransaction", mock.AnythingOfType("func(*gorm.DB) error")).Return(errors.New("transaction error"))
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil
			mockTxManager.ExpectedCalls = nil

			tt.setupMocks()

			err := service.LogoutUser(tt.accessToken, tt.refreshToken)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
			mockTxManager.AssertExpectations(t)
		})
	}
}

func TestJWTService_IsTokenBlacklisted(t *testing.T) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"
	token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)

	tests := []struct {
		name       string
		setupMocks func()
		wantResult bool
		wantError  bool
	}{
		{
			name: "Token is blacklisted",
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(true, nil)
			},
			wantResult: true,
			wantError:  false,
		},
		{
			name: "Token is not blacklisted",
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)
			},
			wantResult: false,
			wantError:  false,
		},
		{
			name: "Repository error",
			setupMocks: func() {
				mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, errors.New("database error"))
			},
			wantResult: false,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			tt.setupMocks()

			result, err := service.IsTokenBlacklisted(token)

			assert.Equal(t, tt.wantResult, result)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_CleanupExpiredTokens(t *testing.T) {
	service, _, mockRepo, _ := setupJWTService()

	tests := []struct {
		name         string
		setupMocks   func()
		wantAffected int64
		wantError    bool
	}{
		{
			name: "Successfully cleanup tokens",
			setupMocks: func() {
				mockRepo.On("CleanupExpiredTokens").Return(int64(5), nil)
			},
			wantAffected: 5,
			wantError:    false,
		},
		{
			name: "Repository error",
			setupMocks: func() {
				mockRepo.On("CleanupExpiredTokens").Return(int64(0), errors.New("database error"))
			},
			wantAffected: 0,
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mocks
			mockRepo.ExpectedCalls = nil

			tt.setupMocks()

			affected, err := service.CleanupExpiredTokens()

			assert.Equal(t, tt.wantAffected, affected)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestJWTService_ExtractTokenFromHeader(t *testing.T) {
	service, _, _, _ := setupJWTService()

	tests := []struct {
		name       string
		authHeader string
		wantToken  string
		wantError  bool
	}{
		{
			name:       "Valid Bearer token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantError:  false,
		},
		{
			name:       "Empty header",
			authHeader: "",
			wantToken:  "",
			wantError:  true,
		},
		{
			name:       "Invalid header format",
			authHeader: "Basic token",
			wantToken:  "",
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.ExtractTokenFromHeader(tt.authHeader)

			assert.Equal(t, tt.wantToken, token)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJWTService_GetUserIDFromToken(t *testing.T) {
	service, jwtManager, _, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name       string
		setupToken func() string
		wantUserID uint
		wantError  bool
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			wantUserID: userID,
			wantError:  false,
		},
		{
			name: "Invalid token",
			setupToken: func() string {
				return "invalid-token"
			},
			wantUserID: 0,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()

			extractedUserID, err := service.GetUserIDFromToken(token)

			assert.Equal(t, tt.wantUserID, extractedUserID)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJWTService_GetTokenExpiry(t *testing.T) {
	service, jwtManager, _, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name       string
		setupToken func() string
		wantError  bool
	}{
		{
			name: "Valid token",
			setupToken: func() string {
				token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)
				return token
			},
			wantError: false,
		},
		{
			name: "Invalid token",
			setupToken: func() string {
				return "invalid-token"
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupToken()

			expiry, err := service.GetTokenExpiry(token)

			if tt.wantError {
				assert.Error(t, err)
				assert.True(t, expiry.IsZero())
			} else {
				assert.NoError(t, err)
				assert.False(t, expiry.IsZero())
				assert.True(t, expiry.After(time.Now()))
			}
		})
	}
}

// Benchmark tests
func BenchmarkJWTService_ValidateAccessToken(b *testing.B) {
	service, jwtManager, mockRepo, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"
	token, _ := jwtManager.GenerateToken(userID, email, utils.AccessToken)

	mockRepo.On("IsTokenBlacklisted", mock.AnythingOfType("string")).Return(false, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.ValidateAccessToken(token)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWTService_GenerateTokenPair(b *testing.B) {
	service, _, _, _ := setupJWTService()

	userID := uint(123)
	email := "test@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.GenerateTokenPair(userID, email)
		if err != nil {
			b.Fatal(err)
		}
	}
}
