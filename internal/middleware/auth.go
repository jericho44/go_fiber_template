package middleware

import (
	"context"
	"strings"

	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// UserContextKey is the key used to store user information in the context
const UserContextKey = "user"

// AuthMiddleware creates a middleware for JWT authentication
type AuthMiddleware struct {
	jwtManager         *utils.JWTManager
	userRepo           repositories.UserRepository
	tokenBlacklistRepo repositories.TokenBlacklistRepository
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(
	jwtManager *utils.JWTManager,
	userRepo repositories.UserRepository,
	tokenBlacklistRepo repositories.TokenBlacklistRepository,
) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:         jwtManager,
		userRepo:           userRepo,
		tokenBlacklistRepo: tokenBlacklistRepo,
	}
}

// RequireAuth creates a middleware that requires authentication
func (am *AuthMiddleware) RequireAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Authorization header is required")
		}

		// Extract token from Bearer format
		token, err := am.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		// Validate JWT token
		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return utils.UnauthorizedResponse(c, "Token has expired")
			}
			return utils.UnauthorizedResponse(c, "Invalid token")
		}

		// Check if token is blacklisted
		tokenHash := am.jwtManager.GetTokenHash(token)
		isBlacklisted, err := am.tokenBlacklistRepo.IsTokenBlacklistedWithContext(c.Context(), tokenHash)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to check token blacklist")
		}
		if isBlacklisted {
			return utils.UnauthorizedResponse(c, "Token has been invalidated")
		}

		// Ensure it's an access token
		if claims.TokenType != utils.AccessToken {
			return utils.UnauthorizedResponse(c, "Access token required")
		}

		// Get user from database
		user, err := am.userRepo.GetByIDWithContext(c.Context(), claims.UserID)
		if err != nil {
			return utils.UnauthorizedResponse(c, "User not found")
		}

		// Check if user is active
		if !user.GetIsActive() {
			return utils.UnauthorizedResponse(c, "User account is inactive")
		}

		// Store user in context
		c.Locals(UserContextKey, user)

		return c.Next()
	}
}

// OptionalAuth creates a middleware that optionally authenticates users
// If a valid token is provided, the user is added to context
// If no token or invalid token, the request continues without user context
func (am *AuthMiddleware) OptionalAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			return c.Next()
		}

		// Extract token from Bearer format
		token, err := am.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			// Invalid token format, continue without authentication
			return c.Next()
		}

		// Validate JWT token
		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			// Invalid token, continue without authentication
			return c.Next()
		}

		// Check if token is blacklisted
		tokenHash := am.jwtManager.GetTokenHash(token)
		isBlacklisted, err := am.tokenBlacklistRepo.IsTokenBlacklistedWithContext(c.Context(), tokenHash)
		if err != nil || isBlacklisted {
			// Token blacklisted or error checking, continue without authentication
			return c.Next()
		}

		// Ensure it's an access token
		if claims.TokenType != utils.AccessToken {
			// Wrong token type, continue without authentication
			return c.Next()
		}

		// Get user from database
		user, err := am.userRepo.GetByIDWithContext(c.Context(), claims.UserID)
		if err != nil || !user.GetIsActive() {
			// User not found or inactive, continue without authentication
			return c.Next()
		}

		// Store user in context
		c.Locals(UserContextKey, user)

		return c.Next()
	}
}

// RequireActiveUser creates a middleware that requires an active authenticated user
func (am *AuthMiddleware) RequireActiveUser() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Run the authentication logic inline
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Authorization header is required")
		}

		// Extract token from Bearer format
		token, err := am.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		// Validate JWT token
		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return utils.UnauthorizedResponse(c, "Token has expired")
			}
			return utils.UnauthorizedResponse(c, "Invalid token")
		}

		// Check if token is blacklisted
		tokenHash := am.jwtManager.GetTokenHash(token)
		isBlacklisted, err := am.tokenBlacklistRepo.IsTokenBlacklistedWithContext(c.Context(), tokenHash)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to check token blacklist")
		}
		if isBlacklisted {
			return utils.UnauthorizedResponse(c, "Token has been invalidated")
		}

		// Ensure it's an access token
		if claims.TokenType != utils.AccessToken {
			return utils.UnauthorizedResponse(c, "Access token required")
		}

		// Get user from database
		user, err := am.userRepo.GetByIDWithContext(c.Context(), claims.UserID)
		if err != nil {
			return utils.UnauthorizedResponse(c, "User not found")
		}

		// Check if user is active (this is the main purpose of this middleware)
		if !user.GetIsActive() {
			return utils.UnauthorizedResponse(c, "User account is inactive")
		}

		// Store user in context
		c.Locals(UserContextKey, user)

		return c.Next()
	}
}

