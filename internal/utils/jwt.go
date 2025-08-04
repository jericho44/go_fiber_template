package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenType represents the type of JWT token
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID            uint      `json:"user_id"`
	Email             string    `json:"email"`
	TokenType         TokenType `json:"token_type"`
	SessionID         string    `json:"session_id"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	IPAddress         string    `json:"ip_address"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	TokenType    string `json:"token_type"`
	SessionID    string `json:"session_id"`
}

// TokenContext represents the context for token generation
type TokenContext struct {
	UserID            uint
	Email             string
	SessionID         string
	DeviceFingerprint string
	IPAddress         string
}

// JWTManager handles JWT token operations
type JWTManager struct {
	secret        string
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        secret,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair generates both access and refresh tokens for a user (legacy method)
func (j *JWTManager) GenerateTokenPair(userID uint, email string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := j.GenerateToken(userID, email, AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := j.GenerateToken(userID, email, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(j.accessExpiry).Unix(),
		TokenType:    "Bearer",
	}, nil
}

// GenerateTokenPairWithContext generates both access and refresh tokens with session context
func (j *JWTManager) GenerateTokenPairWithContext(ctx TokenContext) (*TokenPair, error) {
	// Generate access token with context
	accessToken, err := j.GenerateTokenWithContext(ctx, AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token with context
	refreshToken, err := j.GenerateTokenWithContext(ctx, RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(j.accessExpiry).Unix(),
		TokenType:    "Bearer",
		SessionID:    ctx.SessionID,
	}, nil
}

// GenerateToken generates a JWT token for a user (legacy method)
func (j *JWTManager) GenerateToken(userID uint, email string, tokenType TokenType) (string, error) {
	var expiry time.Duration
	switch tokenType {
	case AccessToken:
		expiry = j.accessExpiry
	case RefreshToken:
		expiry = j.refreshExpiry
	default:
		return "", errors.New("invalid token type")
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:    userID,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-fiber-template",
			Subject:   fmt.Sprintf("user:%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateTokenWithContext generates a JWT token with session context
func (j *JWTManager) GenerateTokenWithContext(ctx TokenContext, tokenType TokenType) (string, error) {
	var expiry time.Duration
	switch tokenType {
	case AccessToken:
		expiry = j.accessExpiry
	case RefreshToken:
		expiry = j.refreshExpiry
	default:
		return "", errors.New("invalid token type")
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:            ctx.UserID,
		Email:             ctx.Email,
		TokenType:         tokenType,
		SessionID:         ctx.SessionID,
		DeviceFingerprint: ctx.DeviceFingerprint,
		IPAddress:         ctx.IPAddress,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "go-fiber-template",
			Subject:   fmt.Sprintf("user:%d", ctx.UserID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (j *JWTManager) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Check if header starts with "Bearer "
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := strings.TrimSpace(authHeader[len(bearerPrefix):])
	if token == "" {
		return "", errors.New("token is required")
	}

	return token, nil
}

// GetTokenHash generates a SHA256 hash of the token for blacklisting
func (j *JWTManager) GetTokenHash(tokenString string) string {
	hash := sha256.Sum256([]byte(tokenString))
	return hex.EncodeToString(hash[:])
}

// GetTokenExpiry extracts the expiry time from a token without validating it
func (j *JWTManager) GetTokenExpiry(tokenString string) (time.Time, error) {
	// Parse token without validation to get expiry
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return time.Time{}, errors.New("invalid token claims")
	}

	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("token has no expiry")
	}

	return claims.ExpiresAt.Time, nil
}

// RefreshAccessToken generates a new access token using a valid refresh token
func (j *JWTManager) RefreshAccessToken(refreshTokenString string) (string, error) {
	// Validate the refresh token
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Ensure it's a refresh token
	if claims.TokenType != RefreshToken {
		return "", errors.New("token is not a refresh token")
	}

	// Generate new access token
	accessToken, err := j.GenerateToken(claims.UserID, claims.Email, AccessToken)
	if err != nil {
		return "", fmt.Errorf("failed to generate new access token: %w", err)
	}

	return accessToken, nil
}

// GetUserIDFromToken extracts user ID from a token without full validation
func (j *JWTManager) GetUserIDFromToken(tokenString string) (uint, error) {
	// Parse token without validation to get user ID
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return 0, errors.New("invalid token claims")
	}

	return claims.UserID, nil
}
