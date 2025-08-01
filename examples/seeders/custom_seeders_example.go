package seeders

import (
	"context"
	"fmt"

	"go-fiber-template/internal/models"
	"go-fiber-template/internal/seeders"

	"gorm.io/gorm"
)

// ExampleProductSeeder demonstrates how to create a custom seeder
// This is an example seeder that would seed a hypothetical products table
type ExampleProductSeeder struct {
	*seeders.BaseSeeder
}

// NewExampleProductSeeder creates a new example product seeder
func NewExampleProductSeeder() *ExampleProductSeeder {
	return &ExampleProductSeeder{
		BaseSeeder: seeders.NewBaseSeeder(
			"example_product_seeder",
			"Example seeder that demonstrates how to create custom seeders for a products table",
			// Optional: Add dependencies if this seeder depends on other seeders
			// "user_seeder", "category_seeder",
		),
	}
}

// Seed implements the Seeder interface
func (s *ExampleProductSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	// Example products data
	// Note: This assumes you have a Product model - this is just for demonstration
	products := []struct {
		Name        string
		Description string
		Price       float64
		SKU         string
		IsActive    bool
	}{
		{
			Name:        "Wireless Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			Price:       199.99,
			SKU:         "WH001",
			IsActive:    true,
		},
		{
			Name:        "Bluetooth Speaker",
			Description: "Portable Bluetooth speaker with excellent sound quality",
			Price:       89.99,
			SKU:         "BS001",
			IsActive:    true,
		},
		{
			Name:        "USB-C Cable",
			Description: "Fast charging USB-C cable - 6 feet long",
			Price:       19.99,
			SKU:         "UC001",
			IsActive:    true,
		},
		{
			Name:        "Laptop Stand",
			Description: "Adjustable aluminum laptop stand for better ergonomics",
			Price:       49.99,
			SKU:         "LS001",
			IsActive:    false, // Example of inactive product
		},
	}

	// Note: Since we don't have a Product model in this template,
	// this is just demonstration code. In a real implementation,
	// you would replace this with your actual model.

	for _, productData := range products {
		// Check if product already exists by SKU
		// var existingProduct models.Product
		// err := db.Where("sku = ?", productData.SKU).First(&existingProduct).Error

		// if err == nil {
		// 	// Product already exists, skip
		// 	continue
		// }

		// if err != gorm.ErrRecordNotFound {
		// 	return fmt.Errorf("failed to check existing product %s: %w", productData.SKU, err)
		// }

		// Create product
		// product := models.Product{
		// 	Name:        productData.Name,
		// 	Description: productData.Description,
		// 	Price:       productData.Price,
		// 	SKU:         productData.SKU,
		// 	IsActive:    productData.IsActive,
		// }

		// if err := db.Create(&product).Error; err != nil {
		// 	return fmt.Errorf("failed to create product %s: %w", productData.SKU, err)
		// }

		// For demonstration purposes, we'll just log what would be created
		fmt.Printf("Would create product: %s (SKU: %s) - $%.2f\n",
			productData.Name, productData.SKU, productData.Price)
	}

	return nil
}

// Rollback implements the Seeder interface
func (s *ExampleProductSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
	// Define SKUs of seeded products
	seededSKUs := []string{
		"WH001",
		"BS001",
		"UC001",
		"LS001",
	}

	// Delete seeded products
	// Note: This would work if you have a Product model
	// result := db.Where("sku IN ?", seededSKUs).Delete(&models.Product{})
	// if result.Error != nil {
	// 	return fmt.Errorf("failed to rollback product seeder: %w", result.Error)
	// }

	// For demonstration purposes, we'll just log what would be deleted
	fmt.Printf("Would delete products with SKUs: %v\n", seededSKUs)

	return nil
}

// ShouldRun implements custom logic to check if this seeder should run
func (s *ExampleProductSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
	// Check if any of the seeded products already exist
	// var count int64
	// err := db.Model(&models.Product{}).Where("sku IN ?", []string{
	// 	"WH001", "BS001", "UC001", "LS001",
	// }).Count(&count).Error

	// if err != nil {
	// 	return false, fmt.Errorf("failed to check existing products: %w", err)
	// }

	// // If any products exist, don't run the seeder
	// return count == 0, nil

	// For demonstration purposes, always return true
	// In a real implementation, you'd check your actual data
	return true, nil
}

// Example of a more complex seeder with relationships
type ExampleOrderSeeder struct {
	*seeders.BaseSeeder
}

func NewExampleOrderSeeder() *ExampleOrderSeeder {
	return &ExampleOrderSeeder{
		BaseSeeder: seeders.NewBaseSeeder(
			"example_order_seeder",
			"Example seeder that creates sample orders with user and product relationships",
			"user_seeder", "example_product_seeder", // Dependencies
		),
	}
}

func (s *ExampleOrderSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	// Get existing users (seeded by user_seeder)
	var users []models.User
	if err := db.Limit(3).Find(&users).Error; err != nil {
		return fmt.Errorf("failed to get users for order seeding: %w", err)
	}

	if len(users) == 0 {
		return fmt.Errorf("no users found - make sure user_seeder has run first")
	}

	// Create sample orders for each user
	for i, user := range users {
		// In a real implementation, you would:
		// 1. Create an Order record
		// 2. Create OrderItem records linking to products
		// 3. Calculate totals, taxes, etc.

		fmt.Printf("Would create order #%d for user %s (%s)\n",
			i+1, user.Email, user.FirstName+" "+user.LastName)
	}

	return nil
}

func (s *ExampleOrderSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
	// In a real implementation, you would delete the seeded orders
	fmt.Println("Would delete all seeded orders and order items")
	return nil
}

// How to register these example seeders:
//
// In internal/seeders/registry.go, add to registerAllSeeders():
//
// func (r *Registry) registerAllSeeders() {
//     // Existing seeders
//     r.manager.RegisterSeeder(NewUserSeeder())
//
//     // Example seeders (only in development)
//     if r.config.IsDevelopment() {
//         r.manager.RegisterSeeder(NewDemoDataSeeder())
//         r.manager.RegisterSeeder(NewExampleProductSeeder())
//         r.manager.RegisterSeeder(NewExampleOrderSeeder())
//     }
// }
