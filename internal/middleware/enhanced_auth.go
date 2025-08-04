package middleware

import (
	"context"
	"strings"
	"time"

	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// EnhancedAuthMiddleware provides enhanced JWT authentication with session tracking
type EnhancedAuthMiddleware struct {
	jwtService           services.EnhancedJWTService
	fingerprintGenerator *utils.DeviceFingerprintGenerator
	skipPaths            []string
}

// NewEnhancedAuthMiddleware creates a new enhanced authentication middleware
func NewEnhancedAuthMiddleware(jwtService services.EnhancedJWTService) *EnhancedAuthMiddleware {
	return &EnhancedAuthMiddleware{
		jwtService:           jwtService,
		fingerprintGenerator: utils.NewDeviceFingerprintGenerator(),
		skipPaths: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/refresh",
			"/api/v1/health",
			"/docs",
			"/swagger",
		},
	}
}

// RequireAuth middleware that requires valid JWT authentication with enhanced security
func (m *EnhancedAuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip authentication for certain paths
		path := c.Path()
		for _, skipPath := range m.skipPaths {
			if strings.HasPrefix(path, skipPath) {
				return c.Next()
			}
		}

		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Authorization header is required")
		}

		token, err := m.jwtService.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid authorization header format")
		}

		// Create request context
		requestCtx := m.createRequestContext(c)

		// Validate token with enhanced security checks
		claims, err := m.jwtService.ValidateAccessTokenWithContext(context.Background(), token, requestCtx)
		if err != nil {
			if strings.Contains(err.Error(), "suspicious activity") {
				return utils.ErrorResponse(c, 403, utils.CodeForbidden, "Suspicious activity detected", nil)
			}
			return utils.UnauthorizedResponse(c, "Invalid or expired token")
		}

		// Store user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("session_id", claims.SessionID)
		c.Locals("device_fingerprint", claims.DeviceFingerprint)
		c.Locals("jwt_claims", claims)

		return c.Next()
	}
}

// OptionalAuth middleware that optionally validates JWT authentication
func (m *EnhancedAuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next() // No token provided, continue without authentication
		}

		token, err := m.jwtService.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return c.Next() // Invalid header format, continue without authentication
		}

		// Create request context
		requestCtx := m.createRequestContext(c)

		// Validate token with enhanced security checks
		claims, err := m.jwtService.ValidateAccessTokenWithContext(context.Background(), token, requestCtx)
		if err != nil {
			return c.Next() // Invalid token, continue without authentication
		}

		// Store user information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("session_id", claims.SessionID)
		c.Locals("device_fingerprint", claims.DeviceFingerprint)
		c.Locals("jwt_claims", claims)

		return c.Next()
	}
}

// RequireRefreshToken middleware for refresh token endpoints
func (m *EnhancedAuthMiddleware) RequireRefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Parse request body to get refresh token
		var req struct {
			RefreshToken string `json:"refresh_token"`
		}

		if err := c.BodyParser(&req); err != nil {
			return utils.BadRequestResponse(c, "Invalid request body", nil)
		}

		if req.RefreshToken == "" {
			return utils.BadRequestResponse(c, "Refresh token is required", nil)
		}

		// Create request context
		requestCtx := m.createRequestContext(c)

		// Validate refresh token
		claims, err := m.jwtService.ValidateRefreshTokenWithContext(context.Background(), req.RefreshToken, requestCtx)
		if err != nil {
			return utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		}

		// Store information in context
		c.Locals("user_id", claims.UserID)
		c.Locals("user_email", claims.Email)
		c.Locals("session_id", claims.SessionID)
		c.Locals("device_fingerprint", claims.DeviceFingerprint)
		c.Locals("refresh_token", req.RefreshToken)
		c.Locals("jwt_claims", claims)

		return c.Next()
	}
}

// createRequestContext creates a request context from Fiber context
func (m *EnhancedAuthMiddleware) createRequestContext(c *fiber.Ctx) services.RequestContext {
	// Get IP address (handle proxy headers)
	ipAddress := c.IP()
	if forwarded := c.Get("X-Forwarded-For"); forwarded != "" {
		// Take the first IP in the chain
		if idx := strings.Index(forwarded, ","); idx != -1 {
			ipAddress = strings.TrimSpace(forwarded[:idx])
		} else {
			ipAddress = strings.TrimSpace(forwarded)
		}
	}

	// Get user agent
	userAgent := c.Get("User-Agent")

	// Get endpoint
	endpoint := c.Method() + " " + c.Path()

	// Extract location from headers (if available)
	location := c.Get("CF-IPCountry") // Cloudflare country header
	if location == "" {
		location = c.Get("X-Country-Code") // Alternative country header
	}

	return services.RequestContext{
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Endpoint:  endpoint,
		Location:  location,
		Timestamp: time.Now(),
	}
}

// GenerateDeviceFingerprint generates a device fingerprint from request headers
func (m *EnhancedAuthMiddleware) GenerateDeviceFingerprint(c *fiber.Ctx) string {
	headers := map[string]string{
		"User-Agent":      c.Get("User-Agent"),
		"Accept-Language": c.Get("Accept-Language"),
		"Accept-Encoding": c.Get("Accept-Encoding"),
		"Accept-Charset":  c.Get("Accept-Charset"),
	}

	ipAddress := c.IP()
	return m.fingerprintGenerator.GenerateFingerprintFromHeaders(headers, ipAddress)
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) (uint, bool) {
	userID, ok := c.Locals("user_id").(uint)
	return userID, ok
}

// GetUserEmail extracts user email from context
func GetUserEmail(c *fiber.Ctx) (string, bool) {
	email, ok := c.Locals("user_email").(string)
	return email, ok
}

// GetSessionID extracts session ID from context
func GetSessionID(c *fiber.Ctx) (string, bool) {
	sessionID, ok := c.Locals("session_id").(string)
	return sessionID, ok
}

// GetDeviceFingerprint extracts device fingerprint from context
func GetDeviceFingerprint(c *fiber.Ctx) (string, bool) {
	fingerprint, ok := c.Locals("device_fingerprint").(string)
	return fingerprint, ok
}

// GetJWTClaims extracts JWT claims from context
func GetJWTClaims(c *fiber.Ctx) (*utils.JWTClaims, bool) {
	claims, ok := c.Locals("jwt_claims").(*utils.JWTClaims)
	return claims, ok
}
