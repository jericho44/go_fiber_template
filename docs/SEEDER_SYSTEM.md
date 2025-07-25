# Database Seeder System

The Go Fiber Template includes a comprehensive database seeding system that allows you to populate your database with initial or test data in a controlled and reproducible manner.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [Creating Custom Seeders](#creating-custom-seeders)
- [Seeder Dependencies](#seeder-dependencies)
- [Environment-Specific Seeders](#environment-specific-seeders)
- [Best Practices](#best-practices)
- [Examples](#examples)

## Overview

The seeder system provides:

- **Dependency Management**: Seeders can declare dependencies on other seeders
- **Execution Tracking**: Automatic tracking of seeder execution history
- **Rollback Support**: Ability to rollback seeded data
- **Environment Awareness**: Different seeders for different environments
- **Idempotent Operations**: Safe to run multiple times
- **Transaction Safety**: All operations wrapped in database transactions

## Quick Start

### 1. Run All Seeders

```bash
go run cmd/seed/main.go run --all
```

### 2. Run Environment-Specific Seeders

```bash
go run cmd/seed/main.go run --env
```

### 3. Check Seeder Status

```bash
go run cmd/seed/main.go status
```

### 4. List Available Seeders

```bash
go run cmd/seed/main.go list
```

## CLI Commands

### Run Command

Run database seeders with various options:

```bash
# Run all registered seeders
go run cmd/seed/main.go run --all

# Run environment-appropriate seeders
go run cmd/seed/main.go run --env

# Run specific seeders
go run cmd/seed/main.go run user_seeder demo_data_seeder

# Run specific seeders using flag
go run cmd/seed/main.go run --seeders user_seeder,demo_data_seeder
```

### Rollback Command

Rollback seeded data:

```bash
# Rollback all seeders
go run cmd/seed/main.go rollback --all

# Rollback specific seeders
go run cmd/seed/main.go rollback user_seeder demo_data_seeder

# Rollback specific seeders using flag
go run cmd/seed/main.go rollback --seeders user_seeder,demo_data_seeder
```

### Status Command

Check the execution status of all seeders:

```bash
go run cmd/seed/main.go status
```

Output example:
```
SEEDER                   STATUS     LAST RUN             DESCRIPTION
--------------------------------------------------------------------------------
demo_data_seeder         Completed  2025-01-25 10:30:15  Seeds comprehensive demo data
user_seeder              Completed  2025-01-25 10:30:10  Seeds sample admin and regular users
```

### List Command

List all available seeders with their dependencies:

```bash
go run cmd/seed/main.go list
```

Output example:
```
SEEDER                   DEPENDENCIES    DESCRIPTION
--------------------------------------------------------------------------------
demo_data_seeder         user_seeder     Seeds comprehensive demo data
user_seeder              None            Seeds sample admin and regular users
```

## Creating Custom Seeders

### Basic Seeder Structure

Create a new seeder by implementing the `Seeder` interface or extending `BaseSeeder`:

```go
package seeders

import (
    "context"
    "go-fiber-template/internal/models"
    "gorm.io/gorm"
)

type MyCustomSeeder struct {
    *BaseSeeder
}

func NewMyCustomSeeder() *MyCustomSeeder {
    return &MyCustomSeeder{
        BaseSeeder: NewBaseSeeder(
            "my_custom_seeder",
            "Description of what this seeder does",
            // Optional dependencies
            "user_seeder",
        ),
    }
}

func (s *MyCustomSeeder) Seed(ctx context.Context, db *gorm.DB) error {
    // Your seeding logic here
    // Example: Create sample data
    data := models.MyModel{
        Name: "Sample Data",
        // ... other fields
    }
    
    return db.Create(&data).Error
}

func (s *MyCustomSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
    // Optional: Implement rollback logic
    return db.Where("name = ?", "Sample Data").Delete(&models.MyModel{}).Error
}
```

### Register Your Seeder

Add your seeder to the registry in `internal/seeders/registry.go`:

```go
func (r *Registry) registerAllSeeders() {
    // Existing seeders
    r.manager.RegisterSeeder(NewUserSeeder())
    
    // Add your custom seeder
    r.manager.RegisterSeeder(NewMyCustomSeeder())
    
    // Environment-specific seeders
    if r.config.IsDevelopment() {
        r.manager.RegisterSeeder(NewDemoDataSeeder())
    }
}
```

### Advanced Seeder with Custom Logic

```go
func (s *MyCustomSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
    // Custom logic to determine if seeder should run
    var count int64
    err := db.Model(&models.MyModel{}).Count(&count).Error
    if err != nil {
        return false, err
    }
    
    // Only run if no data exists
    return count == 0, nil
}
```

## Seeder Dependencies

Seeders can declare dependencies to ensure proper execution order:

```go
func NewOrderSeeder() *OrderSeeder {
    return &OrderSeeder{
        BaseSeeder: NewBaseSeeder(
            "order_seeder",
            "Seeds sample orders",
            "user_seeder", "product_seeder", // Dependencies
        ),
    }
}
```

The system will automatically:
- Resolve dependency order
- Execute dependencies first
- Detect circular dependencies
- Fail gracefully if dependencies are missing

## Environment-Specific Seeders

The system supports different seeders for different environments:

### Current Environment Configuration

- **All Environments**: `user_seeder`
- **Development Only**: `demo_data_seeder`

### Customizing Environment Seeders

Modify `GetSeedersByEnvironment()` in `registry.go`:

```go
func (r *Registry) GetSeedersByEnvironment() []string {
    var seeders []string
    
    // Core seeders for all environments
    seeders = append(seeders, "user_seeder")
    
    // Development-specific seeders
    if r.config.IsDevelopment() {
        seeders = append(seeders, "demo_data_seeder", "test_data_seeder")
    }
    
    // Production-specific seeders
    if r.config.IsProduction() {
        seeders = append(seeders, "production_data_seeder")
    }
    
    return seeders
}
```

## Best Practices

### 1. Idempotent Seeders

Always check if data exists before creating:

```go
func (s *MySeeder) Seed(ctx context.Context, db *gorm.DB) error {
    // Check if data already exists
    var existing models.MyModel
    err := db.Where("unique_field = ?", "value").First(&existing).Error
    
    if err == nil {
        // Data exists, skip
        return nil
    }
    
    if err != gorm.ErrRecordNotFound {
        return err
    }
    
    // Create new data
    return db.Create(&models.MyModel{...}).Error
}
```

### 2. Use Transactions

The system automatically wraps seeders in transactions, but for complex operations:

```go
func (s *MySeeder) Seed(ctx context.Context, db *gorm.DB) error {
    return db.Transaction(func(tx *gorm.DB) error {
        // Multiple operations
        if err := tx.Create(&model1).Error; err != nil {
            return err
        }
        
        if err := tx.Create(&model2).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

### 3. Implement Rollbacks

Always implement rollback methods for data cleanup:

```go
func (s *MySeeder) Rollback(ctx context.Context, db *gorm.DB) error {
    // Remove specific seeded data
    return db.Where("created_by_seeder = ?", true).Delete(&models.MyModel{}).Error
}
```

### 4. Use Meaningful Names and Descriptions

```go
func NewUserSeeder() *UserSeeder {
    return &UserSeeder{
        BaseSeeder: NewBaseSeeder(
            "user_seeder",
            "Seeds the users table with sample admin and regular users for testing and development",
        ),
    }
}
```

## Examples

### Example 1: Product Seeder

```go
type ProductSeeder struct {
    *BaseSeeder
}

func NewProductSeeder() *ProductSeeder {
    return &ProductSeeder{
        BaseSeeder: NewBaseSeeder(
            "product_seeder",
            "Seeds sample products for the e-commerce system",
        ),
    }
}

func (s *ProductSeeder) Seed(ctx context.Context, db *gorm.DB) error {
    products := []models.Product{
        {Name: "Laptop", Price: 999.99, SKU: "LAP001"},
        {Name: "Mouse", Price: 29.99, SKU: "MOU001"},
        {Name: "Keyboard", Price: 79.99, SKU: "KEY001"},
    }
    
    for _, product := range products {
        var existing models.Product
        err := db.Where("sku = ?", product.SKU).First(&existing).Error
        
        if err == gorm.ErrRecordNotFound {
            if err := db.Create(&product).Error; err != nil {
                return err
            }
        } else if err != nil {
            return err
        }
    }
    
    return nil
}
```

### Example 2: Order Seeder with Dependencies

```go
type OrderSeeder struct {
    *BaseSeeder
}

func NewOrderSeeder() *OrderSeeder {
    return &OrderSeeder{
        BaseSeeder: NewBaseSeeder(
            "order_seeder",
            "Seeds sample orders with user and product relationships",
            "user_seeder", "product_seeder", // Dependencies
        ),
    }
}

func (s *OrderSeeder) Seed(ctx context.Context, db *gorm.DB) error {
    // Get seeded users and products
    var users []models.User
    if err := db.Find(&users).Error; err != nil {
        return err
    }
    
    var products []models.Product
    if err := db.Find(&products).Error; err != nil {
        return err
    }
    
    // Create sample orders
    for i, user := range users[:3] { // First 3 users
        order := models.Order{
            UserID: user.ID,
            Total:  100.00 * float64(i+1),
            Status: "completed",
        }
        
        if err := db.Create(&order).Error; err != nil {
            return err
        }
    }
    
    return nil
}
```

## Troubleshooting

### Common Issues

1. **Circular Dependencies**: Ensure your seeder dependencies don't form cycles
2. **Missing Dependencies**: Make sure all referenced seeders are registered
3. **Database Connection**: Verify your database configuration is correct
4. **Permission Issues**: Ensure the database user has necessary permissions

### Debug Mode

Run with verbose output to see detailed execution information:

```bash
# The CLI automatically provides detailed logging
go run cmd/seed/main.go run --all
```

### Manual Execution Tracking

If you need to manually mark a seeder as executed or reset its status:

```sql
-- Mark as executed
INSERT INTO seeder_executions (name, run_at, success) 
VALUES ('my_seeder', EXTRACT(EPOCH FROM NOW()), true);

-- Reset seeder status
DELETE FROM seeder_executions WHERE name = 'my_seeder';
```

## Integration with CI/CD

### Example GitHub Actions Workflow

```yaml
name: Database Setup
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:13
        env:
          POSTGRES_PASSWORD: postgres
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5

    steps:
      - uses: actions/checkout@v2
      
      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.21
          
      - name: Run Migrations
        run: go run cmd/migrate/main.go up
        
      - name: Run Seeders
        run: go run cmd/seed/main.go run --env
        
      - name: Run Tests
        run: go test ./...
```

This documentation provides a comprehensive guide to using and extending the database seeder system in your Go Fiber Template project.