// RequireRefreshToken creates a middleware that validates refresh tokens
func (am *AuthMiddleware) RequireRefreshToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Authorization header is required")
		}

		// Extract token from Bearer format
		token, err := am.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		// Validate JWT token
		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return utils.UnauthorizedResponse(c, "Refresh token has expired")
			}
			return utils.UnauthorizedResponse(c, "Invalid refresh token")
		}

		// Check if token is blacklisted
		tokenHash := am.jwtManager.GetTokenHash(token)
		isBlacklisted, err := am.tokenBlacklistRepo.IsTokenBlacklistedWithContext(c.Context(), tokenHash)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to check token blacklist")
		}
		if isBlacklisted {
			return utils.UnauthorizedResponse(c, "Refresh token has been invalidated")
		}

		// Ensure it's a refresh token
		if claims.TokenType != utils.RefreshToken {
			return utils.UnauthorizedResponse(c, "Refresh token required")
		}

		// Get user from database
		user, err := am.userRepo.GetByIDWithContext(c.Context(), claims.UserID)
		if err != nil {
			return utils.UnauthorizedResponse(c, "User not found")
		}

		// Check if user is active
		if !user.GetIsActive() {
			return utils.UnauthorizedResponse(c, "User account is inactive")
		}

		// Store user and token claims in context
		c.Locals(UserContextKey, user)
		c.Locals("token_claims", claims)
		c.Locals("refresh_token", token)

		return c.Next()
	}
}

// GetUserFromContext retrieves the authenticated user from the Fiber context
func GetUserFromContext(c *fiber.Ctx) repositories.UserEntity {
	user := c.Locals(UserContextKey)
	if user == nil {
		return nil
	}

	userEntity, ok := user.(repositories.UserEntity)
	if !ok {
		return nil
	}

	return userEntity
}

// GetUserIDFromContext retrieves the authenticated user ID from the Fiber context
func GetUserIDFromContext(c *fiber.Ctx) uint {
	user := GetUserFromContext(c)
	if user == nil {
		return 0
	}
	return user.GetID()
}

// GetTokenClaimsFromContext retrieves the token claims from the Fiber context
func GetTokenClaimsFromContext(c *fiber.Ctx) *utils.JWTClaims {
	claims := c.Locals("token_claims")
	if claims == nil {
		return nil
	}

	jwtClaims, ok := claims.(*utils.JWTClaims)
	if !ok {
		return nil
	}

	return jwtClaims
}

// GetRefreshTokenFromContext retrieves the refresh token from the Fiber context
func GetRefreshTokenFromContext(c *fiber.Ctx) string {
	token := c.Locals("refresh_token")
	if token == nil {
		return ""
	}

	tokenStr, ok := token.(string)
	if !ok {
		return ""
	}

	return tokenStr
}

// IsAuthenticated checks if the current request is authenticated
func IsAuthenticated(c *fiber.Ctx) bool {
	return GetUserFromContext(c) != nil
}

// RequireUserID creates a middleware that ensures the authenticated user matches a specific user ID
// This is useful for endpoints where users can only access their own resources
func (am *AuthMiddleware) RequireUserID(userIDParam string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// First run the authentication logic inline
		// Extract token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return utils.UnauthorizedResponse(c, "Authorization header is required")
		}

		// Extract token from Bearer format
		token, err := am.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			return utils.UnauthorizedResponse(c, err.Error())
		}

		// Validate JWT token
		claims, err := am.jwtManager.ValidateToken(token)
		if err != nil {
			if strings.Contains(err.Error(), "expired") {
				return utils.UnauthorizedResponse(c, "Token has expired")
			}
			return utils.UnauthorizedResponse(c, "Invalid token")
		}

		// Check if token is blacklisted
		tokenHash := am.jwtManager.GetTokenHash(token)
		isBlacklisted, err := am.tokenBlacklistRepo.IsTokenBlacklistedWithContext(c.Context(), tokenHash)
		if err != nil {
			return utils.InternalServerErrorResponse(c, "Failed to check token blacklist")
		}
		if isBlacklisted {
			return utils.UnauthorizedResponse(c, "Token has been invalidated")
		}

		// Ensure it's an access token
		if claims.TokenType != utils.AccessToken {
			return utils.UnauthorizedResponse(c, "Access token required")
		}

		// Get user from database
		user, err := am.userRepo.GetByIDWithContext(c.Context(), claims.UserID)
		if err != nil {
			return utils.UnauthorizedResponse(c, "User not found")
		}

		// Check if user is active
		if !user.GetIsActive() {
			return utils.UnauthorizedResponse(c, "User account is inactive")
		}

		// Store user in context
		c.Locals(UserContextKey, user)

		// Get authenticated user ID
		authenticatedUserID := user.GetID()
		if authenticatedUserID == 0 {
			return utils.UnauthorizedResponse(c, "User not found in context")
		}

		// Get user ID from URL parameter
		paramUserID, err := c.ParamsInt(userIDParam)
		if err != nil {
			return utils.BadRequestResponse(c, "Invalid user ID parameter", nil)
		}

		// Check if the authenticated user matches the requested user ID
		if authenticatedUserID != uint(paramUserID) {
			return utils.ForbiddenResponse(c, "You can only access your own resources")
		}

		return c.Next()
	}
}

// WithUserContext adds user context to the standard Go context
// This is useful for passing user information to services and repositories
func WithUserContext(ctx context.Context, user repositories.UserEntity) context.Context {
	return context.WithValue(ctx, UserContextKey, user)
}

// GetUserFromGoContext retrieves the user from a standard Go context
func GetUserFromGoContext(ctx context.Context) repositories.UserEntity {
	user := ctx.Value(UserContextKey)
	if user == nil {
		return nil
	}

	userEntity, ok := user.(repositories.UserEntity)
	if !ok {
		return nil
	}

	return userEntity
}
