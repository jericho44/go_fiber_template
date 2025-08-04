package controllers

import (
	"context"
	"strings"
	"time"

	"go-fiber-template/internal/middleware"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// EnhancedAuthController handles authentication with enhanced security features
type EnhancedAuthController struct {
	authService          services.AuthService
	enhancedJWTService   services.EnhancedJWTService
	validator            *utils.ValidationErrorFormatter
	fingerprintGenerator *utils.DeviceFingerprintGenerator
}

// NewEnhancedAuthController creates a new enhanced authentication controller
func NewEnhancedAuthController(
	authService services.AuthService,
	enhancedJWTService services.EnhancedJWTService,
) *EnhancedAuthController {
	return &EnhancedAuthController{
		authService:          authService,
		enhancedJWTService:   enhancedJWTService,
		validator:            utils.NewValidationErrorFormatter(),
		fingerprintGenerator: utils.NewDeviceFingerprintGenerator(),
	}
}

// EnhancedAuthResponse represents the enhanced authentication response
type EnhancedAuthResponse struct {
	User              interface{} `json:"user"`
	AccessToken       string      `json:"access_token"`
	RefreshToken      string      `json:"refresh_token"`
	ExpiresAt         int64       `json:"expires_at"`
	RefreshExpiresAt  int64       `json:"refresh_expires_at"`
	TokenType         string      `json:"token_type"`
	SessionID         string      `json:"session_id"`
	DeviceFingerprint string      `json:"device_fingerprint"`
	SecurityLevel     string      `json:"security_level"`
}

// SessionInfo represents session information
type SessionInfo struct {
	SessionID         string    `json:"session_id"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	IPAddress         string    `json:"ip_address"`
	UserAgent         string    `json:"user_agent"`
	Location          string    `json:"location,omitempty"`
	LastUsedAt        time.Time `json:"last_used_at"`
	ExpiresAt         time.Time `json:"expires_at"`
	IsActive          bool      `json:"is_active"`
}

// Register handles user registration with enhanced security
// @Summary Register a new user with enhanced security
// @Description Create a new user account with enhanced JWT tokens and session tracking
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request with user details"
// @Success 201 {object} utils.APIResponse{data=EnhancedAuthResponse} "User registered successfully with enhanced tokens"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 409 {object} utils.APIResponse{error=utils.APIError} "Email already exists"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/register [post]
func (ac *EnhancedAuthController) Register(c *fiber.Ctx) error {
	var req RegisterRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request body", nil)
	}

	// Validate request
	if validationErr := ac.validator.ValidateStruct(req); validationErr != nil {
		return utils.ValidationErrorResponse(c, validationErr.Details)
	}

	// Convert to service request
	serviceReq := services.RegisterRequest{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Register user
	authResponse, err := ac.authService.Register(context.Background(), serviceReq)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case "VALIDATION_ERROR":
				return utils.ValidationErrorResponse(c, appErr.Details)
			case "CONFLICT":
				return utils.ErrorResponse(c, 409, utils.CodeBadRequest, appErr.Message, nil)
			default:
				return utils.InternalServerErrorResponse(c, "Registration failed")
			}
		}
		return utils.InternalServerErrorResponse(c, "Registration failed")
	}

	// Generate device fingerprint
	deviceFingerprint := ac.generateDeviceFingerprint(c)

	// Create token context
	tokenCtx := utils.TokenContext{
		UserID:            authResponse.User.ID,
		Email:             authResponse.User.Email,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         c.IP(),
	}

	// Generate enhanced JWT tokens
	tokenPair, err := ac.enhancedJWTService.GenerateTokenPairWithContext(context.Background(), tokenCtx)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to generate tokens")
	}

	// Prepare enhanced response
	response := EnhancedAuthResponse{
		User:              authResponse.User,
		AccessToken:       tokenPair.AccessToken,
		RefreshToken:      tokenPair.RefreshToken,
		ExpiresAt:         tokenPair.ExpiresAt.Unix(),
		RefreshExpiresAt:  tokenPair.RefreshExpiresAt.Unix(),
		TokenType:         tokenPair.TokenType,
		SessionID:         tokenPair.SessionID,
		DeviceFingerprint: tokenPair.DeviceFingerprint,
		SecurityLevel:     tokenPair.SecurityLevel,
	}

	return utils.CreatedResponse(c, "User registered successfully", response)
}

// Login handles user authentication with enhanced security
// @Summary Login user with enhanced security
// @Description Authenticate user with enhanced JWT tokens and session tracking
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.APIResponse{data=EnhancedAuthResponse} "Login successful with enhanced tokens"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid credentials"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/login [post]
func (ac *EnhancedAuthController) Login(c *fiber.Ctx) error {
	var req LoginRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request body", nil)
	}

	// Validate request
	if validationErr := ac.validator.ValidateStruct(req); validationErr != nil {
		return utils.ValidationErrorResponse(c, validationErr.Details)
	}

	// Convert to service request
	serviceReq := services.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	}

	// Authenticate user
	authResponse, err := ac.authService.Login(context.Background(), serviceReq)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case "VALIDATION_ERROR":
				return utils.ValidationErrorResponse(c, appErr.Details)
			case "UNAUTHORIZED":
				return utils.UnauthorizedResponse(c, appErr.Message)
			default:
				return utils.InternalServerErrorResponse(c, "Login failed")
			}
		}
		return utils.InternalServerErrorResponse(c, "Login failed")
	}

	// Generate device fingerprint
	deviceFingerprint := ac.generateDeviceFingerprint(c)

	// Create token context
	tokenCtx := utils.TokenContext{
		UserID:            authResponse.User.ID,
		Email:             authResponse.User.Email,
		DeviceFingerprint: deviceFingerprint,
		IPAddress:         c.IP(),
	}

	// Generate enhanced JWT tokens
	tokenPair, err := ac.enhancedJWTService.GenerateTokenPairWithContext(context.Background(), tokenCtx)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to generate tokens")
	}

	// Prepare enhanced response
	response := EnhancedAuthResponse{
		User:              authResponse.User,
		AccessToken:       tokenPair.AccessToken,
		RefreshToken:      tokenPair.RefreshToken,
		ExpiresAt:         tokenPair.ExpiresAt.Unix(),
		RefreshExpiresAt:  tokenPair.RefreshExpiresAt.Unix(),
		TokenType:         tokenPair.TokenType,
		SessionID:         tokenPair.SessionID,
		DeviceFingerprint: tokenPair.DeviceFingerprint,
		SecurityLevel:     tokenPair.SecurityLevel,
	}

	return utils.SuccessResponse(c, "Login successful", response)
}

// RefreshToken handles token refresh with rotation
// @Summary Refresh access token with rotation
// @Description Generate new access and refresh tokens with rotation for enhanced security
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} utils.APIResponse{data=EnhancedAuthResponse} "Tokens refreshed successfully"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or expired refresh token"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/refresh [post]
func (ac *EnhancedAuthController) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request body", nil)
	}

	// Validate request
	if validationErr := ac.validator.ValidateStruct(req); validationErr != nil {
		return utils.ValidationErrorResponse(c, validationErr.Details)
	}

	// Create request context
	requestCtx := services.RequestContext{
		IPAddress: c.IP(),
		UserAgent: c.Get("User-Agent"),
		Endpoint:  "POST /auth/v2/refresh",
		Timestamp: time.Now(),
	}

	// Rotate tokens
	tokenPair, err := ac.enhancedJWTService.RotateTokens(context.Background(), req.RefreshToken, requestCtx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid refresh token") ||
			strings.Contains(err.Error(), "token has been revoked") ||
			strings.Contains(err.Error(), "token has expired") ||
			strings.Contains(err.Error(), "session not found") {
			return utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		}
		return utils.InternalServerErrorResponse(c, "Failed to refresh token")
	}

	// Prepare enhanced response
	response := EnhancedAuthResponse{
		AccessToken:       tokenPair.AccessToken,
		RefreshToken:      tokenPair.RefreshToken,
		ExpiresAt:         tokenPair.ExpiresAt.Unix(),
		RefreshExpiresAt:  tokenPair.RefreshExpiresAt.Unix(),
		TokenType:         tokenPair.TokenType,
		SessionID:         tokenPair.SessionID,
		DeviceFingerprint: tokenPair.DeviceFingerprint,
		SecurityLevel:     tokenPair.SecurityLevel,
	}

	return utils.SuccessResponse(c, "Tokens refreshed successfully", response)
}

// Logout handles enhanced logout with session cleanup
// @Summary Logout user with session cleanup
// @Description Logout user and cleanup session with enhanced security
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT access token" Format(Bearer {token})
// @Param request body LogoutRequest false "Optional logout request with refresh token"
// @Success 200 {object} utils.APIResponse "Logout successful"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or missing authorization header"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/logout [post]
// @Security BearerAuth
func (ac *EnhancedAuthController) Logout(c *fiber.Ctx) error {
	// Get access token from Authorization header
	authHeader := c.Get("Authorization")
	accessToken, err := ac.enhancedJWTService.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid authorization header")
	}

	// Parse request body for refresh token (optional)
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		// If body parsing fails, continue with just access token logout
		req.RefreshToken = ""
	}

	// Logout session with enhanced cleanup
	err = ac.enhancedJWTService.LogoutSession(context.Background(), accessToken, req.RefreshToken)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to logout user")
	}

	return utils.SuccessResponse(c, "Logout successful", nil)
}

// LogoutAllSessions handles logout from all sessions
// @Summary Logout from all sessions
// @Description Logout user from all active sessions
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT access token" Format(Bearer {token})
// @Success 200 {object} utils.APIResponse "Logout from all sessions successful"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or missing authorization header"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/logout-all [post]
// @Security BearerAuth
func (ac *EnhancedAuthController) LogoutAllSessions(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	// Logout from all sessions
	err := ac.enhancedJWTService.LogoutAllSessions(context.Background(), userID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to logout from all sessions")
	}

	return utils.SuccessResponse(c, "Logout from all sessions successful", nil)
}

// GetActiveSessions retrieves user's active sessions
// @Summary Get active sessions
// @Description Retrieve all active sessions for the authenticated user
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT access token" Format(Bearer {token})
// @Success 200 {object} utils.APIResponse{data=[]SessionInfo} "Active sessions retrieved successfully"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or missing authorization header"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/sessions [get]
// @Security BearerAuth
func (ac *EnhancedAuthController) GetActiveSessions(c *fiber.Ctx) error {
	// Get user ID from context (set by middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	// Get active sessions
	sessions, err := ac.enhancedJWTService.GetUserActiveSessions(context.Background(), userID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to retrieve sessions")
	}

	// Convert to response format
	var sessionInfos []SessionInfo
	for _, session := range sessions {
		sessionInfos = append(sessionInfos, SessionInfo{
			SessionID:         session.SessionID,
			DeviceFingerprint: session.DeviceFingerprint,
			IPAddress:         session.IPAddress,
			UserAgent:         session.UserAgent,
			Location:          session.Location,
			LastUsedAt:        session.LastUsedAt,
			ExpiresAt:         session.ExpiresAt,
			IsActive:          session.IsActive,
		})
	}

	return utils.SuccessResponse(c, "Active sessions retrieved successfully", sessionInfos)
}

// RevokeSession revokes a specific session
// @Summary Revoke a specific session
// @Description Revoke a specific session by session ID
// @Tags Enhanced Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT access token" Format(Bearer {token})
// @Param sessionId path string true "Session ID to revoke"
// @Success 200 {object} utils.APIResponse "Session revoked successfully"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or missing authorization header"
// @Failure 404 {object} utils.APIResponse{error=utils.APIError} "Session not found"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/v2/sessions/{sessionId}/revoke [post]
// @Security BearerAuth
func (ac *EnhancedAuthController) RevokeSession(c *fiber.Ctx) error {
	// Get session ID from URL parameter
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return utils.BadRequestResponse(c, "Session ID is required", nil)
	}

	// Get user ID from context (set by middleware)
	userID, ok := middleware.GetUserID(c)
	if !ok {
		return utils.UnauthorizedResponse(c, "User not authenticated")
	}

	// Verify session belongs to user
	session, err := ac.enhancedJWTService.GetActiveSession(context.Background(), sessionID)
	if err != nil {
		return utils.NotFoundResponse(c, "Session not found")
	}

	if session.UserID != userID {
		return utils.ForbiddenResponse(c, "Access denied")
	}

	// Deactivate session
	err = ac.enhancedJWTService.DeactivateSession(context.Background(), sessionID)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to revoke session")
	}

	return utils.SuccessResponse(c, "Session revoked successfully", nil)
}

// generateDeviceFingerprint generates a device fingerprint from request headers
func (ac *EnhancedAuthController) generateDeviceFingerprint(c *fiber.Ctx) string {
	headers := map[string]string{
		"User-Agent":      c.Get("User-Agent"),
		"Accept-Language": c.Get("Accept-Language"),
		"Accept-Encoding": c.Get("Accept-Encoding"),
		"Accept-Charset":  c.Get("Accept-Charset"),
	}

	ipAddress := c.IP()
	return ac.fingerprintGenerator.GenerateFingerprintFromHeaders(headers, ipAddress)
}
