package seeders

import (
	"go-fiber-template/internal/config"

	"gorm.io/gorm"
)

// Registry holds all available seeders and provides easy access
type Registry struct {
	manager *DefaultSeederManager
	config  *config.Config
}

// NewRegistry creates a new seeder registry
func NewRegistry(db *gorm.DB, cfg *config.Config) *Registry {
	manager := NewSeederManager(db)

	registry := &Registry{
		manager: manager,
		config:  cfg,
	}

	// Register all available seeders
	registry.registerAllSeeders()

	return registry
}

// GetManager returns the seeder manager
func (r *Registry) GetManager() SeederManager {
	return r.manager
}

// registerAllSeeders registers all available seeders
func (r *Registry) registerAllSeeders() {
	// Register core seeders
	r.manager.RegisterSeeder(NewUserSeeder())

	// Register demo data seeders (only in development)
	if r.config.IsDevelopment() {
		r.manager.RegisterSeeder(NewDemoDataSeeder())
	}
}

// GetAvailableSeeders returns a list of all available seeders
func (r *Registry) GetAvailableSeeders() []string {
	infos := r.manager.ListSeeders()
	names := make([]string, len(infos))

	for i, info := range infos {
		names[i] = info.Name
	}

	return names
}

// GetSeedersByEnvironment returns seeders appropriate for the current environment
func (r *Registry) GetSeedersByEnvironment() []string {
	var seeders []string

	// Core seeders for all environments
	seeders = append(seeders, "user_seeder")

	// Development-specific seeders
	if r.config.IsDevelopment() {
		seeders = append(seeders, "demo_data_seeder")
	}

	return seeders
}
