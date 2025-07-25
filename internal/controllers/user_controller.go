package controllers

import (
	"context"
	"strconv"

	_ "go-fiber-template/internal/models"
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/gofiber/fiber/v2"
)

// UserController handles user management HTTP requests
type UserController struct {
	userService services.UserService
	validator   *utils.ValidationErrorFormatter
}

// NewUserController creates a new user management controller
func NewUserController(userService services.UserService) *UserController {
	return &UserController{
		userService: userService,
		validator:   utils.NewValidationErrorFormatter(),
	}
}

// UpdateUserProfileRequest represents the request body for updating user profile
type UpdateUserProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// ListUsersQueryParams represents query parameters for listing users
type ListUsersQueryParams struct {
	Page      int    `query:"page"`
	Limit     int    `query:"limit"`
	Search    string `query:"search"`
	IsActive  *bool  `query:"is_active"`
	SortBy    string `query:"sort_by"`
	SortOrder string `query:"sort_order"`
}

// GetUserProfile retrieves the current user's profile or a specific user's profile
// @Summary Get user profile
// @Description Get user profile information by ID. Requires valid JWT token for authentication.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Success 200 {object} utils.APIResponse{data=models.User} "User profile retrieved successfully"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid user ID"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Unauthorized - invalid or missing JWT token"
// @Failure 404 {object} utils.APIResponse{error=utils.APIError} "User not found"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /users/{id} [get]
// @Security BearerAuth
func (uc *UserController) GetUserProfile(c *fiber.Ctx) error {
	// Get user ID from path parameter
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", nil)
	}

	// Get user profile
	user, err := uc.userService.GetUserProfile(context.Background(), uint(userID))
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case "NOT_FOUND":
				return utils.NotFoundResponse(c, "User not found")
			default:
				return utils.InternalServerErrorResponse(c, "Failed to get user profile")
			}
		}
		return utils.InternalServerErrorResponse(c, "Failed to get user profile")
	}

	return utils.SuccessResponse(c, "User profile retrieved successfully", user)
}

// UpdateUserProfile updates a user's profile information
// @Summary Update user profile
// @Description Update user profile information. All fields are optional. Requires valid JWT token for authentication.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Param request body UpdateUserProfileRequest true "Update profile request with optional fields"
// @Success 200 {object} utils.APIResponse{data=models.User} "User profile updated successfully"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid user ID, request body, or validation errors"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Unauthorized - invalid or missing JWT token"
// @Failure 404 {object} utils.APIResponse{error=utils.APIError} "User not found"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /users/{id} [put]
// @Security BearerAuth
func (uc *UserController) UpdateUserProfile(c *fiber.Ctx) error {
	// Get user ID from path parameter
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", nil)
	}

	// Parse request body
	var req UpdateUserProfileRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(c, "Invalid request body", nil)
	}

	// Validate request
	if validationErr := uc.validator.ValidateStruct(req); validationErr != nil {
		return utils.ValidationErrorResponse(c, validationErr.Details)
	}

	// Convert to service request
	serviceReq := services.UpdateUserProfileRequest{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  req.IsActive,
	}

	// Update user profile
	user, err := uc.userService.UpdateUserProfile(context.Background(), uint(userID), serviceReq)
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case "VALIDATION_ERROR":
				return utils.ValidationErrorResponse(c, appErr.Details)
			case "NOT_FOUND":
				return utils.NotFoundResponse(c, appErr.Message)
			default:
				return utils.InternalServerErrorResponse(c, "Failed to update user profile")
			}
		}
		return utils.InternalServerErrorResponse(c, "Failed to update user profile")
	}

	return utils.SuccessResponse(c, "User profile updated successfully", user)
}

// ListUsers retrieves users with pagination and filtering
// @Summary List users
// @Description Get a paginated list of users with optional filtering and sorting. Requires valid JWT token for authentication.
// @Tags Users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)" minimum(1)
// @Param limit query int false "Items per page (default: 10, max: 100)" minimum(1) maximum(100)
// @Param search query string false "Search term for name or email"
// @Param is_active query bool false "Filter by active status"
// @Param sort_by query string false "Sort field" Enums(id, email, first_name, last_name, is_active, created_at, updated_at)
// @Param sort_order query string false "Sort order" Enums(asc, desc)
// @Success 200 {object} utils.APIResponse{data=[]models.User,meta=utils.Meta} "Users retrieved successfully with pagination metadata"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid query parameters"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Unauthorized - invalid or missing JWT token"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /users [get]
// @Security BearerAuth
func (uc *UserController) ListUsers(c *fiber.Ctx) error {
	// Get pagination parameters
	paginationParams := utils.GetPaginationParams(c)

	// Parse additional query parameters
	var queryParams ListUsersQueryParams
	if err := c.QueryParser(&queryParams); err != nil {
		return utils.BadRequestResponse(c, "Invalid query parameters", nil)
	}

	// Use pagination params if not provided in query
	if queryParams.Page == 0 {
		queryParams.Page = paginationParams.Page
	}
	if queryParams.Limit == 0 {
		queryParams.Limit = paginationParams.Limit
	}

	// Convert to service params
	serviceParams := services.ListUsersParams{
		Page:      queryParams.Page,
		Limit:     queryParams.Limit,
		Search:    queryParams.Search,
		IsActive:  queryParams.IsActive,
		SortBy:    queryParams.SortBy,
		SortOrder: queryParams.SortOrder,
	}

	// Get users list
	response, err := uc.userService.ListUsers(context.Background(), serviceParams)
	if err != nil {
		return utils.InternalServerErrorResponse(c, "Failed to list users")
	}

	// Return paginated response
	return utils.PaginatedResponse(c, "Users retrieved successfully", response.Users, queryParams.Page, queryParams.Limit, response.Total)
}

// DeleteUser soft deletes a user
// @Summary Delete user
// @Description Soft delete a user by ID. This performs a soft delete, marking the user as deleted without removing the record. Requires valid JWT token for authentication.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "User ID" minimum(1)
// @Success 200 {object} utils.APIResponse "User deleted successfully"
// @Failure 400 {object} utils.APIResponse{error=utils.APIError} "Invalid user ID"
// @Failure 401 {object} utils.APIResponse{error=utils.APIError} "Unauthorized - invalid or missing JWT token"
// @Failure 404 {object} utils.APIResponse{error=utils.APIError} "User not found"
// @Failure 500 {object} utils.APIResponse{error=utils.APIError} "Internal server error"
// @Router /users/{id} [delete]
// @Security BearerAuth
func (uc *UserController) DeleteUser(c *fiber.Ctx) error {
	// Get user ID from path parameter
	userIDStr := c.Params("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return utils.BadRequestResponse(c, "Invalid user ID", nil)
	}

	// Delete user
	err = uc.userService.DeleteUser(context.Background(), uint(userID))
	if err != nil {
		if appErr, ok := err.(*utils.AppError); ok {
			switch appErr.Code {
			case "NOT_FOUND":
				return utils.NotFoundResponse(c, appErr.Message)
			default:
				return utils.InternalServerErrorResponse(c, "Failed to delete user")
			}
		}
		return utils.InternalServerErrorResponse(c, "Failed to delete user")
	}

	return utils.SuccessResponse(c, "User deleted successfully", nil)
}
