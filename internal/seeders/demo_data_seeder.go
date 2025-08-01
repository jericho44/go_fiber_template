package seeders

import (
	"context"
	"fmt"

	"go-fiber-template/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// DemoDataSeeder seeds the database with demo data for testing and development
type DemoDataSeeder struct {
	*BaseSeeder
}

// NewDemoDataSeeder creates a new DemoDataSeeder
func NewDemoDataSeeder() *DemoDataSeeder {
	return &DemoDataSeeder{
		BaseSeeder: NewBaseSeeder(
			"demo_data_seeder",
			"Seeds the database with comprehensive demo data for testing and development",
			"user_seeder", // Depends on user_seeder
		),
	}
}

// Seed implements the Seeder interface
func (s *DemoDataSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	// Additional demo users
	demoUsers := []struct {
		Email     string
		Password  string
		FirstName string
		LastName  string
		IsActive  bool
	}{
		{
			Email:     "demo.user1@example.com",
			Password:  "demo123456",
			FirstName: "Demo",
			LastName:  "User One",
			IsActive:  true,
		},
		{
			Email:     "demo.user2@example.com",
			Password:  "demo123456",
			FirstName: "Demo",
			LastName:  "User Two",
			IsActive:  true,
		},
		{
			Email:     "demo.user3@example.com",
			Password:  "demo123456",
			FirstName: "Demo",
			LastName:  "User Three",
			IsActive:  false,
		},
		{
			Email:     "test.manager@example.com",
			Password:  "manager123",
			FirstName: "Test",
			LastName:  "Manager",
			IsActive:  true,
		},
		{
			Email:     "support@example.com",
			Password:  "support123",
			FirstName: "Support",
			LastName:  "Team",
			IsActive:  true,
		},
	}

	for _, userData := range demoUsers {
		// Check if user already exists
		var existingUser models.User
		err := db.Where("email = ?", userData.Email).First(&existingUser).Error

		if err == nil {
			// User already exists, skip
			continue
		}

		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to check existing demo user %s: %w", userData.Email, err)
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userData.Password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash password for demo user %s: %w", userData.Email, err)
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
			return fmt.Errorf("failed to create demo user %s: %w", userData.Email, err)
		}
	}

	return nil
}

// Rollback implements the Seeder interface
func (s *DemoDataSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
	// Define emails of demo users
	demoEmails := []string{
		"demo.user1@example.com",
		"demo.user2@example.com",
		"demo.user3@example.com",
		"test.manager@example.com",
		"support@example.com",
	}

	// Delete demo users
	result := db.Where("email IN ?", demoEmails).Delete(&models.User{})
	if result.Error != nil {
		return fmt.Errorf("failed to rollback demo data seeder: %w", result.Error)
	}

	return nil
}

// ShouldRun implements custom logic to check if this seeder should run
func (s *DemoDataSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
	// Check if any of the demo users already exist
	var count int64
	err := db.Model(&models.User{}).Where("email IN ?", []string{
		"demo.user1@example.com",
		"demo.user2@example.com",
		"demo.user3@example.com",
		"test.manager@example.com",
		"support@example.com",
	}).Count(&count).Error

	if err != nil {
		return false, fmt.Errorf("failed to check existing demo users: %w", err)
	}

	// If any demo users exist, don't run the seeder
	return count == 0, nil
}
