package seeders

import (
	"context"
	"testing"
	"time"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	
	// Auto-migrate test tables
	err = db.AutoMigrate(&SeederExecution{})
	require.NoError(t, err)
	
	return db
}

// TestSeederExecution tests the SeederExecution model
func TestSeederExecution(t *testing.T) {
	db := setupTestDB(t)
	
	execution := SeederExecution{
		Name:    "test_seeder",
		RunAt:   time.Now().Unix(),
		Success: true,
	}
	
	err := db.Create(&execution).Error
	assert.NoError(t, err)
	assert.NotZero(t, execution.ID)
	
	// Test retrieval
	var retrieved SeederExecution
	err = db.Where("name = ?", "test_seeder").First(&retrieved).Error
	assert.NoError(t, err)
	assert.Equal(t, "test_seeder", retrieved.Name)
	assert.True(t, retrieved.Success)
}

// TestBaseSeeder tests the BaseSeeder functionality
func TestBaseSeeder(t *testing.T) {
	seeder := NewBaseSeeder("test_seeder", "Test seeder description", "dependency1")
	
	assert.Equal(t, "test_seeder", seeder.GetName())
	assert.Equal(t, "Test seeder description", seeder.GetDescription())
	assert.Equal(t, []string{"dependency1"}, seeder.GetDependencies())
}

// TestSeederManager tests the seeder manager functionality
func TestSeederManager(t *testing.T) {
	db := setupTestDB(t)
	manager := NewSeederManager(db)
	
	// Create a test seeder
	testSeeder := &TestSeeder{
		BaseSeeder: NewBaseSeeder("test_seeder", "Test seeder"),
	}
	
	// Test registration
	err := manager.RegisterSeeder(testSeeder)
	assert.NoError(t, err)
	
	// Test duplicate registration
	err = manager.RegisterSeeder(testSeeder)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")
	
	// Test listing seeders
	infos := manager.ListSeeders()
	assert.Len(t, infos, 1)
	assert.Equal(t, "test_seeder", infos[0].Name)
	assert.Equal(t, "Test seeder", infos[0].Description)
}

// TestSeederDependencyResolution tests dependency resolution
func TestSeederDependencyResolution(t *testing.T) {
	db := setupTestDB(t)
	manager := NewSeederManager(db)
	
	// Create seeders with dependencies
	seeder1 := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder1", "First seeder"),
	}
	seeder2 := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder2", "Second seeder", "seeder1"),
	}
	seeder3 := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder3", "Third seeder", "seeder2"),
	}
	
	// Register in reverse order to test dependency resolution
	manager.RegisterSeeder(seeder3)
	manager.RegisterSeeder(seeder2)
	manager.RegisterSeeder(seeder1)
	
	ctx := context.WithValue(context.Background(), "run_time", time.Now().Unix())
	
	// Run all seeders
	err := manager.RunAll(ctx)
	assert.NoError(t, err)
	
	// Verify execution order by checking the database
	var executions []SeederExecution
	err = db.Order("created_at ASC").Find(&executions).Error
	assert.NoError(t, err)
	assert.Len(t, executions, 3)
	
	// Should execute in dependency order: seeder1, seeder2, seeder3
	assert.Equal(t, "seeder1", executions[0].Name)
	assert.Equal(t, "seeder2", executions[1].Name)
	assert.Equal(t, "seeder3", executions[2].Name)
}

// TestSeederRegistry tests the seeder registry
func TestSeederRegistry(t *testing.T) {
	db := setupTestDB(t)
	
	// Create a test config
	cfg := &config.Config{
		Env: config.Development,
	}
	
	registry := NewRegistry(db, cfg)
	manager := registry.GetManager()
	
	// In development, should have both user_seeder and demo_data_seeder
	infos := manager.ListSeeders()
	assert.Len(t, infos, 2)
	
	seederNames := make([]string, len(infos))
	for i, info := range infos {
		seederNames[i] = info.Name
	}
	
	assert.Contains(t, seederNames, "user_seeder")
	assert.Contains(t, seederNames, "demo_data_seeder")
	
	// Test environment-specific seeders
	envSeeders := registry.GetSeedersByEnvironment()
	assert.Contains(t, envSeeders, "user_seeder")
	assert.Contains(t, envSeeders, "demo_data_seeder")
}

// TestSeeder is a simple test seeder for testing purposes
type TestSeeder struct {
	*BaseSeeder
	executed bool
}

func (s *TestSeeder) Seed(ctx context.Context, db *gorm.DB) error {
	s.executed = true
	return nil
}

func (s *TestSeeder) ShouldRun(ctx context.Context, db *gorm.DB) (bool, error) {
	return !s.executed, nil
}

// TestCircularDependency tests circular dependency detection
func TestCircularDependency(t *testing.T) {
	db := setupTestDB(t)
	manager := NewSeederManager(db)
	
	// Create seeders with circular dependencies
	seeder1 := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder1", "First seeder", "seeder2"),
	}
	seeder2 := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder2", "Second seeder", "seeder1"),
	}
	
	manager.RegisterSeeder(seeder1)
	manager.RegisterSeeder(seeder2)
	
	ctx := context.WithValue(context.Background(), "run_time", time.Now().Unix())
	
	// Should detect circular dependency
	err := manager.RunAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

// TestMissingDependency tests missing dependency detection
func TestMissingDependency(t *testing.T) {
	db := setupTestDB(t)
	manager := NewSeederManager(db)
	
	// Create seeder with missing dependency
	seeder := &TestSeeder{
		BaseSeeder: NewBaseSeeder("seeder1", "First seeder", "missing_seeder"),
	}
	
	manager.RegisterSeeder(seeder)
	
	ctx := context.WithValue(context.Background(), "run_time", time.Now().Unix())
	
	// Should detect missing dependency
	err := manager.RunAll(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}