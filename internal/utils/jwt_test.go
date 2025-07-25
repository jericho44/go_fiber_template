package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJWTManager(t *testing.T) {
	secret := "test-secret-key-that-is-long-enough-for-testing"
	accessExpiry := 15 * time.Minute
	refreshExpiry := 7 * 24 * time.Hour

	manager := NewJWTManager(secret, accessExpiry, refreshExpiry)

	assert.NotNil(t, manager)
	assert.Equal(t, secret, manager.secret)
	assert.Equal(t, accessExpiry, manager.accessExpiry)
	assert.Equal(t, refreshExpiry, manager.refreshExpiry)
}

func TestJWTManager_GenerateToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)
	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name      string
		tokenType TokenType
		wantError bool
	}{
		{
			name:      "Generate access token",
			tokenType: AccessToken,
			wantError: false,
		},
		{
			name:      "Generate refresh token",
			tokenType: RefreshToken,
			wantError: false,
		},
		{
			name:      "Invalid token type",
			tokenType: TokenType("invalid"),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.GenerateToken(userID, email, tt.tokenType)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)

				// Verify token can be parsed
				claims, err := manager.ValidateToken(token)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, tt.tokenType, claims.TokenType)
			}
		})
	}
}

func TestJWTManager_GenerateTokenPair(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)
	userID := uint(123)
	email := "test@example.com"

	tokenPair, err := manager.GenerateTokenPair(userID, email)

	require.NoError(t, err)
	require.NotNil(t, tokenPair)

	// Verify access token
	assert.NotEmpty(t, tokenPair.AccessToken)
	accessClaims, err := manager.ValidateToken(tokenPair.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, email, accessClaims.Email)
	assert.Equal(t, AccessToken, accessClaims.TokenType)

	// Verify refresh token
	assert.NotEmpty(t, tokenPair.RefreshToken)
	refreshClaims, err := manager.ValidateToken(tokenPair.RefreshToken)
	assert.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, email, refreshClaims.Email)
	assert.Equal(t, RefreshToken, refreshClaims.TokenType)

	// Verify token pair metadata
	assert.Equal(t, "Bearer", tokenPair.TokenType)
	assert.True(t, tokenPair.ExpiresAt > time.Now().Unix())
}

func TestJWTManager_ValidateToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)
	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name      string
		setupFunc func() string
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid access token",
			setupFunc: func() string {
				token, _ := manager.GenerateToken(userID, email, AccessToken)
				return token
			},
			wantError: false,
		},
		{
			name: "Valid refresh token",
			setupFunc: func() string {
				token, _ := manager.GenerateToken(userID, email, RefreshToken)
				return token
			},
			wantError: false,
		},
		{
			name: "Invalid token format",
			setupFunc: func() string {
				return "invalid-token"
			},
			wantError: true,
			errorMsg:  "failed to parse token",
		},
		{
			name: "Empty token",
			setupFunc: func() string {
				return ""
			},
			wantError: true,
			errorMsg:  "failed to parse token",
		},
		{
			name: "Token with wrong secret",
			setupFunc: func() string {
				wrongManager := NewJWTManager("wrong-secret-key-that-is-different", 15*time.Minute, 7*24*time.Hour)
				token, _ := wrongManager.GenerateToken(userID, email, AccessToken)
				return token
			},
			wantError: true,
			errorMsg:  "failed to parse token",
		},
		{
			name: "Expired token",
			setupFunc: func() string {
				expiredManager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", -1*time.Hour, 7*24*time.Hour)
				token, _ := expiredManager.GenerateToken(userID, email, AccessToken)
				return token
			},
			wantError: true,
			errorMsg:  "token is expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token := tt.setupFunc()
			claims, err := manager.ValidateToken(token)

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, claims)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
			}
		})
	}
}

func TestJWTManager_ExtractTokenFromHeader(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name       string
		authHeader string
		wantError  bool
		errorMsg   string
	}{
		{
			name:       "Valid Bearer token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantError:  false,
		},
		{
			name:       "Empty header",
			authHeader: "",
			wantError:  true,
			errorMsg:   "authorization header is required",
		},
		{
			name:       "Missing Bearer prefix",
			authHeader: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantError:  true,
			errorMsg:   "authorization header must start with 'Bearer '",
		},
		{
			name:       "Bearer without token",
			authHeader: "Bearer ",
			wantError:  true,
			errorMsg:   "token is required",
		},
		{
			name:       "Bearer with only spaces",
			authHeader: "Bearer    ",
			wantError:  true,
			errorMsg:   "token is required",
		},
		{
			name:       "Wrong prefix",
			authHeader: "Basic eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			wantError:  true,
			errorMsg:   "authorization header must start with 'Bearer '",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := manager.ExtractTokenFromHeader(tt.authHeader)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, token)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token)
			}
		})
	}
}

func TestJWTManager_GetTokenHash(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "Valid token",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
		},
		{
			name:  "Empty token",
			token: "",
		},
		{
			name:  "Short token",
			token: "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := manager.GetTokenHash(tt.token)
			hash2 := manager.GetTokenHash(tt.token)

			// Hash should be consistent
			assert.Equal(t, hash1, hash2)
			// Hash should be 64 characters (SHA256 hex)
			assert.Len(t, hash1, 64)
			// Hash should be different for different tokens
			if tt.token != "" {
				differentHash := manager.GetTokenHash(tt.token + "different")
				assert.NotEqual(t, hash1, differentHash)
			}
		})
	}
}

func TestJWTManager_RefreshAccessToken(t *testing.T) {
	manager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, 7*24*time.Hour)
	userID := uint(123)
	email := "test@example.com"

	tests := []struct {
		name      string
		setupFunc func() string
		wantError bool
		errorMsg  string
	}{
		{
			name: "Valid refresh token",
			setupFunc: func() string {
				token, _ := manager.GenerateToken(userID, email, RefreshToken)
				return token
			},
			wantError: false,
		},
		{
			name: "Access token instead of refresh token",
			setupFunc: func() string {
				token, _ := manager.GenerateToken(userID, email, AccessToken)
				return token
			},
			wantError: true,
			errorMsg:  "token is not a refresh token",
		},
		{
			name: "Invalid token",
			setupFunc: func() string {
				return "invalid-token"
			},
			wantError: true,
			errorMsg:  "invalid refresh token",
		},
		{
			name: "Expired refresh token",
			setupFunc: func() string {
				expiredManager := NewJWTManager("test-secret-key-that-is-long-enough-for-testing", 15*time.Minute, -1*time.Hour)
				token, _ := expiredManager.GenerateToken(userID, email, RefreshToken)
				return token
			},
			wantError: true,
			errorMsg:  "invalid refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			refreshToken := tt.setupFunc()
			newAccessToken, err := manager.RefreshAccessToken(refreshToken)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, newAccessToken)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, newAccessToken)

				// Verify the new access token
				claims, err := manager.ValidateToken(newAccessToken)
				assert.NoError(t, err)
				assert.Equal(t, userID, claims.UserID)
				assert.Equal(t, email, claims.Email)
				assert.Equal(t, AccessToken, claims.TokenType)
			}
		})
	}
}
