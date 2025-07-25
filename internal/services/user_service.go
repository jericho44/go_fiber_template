package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"
)

// UserService defines the interface for user management operations
type UserService interface {
	// GetUserProfile retrieves a user's profile by ID
	GetUserProfile(ctx context.Context, userID uint) (*models.User, error)

	// UpdateUserProfile updates a user's profile information
	UpdateUserProfile(ctx context.Context, userID uint, req UpdateUserProfileRequest) (*models.User, error)

	// ListUsers retrieves users with pagination
	ListUsers(ctx context.Context, params ListUsersParams) (*ListUsersResponse, error)

	// DeleteUser soft deletes a user
	DeleteUser(ctx context.Context, userID uint) error
}

// UpdateUserProfileRequest represents the data for updating user profile
type UpdateUserProfileRequest struct {
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=1,max=100"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// ListUsersParams represents parameters for listing users
type ListUsersParams struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	Search    string `json:"search,omitempty"`
	IsActive  *bool  `json:"is_active,omitempty"`
	SortBy    string `json:"sort_by,omitempty"`
	SortOrder string `json:"sort_order,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users []models.User `json:"users"`
	Total int64         `json:"total"`
}

// userServiceImpl is the concrete implementation of UserService
type userServiceImpl struct {
	db        *gorm.DB
	userRepo  repositories.UserRepository
	txManager repositories.TransactionManager
	validator *utils.ValidationErrorFormatter
}

// NewUserService creates a new instance of UserService
func NewUserService(db *gorm.DB, userRepo repositories.UserRepository, txManager repositories.TransactionManager) UserService {
	return &userServiceImpl{
		db:        db,
		userRepo:  userRepo,
		txManager: txManager,
		validator: utils.NewValidationErrorFormatter(),
	}
}

// GetUserProfile retrieves a user's profile by ID
func (s *userServiceImpl) GetUserProfile(ctx context.Context, userID uint) (*models.User, error) {
	// Get user by ID
	userEntity, err := s.userRepo.GetByIDWithContext(ctx, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, utils.NewNotFoundError("User not found")
		}
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Convert UserEntity to models.User
	user := &models.User{
		ID:        userEntity.GetID(),
		Email:     userEntity.GetEmail(),
		FirstName: userEntity.GetFirstName(),
		LastName:  userEntity.GetLastName(),
		IsActive:  userEntity.GetIsActive(),
	}

	return user, nil
}

// UpdateUserProfile updates a user's profile information
func (s *userServiceImpl) UpdateUserProfile(ctx context.Context, userID uint, req UpdateUserProfileRequest) (*models.User, error) {
	// Validate input
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	var updatedUser *models.User

	// Use transaction for profile update
	err := s.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Get existing user
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.NewNotFoundError("User not found")
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Update fields if provided
		if req.FirstName != nil {
			user.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			user.LastName = *req.LastName
		}
		if req.IsActive != nil {
			user.IsActive = *req.IsActive
		}

		// Save updated user
		if err := tx.Save(&user).Error; err != nil {
			return fmt.Errorf("failed to update user profile: %w", err)
		}

		updatedUser = &user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

// ListUsers retrieves users with pagination
func (s *userServiceImpl) ListUsers(ctx context.Context, params ListUsersParams) (*ListUsersResponse, error) {
	var users []models.User
	var total int64

	// Use transaction for consistent read
	err := s.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		query := tx.Model(&models.User{})

		// Apply filters
		if params.IsActive != nil {
			query = query.Where("is_active = ?", *params.IsActive)
		}

		// Apply search filter
		if params.Search != "" {
			searchTerm := "%" + strings.ToLower(params.Search) + "%"
			query = query.Where(
				"LOWER(first_name) LIKE ? OR LOWER(last_name) LIKE ? OR LOWER(email) LIKE ?",
				searchTerm, searchTerm, searchTerm,
			)
		}

		// Get total count
		if err := query.Count(&total).Error; err != nil {
			return fmt.Errorf("failed to count users: %w", err)
		}

		// Apply sorting
		orderBy := "created_at DESC" // default sorting
		if params.SortBy != "" {
			validSortFields := map[string]bool{
				"id":         true,
				"email":      true,
				"first_name": true,
				"last_name":  true,
				"is_active":  true,
				"created_at": true,
				"updated_at": true,
			}

			if validSortFields[params.SortBy] {
				sortOrder := "ASC"
				if params.SortOrder == "desc" || params.SortOrder == "DESC" {
					sortOrder = "DESC"
				}
				orderBy = fmt.Sprintf("%s %s", params.SortBy, sortOrder)
			}
		}

		// Apply pagination and get results
		offset := (params.Page - 1) * params.Limit
		if err := query.Order(orderBy).Limit(params.Limit).Offset(offset).Find(&users).Error; err != nil {
			return fmt.Errorf("failed to list users: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &ListUsersResponse{
		Users: users,
		Total: total,
	}, nil
}

// DeleteUser soft deletes a user
func (s *userServiceImpl) DeleteUser(ctx context.Context, userID uint) error {
	// Use transaction for user deletion
	return s.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Check if user exists
		var user models.User
		if err := tx.First(&user, userID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.NewNotFoundError("User not found")
			}
			return fmt.Errorf("failed to get user: %w", err)
		}

		// Soft delete the user
		if err := tx.Delete(&user).Error; err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})
}
