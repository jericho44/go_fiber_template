package controllers

import (
	"context"
	"strings"

	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	authService services.AuthService
	jwtService  services.JWTServiceInterface
	validator   *utils.ValidationErrorFormatter
}

// NewAuthController creates a new authentication controller
func NewAuthController(authService services.AuthService, jwtService services.JWTServiceInterface) *AuthController {
	return &AuthController{
		authService: authService,
		jwtService:  jwtService,
		validator:   utils.NewValidationErrorFormatter(),
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshTokenRequest represents the request body for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	User         interface{} `json:"user"`
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresAt    int64       `json:"expires_at"`
	TokenType    string      `json:"token_type"`
}

// Register handles user registration
// @Summary Register a new user
// @Description Create a new user account with email and password. Password must be at least 8 characters long.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RegisterRequest true "Registration request with user details"
// @Success 201 {object} utils.APIResponse{data=AuthResponse} "User registered successfully with authentication tokens"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 409 {object} utils.APIResponse{error=utils.APIError} "Email already exists"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/register [post]
func (ac *AuthController) Register(c *fiber.Ctx) error {
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
				return utils.ErrorResponse(c, 409, "CONFLICT", appErr.Message, nil)
			default:
				return utils.InternalServerErrorResponse(c, "Registration failed")
			}
		}
		return utils.InternalServerErrorResponse(c, "Registration failed")
	}

	// Generate JWT tokens
	tokenPair, err := ac.jwtService.GenerateTokenPair(authResponse.User.ID, authResponse.User.Email)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to generate tokens")
	}

	// Prepare response
	response := AuthResponse{
		User:         authResponse.User,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	}

	return utils.CreatedResponse(c, "User registered successfully", response)
}

// Login handles user authentication
// @Summary Login user
// @Description Authenticate user with email and password. Returns JWT access and refresh tokens on successful authentication.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} utils.APIResponse{data=AuthResponse} "Login successful with authentication tokens"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid credentials"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/login [post]
func (ac *AuthController) Login(c *fiber.Ctx) error {
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

	// Generate JWT tokens
	tokenPair, err := ac.jwtService.GenerateTokenPair(authResponse.User.ID, authResponse.User.Email)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to generate tokens")
	}

	// Prepare response
	response := AuthResponse{
		User:         authResponse.User,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    tokenPair.ExpiresAt,
		TokenType:    tokenPair.TokenType,
	}

	return utils.SuccessResponse(c, "Login successful", response)
}

// Logout handles user logout and token invalidation
// @Summary Logout user
// @Description Invalidate user tokens and logout. Requires valid JWT token in Authorization header. Optionally accepts refresh token in request body for complete logout.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer JWT access token" Format(Bearer {token})
// @Param request body LogoutRequest false "Optional logout request with refresh token"
// @Success 200 {object} utils.APIResponse "Logout successful"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or missing authorization header"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/logout [post]
// @Security BearerAuth
func (ac *AuthController) Logout(c *fiber.Ctx) error {
	// Get access token from Authorization header
	authHeader := c.Get("Authorization")
	accessToken, err := ac.jwtService.ExtractTokenFromHeader(authHeader)
	if err != nil {
		return utils.UnauthorizedResponse(c, "Invalid authorization header")
	}

	// Parse request body for refresh token (optional)
	var req LogoutRequest
	if err := c.BodyParser(&req); err != nil {
		// If body parsing fails, continue with just access token logout
		req.RefreshToken = ""
	}

	// Logout user (blacklist tokens)
	err = ac.jwtService.LogoutUser(accessToken, req.RefreshToken)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to logout user")
	}

	return utils.SuccessResponse(c, "Logout successful", nil)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate a new access token using a valid refresh token. The refresh token must not be expired or blacklisted.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "Refresh token request"
// @Success 200 {object} utils.APIResponse{data=map[string]interface{}} "Token refreshed successfully with new access token"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid request body or validation errors"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Invalid or expired refresh token"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /auth/refresh [post]
func (ac *AuthController) RefreshToken(c *fiber.Ctx) error {
	var req RefreshTokenRequest

	// Parse request body
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request body", nil)
	}

	// Validate request
	if validationErr := ac.validator.ValidateStruct(req); validationErr != nil {
		return utils.ValidationErrorResponse(c, validationErr.Details)
	}

	// Refresh access token
	newAccessToken, err := ac.jwtService.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		if strings.Contains(err.Error(), "invalid refresh token") ||
			strings.Contains(err.Error(), "token has been revoked") ||
			strings.Contains(err.Error(), "token has expired") {
			return utils.UnauthorizedResponse(c, "Invalid or expired refresh token")
		}
		return utils.InternalServerErrorResponse(c, "Failed to refresh token")
	}

	// Get token expiry for response
	tokenExpiry, err := ac.jwtService.GetTokenExpiry(newAccessToken)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to get token expiry")
	}

	// Prepare response
	response := map[string]interface{}{
		"access_token": newAccessToken,
		"expires_at":   tokenExpiry.Unix(),
		"token_type":   "Bearer",
	}

	return utils.SuccessResponse(c, "Token refreshed successfully", response)
}
