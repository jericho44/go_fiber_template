package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-fiber-template/internal/controllers"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthenticationFlow tests the complete authentication flow
func TestAuthenticationFlow(t *testing.T) {
	// Setup test database
	testContainer := SetupTestDatabase(t)
	defer testContainer.TeardownTestDatabase(t)

	// Setup services and controllers
	tokenBlacklistRepo := repositories.NewTokenBlacklistRepository(testContainer.DB)
	txManager := repositories.NewTransactionManager(testContainer.DB)

	authService := services.NewAuthService(testContainer.DB, txManager)
	jwtManager := utils.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	jwtService := services.NewJWTService(jwtManager, tokenBlacklistRepo, txManager)

	authController := controllers.NewAuthController(authService, jwtService)

	// Setup Fiber app
	app := fiber.New()
	auth := app.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/logout", authController.Logout)
	auth.Post("/refresh", authController.RefreshToken)

	t.Run("Complete Authentication Flow", func(t *testing.T) {
		// Clean up before test
		require.NoError(t, testContainer.CleanupTestData())

		// Step 1: Register a new user
		registerReq := controllers.RegisterRequest{
			Email:     "test@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		registerBody, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(registerBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var registerResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&registerResponse)
		require.NoError(t, err)
		assert.True(t, registerResponse.Success)

		// Extract tokens from registration response
		data, ok := registerResponse.Data.(map[string]interface{})
		require.True(t, ok)
		accessToken := data["access_token"].(string)
		refreshToken := data["refresh_token"].(string)

		assert.NotEmpty(t, accessToken)
		assert.NotEmpty(t, refreshToken)

		// Step 2: Login with the same credentials
		loginReq := controllers.LoginRequest{
			Email:    "test@example.com",
			Password: "Password123!",
		}

		loginBody, _ := json.Marshal(loginReq)
		req = httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(loginBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var loginResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&loginResponse)
		require.NoError(t, err)
		assert.True(t, loginResponse.Success)

		// Extract new tokens from login response
		loginData, ok := loginResponse.Data.(map[string]interface{})
		require.True(t, ok)
		newAccessToken := loginData["access_token"].(string)
		newRefreshToken := loginData["refresh_token"].(string)

		assert.NotEmpty(t, newAccessToken)
		assert.NotEmpty(t, newRefreshToken)

		// Step 3: Refresh the access token
		refreshReq := controllers.RefreshTokenRequest{
			RefreshToken: newRefreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		req = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&refreshResponse)
		require.NoError(t, err)
		assert.True(t, refreshResponse.Success)

		// Extract refreshed access token
		refreshData, ok := refreshResponse.Data.(map[string]interface{})
		require.True(t, ok)
		refreshedAccessToken := refreshData["access_token"].(string)

		assert.NotEmpty(t, refreshedAccessToken)
		assert.NotEqual(t, newAccessToken, refreshedAccessToken)

		// Step 4: Logout using the refreshed access token
		logoutReq := controllers.LogoutRequest{
			RefreshToken: newRefreshToken,
		}

		logoutBody, _ := json.Marshal(logoutReq)
		req = httptest.NewRequest("POST", "/auth/logout", bytes.NewBuffer(logoutBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", refreshedAccessToken))

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var logoutResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&logoutResponse)
		require.NoError(t, err)
		assert.True(t, logoutResponse.Success)
		assert.Equal(t, "Logout successful", logoutResponse.Message)

		// Step 5: Verify tokens are blacklisted by trying to use them
		req = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Registration Validation", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		tests := []struct {
			name           string
			request        controllers.RegisterRequest
			expectedStatus int
			expectedError  string
		}{
			{
				name: "Invalid email format",
				request: controllers.RegisterRequest{
					Email:     "invalid-email",
					Password:  "Password123!",
					FirstName: "John",
					LastName:  "Doe",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "VALIDATION_ERROR",
			},
			{
				name: "Weak password",
				request: controllers.RegisterRequest{
					Email:     "test@example.com",
					Password:  "weak",
					FirstName: "John",
					LastName:  "Doe",
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "VALIDATION_ERROR",
			},
			{
				name: "Missing required fields",
				request: controllers.RegisterRequest{
					Email:    "test@example.com",
					Password: "Password123!",
					// FirstName and LastName missing
				},
				expectedStatus: http.StatusBadRequest,
				expectedError:  "VALIDATION_ERROR",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body, _ := json.Marshal(tt.request)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)

				var response utils.APIResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error.Code)
			})
		}
	})

	t.Run("Duplicate Registration", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		// Create a user first
		_, err := testContainer.CreateTestUser("existing@example.com", "Existing", "User")
		require.NoError(t, err)

		// Try to register with the same email
		registerReq := controllers.RegisterRequest{
			Email:     "existing@example.com",
			Password:  "Password123!",
			FirstName: "John",
			LastName:  "Doe",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)
		assert.False(t, response.Success)
		assert.Equal(t, "CONFLICT", response.Error.Code)
	})

	t.Run("Login with Invalid Credentials", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		// Create a user
		_, err := testContainer.CreateTestUser("test@example.com", "Test", "User")
		require.NoError(t, err)

		tests := []struct {
			name     string
			email    string
			password string
		}{
			{
				name:     "Wrong email",
				email:    "wrong@example.com",
				password: "Password123!",
			},
			{
				name:     "Wrong password",
				email:    "test@example.com",
				password: "WrongPassword123!",
			},
			{
				name:     "Non-existent user",
				email:    "nonexistent@example.com",
				password: "Password123!",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				loginReq := controllers.LoginRequest{
					Email:    tt.email,
					Password: tt.password,
				}

				body, _ := json.Marshal(loginReq)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				require.NoError(t, err)
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

				var response utils.APIResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, "UNAUTHORIZED", response.Error.Code)
			})
		}
	})

	t.Run("Token Expiry and Refresh", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		// Register a user
		registerReq := controllers.RegisterRequest{
			Email:     "token@example.com",
			Password:  "Password123!",
			FirstName: "Token",
			LastName:  "User",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var registerResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&registerResponse)
		require.NoError(t, err)

		data, ok := registerResponse.Data.(map[string]interface{})
		require.True(t, ok)
		refreshToken := data["refresh_token"].(string)

		// Test refresh token functionality
		refreshReq := controllers.RefreshTokenRequest{
			RefreshToken: refreshToken,
		}

		refreshBody, _ := json.Marshal(refreshReq)
		req = httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(refreshBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var refreshResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&refreshResponse)
		require.NoError(t, err)
		assert.True(t, refreshResponse.Success)

		// Verify new access token is different
		refreshData, ok := refreshResponse.Data.(map[string]interface{})
		require.True(t, ok)
		newAccessToken := refreshData["access_token"].(string)
		assert.NotEmpty(t, newAccessToken)
	})
}

