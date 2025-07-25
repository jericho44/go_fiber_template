package seeders

import (
	"context"
	"fmt"

	"go-fiber-template/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserSeeder seeds the users table with sample data
type UserSeeder struct {
	*BaseSeeder
}

// NewUserSeeder creates a new UserSeeder
func NewUserSeeder() *UserSeeder {
	return &UserSeeder{
		BaseSeeder: NewBaseSeeder(
			"user_seeder",
			"Seeds the users table with sample admin and regular users",
		),
	}
}

// Seed implements the Seeder interface
func (s *UserSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	// Sample users to create
	users := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
		IsActive  bool
	}{
		{
			Email:     "admin@example.com",
			Password:  "admin123456",
			FirstName: "Admin",
			LastName:  "User",
			IsActive:  true,
		},
		{
			Email:     "john.doe@example.com",
			Password:  "password123",
			FirstName: "John",
			LastName:  "Doe",
			IsActive:  true,
		},
		{
			Email:     "jane.smith@example.com",
			Password:  "password123",
			FirstName: "Jane",
			LastName:  "Smith",
			IsActive:  true,
		},
		{
			Email:     "bob.wilson@example.com",
			Password:  "password123",
			FirstName: "Bob",
			LastName:  "Wilson",
			IsActive:  false,
		},
		{
			Email:     "alice.johnson@example.com",
			Password:  "password123",
			FirstName: "Alice",
			LastName:  "Johnson",
			IsActive:  true,
		},
	}

	for _, userData := range users {
		// Check if user already exists
		var existingUser models.User
		err := db.Where("email = ?", userData.Email).First(&existingUser).Error
		
		if err == nil {
			// User already exists, skip
			continue
		}
		
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing user %s: %w", userData.Email, err)
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for user %s: %w", userData.Email, err)
		}

		// Create user
		user := models.User{
			Email:     userData.Email,
			Password:  string(hashedPassword),
			FirstName: userData.FirstName,
			LastName:  userData.LastName,
			IsActive:  userData.IsActive,
		}

		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user %s: %w", userData.Email, err)
		}
	}

	return nil
}

// Rollback implements the Seeder interface
func (s *UserSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
	// Define emails of seeded users
	seededEmails := []string{
		"admin@example.com",
		"john.doe@example.com",
		"jane.smith@example.com",
		"bob.wilson@example.com",
		"alice.johnson@example.com",
	}

	// Delete seeded users
	result := db.Where("email IN ?", seededEmails).Delete(&models.User{})
	if result.Error != nil {
		return fmt.Errorf("failed to rollback user seeder: %w", result.Error)
	}

	return nil
}

// ShouldRun implements custom logic to check if this seeder should run
func (s *UserSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
	// Check if any of the seeded users already exist
	var count int64
	err := db.Model(&models.User{}).Where("email IN ?", []string{
		"admin@example.com",
		"john.doe@example.com",
		"jane.smith@example.com",
		"bob.wilson@example.com",
		"alice.johnson@example.com",
	}).Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check existing users: %w", err)
	}

	// If any users exist, don't run the seeder
	return count == 0, nil
}