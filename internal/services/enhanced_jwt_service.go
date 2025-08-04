package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"

	"gorm.io/gorm"
)

// EnhancedJWTService provides advanced JWT token management with rotation, fingerprinting, and anomaly detection
type EnhancedJWTService interface {
	// Enhanced token generation with session context
	GenerateTokenPairWithContext(ctx context.Context, tokenCtx utils.TokenContext) (*EnhancedTokenPair, error)

	// Token validation with enhanced security checks
	ValidateAccessTokenWithContext(ctx context.Context, tokenString string, requestCtx RequestContext) (*utils.JWTClaims, error)
	ValidateRefreshTokenWithContext(ctx context.Context, tokenString string, requestCtx RequestContext) (*utils.JWTClaims, error)

	// Token rotation
	RotateTokens(ctx context.Context, refreshToken string, requestCtx RequestContext) (*EnhancedTokenPair, error)

	// Session management
	CreateTokenSession(ctx context.Context, tokenCtx utils.TokenContext, refreshTokenHash string) (*models.TokenSession, error)
	GetActiveSession(ctx context.Context, sessionID string) (*models.TokenSession, error)
	DeactivateSession(ctx context.Context, sessionID string) error
	DeactivateAllUserSessions(ctx context.Context, userID uint) error

	// Enhanced logout with session cleanup
	LogoutSession(ctx context.Context, accessToken, refreshToken string) error
	LogoutAllSessions(ctx context.Context, userID uint) error

	// Token usage tracking
	RecordTokenUsage(ctx context.Context, claims *utils.JWTClaims, requestCtx RequestContext) error

	// Anomaly detection
	DetectAnomalies(ctx context.Context, claims *utils.JWTClaims, requestCtx RequestContext) (*AnomalyResult, error)

	// Redis-based blacklisting
	BlacklistTokenInRedis(ctx context.Context, tokenString string, expiration time.Duration) error
	IsTokenBlacklistedInRedis(ctx context.Context, tokenString string) (bool, error)

	// Session queries
	GetUserActiveSessions(ctx context.Context, userID uint) ([]*models.TokenSession, error)

	// Cleanup operations
	CleanupExpiredSessions(ctx context.Context) (int64, error)

	// Legacy compatibility
	JWTServiceInterface
}