// TestConcurrentAuthentication tests concurrent authentication operations
func TestConcurrentAuthentication(t *testing.T) {
	testContainer := SetupTestDatabase(t)
	defer testContainer.TeardownTestDatabase(t)

	// Setup services
	tokenBlacklistRepo := repositories.NewTokenBlacklistRepository(testContainer.DB)
	txManager := repositories.NewTransactionManager(testContainer.DB)

	authService := services.NewAuthService(testContainer.DB, txManager)
	jwtManager := utils.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	jwtService := services.NewJWTService(jwtManager, tokenBlacklistRepo, txManager)

	authController := controllers.NewAuthController(authService, jwtService)

	// Setup Fiber app
	app := fiber.New()
	auth := app.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)

	t.Run("Concurrent User Registration", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		const numUsers = 10
		results := make(chan error, numUsers)

		// Register multiple users concurrently
		for i := 0; i < numUsers; i++ {
			go func(id int) {
				registerReq := controllers.RegisterRequest{
					Email:     fmt.Sprintf("user%d@example.com", id),
					Password:  "Password123!",
					FirstName: fmt.Sprintf("User%d", id),
					LastName:  "Test",
				}

				body, _ := json.Marshal(registerReq)
				req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				if err != nil {
					results <- err
					return
				}

				if resp.StatusCode != http.StatusCreated {
					results <- fmt.Errorf("expected status 201, got %d", resp.StatusCode)
					return
				}

				results <- nil
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numUsers; i++ {
			err := <-results
			assert.NoError(t, err)
		}

		// Verify all users were created
		var count int64
		err := testContainer.DB.Model(&repositories.UserModel{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(numUsers), count)
	})

	t.Run("Concurrent Login Attempts", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		// Create a test user
		_, err := testContainer.CreateTestUser("concurrent@example.com", "Concurrent", "User")
		require.NoError(t, err)

		const numAttempts = 5
		results := make(chan error, numAttempts)

		// Perform concurrent login attempts
		for i := 0; i < numAttempts; i++ {
			go func() {
				loginReq := controllers.LoginRequest{
					Email:    "concurrent@example.com",
					Password: "Password123!", // This won't match the pre-hashed password, but tests concurrent access
				}

				body, _ := json.Marshal(loginReq)
				req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				if err != nil {
					results <- err
					return
				}

				// We expect unauthorized since the password won't match the hash
				if resp.StatusCode != http.StatusUnauthorized {
					results <- fmt.Errorf("expected status 401, got %d", resp.StatusCode)
					return
				}

				results <- nil
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < numAttempts; i++ {
			err := <-results
			assert.NoError(t, err)
		}
	})
}

// TestAuthenticationEdgeCases tests edge cases in authentication
func TestAuthenticationEdgeCases(t *testing.T) {
	testContainer := SetupTestDatabase(t)
	defer testContainer.TeardownTestDatabase(t)

	// Setup services
	tokenBlacklistRepo := repositories.NewTokenBlacklistRepository(testContainer.DB)
	txManager := repositories.NewTransactionManager(testContainer.DB)

	authService := services.NewAuthService(testContainer.DB, txManager)
	jwtManager := utils.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	jwtService := services.NewJWTService(jwtManager, tokenBlacklistRepo, txManager)

	authController := controllers.NewAuthController(authService, jwtService)

	// Setup Fiber app
	app := fiber.New()
	auth := app.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/logout", authController.Logout)
	auth.Post("/refresh", authController.RefreshToken)

	t.Run("Malformed JSON Requests", func(t *testing.T) {
		tests := []struct {
			name     string
			endpoint string
			body     string
		}{
			{
				name:     "Invalid JSON in register",
				endpoint: "/auth/register",
				body:     `{"email": "test@example.com", "password": "Password123!", "firstName":}`,
			},
			{
				name:     "Invalid JSON in login",
				endpoint: "/auth/login",
				body:     `{"email": "test@example.com", "password":}`,
			},
			{
				name:     "Empty body in register",
				endpoint: "/auth/register",
				body:     ``,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest("POST", tt.endpoint, bytes.NewBufferString(tt.body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				require.NoError(t, err)
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

				var response utils.APIResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			})
		}
	})

	t.Run("Invalid Token Formats", func(t *testing.T) {
		tests := []struct {
			name         string
			refreshToken string
		}{
			{
				name:         "Empty refresh token",
				refreshToken: "",
			},
			{
				name:         "Invalid JWT format",
				refreshToken: "invalid.jwt.token",
			},
			{
				name:         "Malformed JWT",
				refreshToken: "not-a-jwt-at-all",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				refreshReq := controllers.RefreshTokenRequest{
					RefreshToken: tt.refreshToken,
				}

				body, _ := json.Marshal(refreshReq)
				req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")

				resp, err := app.Test(req)
				require.NoError(t, err)

				// Should be either validation error (400) or unauthorized (401)
				assert.True(t, resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnauthorized)

				var response utils.APIResponse
				err = json.NewDecoder(resp.Body).Decode(&response)
				require.NoError(t, err)
				assert.False(t, response.Success)
			})
		}
	})

	t.Run("Multiple Logout Attempts", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		// Register and login to get tokens
		registerReq := controllers.RegisterRequest{
			Email:     "logout@example.com",
			Password:  "Password123!",
			FirstName: "Logout",
			LastName:  "Test",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var registerResponse utils.APIResponse
		err = json.NewDecoder(resp.Body).Decode(&registerResponse)
		require.NoError(t, err)

		data, ok := registerResponse.Data.(map[string]interface{})
		require.True(t, ok)
		accessToken := data["access_token"].(string)
		refreshToken := data["refresh_token"].(string)

		// First logout - should succeed
		logoutReq := controllers.LogoutRequest{
			RefreshToken: refreshToken,
		}

		logoutBody, _ := json.Marshal(logoutReq)
		req = httptest.NewRequest("POST", "/auth/logout", bytes.NewBuffer(logoutBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Second logout with same tokens - should fail
		req = httptest.NewRequest("POST", "/auth/logout", bytes.NewBuffer(logoutBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

		resp, err = app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestAuthenticationPerformance tests authentication performance
func TestAuthenticationPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	testContainer := SetupTestDatabase(t)
	defer testContainer.TeardownTestDatabase(t)

	// Setup services
	tokenBlacklistRepo := repositories.NewTokenBlacklistRepository(testContainer.DB)
	txManager := repositories.NewTransactionManager(testContainer.DB)

	authService := services.NewAuthService(testContainer.DB, txManager)
	jwtManager := utils.NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	jwtService := services.NewJWTService(jwtManager, tokenBlacklistRepo, txManager)

	authController := controllers.NewAuthController(authService, jwtService)

	// Setup Fiber app
	app := fiber.New()
	auth := app.Group("/auth")
	auth.Post("/register", authController.Register)

	t.Run("Registration Performance", func(t *testing.T) {
		require.NoError(t, testContainer.CleanupTestData())

		start := time.Now()
		const numRegistrations = 100

		for i := 0; i < numRegistrations; i++ {
			registerReq := controllers.RegisterRequest{
				Email:     fmt.Sprintf("perf%d@example.com", i),
				Password:  "Password123!",
				FirstName: fmt.Sprintf("User%d", i),
				LastName:  "Performance",
			}

			body, _ := json.Marshal(registerReq)
			req := httptest.NewRequest("POST", "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			require.NoError(t, err)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
		}

		duration := time.Since(start)
		t.Logf("Registered %d users in %v (%.2f users/second)",
			numRegistrations, duration, float64(numRegistrations)/duration.Seconds())

		// Verify all users were created
		var count int64
		err := testContainer.DB.Model(&repositories.UserModel{}).Count(&count).Error
		require.NoError(t, err)
		assert.Equal(t, int64(numRegistrations), count)
	})
}
