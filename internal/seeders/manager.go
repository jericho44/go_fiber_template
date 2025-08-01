package seeders

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"gorm.io/gorm"
)

// DefaultSeederManager implements SeederManager interface
type DefaultSeederManager struct {
	db      *gorm.DB
	seeders map[string]Seeder
	logger  *log.Logger
}

// NewSeederManager creates a new seeder manager
func NewSeederManager(db *gorm.DB) *DefaultSeederManager {
	return &DefaultSeederManager{
		db:      db,
		seeders: make(map[string]Seeder),
		logger:  log.New(log.Writer(), "[SEEDER] ", log.LstdFlags),
	}
}

// RegisterSeeder registers a new seeder
func (sm *DefaultSeederManager) RegisterSeeder(seeder Seeder) error {
	name := seeder.GetName()
	if name == "" {
		return fmt.Errorf("seeder name cannot be empty")
	}

	if _, exists := sm.seeders[name]; exists {
		return fmt.Errorf("seeder with name '%s' already registered", name)
	}

	sm.seeders[name] = seeder
	sm.logger.Printf("Registered seeder: %s", name)
	return nil
}

// RunAll runs all registered seeders in dependency order
func (sm *DefaultSeederManager) RunAll(ctx context.Context) error {
	if err := sm.ensureSeederTable(); err != nil {
		return fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	orderedSeeders, err := sm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	sm.logger.Printf("Running %d seeders in dependency order", len(orderedSeeders))

	ctx = context.WithValue(ctx, "run_time", time.Now().Unix())

	for _, seeder := range orderedSeeders {
		if err := sm.runSingleSeeder(ctx, seeder); err != nil {
			return fmt.Errorf("failed to run seeder '%s': %w", seeder.GetName(), err)
		}
	}

	sm.logger.Println("All seeders completed successfully")
	return nil
}

// RunSeeder runs a specific seeder by name
func (sm *DefaultSeederManager) RunSeeder(ctx context.Context, name string) error {
	if err := sm.ensureSeederTable(); err != nil {
		return fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	seeder, exists := sm.seeders[name]
	if !exists {
		return fmt.Errorf("seeder '%s' not found", name)
	}

	ctx = context.WithValue(ctx, "run_time", time.Now().Unix())
	return sm.runSingleSeeder(ctx, seeder)
}

// RunSeeders runs specific seeders by names
func (sm *DefaultSeederManager) RunSeeders(ctx context.Context, names []string) error {
	if err := sm.ensureSeederTable(); err != nil {
		return fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	// Collect requested seeders
	requestedSeeders := make([]Seeder, 0, len(names))
	for _, name := range names {
		seeder, exists := sm.seeders[name]
		if !exists {
			return fmt.Errorf("seeder '%s' not found", name)
		}
		requestedSeeders = append(requestedSeeders, seeder)
	}

	// Resolve dependencies for requested seeders
	orderedSeeders, err := sm.resolveDependenciesForSeeders(requestedSeeders)
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	ctx = context.WithValue(ctx, "run_time", time.Now().Unix())

	for _, seeder := range orderedSeeders {
		if err := sm.runSingleSeeder(ctx, seeder); err != nil {
			return fmt.Errorf("failed to run seeder '%s': %w", seeder.GetName(), err)
		}
	}

	return nil
}

// RollbackSeeder rollbacks a specific seeder by name
func (sm *DefaultSeederManager) RollbackSeeder(ctx context.Context, name string) error {
	seeder, exists := sm.seeders[name]
	if !exists {
		return fmt.Errorf("seeder '%s' not found", name)
	}

	sm.logger.Printf("Rolling back seeder: %s", name)

	if err := seeder.Rollback(ctx, sm.db); err != nil {
		sm.logger.Printf("Failed to rollback seeder '%s': %v", name, err)
		return fmt.Errorf("rollback failed: %w", err)
	}

	// Remove execution record
	if baseSeeder, ok := seeder.(*BaseSeeder); ok {
		if err := baseSeeder.RemoveExecutionRecord(ctx, sm.db); err != nil {
			sm.logger.Printf("Failed to remove execution record for '%s': %v", name, err)
		}
	}

	sm.logger.Printf("Successfully rolled back seeder: %s", name)
	return nil
}

// RollbackAll rollbacks all seeders in reverse dependency order
func (sm *DefaultSeederManager) RollbackAll(ctx context.Context) error {
	orderedSeeders, err := sm.resolveDependencies()
	if err != nil {
		return fmt.Errorf("failed to resolve dependencies: %w", err)
	}

	// Reverse the order for rollback
	for i := len(orderedSeeders) - 1; i >= 0; i-- {
		seeder := orderedSeeders[i]
		if err := sm.RollbackSeeder(ctx, seeder.GetName()); err != nil {
			sm.logger.Printf("Failed to rollback seeder '%s': %v", seeder.GetName(), err)
			// Continue with other rollbacks even if one fails
		}
	}

	return nil
}

// ListSeeders returns information about all registered seeders
func (sm *DefaultSeederManager) ListSeeders() []SeederInfo {
	infos := make([]SeederInfo, 0, len(sm.seeders))

	for _, seeder := range sm.seeders {
		infos = append(infos, SeederInfo{
			Name:         seeder.GetName(),
			Description:  seeder.GetDescription(),
			Dependencies: seeder.GetDependencies(),
		})
	}

	// Sort by name for consistent output
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// GetSeederStatus returns the status of seeders
func (sm *DefaultSeederManager) GetSeederStatus(ctx context.Context) ([]SeederStatus, error) {
	if err := sm.ensureSeederTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure seeder table: %w", err)
	}

	statuses := make([]SeederStatus, 0, len(sm.seeders))

	for _, seeder := range sm.seeders {
		var execution SeederExecution
		err := sm.db.Where("name = ? AND success = ?", seeder.GetName(), true).
			Order("run_at DESC").
			First(&execution).Error

		status := SeederStatus{
			Name:        seeder.GetName(),
			Description: seeder.GetDescription(),
			HasRun:      false,
		}

		if err == nil {
			status.HasRun = true
			status.LastRunAt = &execution.RunAt
		} else if err != gorm.ErrRecordNotFound {
			status.Error = err.Error()
		}

		statuses = append(statuses, status)
	}

	// Sort by name for consistent output
	sort.Slice(statuses, func(i, j int) bool {
		return statuses[i].Name < statuses[j].Name
	})

	return statuses, nil
}

// runSingleSeeder runs a single seeder with proper error handling
func (sm *DefaultSeederManager) runSingleSeeder(ctx context.Context, seeder Seeder) error {
	name := seeder.GetName()

	// Check if seeder should run
	shouldRun, err := seeder.ShouldRun(ctx, sm.db)
	if err != nil {
		return fmt.Errorf("failed to check if seeder should run: %w", err)
	}

	if !shouldRun {
		sm.logger.Printf("Skipping seeder '%s' (already executed)", name)
		return nil
	}

	sm.logger.Printf("Running seeder: %s", name)

	// Run seeder in transaction
	err = sm.db.Transaction(func(tx *gorm.DB) error {
		if err := seeder.Seed(ctx, tx); err != nil {
			return err
		}

		// Mark as executed if seeder is BaseSeeder
		if baseSeeder, ok := seeder.(*BaseSeeder); ok {
			return baseSeeder.MarkAsExecuted(ctx, tx)
		}

		return nil
	})

	if err != nil {
		sm.logger.Printf("Failed to run seeder '%s': %v", name, err)

		// Mark as failed if seeder is BaseSeeder
		if baseSeeder, ok := seeder.(*BaseSeeder); ok {
			baseSeeder.MarkAsFailed(ctx, sm.db, err)
		}

		return err
	}

	sm.logger.Printf("Successfully completed seeder: %s", name)
	return nil
}

// ensureSeederTable ensures the seeder execution table exists
func (sm *DefaultSeederManager) ensureSeederTable() error {
	return sm.db.AutoMigrate(&SeederExecution{})
}

// resolveDependencies resolves seeder dependencies and returns them in execution order
func (sm *DefaultSeederManager) resolveDependencies() ([]Seeder, error) {
	return sm.resolveDependenciesForSeeders(sm.getAllSeeders())
}

// resolveDependenciesForSeeders resolves dependencies for specific seeders
func (sm *DefaultSeederManager) resolveDependenciesForSeeders(seeders []Seeder) ([]Seeder, error) {
	visited := make(map[string]bool)
	visiting := make(map[string]bool)
	result := make([]Seeder, 0)

	var visit func(seeder Seeder) error
	visit = func(seeder Seeder) error {
		name := seeder.GetName()

		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving seeder '%s'", name)
		}

		if visited[name] {
			return nil
		}

		visiting[name] = true

		// Visit dependencies first
		for _, depName := range seeder.GetDependencies() {
			depSeeder, exists := sm.seeders[depName]
			if !exists {
				return fmt.Errorf("dependency '%s' not found for seeder '%s'", depName, name)
			}

			if err := visit(depSeeder); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		result = append(result, seeder)

		return nil
	}

	for _, seeder := range seeders {
		if err := visit(seeder); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// getAllSeeders returns all registered seeders as a slice
func (sm *DefaultSeederManager) getAllSeeders() []Seeder {
	seeders := make([]Seeder, 0, len(sm.seeders))
	for _, seeder := range sm.seeders {
		seeders = append(seeders, seeder)
	}
	return seeders
}
