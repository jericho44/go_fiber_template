package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, req services.RegisterRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, req services.LoginRequest) (*services.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuthResponse), args.Error(1)
}

func (m *MockAuthService) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ComparePassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// MockJWTService is a mock implementation of JWTService
type MockJWTService struct {
	mock.Mock
}

func (m *MockJWTService) GenerateTokenPair(userID uint, email string) (*utils.TokenPair, error) {
	args := m.Called(userID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.TokenPair), args.Error(1)
}

func (m *MockJWTService) ValidateAccessToken(tokenString string) (*utils.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.JWTClaims), args.Error(1)
}

func (m *MockJWTService) ValidateRefreshToken(tokenString string) (*utils.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.JWTClaims), args.Error(1)
}

func (m *MockJWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	args := m.Called(refreshTokenString)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) BlacklistToken(tokenString string, reason string) error {
	args := m.Called(tokenString, reason)
	return args.Error(0)
}

func (m *MockJWTService) LogoutUser(accessToken, refreshToken string) error {
	args := m.Called(accessToken, refreshToken)
	return args.Error(0)
}

func (m *MockJWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	args := m.Called(tokenString)
	return args.Bool(0), args.Error(1)
}

func (m *MockJWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	args := m.Called(authHeader)
	return args.String(0), args.Error(1)
}

func (m *MockJWTService) GetTokenExpiry(tokenString string) (time.Time, error) {
	args := m.Called(tokenString)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockJWTService) GetUserIDFromToken(tokenString string) (uint, error) {
	args := m.Called(tokenString)
	return args.Get(0).(uint), args.Error(1)
}

// Helper function to create a test Fiber app with the auth controller
func setupTestApp(authService services.AuthService, jwtService services.JWTServiceInterface) *fiber.App {
	app := fiber.New()
	controller := NewAuthController(authService, jwtService)

	auth := app.Group("/auth")
	auth.Post("/register", controller.Register)
	auth.Post("/login", controller.Login)
	auth.Post("/logout", controller.Logout)
	auth.Post("/refresh", controller.RefreshToken)

	return app
}

// Helper function to make HTTP requests
func makeRequest(app *fiber.App, method, url string, body interface{}, headers map[string]string) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return app.Test(req)
}

