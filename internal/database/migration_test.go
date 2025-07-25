package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrationManager_Create(t *testing.T) {
	// Create temporary directory for test migrations
	tempDir := t.TempDir()

	// Create a test migration manager with custom migrations path
	mm := &MigrationManager{
		migrationsPath: tempDir,
	}

	// Test creating a migration
	migrationName := "create_test_table"
	err := mm.Create(migrationName)
	require.NoError(t, err)

	// Check if files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "*.sql"))
	require.NoError(t, err)
	assert.Len(t, files, 2) // Should have up and down files

	// Check file names contain timestamp and migration name
	upFile := filepath.Join(tempDir, fmt.Sprintf("*_%s.up.sql", migrationName))
	downFile := filepath.Join(tempDir, fmt.Sprintf("*_%s.down.sql", migrationName))

	upFiles, err := filepath.Glob(upFile)
	require.NoError(t, err)
	assert.Len(t, upFiles, 1)

	downFiles, err := filepath.Glob(downFile)
	require.NoError(t, err)
	assert.Len(t, downFiles, 1)

	// Check file contents
	upContent, err := os.ReadFile(upFiles[0])
	require.NoError(t, err)
	assert.Contains(t, string(upContent), migrationName)
	assert.Contains(t, string(upContent), "Add your up migration SQL here")

	downContent, err := os.ReadFile(downFiles[0])
	require.NoError(t, err)
	assert.Contains(t, string(downContent), migrationName)
	assert.Contains(t, string(downContent), "Add your down migration SQL here")
}

func TestMigrationManager_CreateEmptyName(t *testing.T) {
	mm := &MigrationManager{
		migrationsPath: t.TempDir(),
	}

	err := mm.Create("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "migration name cannot be empty")
}

func TestMigrationManager_CreateDirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	nonExistentDir := filepath.Join(tempDir, "migrations", "nested")

	mm := &MigrationManager{
		migrationsPath: nonExistentDir,
	}

	err := mm.Create("test_migration")
	require.NoError(t, err)

	// Check if directory was created
	_, err = os.Stat(nonExistentDir)
	assert.NoError(t, err)
}

// Integration test for migration manager (requires database)
func TestMigrationManager_Integration(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Load test configuration
	cfg, err := config.Load()
	if err != nil {
		t.Skip("Skipping integration test: failed to load config")
	}

	// Use test database
	cfg.Database.Name = cfg.Database.Name + "_migration_test"

	// Create test database
	adminDB, err := sql.Open("postgres", cfg.Database.GetAdminDSN())
	if err != nil {
		t.Skip("Skipping integration test: failed to connect to admin database")
	}
	defer adminDB.Close()

	// Drop test database if exists and create new one
	_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.Database.Name))
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", cfg.Database.Name))
	if err != nil {
		t.Skip("Skipping integration test: failed to create test database")
	}

	// Clean up test database after test
	defer func() {
		_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", cfg.Database.Name))
	}()

	// Create migration manager
	mm, err := NewMigrationManager(cfg.Database)
	require.NoError(t, err)
	defer mm.Close()

	// Create temporary migrations directory
	tempDir := t.TempDir()
	mm.migrationsPath = tempDir

	// Test creating a migration
	err = mm.Create("test_table")
	require.NoError(t, err)

	// Create a simple test migration
	timestamp := time.Now().Format("20060102150405")
	upFile := filepath.Join(tempDir, fmt.Sprintf("%s_test_table.up.sql", timestamp))
	downFile := filepath.Join(tempDir, fmt.Sprintf("%s_test_table.down.sql", timestamp))

	upSQL := `CREATE TABLE test_table (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	downSQL := `DROP TABLE IF EXISTS test_table;`

	err = os.WriteFile(upFile, []byte(upSQL), 0644)
	require.NoError(t, err)

	err = os.WriteFile(downFile, []byte(downSQL), 0644)
	require.NoError(t, err)

	// Test version before migration
	err = mm.Version()
	require.NoError(t, err)

	// Test status before migration
	err = mm.Status()
	require.NoError(t, err)

	// Test running migration up
	err = mm.Up(1)
	require.NoError(t, err)

	// Verify table was created
	var exists bool
	err = mm.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_table')").Scan(&exists)
	require.NoError(t, err)
	assert.True(t, exists)

	// Test version after migration
	err = mm.Version()
	require.NoError(t, err)

	// Test status after migration
	err = mm.Status()
	require.NoError(t, err)

	// Test rolling back migration
	err = mm.Down(1)
	require.NoError(t, err)

	// Verify table was dropped
	err = mm.db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'test_table')").Scan(&exists)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestMigrationManager_FileOperations(t *testing.T) {
	tempDir := t.TempDir()
	mm := &MigrationManager{
		migrationsPath: tempDir,
	}

	// Test creating multiple migrations
	migrations := []string{"create_users", "create_posts", "add_indexes"}

	for _, name := range migrations {
		err := mm.Create(name)
		require.NoError(t, err)
	}

	// Check all files were created
	files, err := filepath.Glob(filepath.Join(tempDir, "*.sql"))
	require.NoError(t, err)
	assert.Len(t, files, len(migrations)*2) // Each migration has up and down files

	// Check file naming pattern
	for _, name := range migrations {
		upFiles, err := filepath.Glob(filepath.Join(tempDir, fmt.Sprintf("*_%s.up.sql", name)))
		require.NoError(t, err)
		assert.Len(t, upFiles, 1)

		downFiles, err := filepath.Glob(filepath.Join(tempDir, fmt.Sprintf("*_%s.down.sql", name)))
		require.NoError(t, err)
		assert.Len(t, downFiles, 1)
	}
}
