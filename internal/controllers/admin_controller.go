package controllers

import (
	"go-fiber-template/internal/services"
	"go-fiber-template/internal/utils"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

// AdminController handles administrative operations
type AdminController struct {
	accountLocker services.AccountLocker
}

// NewAdminController creates a new AdminController instance
func NewAdminController(accountLocker services.AccountLocker) *AdminController {
	return &AdminController{
		accountLocker: accountLocker,
	}
}

// UnlockAccountRequest represents the request to unlock an account
type UnlockAccountRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// LockoutInfoResponse represents the response containing lockout information
type LockoutInfoResponse struct {
	Email       string                `json:"email"`
	LockoutInfo *services.LockoutInfo `json:"lockout_info"`
	Message     string                `json:"message"`
}

// UnlockAccount unlocks a user account
// @Summary Unlock user account
// @Description Unlocks a user account that has been locked due to failed login attempts (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param request body UnlockAccountRequest true "Unlock account request"
// @Success 200 {object} utils.APIResponse{data=LockoutInfoResponse} "Account unlocked successfully"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden - Admin access required"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /admin/unlock-account [post]
func (c *AdminController) UnlockAccount(ctx *fiber.Ctx) error {
	var req UnlockAccountRequest
	if err := ctx.BodyParser(&req); err != nil {
		return utils.BadRequestResponse(ctx, "Invalid request body", nil)
	}

	// Validate request
	validationFormatter := utils.NewValidationErrorFormatter()
	if err := validationFormatter.ValidateStruct(req); err != nil {
		if appErr, ok := utils.AsAppError(err); ok {
			return utils.ValidationErrorResponse(ctx, appErr.Details)
		}
		return utils.BadRequestResponse(ctx, "Validation failed", nil)
	}

	// Unlock the account
	if err := c.accountLocker.UnlockAccount(ctx.Context(), req.Email); err != nil {
		return utils.InternalServerErrorResponse(ctx, "Failed to unlock account")
	}

	// Get updated lockout info
	updatedLockoutInfo, err := c.accountLocker.GetLockoutInfo(ctx.Context(), req.Email)
	if err != nil {
		return utils.InternalServerErrorResponse(ctx, "Failed to get updated lockout information")
	}

	response := LockoutInfoResponse{
		Email:       req.Email,
		LockoutInfo: updatedLockoutInfo,
		Message:     "Account unlocked successfully",
	}

	return utils.SuccessResponse(ctx, "Account unlocked successfully", response)
}

// GetLockoutInfo retrieves lockout information for a user account
// @Summary Get account lockout information
// @Description Retrieves lockout information for a user account (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param email query string true "User email address"
// @Success 200 {object} utils.APIResponse{data=LockoutInfoResponse} "Lockout information retrieved successfully"
// @Failure 400 {object} utils.APIResponse "Invalid request"
// @Failure 401 {object} utils.APIResponse "Unauthorized"
// @Failure 403 {object} utils.APIResponse "Forbidden - Admin access required"
// @Failure 500 {object} utils.APIResponse "Internal server error"
// @Security BearerAuth
// @Router /admin/lockout-info [get]
func (c *AdminController) GetLockoutInfo(ctx *fiber.Ctx) error {
	email := ctx.Query("email")
	if email == "" {
		return utils.BadRequestResponse(ctx, "Email parameter is required", nil)
	}

	// Validate email format using go-playground/validator directly
	validate := validator.New()
	if err := validate.Var(email, "required,email"); err != nil {
		return utils.BadRequestResponse(ctx, "Invalid email format", nil)
	}

	// Get lockout info
	lockoutInfo, err := c.accountLocker.GetLockoutInfo(ctx.Context(), email)
	if err != nil {
		return utils.InternalServerErrorResponse(ctx, "Failed to get lockout information")
	}

	response := LockoutInfoResponse{
		Email:       email,
		LockoutInfo: lockoutInfo,
		Message:     "Lockout information retrieved successfully",
	}

	return utils.SuccessResponse(ctx, "Lockout information retrieved successfully", response)
}