// EnhancedTokenPair extends TokenPair with additional security information
type EnhancedTokenPair struct {
	*utils.TokenPair
	SessionID         string    `json:"session_id"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	ExpiresAt         time.Time `json:"expires_at"`
	RefreshExpiresAt  time.Time `json:"refresh_expires_at"`
	IsRotated         bool      `json:"is_rotated"`
	SecurityLevel     string    `json:"security_level"`
}

// RequestContext contains information about the current request
type RequestContext struct {
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	Endpoint  string    `json:"endpoint"`
	Location  string    `json:"location,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// enhancedJWTServiceImpl implements EnhancedJWTService
type enhancedJWTServiceImpl struct {
	// Legacy components
	jwtManager         *utils.JWTManager
	tokenBlacklistRepo repositories.TokenBlacklistRepository
	transactionManager repositories.TransactionManager

	// Enhanced components
	tokenSessionRepo repositories.TokenSessionRepository
	redisService     RedisService
	anomalyDetector  AnomalyDetector
}

// NewEnhancedJWTService creates a new enhanced JWT service
func NewEnhancedJWTService(
	jwtManager *utils.JWTManager,
	tokenBlacklistRepo repositories.TokenBlacklistRepository,
	transactionManager repositories.TransactionManager,
	tokenSessionRepo repositories.TokenSessionRepository,
	redisService RedisService,
	anomalyDetector AnomalyDetector,
) EnhancedJWTService {
	return &enhancedJWTServiceImpl{
		jwtManager:         jwtManager,
		tokenBlacklistRepo: tokenBlacklistRepo,
		transactionManager: transactionManager,
		tokenSessionRepo:   tokenSessionRepo,
		redisService:       redisService,
		anomalyDetector:    anomalyDetector,
	}
}

// GenerateTokenPairWithContext generates tokens with enhanced session context
func (s *enhancedJWTServiceImpl) GenerateTokenPairWithContext(ctx context.Context, tokenCtx utils.TokenContext) (*EnhancedTokenPair, error) {
	// Generate session ID if not provided
	if tokenCtx.SessionID == "" {
		sessionID, err := s.generateSessionID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate session ID: %w", err)
		}
		tokenCtx.SessionID = sessionID
	}

	// Generate token pair with context
	tokenPair, err := s.jwtManager.GenerateTokenPairWithContext(tokenCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Create token session
	refreshTokenHash := s.jwtManager.GetTokenHash(tokenPair.RefreshToken)
	session, err := s.CreateTokenSession(ctx, tokenCtx, refreshTokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to create token session: %w", err)
	}

	// Store refresh token hash in Redis for fast lookup
	err = s.redisService.Set(ctx,
		fmt.Sprintf("session:refresh:%s", refreshTokenHash),
		tokenCtx.SessionID,
		time.Until(session.ExpiresAt))
	if err != nil {
		// Log error but don't fail - database is primary source
	}

	return &EnhancedTokenPair{
		TokenPair:         tokenPair,
		SessionID:         tokenCtx.SessionID,
		DeviceFingerprint: tokenCtx.DeviceFingerprint,
		ExpiresAt:         time.Now().Add(15 * time.Minute), // Access token expiry
		RefreshExpiresAt:  session.ExpiresAt,
		IsRotated:         false,
		SecurityLevel:     "standard",
	}, nil
}

// ValidateAccessTokenWithContext validates access token with enhanced security checks
func (s *enhancedJWTServiceImpl) ValidateAccessTokenWithContext(ctx context.Context, tokenString string, requestCtx RequestContext) (*utils.JWTClaims, error) {
	// First validate the token structure and signature
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		// Record failed validation attempt
		s.recordSecurityEvent(ctx, "token_validation_failed", map[string]interface{}{
			"error":      err.Error(),
			"ip_address": requestCtx.IPAddress,
			"user_agent": requestCtx.UserAgent,
			"endpoint":   requestCtx.Endpoint,
		})
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Check if token type is access
	if claims.TokenType != utils.AccessToken {
		return nil, fmt.Errorf("invalid token type: expected access, got %s", claims.TokenType)
	}

	tokenHash := s.jwtManager.GetTokenHash(tokenString)

	// Check Redis blacklist first (faster) with enhanced metadata
	isBlacklisted, err := s.IsTokenBlacklistedInRedis(ctx, tokenString)
	if err == nil && isBlacklisted {
		// Record blacklisted token usage attempt
		s.recordSecurityEvent(ctx, "blacklisted_token_usage", map[string]interface{}{
			"token_hash": tokenHash,
			"user_id":    claims.UserID,
			"ip_address": requestCtx.IPAddress,
			"user_agent": requestCtx.UserAgent,
			"endpoint":   requestCtx.Endpoint,
		})
		return nil, fmt.Errorf("token has been revoked")
	}

	// Fallback to database blacklist check
	isBlacklisted, err = s.tokenBlacklistRepo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		// Record blacklisted token usage attempt
		s.recordSecurityEvent(ctx, "blacklisted_token_usage", map[string]interface{}{
			"token_hash": tokenHash,
			"user_id":    claims.UserID,
			"ip_address": requestCtx.IPAddress,
			"user_agent": requestCtx.UserAgent,
			"endpoint":   requestCtx.Endpoint,
		})
		return nil, fmt.Errorf("token has been revoked")
	}

	// Track token usage in Redis for analytics
	usageMetadata := map[string]interface{}{
		"ip_address": requestCtx.IPAddress,
		"user_agent": requestCtx.UserAgent,
		"endpoint":   requestCtx.Endpoint,
		"timestamp":  requestCtx.Timestamp.Unix(),
		"user_id":    claims.UserID,
	}
	_ = s.redisService.TrackTokenUsage(ctx, tokenHash, usageMetadata)

	// Verify session is still active
	if claims.SessionID != "" {
		session, err := s.GetActiveSession(ctx, claims.SessionID)
		if err != nil {
			// Record session validation failure
			s.recordSecurityEvent(ctx, "session_validation_failed", map[string]interface{}{
				"session_id": claims.SessionID,
				"user_id":    claims.UserID,
				"error":      err.Error(),
				"ip_address": requestCtx.IPAddress,
			})
			return nil, fmt.Errorf("session not found or inactive: %w", err)
		}

		// Enhanced session validation
		if err := s.validateSessionContext(ctx, session, requestCtx, claims); err != nil {
			return nil, fmt.Errorf("session context validation failed: %w", err)
		}

		// Update session last used time
		session.UpdateLastUsed()
		err = s.tokenSessionRepo.UpdateSession(ctx, session)
		if err != nil {
			// Log error but don't fail validation
		}
	}

	// Record token usage for anomaly detection
	err = s.RecordTokenUsage(ctx, claims, requestCtx)
	if err != nil {
		// Log error but don't fail validation
	}

	// Perform enhanced anomaly detection
	anomalyResult, err := s.DetectAnomalies(ctx, claims, requestCtx)
	if err != nil {
		// Log error but don't fail validation
	} else if anomalyResult != nil && anomalyResult.IsAnomaly {
		// Record anomaly detection
		s.recordSecurityEvent(ctx, "anomaly_detected", map[string]interface{}{
			"anomaly_type": anomalyResult.AnomalyType,
			"severity":     anomalyResult.Severity,
			"confidence":   anomalyResult.Confidence,
			"description":  anomalyResult.Description,
			"user_id":      claims.UserID,
			"session_id":   claims.SessionID,
			"ip_address":   requestCtx.IPAddress,
			"user_agent":   requestCtx.UserAgent,
		})

		// For critical anomalies, fail validation
		if anomalyResult.Severity == SeverityCritical {
			return nil, fmt.Errorf("suspicious activity detected: %s", anomalyResult.Description)
		}

		// For high severity anomalies, consider additional security measures
		if anomalyResult.Severity == SeverityHigh {
			// Could implement additional checks here, like requiring re-authentication
			// For now, we'll allow but log the event
		}
	}

	// Record successful token validation
	s.recordSecurityEvent(ctx, "token_validated", map[string]interface{}{
		"user_id":    claims.UserID,
		"session_id": claims.SessionID,
		"ip_address": requestCtx.IPAddress,
		"endpoint":   requestCtx.Endpoint,
	})

	return claims, nil
}

