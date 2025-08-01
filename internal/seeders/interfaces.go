package seeders

import (
	"context"

	"gorm.io/gorm"
)

// Seeder defines the interface for all seeders
type Seeder interface {
	// GetName returns the unique name of the seeder
	GetName() string

	// GetDescription returns a description of what this seeder does
	GetDescription() string

	// GetDependencies returns a list of seeder names that must run before this one
	GetDependencies() []string

	// Seed executes the seeding logic
	Seed(ctx context.Context, db *gorm.DB) error

	// Rollback removes the seeded data (optional, can return nil if not supported)
	Rollback(ctx context.Context, db *gorm.DB) error

	// ShouldRun determines if this seeder should run based on current database state
	ShouldRun(ctx context.Context, db *gorm.DB) (bool, error)
}

// SeederManager manages the execution of seeders
type SeederManager interface {
	// RegisterSeeder registers a new seeder
	RegisterSeeder(seeder Seeder) error

	// RunAll runs all registered seeders in dependency order
	RunAll(ctx context.Context) error

	// RunSeeder runs a specific seeder by name
	RunSeeder(ctx context.Context, name string) error

	// RunSeeders runs specific seeders by names
	RunSeeders(ctx context.Context, names []string) error

	// RollbackSeeder rollbacks a specific seeder by name
	RollbackSeeder(ctx context.Context, name string) error

	// RollbackAll rollbacks all seeders in reverse dependency order
	RollbackAll(ctx context.Context) error

	// ListSeeders returns information about all registered seeders
	ListSeeders() []SeederInfo

	// GetSeederStatus returns the status of seeders
	GetSeederStatus(ctx context.Context) ([]SeederStatus, error)
}

// SeederInfo contains information about a seeder
type SeederInfo struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies"`
}

// SeederStatus represents the execution status of a seeder
type SeederStatus struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	HasRun      bool   `json:"has_run"`
	LastRunAt   *int64 `json:"last_run_at,omitempty"`
	Error       string `json:"error,omitempty"`
}

// SeederExecution tracks seeder execution history
type SeederExecution struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Name      string `json:"name" gorm:"unique;not null;size:255"`
	RunAt     int64  `json:"run_at" gorm:"not null"`
	Success   bool   `json:"success" gorm:"not null;default:false"`
	Error     string `json:"error,omitempty" gorm:"type:text"`
	CreatedAt int64  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName returns the table name for SeederExecution
func (SeederExecution) TableName() string {
	return "seeder_executions"
}
