package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(entity *repositories.UserEntity) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id uint) (repositories.UserEntity, error) {
	args := m.Called(id)
	return args.Get(0).(repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetByIDWithContext(ctx context.Context, id uint) (repositories.UserEntity, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) Update(entity *repositories.UserEntity) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) List(limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CreateTx(tx *gorm.DB, entity *repositories.UserEntity) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateTx(tx *gorm.DB, entity *repositories.UserEntity) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateWithContext(ctx context.Context, entity *repositories.UserEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateWithContext(ctx context.Context, entity *repositories.UserEntity) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) ListWithContext(ctx context.Context, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) CountWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(email string) (repositories.UserEntity, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetByEmailTx(tx *gorm.DB, email string) (repositories.UserEntity, error) {
	args := m.Called(tx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetByEmailWithContext(ctx context.Context, email string) (repositories.UserEntity, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsers(limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsersTx(tx *gorm.DB, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(tx, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetActiveUsersWithContext(ctx context.Context, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(email string) (bool, error) {
	args := m.Called(email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmailTx(tx *gorm.DB, email string) (bool, error) {
	args := m.Called(tx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmailWithContext(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CreateBatch(users []repositories.UserEntity) error {
	args := m.Called(users)
	return args.Error(0)
}

func (m *MockUserRepository) CreateBatchTx(tx *gorm.DB, users []repositories.UserEntity) error {
	args := m.Called(tx, users)
	return args.Error(0)
}

func (m *MockUserRepository) CreateBatchWithContext(ctx context.Context, users []repositories.UserEntity) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateBatch(users []repositories.UserEntity) error {
	args := m.Called(users)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateBatchTx(tx *gorm.DB, users []repositories.UserEntity) error {
	args := m.Called(tx, users)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateBatchWithContext(ctx context.Context, users []repositories.UserEntity) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDelete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) SoftDeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) Restore(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) RestoreTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) RestoreWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetDeletedUsers(limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetDeletedUsersTx(tx *gorm.DB, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(tx, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) GetDeletedUsersWithContext(ctx context.Context, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) HardDelete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) HardDeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockUserRepository) HardDeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) SearchByName(firstName, lastName string, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(firstName, lastName, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) SearchByNameTx(tx *gorm.DB, firstName, lastName string, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(tx, firstName, lastName, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) SearchByNameWithContext(ctx context.Context, firstName, lastName string, limit, offset int) ([]repositories.UserEntity, error) {
	args := m.Called(ctx, firstName, lastName, limit, offset)
	return args.Get(0).([]repositories.UserEntity), args.Error(1)
}

func (m *MockUserRepository) CountActiveUsers() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountActiveUsersTx(tx *gorm.DB) (int64, error) {
	args := m.Called(tx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountActiveUsersWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountDeletedUsers() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountDeletedUsersTx(tx *gorm.DB) (int64, error) {
	args := m.Called(tx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) CountDeletedUsersWithContext(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

// MockTokenBlacklistRepository is a mock implementation of TokenBlacklistRepository
type MockTokenBlacklistRepository struct {
	mock.Mock
}

func (m *MockTokenBlacklistRepository) Create(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) GetByID(id uint) (models.TokenBlacklist, error) {
	args := m.Called(id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) Update(entity *models.TokenBlacklist) error {
	args := m.Called(entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) List(limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTokenBlacklistRepository) CreateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) UpdateTx(tx *gorm.DB, entity *models.TokenBlacklist) error {
	args := m.Called(tx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) DeleteTx(tx *gorm.DB, id uint) error {
	args := m.Called(tx, id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) CreateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) GetByIDWithContext(ctx context.Context, id uint) (models.TokenBlacklist, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.TokenBlacklist), args.Error(1)
}

func (m *MockTokenBlacklistRepository) UpdateWithContext(ctx context.Context, entity *models.TokenBlacklist) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) DeleteWithContext(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTokenBlacklistRepository) ListWithContext(ctx context.Context, limit, offset int) ([]models.TokenBlacklist, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]models.TokenBlacklist), args.Error(1)
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

// MockUser is a mock implementation of UserEntity
type MockUser struct {
	ID        uint
	Email     string
	FirstName string
	LastName  string
	IsActive  bool
}

func (u *MockUser) GetID() uint {
	return u.ID
}

func (u *MockUser) GetEmail() string {
	return u.Email
}

func (u *MockUser) SetEmail(email string) {
	u.Email = email
}

func (u *MockUser) GetFirstName() string {
	return u.FirstName
}

func (u *MockUser) GetLastName() string {
	return u.LastName
}

func (u *MockUser) GetIsActive() bool {
	return u.IsActive
}

func (u *MockUser) SetIsActive(active bool) {
	u.IsActive = active
}

// Test setup helpers
func setupTestApp() *fiber.App {
	app := fiber.New()
	return app
}

func setupAuthMiddleware() (*AuthMiddleware, *MockUserRepository, *MockTokenBlacklistRepository, *utils.JWTManager) {
	mockUserRepo := &MockUserRepository{}
	mockTokenRepo := &MockTokenBlacklistRepository{}
	jwtManager := utils.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)

	authMiddleware := NewAuthMiddleware(jwtManager, mockUserRepo, mockTokenRepo)

	return authMiddleware, mockUserRepo, mockTokenRepo, jwtManager
}

func createValidToken(jwtManager *utils.JWTManager, userID uint, email string, tokenType utils.TokenType) string {
	token, _ := jwtManager.GenerateToken(userID, email, tokenType)
	return token
}

func TestNewAuthMiddleware(t *testing.T) {
	authMiddleware, _, _, jwtManager := setupAuthMiddleware()

	assert.NotNil(t, authMiddleware)
	assert.Equal(t, jwtManager, authMiddleware.jwtManager)
}

func TestRequireAuth_Success(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.AccessToken)

	// Setup mocks
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: true,
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		assert.NotNil(t, user)
		assert.Equal(t, userID, user.GetID())
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRequireAuth_MissingToken(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, _, _ := setupAuthMiddleware()

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/protected", nil)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Authorization header is required", response.Error.Message)
}

func TestRequireAuth_InvalidTokenFormat(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, _, _ := setupAuthMiddleware()

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request with invalid token format
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, _, _ := setupAuthMiddleware()

	// Create JWT manager with very short expiry
	shortJwtManager := utils.NewJWTManager("test-secret", -1*time.Hour, 7*24*time.Hour)
	expiredToken := createValidToken(shortJwtManager, 1, "test@example.com", utils.AccessToken)

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request with expired token
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Token has expired", response.Error.Message)
}

func TestRequireAuth_BlacklistedToken(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token
	token := createValidToken(jwtManager, 1, "test@example.com", utils.AccessToken)

	// Setup mocks - token is blacklisted
	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(true, nil)

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Token has been invalidated", response.Error.Message)

	mockTokenRepo.AssertExpectations(t)
}

func TestRequireAuth_RefreshTokenNotAllowed(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a refresh token instead of access token
	token := createValidToken(jwtManager, 1, "test@example.com", utils.RefreshToken)

	// Setup mocks
	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Access token required", response.Error.Message)

	mockTokenRepo.AssertExpectations(t)
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token
	userID := uint(1)
	token := createValidToken(jwtManager, userID, "test@example.com", utils.AccessToken)

	// Setup mocks - user not found
	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(nil, fmt.Errorf("user not found"))

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRequireAuth_InactiveUser(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.AccessToken)

	// Setup mocks - inactive user
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: false, // User is inactive
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware
	app.Get("/protected", authMiddleware.RequireAuth(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "User account is inactive", response.Error.Message)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestOptionalAuth_WithValidToken(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.AccessToken)

	// Setup mocks
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: true,
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware
	app.Get("/optional", authMiddleware.OptionalAuth(), func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		isAuth := IsAuthenticated(c)
		return c.JSON(fiber.Map{
			"authenticated": isAuth,
			"user_id":       user.GetID(),
		})
	})

	// Create request
	req := httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check response body
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["authenticated"].(bool))
	assert.Equal(t, float64(userID), response["user_id"].(float64))

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestOptionalAuth_WithoutToken(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, _, _ := setupAuthMiddleware()

	// Setup route with middleware
	app.Get("/optional", authMiddleware.OptionalAuth(), func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		isAuth := IsAuthenticated(c)
		return c.JSON(fiber.Map{
			"authenticated": isAuth,
			"user":          user,
		})
	})

	// Create request without token
	req := httptest.NewRequest("GET", "/optional", nil)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check response body
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response["authenticated"].(bool))
	assert.Nil(t, response["user"])
}

func TestRequireRefreshToken_Success(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid refresh token
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.RefreshToken)

	// Setup mocks
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: true,
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware
	app.Post("/refresh", authMiddleware.RequireRefreshToken(), func(c *fiber.Ctx) error {
		user := GetUserFromContext(c)
		claims := GetTokenClaimsFromContext(c)
		refreshToken := GetRefreshTokenFromContext(c)

		assert.NotNil(t, user)
		assert.NotNil(t, claims)
		assert.NotEmpty(t, refreshToken)
		assert.Equal(t, utils.RefreshToken, claims.TokenType)

		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRequireRefreshToken_AccessTokenNotAllowed(t *testing.T) {
	app := setupTestApp()
	authMiddleware, _, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create an access token instead of refresh token
	token := createValidToken(jwtManager, 1, "test@example.com", utils.AccessToken)

	// Setup mocks
	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)

	// Setup route with middleware
	app.Post("/refresh", authMiddleware.RequireRefreshToken(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request
	req := httptest.NewRequest("POST", "/refresh", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "Refresh token required", response.Error.Message)

	mockTokenRepo.AssertExpectations(t)
}

func TestRequireUserID_Success(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token for user ID 1
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.AccessToken)

	// Setup mocks
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: true,
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware - user can access their own resource
	app.Get("/users/:id", authMiddleware.RequireUserID("id"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request for the same user ID
	req := httptest.NewRequest("GET", "/users/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestRequireUserID_AccessDenied(t *testing.T) {
	app := setupTestApp()
	authMiddleware, mockUserRepo, mockTokenRepo, jwtManager := setupAuthMiddleware()

	// Create a valid token for user ID 1
	userID := uint(1)
	email := "test@example.com"
	token := createValidToken(jwtManager, userID, email, utils.AccessToken)

	// Setup mocks
	mockUser := &MockUser{
		ID:       userID,
		Email:    email,
		IsActive: true,
	}

	mockTokenRepo.On("IsTokenBlacklistedWithContext", mock.Anything, mock.AnythingOfType("string")).Return(false, nil)
	mockUserRepo.On("GetByIDWithContext", mock.Anything, userID).Return(mockUser, nil)

	// Setup route with middleware
	app.Get("/users/:id", authMiddleware.RequireUserID("id"), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "success"})
	})

	// Create request for a different user ID (2)
	req := httptest.NewRequest("GET", "/users/2", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// Check response body
	var response utils.APIResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "You can only access your own resources", response.Error.Message)

	mockUserRepo.AssertExpectations(t)
	mockTokenRepo.AssertExpectations(t)
}

func TestContextHelpers(t *testing.T) {
	// Test WithUserContext and GetUserFromGoContext
	mockUser := &MockUser{
		ID:       1,
		Email:    "test@example.com",
		IsActive: true,
	}

	ctx := context.Background()
	userCtx := WithUserContext(ctx, mockUser)

	retrievedUser := GetUserFromGoContext(userCtx)
	assert.NotNil(t, retrievedUser)
	assert.Equal(t, mockUser.GetID(), retrievedUser.GetID())
	assert.Equal(t, mockUser.GetEmail(), retrievedUser.GetEmail())

	// Test with nil context
	nilUser := GetUserFromGoContext(context.Background())
	assert.Nil(t, nilUser)
}

func TestGetUserIDFromContext(t *testing.T) {
	app := setupTestApp()

	mockUser := &MockUser{
		ID:       123,
		Email:    "test@example.com",
		IsActive: true,
	}

	app.Get("/test", func(c *fiber.Ctx) error {
		// Test with user in context
		c.Locals(UserContextKey, mockUser)
		userID := GetUserIDFromContext(c)
		assert.Equal(t, uint(123), userID)

		// Test without user in context
		c.Locals(UserContextKey, nil)
		userID = GetUserIDFromContext(c)
		assert.Equal(t, uint(0), userID)

		return c.JSON(fiber.Map{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