// ValidateRefreshTokenWithContext validates refresh token with enhanced security checks
func (s *enhancedJWTServiceImpl) ValidateRefreshTokenWithContext(ctx context.Context, tokenString string, requestCtx RequestContext) (*utils.JWTClaims, error) {
	// First validate the token structure and signature
	claims, err := s.jwtManager.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Check if token type is refresh
	if claims.TokenType != utils.RefreshToken {
		return nil, fmt.Errorf("invalid token type: expected refresh, got %s", claims.TokenType)
	}

	// Check Redis blacklist first
	isBlacklisted, err := s.IsTokenBlacklistedInRedis(ctx, tokenString)
	if err == nil && isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Check database blacklist
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	isBlacklisted, err = s.tokenBlacklistRepo.IsTokenBlacklisted(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Verify session exists and is active
	if claims.SessionID != "" {
		session, err := s.tokenSessionRepo.GetSessionByRefreshToken(ctx, tokenHash)
		if err != nil {
			return nil, fmt.Errorf("session not found: %w", err)
		}

		if !session.IsActive || session.IsExpired() {
			return nil, fmt.Errorf("session is inactive or expired")
		}
	}

	return claims, nil
}

// RotateTokens performs token rotation with enhanced security tracking
func (s *enhancedJWTServiceImpl) RotateTokens(ctx context.Context, refreshToken string, requestCtx RequestContext) (*EnhancedTokenPair, error) {
	// Validate refresh token
	claims, err := s.ValidateRefreshTokenWithContext(ctx, refreshToken, requestCtx)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get current session
	oldRefreshHash := s.jwtManager.GetTokenHash(refreshToken)
	session, err := s.tokenSessionRepo.GetSessionByRefreshToken(ctx, oldRefreshHash)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	// Enhanced security checks before rotation
	if err := s.performPreRotationSecurityChecks(ctx, session, requestCtx); err != nil {
		return nil, fmt.Errorf("security check failed: %w", err)
	}

	var newTokenPair *EnhancedTokenPair

	// Perform rotation in transaction
	err = s.transactionManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Generate new session ID for enhanced security
		newSessionID, err := s.generateSessionID()
		if err != nil {
			return fmt.Errorf("failed to generate new session ID: %w", err)
		}

		// Generate new token pair with new session ID
		tokenCtx := utils.TokenContext{
			UserID:            claims.UserID,
			Email:             claims.Email,
			SessionID:         newSessionID,
			DeviceFingerprint: claims.DeviceFingerprint,
			IPAddress:         requestCtx.IPAddress,
		}

		newPair, err := s.jwtManager.GenerateTokenPairWithContext(tokenCtx)
		if err != nil {
			return fmt.Errorf("failed to generate new token pair: %w", err)
		}

		// Update session with new refresh token hash and session ID
		newRefreshHash := s.jwtManager.GetTokenHash(newPair.RefreshToken)
		session.SessionID = newSessionID
		session.RefreshTokenHash = newRefreshHash
		session.IPAddress = requestCtx.IPAddress
		session.UserAgent = requestCtx.UserAgent
		session.UpdateLastUsed()

		err = s.tokenSessionRepo.UpdateSessionTx(tx, session)
		if err != nil {
			return fmt.Errorf("failed to update session: %w", err)
		}

		// Record token rotation with enhanced metadata
		rotation := &models.TokenRotation{
			SessionID:      newSessionID,
			OldTokenHash:   oldRefreshHash,
			NewTokenHash:   newRefreshHash,
			RotationReason: s.determineRotationReason(ctx, session, requestCtx),
			IPAddress:      requestCtx.IPAddress,
			UserAgent:      requestCtx.UserAgent,
			RotatedAt:      time.Now(),
		}

		err = s.tokenSessionRepo.RecordTokenRotationTx(tx, rotation)
		if err != nil {
			return fmt.Errorf("failed to record token rotation: %w", err)
		}

		// Blacklist old refresh token with extended expiry for security
		err = s.tokenBlacklistRepo.BlacklistTokenTx(tx, oldRefreshHash, claims.UserID,
			time.Now().Add(7*24*time.Hour), "rotation")
		if err != nil {
			return fmt.Errorf("failed to blacklist old token: %w", err)
		}

		// Determine security level based on rotation context
		securityLevel := s.determineSecurityLevel(ctx, session, requestCtx)

		newTokenPair = &EnhancedTokenPair{
			TokenPair:         newPair,
			SessionID:         newSessionID,
			DeviceFingerprint: claims.DeviceFingerprint,
			ExpiresAt:         time.Now().Add(15 * time.Minute),
			RefreshExpiresAt:  session.ExpiresAt,
			IsRotated:         true,
			SecurityLevel:     securityLevel,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Update Redis cache with new session mapping
	newRefreshHash := s.jwtManager.GetTokenHash(newTokenPair.RefreshToken)
	_ = s.redisService.Set(ctx,
		fmt.Sprintf("session:refresh:%s", newRefreshHash),
		newTokenPair.SessionID,
		time.Until(session.ExpiresAt))

	// Remove old session mapping from Redis
	_ = s.redisService.Del(ctx, fmt.Sprintf("session:refresh:%s", oldRefreshHash))

	// Blacklist old token in Redis with extended expiry
	_ = s.BlacklistTokenInRedis(ctx, refreshToken, 7*24*time.Hour)

	// Record successful rotation for analytics
	_ = s.recordRotationMetrics(ctx, claims.UserID, requestCtx)

	return newTokenPair, nil
}

// CreateTokenSession creates a new token session
func (s *enhancedJWTServiceImpl) CreateTokenSession(ctx context.Context, tokenCtx utils.TokenContext, refreshTokenHash string) (*models.TokenSession, error) {
	session := &models.TokenSession{
		UserID:            tokenCtx.UserID,
		SessionID:         tokenCtx.SessionID,
		RefreshTokenHash:  refreshTokenHash,
		DeviceFingerprint: tokenCtx.DeviceFingerprint,
		IPAddress:         tokenCtx.IPAddress,
		IsActive:          true,
		LastUsedAt:        time.Now(),
		ExpiresAt:         time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	err := s.tokenSessionRepo.CreateSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetActiveSession retrieves an active session
func (s *enhancedJWTServiceImpl) GetActiveSession(ctx context.Context, sessionID string) (*models.TokenSession, error) {
	return s.tokenSessionRepo.GetSessionByID(ctx, sessionID)
}

// DeactivateSession deactivates a session
func (s *enhancedJWTServiceImpl) DeactivateSession(ctx context.Context, sessionID string) error {
	return s.tokenSessionRepo.DeactivateSession(ctx, sessionID)
}

// DeactivateAllUserSessions deactivates all sessions for a user
func (s *enhancedJWTServiceImpl) DeactivateAllUserSessions(ctx context.Context, userID uint) error {
	return s.tokenSessionRepo.DeactivateUserSessions(ctx, userID)
}

// LogoutSession logs out a specific session
func (s *enhancedJWTServiceImpl) LogoutSession(ctx context.Context, accessToken, refreshToken string) error {
	return s.transactionManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		var sessionID string

		// Get session ID from access token
		if accessToken != "" {
			claims, err := s.jwtManager.ValidateToken(accessToken)
			if err == nil {
				sessionID = claims.SessionID
			}
		}

		// Get session ID from refresh token if not found
		if sessionID == "" && refreshToken != "" {
			claims, err := s.jwtManager.ValidateToken(refreshToken)
			if err == nil {
				sessionID = claims.SessionID
			}
		}

		// Deactivate session
		if sessionID != "" {
			err := s.tokenSessionRepo.DeactivateSessionTx(tx, sessionID)
			if err != nil {
				return fmt.Errorf("failed to deactivate session: %w", err)
			}
		}

		// Blacklist tokens
		if accessToken != "" {
			accessTokenHash := s.jwtManager.GetTokenHash(accessToken)
			accessExpiresAt, _ := s.jwtManager.GetTokenExpiry(accessToken)
			userID, _ := s.jwtManager.GetUserIDFromToken(accessToken)

			_ = s.tokenBlacklistRepo.BlacklistTokenTx(tx, accessTokenHash, userID, accessExpiresAt, "logout")
			_ = s.BlacklistTokenInRedis(ctx, accessToken, time.Until(accessExpiresAt))
		}

		if refreshToken != "" {
			refreshTokenHash := s.jwtManager.GetTokenHash(refreshToken)
			refreshExpiresAt, _ := s.jwtManager.GetTokenExpiry(refreshToken)
			userID, _ := s.jwtManager.GetUserIDFromToken(refreshToken)

			_ = s.tokenBlacklistRepo.BlacklistTokenTx(tx, refreshTokenHash, userID, refreshExpiresAt, "logout")
			_ = s.BlacklistTokenInRedis(ctx, refreshToken, time.Until(refreshExpiresAt))
		}

		return nil
	})
}

// LogoutAllSessions logs out all sessions for a user
func (s *enhancedJWTServiceImpl) LogoutAllSessions(ctx context.Context, userID uint) error {
	return s.DeactivateAllUserSessions(ctx, userID)
}

// RecordTokenUsage records token usage for tracking
func (s *enhancedJWTServiceImpl) RecordTokenUsage(ctx context.Context, claims *utils.JWTClaims, requestCtx RequestContext) error {
	if claims.SessionID == "" {
		return nil // Skip if no session ID
	}

	usage := &models.TokenUsage{
		SessionID: claims.SessionID,
		TokenType: string(claims.TokenType),
		IPAddress: requestCtx.IPAddress,
		UserAgent: requestCtx.UserAgent,
		Endpoint:  requestCtx.Endpoint,
		UsedAt:    requestCtx.Timestamp,
	}

	return s.tokenSessionRepo.RecordTokenUsage(ctx, usage)
}

// DetectAnomalies performs anomaly detection on token usage
func (s *enhancedJWTServiceImpl) DetectAnomalies(ctx context.Context, claims *utils.JWTClaims, requestCtx RequestContext) (*AnomalyResult, error) {
	if s.anomalyDetector == nil {
		return nil, nil
	}

	usage := &models.TokenUsage{
		SessionID: claims.SessionID,
		TokenType: string(claims.TokenType),
		IPAddress: requestCtx.IPAddress,
		UserAgent: requestCtx.UserAgent,
		Endpoint:  requestCtx.Endpoint,
		UsedAt:    requestCtx.Timestamp,
	}

	return s.anomalyDetector.AnalyzeTokenUsage(ctx, usage)
}

// BlacklistTokenInRedis blacklists a token in Redis
func (s *enhancedJWTServiceImpl) BlacklistTokenInRedis(ctx context.Context, tokenString string, expiration time.Duration) error {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	return s.redisService.BlacklistToken(ctx, tokenHash, expiration)
}

// IsTokenBlacklistedInRedis checks if a token is blacklisted in Redis
func (s *enhancedJWTServiceImpl) IsTokenBlacklistedInRedis(ctx context.Context, tokenString string) (bool, error) {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	return s.redisService.IsTokenBlacklisted(ctx, tokenHash)
}

// GetUserActiveSessions retrieves active sessions for a user
func (s *enhancedJWTServiceImpl) GetUserActiveSessions(ctx context.Context, userID uint) ([]*models.TokenSession, error) {
	return s.tokenSessionRepo.GetActiveSessions(ctx, userID)
}

// CleanupExpiredSessions removes expired sessions
func (s *enhancedJWTServiceImpl) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.tokenSessionRepo.CleanupExpiredSessions(ctx)
}

// generateSessionID generates a unique session ID
func (s *enhancedJWTServiceImpl) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Legacy compatibility methods - delegate to original JWT service

func (s *enhancedJWTServiceImpl) GenerateTokenPair(userID uint, email string) (*utils.TokenPair, error) {
	return s.jwtManager.GenerateTokenPair(userID, email)
}

func (s *enhancedJWTServiceImpl) ValidateAccessToken(tokenString string) (*utils.JWTClaims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

func (s *enhancedJWTServiceImpl) ValidateRefreshToken(tokenString string) (*utils.JWTClaims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}

func (s *enhancedJWTServiceImpl) RefreshAccessToken(refreshTokenString string) (string, error) {
	return s.jwtManager.RefreshAccessToken(refreshTokenString)
}

func (s *enhancedJWTServiceImpl) BlacklistToken(tokenString string, reason string) error {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	expiresAt, _ := s.jwtManager.GetTokenExpiry(tokenString)
	userID, _ := s.jwtManager.GetUserIDFromToken(tokenString)
	return s.tokenBlacklistRepo.BlacklistToken(tokenHash, userID, expiresAt, reason)
}

func (s *enhancedJWTServiceImpl) LogoutUser(accessToken, refreshToken string) error {
	return s.LogoutSession(context.Background(), accessToken, refreshToken)
}

func (s *enhancedJWTServiceImpl) IsTokenBlacklisted(tokenString string) (bool, error) {
	tokenHash := s.jwtManager.GetTokenHash(tokenString)
	return s.tokenBlacklistRepo.IsTokenBlacklisted(tokenHash)
}

func (s *enhancedJWTServiceImpl) ExtractTokenFromHeader(authHeader string) (string, error) {
	return s.jwtManager.ExtractTokenFromHeader(authHeader)
}

func (s *enhancedJWTServiceImpl) GetTokenExpiry(tokenString string) (time.Time, error) {
	return s.jwtManager.GetTokenExpiry(tokenString)
}

func (s *enhancedJWTServiceImpl) GetUserIDFromToken(tokenString string) (uint, error) {
	return s.jwtManager.GetUserIDFromToken(tokenString)
}

// Enhanced security helper methods

// performPreRotationSecurityChecks performs security checks before token rotation
func (s *enhancedJWTServiceImpl) performPreRotationSecurityChecks(ctx context.Context, session *models.TokenSession, requestCtx RequestContext) error {
	// Check for suspicious IP changes
	if session.IPAddress != requestCtx.IPAddress {
		// Allow IP changes but record them for anomaly detection
		if s.anomalyDetector != nil {
			anomalyResult, err := s.anomalyDetector.CheckIPAnomalies(ctx, session.UserID, requestCtx.IPAddress)
			if err == nil && anomalyResult.IsAnomaly && anomalyResult.Severity == SeverityCritical {
				return fmt.Errorf("suspicious IP change detected: %s", anomalyResult.Description)
			}
		}
	}

	// Check for suspicious user agent changes
	if session.UserAgent != requestCtx.UserAgent {
		if s.anomalyDetector != nil {
			anomalyResult, err := s.anomalyDetector.CheckUserAgentAnomalies(ctx, session.UserID, requestCtx.UserAgent)
			if err == nil && anomalyResult.IsAnomaly && anomalyResult.Severity >= SeverityHigh {
				return fmt.Errorf("suspicious user agent change detected: %s", anomalyResult.Description)
			}
		}
	}

	// Check rotation frequency
	rotations, err := s.tokenSessionRepo.GetTokenRotations(ctx, session.SessionID, 10)
	if err == nil && len(rotations) > 0 {
		// Check if last rotation was too recent (less than 1 minute ago)
		lastRotation := rotations[0]
		if time.Since(lastRotation.RotatedAt) < time.Minute {
			return fmt.Errorf("token rotation too frequent, please wait before rotating again")
		}

		// Check for excessive rotations in the last hour
		recentRotations := 0
		oneHourAgo := time.Now().Add(-time.Hour)
		for _, rotation := range rotations {
			if rotation.RotatedAt.After(oneHourAgo) {
				recentRotations++
			}
		}

		if recentRotations > 10 {
			return fmt.Errorf("too many token rotations in the last hour")
		}
	}

	return nil
}

// determineRotationReason determines the reason for token rotation based on context
func (s *enhancedJWTServiceImpl) determineRotationReason(ctx context.Context, session *models.TokenSession, requestCtx RequestContext) string {
	// Check for IP change
	if session.IPAddress != requestCtx.IPAddress {
		return "ip_change"
	}

	// Check for user agent change
	if session.UserAgent != requestCtx.UserAgent {
		return "user_agent_change"
	}

	// Check if this is a scheduled rotation (based on session age)
	sessionAge := time.Since(session.CreatedAt)
	if sessionAge > 24*time.Hour {
		return "scheduled"
	}

	// Check for anomaly-triggered rotation
	if s.anomalyDetector != nil {
		usage := &models.TokenUsage{
			SessionID: session.SessionID,
			TokenType: "refresh",
			IPAddress: requestCtx.IPAddress,
			UserAgent: requestCtx.UserAgent,
			Endpoint:  requestCtx.Endpoint,
			UsedAt:    requestCtx.Timestamp,
		}

		anomalyResult, err := s.anomalyDetector.AnalyzeTokenUsage(ctx, usage)
		if err == nil && anomalyResult.IsAnomaly {
			return "anomaly_detected"
		}
	}

	// Default reason
	return "refresh"
}

// determineSecurityLevel determines the security level based on rotation context
func (s *enhancedJWTServiceImpl) determineSecurityLevel(ctx context.Context, session *models.TokenSession, requestCtx RequestContext) string {
	// Start with standard level
	level := "standard"

	// Upgrade to enhanced if there are context changes
	if session.IPAddress != requestCtx.IPAddress || session.UserAgent != requestCtx.UserAgent {
		level = "enhanced"
	}

	// Check for anomalies
	if s.anomalyDetector != nil {
		usage := &models.TokenUsage{
			SessionID: session.SessionID,
			TokenType: "refresh",
			IPAddress: requestCtx.IPAddress,
			UserAgent: requestCtx.UserAgent,
			Endpoint:  requestCtx.Endpoint,
			UsedAt:    requestCtx.Timestamp,
		}

		anomalyResult, err := s.anomalyDetector.AnalyzeTokenUsage(ctx, usage)
		if err == nil && anomalyResult.IsAnomaly {
			switch anomalyResult.Severity {
			case SeverityHigh, SeverityCritical:
				level = "high_security"
			case SeverityMedium:
				if level == "standard" {
					level = "enhanced"
				}
			}
		}
	}

	// Check session age - older sessions get higher security
	sessionAge := time.Since(session.CreatedAt)
	if sessionAge > 7*24*time.Hour && level == "standard" {
		level = "enhanced"
	}

	return level
}

// recordRotationMetrics records metrics for token rotation analytics
func (s *enhancedJWTServiceImpl) recordRotationMetrics(ctx context.Context, userID uint, requestCtx RequestContext) error {
	// Record rotation count in Redis for analytics
	key := fmt.Sprintf("metrics:rotation:user:%d:daily", userID)
	_, err := s.redisService.IncrementCounter(ctx, key, 24*time.Hour)
	if err != nil {
		return fmt.Errorf("failed to record rotation metrics: %w", err)
	}

	// Record rotation by IP for security monitoring
	ipKey := fmt.Sprintf("metrics:rotation:ip:%s:hourly", requestCtx.IPAddress)
	_, err = s.redisService.IncrementCounter(ctx, ipKey, time.Hour)
	if err != nil {
		return fmt.Errorf("failed to record IP rotation metrics: %w", err)
	}

	return nil
}

// validateSessionContext performs enhanced session context validation
func (s *enhancedJWTServiceImpl) validateSessionContext(ctx context.Context, session *models.TokenSession, requestCtx RequestContext, claims *utils.JWTClaims) error {
	// Check device fingerprint consistency
	if claims.DeviceFingerprint != "" && session.DeviceFingerprint != claims.DeviceFingerprint {
		// Record device fingerprint mismatch
		s.recordSecurityEvent(ctx, "device_fingerprint_mismatch", map[string]interface{}{
			"session_id":          session.SessionID,
			"user_id":             session.UserID,
			"session_fingerprint": session.DeviceFingerprint,
			"token_fingerprint":   claims.DeviceFingerprint,
			"ip_address":          requestCtx.IPAddress,
		})

		// For now, we'll allow but log - in production you might want to fail
		// return fmt.Errorf("device fingerprint mismatch detected")
	}

	// Check for significant IP address changes (different subnets)
	if session.IPAddress != requestCtx.IPAddress {
		// Check if IPs are from different subnets (basic check)
		if !s.areIPsInSameSubnet(session.IPAddress, requestCtx.IPAddress) {
			s.recordSecurityEvent(ctx, "significant_ip_change", map[string]interface{}{
				"session_id": session.SessionID,
				"user_id":    session.UserID,
				"old_ip":     session.IPAddress,
				"new_ip":     requestCtx.IPAddress,
			})
		}
	}

	// Check session expiry with buffer
	if session.IsExpired() {
		return fmt.Errorf("session has expired")
	}

	// Check if session is close to expiry (within 1 hour)
	if time.Until(session.ExpiresAt) < time.Hour {
		s.recordSecurityEvent(ctx, "session_near_expiry", map[string]interface{}{
			"session_id":     session.SessionID,
			"user_id":        session.UserID,
			"expires_at":     session.ExpiresAt.Unix(),
			"time_remaining": time.Until(session.ExpiresAt).Seconds(),
		})
	}

	return nil
}

// recordSecurityEvent records a security event in Redis
func (s *enhancedJWTServiceImpl) recordSecurityEvent(ctx context.Context, eventType string, metadata map[string]interface{}) {
	if s.redisService == nil {
		return // Skip if Redis is not available
	}

	// Add timestamp to metadata
	metadata["timestamp"] = time.Now().Unix()

	// Record the event with 7-day expiry
	err := s.redisService.RecordSecurityEvent(ctx, eventType, metadata, 7*24*time.Hour)
	if err != nil {
		// Log error but don't fail the operation
		// In production, you'd want proper logging here
	}
}

// areIPsInSameSubnet checks if two IP addresses are in the same subnet (basic implementation)
func (s *enhancedJWTServiceImpl) areIPsInSameSubnet(ip1, ip2 string) bool {
	// Simple implementation - check if first 3 octets are the same for IPv4
	// In production, you'd want more sophisticated subnet checking
	if ip1 == ip2 {
		return true
	}

	// For IPv4, check if first 3 octets match
	parts1 := strings.Split(ip1, ".")
	parts2 := strings.Split(ip2, ".")

	if len(parts1) == 4 && len(parts2) == 4 {
		// Check first 3 octets
		for i := 0; i < 3; i++ {
			if parts1[i] != parts2[i] {
				return false
			}
		}
		return true
	}

	// For IPv6 or other formats, consider them different subnets for now
	return false
}
