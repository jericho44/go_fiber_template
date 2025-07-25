package docs

import (
	"go-fiber-template/internal/controllers"
	"go-fiber-template/internal/models"
	"go-fiber-template/internal/utils"
)

// Swagger model definitions for request/response structures

// User represents the user model for Swagger documentation
// @Description User account information
type User struct {
	ID        uint   `json:"id" example:"1"`                            // User ID
	Email     string `json:"email" example:"john.doe@example.com"`      // User email address
	FirstName string `json:"first_name" example:"John"`                 // User first name
	LastName  string `json:"last_name" example:"Doe"`                   // User last name
	IsActive  bool   `json:"is_active" example:"true"`                  // User active status
	CreatedAt string `json:"created_at" example:"2023-01-01T00:00:00Z"` // Account creation timestamp
	UpdatedAt string `json:"updated_at" example:"2023-01-01T00:00:00Z"` // Last update timestamp
}

// RegisterRequest represents the registration request body
// @Description User registration request
type RegisterRequest struct {
	Email     string `json:"email" example:"john.doe@example.com" validate:"required,email"` // User email address
	Password  string `json:"password" example:"securepassword123" validate:"required,min=8"` // User password (minimum 8 characters)
	FirstName string `json:"first_name" example:"John" validate:"required,min=1,max=100"`    // User first name
	LastName  string `json:"last_name" example:"Doe" validate:"required,min=1,max=100"`      // User last name
}

// LoginRequest represents the login request body
// @Description User login request
type LoginRequest struct {
	Email    string `json:"email" example:"john.doe@example.com" validate:"required,email"` // User email address
	Password string `json:"password" example:"securepassword123" validate:"required"`       // User password
}

// RefreshTokenRequest represents the refresh token request body
// @Description Token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." validate:"required"` // JWT refresh token
}

// LogoutRequest represents the logout request body
// @Description User logout request
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // Optional JWT refresh token
}

// AuthResponse represents the authentication response
// @Description Authentication response with user data and tokens
type AuthResponse struct {
	User         User   `json:"user"`                                                                                                                                                      // User information
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsImV4cCI6MTY3MjUzMjEwMH0.example_token"` // JWT access token (expires in 15 minutes)
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwidHlwZSI6InJlZnJlc2giLCJleHAiOjE2NzMxMzY5MDB9.example_refresh_token"`           // JWT refresh token (expires in 7 days)
	ExpiresAt    int64  `json:"expires_at" example:"1672532100"`                                                                                                                           // Token expiration timestamp (Unix timestamp)
	TokenType    string `json:"token_type" example:"Bearer"`                                                                                                                               // Token type (always "Bearer")
}

// UpdateUserProfileRequest represents the update profile request body
// @Description User profile update request
type UpdateUserProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" example:"John" validate:"omitempty,min=1,max=100"` // User first name (optional)
	LastName  *string `json:"last_name,omitempty" example:"Doe" validate:"omitempty,min=1,max=100"`   // User last name (optional)
	IsActive  *bool   `json:"is_active,omitempty" example:"true"`                                     // User active status (optional)
}

// TokenRefreshResponse represents the token refresh response
// @Description Token refresh response
type TokenRefreshResponse struct {
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsImV4cCI6MTY3MjUzMzAwMH0.new_example_token"` // New JWT access token
	ExpiresAt   int64  `json:"expires_at" example:"1672533000"`                                                                                                                               // Token expiration timestamp
	TokenType   string `json:"token_type" example:"Bearer"`                                                                                                                                   // Token type
}

// APIResponse represents the standard API response structure
// @Description Standard API response wrapper
type APIResponse struct {
	Success bool        `json:"success" example:"true"`                 // Request success status
	Message string      `json:"message" example:"Operation successful"` // Response message
	Data    interface{} `json:"data,omitempty"`                         // Response data (varies by endpoint)
	Error   *APIError   `json:"error,omitempty"`                        // Error information (only present on failure)
	Meta    *Meta       `json:"meta,omitempty"`                         // Pagination metadata (for paginated responses)
}

// APIError represents error information in API responses
// @Description API error details
type APIError struct {
	Code    string            `json:"code" example:"VALIDATION_ERROR"`                                                                     // Error code
	Message string            `json:"message" example:"Validation failed"`                                                                 // Error message
	Details map[string]string `json:"details,omitempty" example:"email:Email is required,password:Password must be at least 8 characters"` // Field-specific error details
}

// Meta represents pagination and additional metadata
// @Description Pagination metadata
type Meta struct {
	Page       int   `json:"page,omitempty" example:"1"`         // Current page number
	Limit      int   `json:"limit,omitempty" example:"10"`       // Items per page
	Total      int64 `json:"total,omitempty" example:"100"`      // Total number of items
	TotalPages int   `json:"total_pages,omitempty" example:"10"` // Total number of pages
}

// Success response examples

// SuccessResponse represents a successful API response
// @Description Successful API response
type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Operation successful"`
	Data    interface{} `json:"data,omitempty"`
}

// CreatedResponse represents a successful creation response
// @Description Successful resource creation response
type CreatedResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Resource created successfully"`
	Data    interface{} `json:"data,omitempty"`
}

// PaginatedResponse represents a paginated response
// @Description Paginated API response
type PaginatedResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"Data retrieved successfully"`
	Data    interface{} `json:"data,omitempty"`
	Meta    Meta        `json:"meta"`
}

// Error response examples

// BadRequestResponse represents a 400 error response
// @Description Bad request error response
type BadRequestResponse struct {
	Success bool     `json:"success" example:"false"`
	Message string   `json:"message" example:"Request failed"`
	Error   APIError `json:"error"`
}

// UnauthorizedResponse represents a 401 error response
// @Description Unauthorized error response
type UnauthorizedResponse struct {
	Success bool     `json:"success" example:"false"`
	Message string   `json:"message" example:"Request failed"`
	Error   APIError `json:"error"`
}

// NotFoundResponse represents a 404 error response
// @Description Not found error response
type NotFoundResponse struct {
	Success bool     `json:"success" example:"false"`
	Message string   `json:"message" example:"Request failed"`
	Error   APIError `json:"error"`
}

// ConflictResponse represents a 409 error response
// @Description Conflict error response
type ConflictResponse struct {
	Success bool     `json:"success" example:"false"`
	Message string   `json:"message" example:"Request failed"`
	Error   APIError `json:"error"`
}

// InternalServerErrorResponse represents a 500 error response
// @Description Internal server error response
type InternalServerErrorResponse struct {
	Success bool     `json:"success" example:"false"`
	Message string   `json:"message" example:"Request failed"`
	Error   APIError `json:"error"`
}

// Ensure the types are used to avoid unused import errors
var (
	_ = controllers.RegisterRequest{}
	_ = controllers.LoginRequest{}
	_ = controllers.RefreshTokenRequest{}
	_ = controllers.LogoutRequest{}
	_ = controllers.AuthResponse{}
	_ = controllers.UpdateUserProfileRequest{}
	_ = models.User{}
	_ = utils.APIResponse{}
	_ = utils.APIError{}
	_ = utils.Meta{}
)
