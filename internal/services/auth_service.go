package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/repositories"
	"go-fiber-template/internal/utils"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	// Register creates a new user account with password hashing
	Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error)

	// Login authenticates a user with email and password
	Login(ctx context.Context, req LoginRequest) (*AuthResponse, error)

	// ValidatePassword checks if a password meets security requirements
	ValidatePassword(password string) error

	// HashPassword hashes a password using bcrypt
	HashPassword(password string) (string, error)

	// ComparePassword compares a plain password with a hashed password
	ComparePassword(hashedPassword, password string) error
}

// RegisterRequest represents the data needed for user registration
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=1,max=100"`
	LastName  string `json:"last_name" validate:"required,min=1,max=100"`
}

// LoginRequest represents the data needed for user login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	User    *models.User `json:"user"`
	Message string       `json:"message"`
}

// authServiceImpl is the concrete implementation of AuthService
type authServiceImpl struct {
	db        *gorm.DB
	txManager repositories.TransactionManager
	validator *utils.ValidationErrorFormatter
}

// NewAuthService creates a new instance of AuthService
func NewAuthService(db *gorm.DB, txManager repositories.TransactionManager) AuthService {
	return &authServiceImpl{
		db:        db,
		txManager: txManager,
		validator: utils.NewValidationErrorFormatter(),
	}
}

// Register creates a new user account with password hashing using transactions
func (s *authServiceImpl) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Validate password strength
	if err := s.ValidatePassword(req.Password); err != nil {
		return nil, utils.NewValidationError(map[string]string{
			"password": err.Error(),
		})
	}

	var user *models.User
	var err error

	// Use transaction for user registration to ensure data consistency
	err = s.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Check if user already exists
		var count int64
		if err := tx.Model(&models.User{}).Where("email = ?", strings.ToLower(req.Email)).Count(&count).Error; err != nil {
			return fmt.Errorf("failed to check if user exists: %w", err)
		}

		if count > 0 {
			return utils.NewConflictError("User with this email already exists")
		}

		// Hash password
		hashedPassword, err := s.HashPassword(req.Password)
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		// Create user model
		user = &models.User{
			Email:     strings.ToLower(req.Email),
			Password:  hashedPassword,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			IsActive:  true,
		}

		// Create user in database within transaction
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Return success response (password is excluded via json:"-" tag)
	return &AuthResponse{
		User:    user,
		Message: "User registered successfully",
	}, nil
}

// Login authenticates a user with email and password
func (s *authServiceImpl) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	// Validate input
	if err := s.validator.ValidateStruct(req); err != nil {
		return nil, err
	}

	// Use transaction for login to ensure consistency
	var user *models.User
	err := s.txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Get user by email directly using GORM
		user = &models.User{}
		if err := tx.Where("email = ?", strings.ToLower(req.Email)).First(user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.NewUnauthorizedError("Invalid email or password")
			}
			return fmt.Errorf("failed to get user by email: %w", err)
		}

		// Check if user is active
		if !user.IsActive {
			return utils.NewUnauthorizedError("Account is deactivated")
		}

		// Verify password
		if err := s.ComparePassword(user.Password, req.Password); err != nil {
			return utils.NewUnauthorizedError("Invalid email or password")
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:    user,
		Message: "Login successful",
	}, nil
}

// ValidatePassword checks if a password meets security requirements
func (s *authServiceImpl) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	// Check for at least one lowercase letter
	hasLower := false
	// Check for at least one digit
	hasDigit := false
	// Check for at least one special character
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		case char >= 32 && char <= 126: // Printable ASCII characters
			if !((char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9')) {
				hasSpecial = true
			}
		}
	}

	var missingRequirements []string
	if !hasUpper {
		missingRequirements = append(missingRequirements, "uppercase letter")
	}
	if !hasLower {
		missingRequirements = append(missingRequirements, "lowercase letter")
	}
	if !hasDigit {
		missingRequirements = append(missingRequirements, "digit")
	}
	if !hasSpecial {
		missingRequirements = append(missingRequirements, "special character")
	}

	if len(missingRequirements) > 0 {
		return fmt.Errorf("password must contain at least one %s", strings.Join(missingRequirements, ", "))
	}

	return nil
}

// HashPassword hashes a password using bcrypt
func (s *authServiceImpl) HashPassword(password string) (string, error) {
	// Use bcrypt with cost 12 for good security/performance balance
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// ComparePassword compares a plain password with a hashed password
func (s *authServiceImpl) ComparePassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("password does not match")
		}
		return fmt.Errorf("failed to compare password: %w", err)
	}
	return nil
}
