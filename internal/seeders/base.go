package seeders

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// BaseSeeder provides a basic implementation of the Seeder interface
type BaseSeeder struct {
	name         string
	description  string
	dependencies []string
}

// NewBaseSeeder creates a new BaseSeeder
func NewBaseSeeder(name, description string, dependencies ...string) *BaseSeeder {
	return &BaseSeeder{
		name:         name,
		description:  description,
		dependencies: dependencies,
	}
}

// GetName returns the seeder name
func (b *BaseSeeder) GetName() string {
	return b.name
}

// GetDescription returns the seeder description
func (b *BaseSeeder) GetDescription() string {
	return b.description
}

// GetDependencies returns the seeder dependencies
func (b *BaseSeeder) GetDependencies() []string {
	return b.dependencies
}

// Seed must be implemented by concrete seeders
func (b *BaseSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	return fmt.Errorf("Seed method must be implemented by concrete seeder")
}

// Rollback provides a default implementation that does nothing
func (b *BaseSeeder) Rollback(ctx context.Context, db *gorm.DB) error {
	// Default implementation does nothing
	return nil
}

// ShouldRun provides a default implementation that checks if the seeder has been executed
func (b *BaseSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
	var execution SeederExecution
	err := db.Where("name = ? AND success = ?", b.name, true).First(&execution).Error

	if err == gorm.ErrRecordNotFound {
		return true, nil // Should run if no successful execution found
	}

	if err != nil {
		return false, fmt.Errorf("failed to check seeder execution status: %w", err)
	}

	return false, nil // Should not run if successful execution found
}

// MarkAsExecuted marks the seeder as successfully executed
func (b *BaseSeeder) MarkAsExecuted(ctx context.Context, db *gorm.DB) error {
	execution := SeederExecution{
		Name:    b.name,
		RunAt:   ctx.Value("run_time").(int64),
		Success: true,
	}

	// Use UPSERT to handle duplicate executions
	return db.Save(&execution).Error
}

// MarkAsFailed marks the seeder as failed with error message
func (b *BaseSeeder) MarkAsFailed(ctx context.Context, db *gorm.DB, err error) error {
	execution := SeederExecution{
		Name:    b.name,
		RunAt:   ctx.Value("run_time").(int64),
		Success: false,
		Error:   err.Error(),
	}

	return db.Save(&execution).Error
}

// RemoveExecutionRecord removes the execution record for this seeder
func (b *BaseSeeder) RemoveExecutionRecord(ctx context.Context, db *gorm.DB) error {
	return db.Where("name = ?", b.name).Delete(&SeederExecution{}).Error
}
