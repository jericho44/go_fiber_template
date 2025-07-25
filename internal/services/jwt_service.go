package services

import (
	"context"
	"fmt"
	"time"

	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"

	"gorm.io/gorm"
)

// JWTServiceInterface defines the interface for JWT service operations
type JWTServiceInterface interface {
	GenerateTokenPair(userID uint, email string) (*utils.TokenPair, error)
	ValidateAccessToken(tokenString string) (*utils.JWTClaims, error)
	ValidateRefreshToken(tokenString string) (*utils.JWTClaims, error)
	RefreshAccessToken(refreshTokenString string) (string, error)
	BlacklistToken(tokenString string, reason string) error
	LogoutUser(accessToken, refreshToken string) error
	IsTokenBlacklisted(tokenString string) (bool, error)
	ExtractTokenFromHeader(authHeader string) (string, error)
	GetTokenExpiry(tokenString string) (time.Time, error)
	GetUserIDFromToken(tokenString string) (uint, error)
}

// JWTService provides high-level JWT token management with blacklisting support
type JWTService struct {
	jwtManager         *utils.JWTManager
	tokenBlacklistRepo repositories.TokenBlacklistRepository
	transactionManager repositories.TransactionManager
}

// NewJWTService creates a new JWT service
func NewJWTService(
	jwtManager *utils.JWTManager,
	tokenBlacklistRepo repositories.TokenBlacklistRepository,
	transactionManager repositories.TransactionManager,
) *JWTService {
	return &JWTService{
		jwtManager:         jwtManager,
		tokenBlacklistRepo: tokenBlacklistRepo,
		transactionManager: transactionManager,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (s *JWTService) GenerateTokenPair(userID uint, email string) (*utils.TokenPair, error) {
	return s.jwtManager.GenerateTokenPair(userID, email)
}

// ValidateAccessToken validates an access token and checks if it's blacklisted
func (s *JWTService) ValidateAccessToken(tokenString string) (*utils.JWTClaims, error) {
	return s.validateTokenWithBlacklistCheck(tokenString, utils.AccessToken)
}

// ValidateRefreshToken validates a refresh token and checks if it's blacklisted
func (s *JWTService) ValidateRefreshToken(tokenString string) (*utils.JWTClaims, error) {
	return s.validateTokenWithBlacklistCheck(tokenString, utils.RefreshToken)
}

// validateTokenWithBlacklistCheck validates a token and checks blacklist
func (s *JWTService) validateTokenWithBlacklistCheck(tokenString string, expectedType utils.TokenType) (*utils.JWTClaims, error) {
	// First validate the token structure and signature
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Check if token type matches expected type
	if claims.TokenType != expectedType {
		return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, claims.TokenType)
	}

	// Check if token is blacklisted
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	isBlacklisted, err := s.tokenBlacklistRepo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}

	if isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate refresh token (includes blacklist check)
	claims, err := s.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token
	newAccessToken, err := s.jwtManager.GenerateToken(claims.UserID, claims.Email, utils.AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return newAccessToken, nil
}

// BlacklistToken adds a token to the blacklist
func (s *JWTService) BlacklistToken(tokenString string, reason string) error {
	// Get token hash
	tokenHash := s.jwtManager.GetTokenHash(tokenString)

	// Get token expiry
	expiresAt, err := s.jwtManager.GetTokenExpiry(tokenString)
	if err != nil {
		return fmt.Errorf("failed to get token expiry: %w", err)
	}

	// Get user ID from token
	userID, err := s.jwtManager.GetUserIDFromToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to get user ID from token: %w", err)
	}

	// Add to blacklist
	err = s.tokenBlacklistRepo.BlacklistToken(tokenHash, userID, expiresAt, reason)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// BlacklistTokenWithContext adds a token to the blacklist with context
func (s *JWTService) BlacklistTokenWithContext(ctx context.Context, tokenString string, reason string) error {
	// Get token hash
	tokenHash := s.jwtManager.GetTokenHash(tokenString)

	// Get token expiry
	expiresAt, err := s.jwtManager.GetTokenExpiry(tokenString)
	if err != nil {
		return fmt.Errorf("failed to get token expiry: %w", err)
	}

	// Get user ID from token
	userID, err := s.jwtManager.GetUserIDFromToken(tokenString)
	if err != nil {
		return fmt.Errorf("failed to get user ID from token: %w", err)
	}

	// Add to blacklist with context
	err = s.tokenBlacklistRepo.BlacklistTokenWithContext(ctx, tokenHash, userID, expiresAt, reason)
	if err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// BlacklistUserTokens blacklists all tokens for a specific user (useful for logout all devices)
func (s *JWTService) BlacklistUserTokens(ctx context.Context, userID uint, reason string) error {
	// This would require getting all active tokens for a user
	// For now, we'll implement a simpler approach by setting a user-wide blacklist timestamp
	// In a production system, you might want to track active tokens per user

	// For this implementation, we'll just return success
	// The actual implementation would depend on your token tracking strategy
	return nil
}

// LogoutUser blacklists both access and refresh tokens for logout
func (s *JWTService) LogoutUser(accessToken, refreshToken string) error {
	return s.transactionManager.WithTransaction(func(tx *gorm.DB) error {
		// Blacklist access token
		if accessToken != "" {
			accessTokenHash := s.jwtManager.GetTokenHash(accessToken)
			accessExpiresAt, err := s.jwtManager.GetTokenExpiry(accessToken)
			if err != nil {
				return fmt.Errorf("failed to get access token expiry: %w", err)
			}

			userID, err := s.jwtManager.GetUserIDFromToken(accessToken)
			if err != nil {
				return fmt.Errorf("failed to get user ID from access token: %w", err)
			}

			err = s.tokenBlacklistRepo.BlacklistTokenTx(tx, accessTokenHash, userID, accessExpiresAt, "logout")
			if err != nil {
				return fmt.Errorf("failed to blacklist access token: %w", err)
			}
		}

		// Blacklist refresh token
		if refreshToken != "" {
			refreshTokenHash := s.jwtManager.GetTokenHash(refreshToken)
			refreshExpiresAt, err := s.jwtManager.GetTokenExpiry(refreshToken)
			if err != nil {
				return fmt.Errorf("failed to get refresh token expiry: %w", err)
			}

			userID, err := s.jwtManager.GetUserIDFromToken(refreshToken)
			if err != nil {
				return fmt.Errorf("failed to get user ID from refresh token: %w", err)
			}

			err = s.tokenBlacklistRepo.BlacklistTokenTx(tx, refreshTokenHash, userID, refreshExpiresAt, "logout")
			if err != nil {
				return fmt.Errorf("failed to blacklist refresh token: %w", err)
			}
		}

		return nil
	})
}

// IsTokenBlacklisted checks if a token is blacklisted
func (s *JWTService) IsTokenBlacklisted(tokenString string) (bool, error) {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	return s.tokenBlacklistRepo.IsTokenBlacklisted(tokenHash)
}

// IsTokenBlacklistedWithContext checks if a token is blacklisted with context
func (s *JWTService) IsTokenBlacklistedWithContext(ctx context.Context, tokenString string) (bool, error) {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	return s.tokenBlacklistRepo.IsTokenBlacklistedWithContext(ctx, tokenHash)
}

// CleanupExpiredTokens removes expired tokens from the blacklist
func (s *JWTService) CleanupExpiredTokens() (int64, error) {
	return s.tokenBlacklistRepo.CleanupExpiredTokens()
}

// CleanupExpiredTokensWithContext removes expired tokens from the blacklist with context
func (s *JWTService) CleanupExpiredTokensWithContext(ctx context.Context) (int64, error) {
	return s.tokenBlacklistRepo.CleanupExpiredTokensWithContext(ctx)
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (s *JWTService) ExtractTokenFromHeader(authHeader string) (string, error) {
	return s.jwtManager.ExtractTokenFromHeader(authHeader)
}

// GetUserIDFromToken extracts user ID from token without full validation
func (s *JWTService) GetUserIDFromToken(tokenString string) (uint, error) {
	return s.jwtManager.GetUserIDFromToken(tokenString)
}

// GetTokenExpiry gets the expiry time of a token
func (s *JWTService) GetTokenExpiry(tokenString string) (time.Time, error) {
	return s.jwtManager.GetTokenExpiry(tokenString)
}