func TestAuthController_Register(t *testing.T) {
	t.Run("Successful registration", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		user := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		authResponse := &services.AuthResponse{
			User:    user,
			Message: "User registered successfully",
		}

		tokenPair := &utils.TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
			TokenType:    "Bearer",
		}

		mockAuthService.On("Register", mock.Anything, mock.MatchedBy(func(req services.RegisterRequest) bool {
			return req.Email == "test@example.com" && req.FirstName == "John"
		})).Return(authResponse, nil)

		mockJWTService.On("GenerateTokenPair", uint(1), "test@example.com").Return(tokenPair, nil)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		resp, err := makeRequest(app, "POST", "/auth/register", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "User registered successfully", response.Message)

		mockAuthService.AssertExpectations(t)
		mockJWTService.AssertExpectations(t)
	})

	t.Run("Invalid request body", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		app := setupTestApp(mockAuthService, mockJWTService)

		resp, err := makeRequest(app, "POST", "/auth/register", "invalid json", nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Invalid request body", response.Error.Message)
	})

	t.Run("Validation error", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RegisterRequest{
			Email:     "invalid-email",
			Password:  "weak",
			FirstName: "",
			LastName:  "Doe",
		}

		resp, err := makeRequest(app, "POST", "/auth/register", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("User already exists", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		conflictErr := utils.NewConflictError("User with this email already exists")
		mockAuthService.On("Register", mock.Anything, mock.Anything).Return(nil, conflictErr)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		resp, err := makeRequest(app, "POST", "/auth/register", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "CONFLICT", response.Error.Code)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("JWT generation failure", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		user := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		authResponse := &services.AuthResponse{
			User:    user,
			Message: "User registered successfully",
		}

		mockAuthService.On("Register", mock.Anything, mock.Anything).Return(authResponse, nil)
		mockJWTService.On("GenerateTokenPair", uint(1), "test@example.com").Return(nil, errors.New("JWT generation failed"))

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		resp, err := makeRequest(app, "POST", "/auth/register", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to generate tokens", response.Error.Message)

		mockAuthService.AssertExpectations(t)
		mockJWTService.AssertExpectations(t)
	})
}

func TestAuthController_Login(t *testing.T) {
	t.Run("Successful login", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		user := &models.User{
			ID:        1,
			Email:     "test@example.com",
			FirstName: "John",
			LastName:  "Doe",
		}

		authResponse := &services.AuthResponse{
			User:    user,
			Message: "Login successful",
		}

		tokenPair := &utils.TokenPair{
			AccessToken:  "access_token",
			RefreshToken: "refresh_token",
			ExpiresAt:    time.Now().Add(15 * time.Minute).Unix(),
			TokenType:    "Bearer",
		}

		mockAuthService.On("Login", mock.Anything, mock.MatchedBy(func(req services.LoginRequest) bool {
			return req.Email == "test@example.com" && req.Password == "Password123!"
		})).Return(authResponse, nil)

		mockJWTService.On("GenerateTokenPair", uint(1), "test@example.com").Return(tokenPair, nil)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := LoginRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		resp, err := makeRequest(app, "POST", "/auth/login", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Login successful", response.Message)

		mockAuthService.AssertExpectations(t)
		mockJWTService.AssertExpectations(t)
	})

	t.Run("Invalid credentials", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		unauthorizedErr := utils.NewUnauthorizedError("Invalid email or password")
		mockAuthService.On("Login", mock.Anything, mock.Anything).Return(nil, unauthorizedErr)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := LoginRequest{
			Email:    "test@example.com",
			Password: "wrongpassword",
		}

		resp, err := makeRequest(app, "POST", "/auth/login", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)

		mockAuthService.AssertExpectations(t)
	})

	t.Run("Validation error", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := LoginRequest{
			Email:    "invalid-email",
			Password: "",
		}

		resp, err := makeRequest(app, "POST", "/auth/login", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})
}

func TestAuthController_Logout(t *testing.T) {
	t.Run("Successful logout with both tokens", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("ExtractTokenFromHeader", "Bearer access_token").Return("access_token", nil)
		mockJWTService.On("LogoutUser", "access_token", "refresh_token").Return(nil)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := LogoutRequest{
			RefreshToken: "refresh_token",
		}

		headers := map[string]string{
			"Authorization": "Bearer access_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/logout", requestBody, headers)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Logout successful", response.Message)

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Successful logout with access token only", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("ExtractTokenFromHeader", "Bearer access_token").Return("access_token", nil)
		mockJWTService.On("LogoutUser", "access_token", "").Return(nil)

		app := setupTestApp(mockAuthService, mockJWTService)

		headers := map[string]string{
			"Authorization": "Bearer access_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/logout", nil, headers)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Logout successful", response.Message)

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Missing authorization header", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("ExtractTokenFromHeader", "").Return("", errors.New("authorization header is required"))

		app := setupTestApp(mockAuthService, mockJWTService)

		resp, err := makeRequest(app, "POST", "/auth/logout", nil, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Logout service failure", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("ExtractTokenFromHeader", "Bearer access_token").Return("access_token", nil)
		mockJWTService.On("LogoutUser", "access_token", "").Return(errors.New("logout failed"))

		app := setupTestApp(mockAuthService, mockJWTService)

		headers := map[string]string{
			"Authorization": "Bearer access_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/logout", nil, headers)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to logout user", response.Error.Message)

		mockJWTService.AssertExpectations(t)
	})
}

func TestAuthController_RefreshToken(t *testing.T) {
	t.Run("Successful token refresh", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		newAccessToken := "new_access_token"
		tokenExpiry := time.Now().Add(15 * time.Minute)

		mockJWTService.On("RefreshAccessToken", "refresh_token").Return(newAccessToken, nil)
		mockJWTService.On("GetTokenExpiry", newAccessToken).Return(tokenExpiry, nil)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RefreshTokenRequest{
			RefreshToken: "refresh_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/refresh", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.True(t, response.Success)
		assert.Equal(t, "Token refreshed successfully", response.Message)

		// Check response data
		data, ok := response.Data.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, newAccessToken, data["access_token"])
		assert.Equal(t, "Bearer", data["token_type"])

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Invalid refresh token", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("RefreshAccessToken", "invalid_token").Return("", errors.New("invalid refresh token"))

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RefreshTokenRequest{
			RefreshToken: "invalid_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/refresh", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "UNAUTHORIZED", response.Error.Code)

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Expired refresh token", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		mockJWTService.On("RefreshAccessToken", "expired_token").Return("", errors.New("token has expired"))

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RefreshTokenRequest{
			RefreshToken: "expired_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/refresh", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Invalid or expired refresh token", response.Error.Message)

		mockJWTService.AssertExpectations(t)
	})

	t.Run("Validation error - missing refresh token", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RefreshTokenRequest{
			RefreshToken: "",
		}

		resp, err := makeRequest(app, "POST", "/auth/refresh", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("Token expiry retrieval failure", func(t *testing.T) {
		mockAuthService := new(MockAuthService)
		mockJWTService := new(MockJWTService)

		newAccessToken := "new_access_token"

		mockJWTService.On("RefreshAccessToken", "refresh_token").Return(newAccessToken, nil)
		mockJWTService.On("GetTokenExpiry", newAccessToken).Return(time.Time{}, errors.New("failed to get expiry"))

		app := setupTestApp(mockAuthService, mockJWTService)

		requestBody := RefreshTokenRequest{
			RefreshToken: "refresh_token",
		}

		resp, err := makeRequest(app, "POST", "/auth/refresh", requestBody, nil)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "Failed to get token expiry", response.Error.Message)

		mockJWTService.AssertExpectations(t)
	})
}

func TestAuthController_NewAuthController(t *testing.T) {
	mockAuthService := new(MockAuthService)
	mockJWTService := new(MockJWTService)

	controller := NewAuthController(mockAuthService, mockJWTService)

	assert.NotNil(t, controller)
	assert.Equal(t, mockAuthService, controller.authService)
	assert.Equal(t, mockJWTService, controller.jwtService)
	assert.NotNil(t, controller.validator)
}

func TestAuthController_RequestStructValidation(t *testing.T) {
	t.Run("RegisterRequest validation", func(t *testing.T) {
		validator := utils.NewValidationErrorFormatter()

		// Valid request
		validReq := RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}
		err := validator.ValidateStruct(validReq)
		assert.Nil(t, err)

		// Invalid email
		invalidEmailReq := RegisterRequest{
			Email:     "invalid-email",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}
		err = validator.ValidateStruct(invalidEmailReq)
		assert.NotNil(t, err)

		// Missing required fields
		missingFieldsReq := RegisterRequest{
			Email:    "test@example.com",
			Password: "Password123!",
			// FirstName and LastName missing
		}
		err = validator.ValidateStruct(missingFieldsReq)
		assert.NotNil(t, err)
	})

	t.Run("LoginRequest validation", func(t *testing.T) {
		validator := utils.NewValidationErrorFormatter()

		// Valid request
		validReq := LoginRequest{
			Email:    "test@example.com",
			Password: "password",
		}
		err := validator.ValidateStruct(validReq)
		assert.Nil(t, err)

		// Invalid email
		invalidEmailReq := LoginRequest{
			Email:    "invalid-email",
			Password: "password",
		}
		err = validator.ValidateStruct(invalidEmailReq)
		assert.NotNil(t, err)

		// Missing password
		missingPasswordReq := LoginRequest{
			Email: "test@example.com",
		}
		err = validator.ValidateStruct(missingPasswordReq)
		assert.NotNil(t, err)
	})

	t.Run("RefreshTokenRequest validation", func(t *testing.T) {
		validator := utils.NewValidationErrorFormatter()

		// Valid request
		validReq := RefreshTokenRequest{
			RefreshToken: "valid_token",
		}
		err := validator.ValidateStruct(validReq)
		assert.Nil(t, err)

		// Missing refresh token
		missingTokenReq := RefreshTokenRequest{}
		err = validator.ValidateStruct(missingTokenReq)
		assert.NotNil(t, err)
	})
}
