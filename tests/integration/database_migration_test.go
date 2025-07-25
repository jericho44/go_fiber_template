package integration

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDatabaseMigrations tests database migration functionality
func TestDatabaseMigrations(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("migration_test"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer postgresContainer.Terminate(ctx)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create database config
	dbConfig := &config.DatabaseConfig{
		Host:     host,
		Port:     port.Int(),
		User:     "testuser",
		Password: "testpass",
		Name:     "migration_test",
		SSLMode:  "disable",
	}

	// Create connection string
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	t.Run("Fresh Migration Up", func(t *testing.T) {
		// Open database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Run migrations up
		err = m.Up()
		require.NoError(t, err)

		// Verify tables were created
		expectedTables := []string{
			"users",
			"token_blacklists",
			"schema_migrations",
		}

		for _, table := range expectedTables {
			var exists bool
			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = $1
			)`
			err = db.QueryRow(query, table).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "Table %s should exist", table)
		}

		// Verify users table structure
		t.Run("Users Table Structure", func(t *testing.T) {
			expectedColumns := map[string]string{
				"id":         "integer",
				"email":      "character varying",
				"password":   "character varying",
				"first_name": "character varying",
				"last_name":  "character varying",
				"is_active":  "boolean",
				"created_at": "timestamp with time zone",
				"updated_at": "timestamp with time zone",
			}

			for column, expectedType := range expectedColumns {
				var dataType string
				query := `SELECT data_type FROM information_schema.columns 
						 WHERE table_name = 'users' AND column_name = $1`
				err = db.QueryRow(query, column).Scan(&dataType)
				require.NoError(t, err, "Column %s should exist", column)
				assert.Contains(t, dataType, expectedType, "Column %s should have type %s", column, expectedType)
			}

			// Check constraints
			var constraintCount int
			query := `SELECT COUNT(*) FROM information_schema.table_constraints 
					 WHERE table_name = 'users' AND constraint_type = 'UNIQUE'`
			err = db.QueryRow(query).Scan(&constraintCount)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, constraintCount, 1, "Users table should have at least one unique constraint")
		})

		// Verify token_blacklists table structure
		t.Run("Token Blacklists Table Structure", func(t *testing.T) {
			expectedColumns := map[string]string{
				"id":         "integer",
				"token_hash": "character varying",
				"reason":     "character varying",
				"expires_at": "timestamp with time zone",
				"created_at": "timestamp with time zone",
			}

			for column, expectedType := range expectedColumns {
				var dataType string
				query := `SELECT data_type FROM information_schema.columns 
						 WHERE table_name = 'token_blacklists' AND column_name = $1`
				err = db.QueryRow(query, column).Scan(&dataType)
				require.NoError(t, err, "Column %s should exist", column)
				assert.Contains(t, dataType, expectedType, "Column %s should have type %s", column, expectedType)
			}
		})

		// Get current migration version
		version, dirty, err := m.Version()
		require.NoError(t, err)
		assert.False(t, dirty, "Migration should not be dirty")
		assert.Greater(t, version, uint(0), "Migration version should be greater than 0")

		// Test migration down
		t.Run("Migration Down", func(t *testing.T) {
			// Run one step down
			err = m.Steps(-1)
			require.NoError(t, err)

			// Verify one table is removed (token_blacklists should be removed first)
			var exists bool
			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'token_blacklists'
			)`
			err = db.QueryRow(query).Scan(&exists)
			require.NoError(t, err)
			assert.False(t, exists, "token_blacklists table should be removed")

			// Users table should still exist
			query = `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'users'
			)`
			err = db.QueryRow(query).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "users table should still exist")
		})

		// Test migration back up
		t.Run("Migration Back Up", func(t *testing.T) {
			err = m.Up()
			require.NoError(t, err)

			// Verify all tables exist again
			expectedTables := []string{
				"users",
				"token_blacklists",
			}

			for _, table := range expectedTables {
				var exists bool
				query := `SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)`
				err = db.QueryRow(query, table).Scan(&exists)
				require.NoError(t, err)
				assert.True(t, exists, "Table %s should exist after migration up", table)
			}
		})

		// Test complete rollback
		t.Run("Complete Rollback", func(t *testing.T) {
			err = m.Down()
			require.NoError(t, err)

			// Verify all tables are removed except schema_migrations
			tables := []string{"users", "token_blacklists"}
			for _, table := range tables {
				var exists bool
				query := `SELECT EXISTS (
					SELECT FROM information_schema.tables 
					WHERE table_schema = 'public' 
					AND table_name = $1
				)`
				err = db.QueryRow(query, table).Scan(&exists)
				require.NoError(t, err)
				assert.False(t, exists, "Table %s should be removed after complete rollback", table)
			}

			// schema_migrations should still exist
			var exists bool
			query := `SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'schema_migrations'
			)`
			err = db.QueryRow(query).Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists, "schema_migrations table should still exist")
		})
	})

	t.Run("Migration with Data", func(t *testing.T) {
		// Open new database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Run migrations up
		err = m.Up()
		require.NoError(t, err)

		// Insert test data
		_, err = db.Exec(`
			INSERT INTO users (email, password, first_name, last_name, is_active, created_at, updated_at)
			VALUES ('test@example.com', 'hashed_password', 'Test', 'User', true, NOW(), NOW())
		`)
		require.NoError(t, err)

		_, err = db.Exec(`
			INSERT INTO token_blacklists (token_hash, reason, expires_at, created_at)
			VALUES ('test_token_hash', 'logout', NOW() + INTERVAL '1 hour', NOW())
		`)
		require.NoError(t, err)

		// Verify data exists
		var userCount, tokenCount int
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		require.NoError(t, err)
		assert.Equal(t, 1, userCount)

		err = db.QueryRow("SELECT COUNT(*) FROM token_blacklists").Scan(&tokenCount)
		require.NoError(t, err)
		assert.Equal(t, 1, tokenCount)

		// Test migration down with data (should handle gracefully)
		err = m.Steps(-1)
		require.NoError(t, err)

		// Users table should still exist with data
		err = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
		require.NoError(t, err)
		assert.Equal(t, 1, userCount)

		// token_blacklists table should be gone
		_, err = db.Query("SELECT COUNT(*) FROM token_blacklists")
		assert.Error(t, err) // Should error because table doesn't exist
	})

	t.Run("Migration Error Handling", func(t *testing.T) {
		// Test with invalid migration path
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Try to create migration with invalid path
		_, err = migrate.NewWithDatabaseInstance(
			"file://nonexistent/path",
			"postgres",
			driver,
		)
		assert.Error(t, err, "Should error with invalid migration path")
	})

	t.Run("Database Connection Integration", func(t *testing.T) {
		// Test that our database package can connect after migrations
		db, err := database.Connect(dbConfig)
		require.NoError(t, err)

		// Get underlying SQL DB
		sqlDB, err := db.DB()
		require.NoError(t, err)
		defer sqlDB.Close()

		// Test connection
		err = sqlDB.Ping()
		require.NoError(t, err)

		// Test that we can perform basic operations
		var result int
		err = sqlDB.QueryRow("SELECT 1").Scan(&result)
		require.NoError(t, err)
		assert.Equal(t, 1, result)
	})
}

// TestMigrationCLI tests the migration CLI commands
func TestMigrationCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CLI test in short mode")
	}

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("cli_test"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer postgresContainer.Terminate(ctx)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Set environment variables for CLI
	t.Setenv("DB_HOST", host)
	t.Setenv("DB_PORT", fmt.Sprintf("%d", port.Int()))
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpass")
	t.Setenv("DB_NAME", "cli_test")
	t.Setenv("DB_SSLMODE", "disable")

	t.Run("Migration Up Command", func(t *testing.T) {
		// This would test the actual CLI command
		// For now, we'll test that the migration logic works

		// Create connection string
		dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%d/cli_test?sslmode=disable",
			host, port.Int())

		// Open database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Test up command
		err = m.Up()
		require.NoError(t, err)

		// Verify tables exist
		var tableCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name IN ('users', 'token_blacklists')
		`).Scan(&tableCount)
		require.NoError(t, err)
		assert.Equal(t, 2, tableCount)
	})

	t.Run("Migration Status", func(t *testing.T) {
		// Create connection string
		dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%d/cli_test?sslmode=disable",
			host, port.Int())

		// Open database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Get version
		version, dirty, err := m.Version()
		require.NoError(t, err)
		assert.False(t, dirty)
		assert.Greater(t, version, uint(0))

		t.Logf("Current migration version: %d, dirty: %v", version, dirty)
	})
}

// TestMigrationRollback tests migration rollback scenarios
func TestMigrationRollback(t *testing.T) {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("rollback_test"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	require.NoError(t, err)
	defer postgresContainer.Terminate(ctx)

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create connection string
	dsn := fmt.Sprintf("postgres://testuser:testpass@%s:%d/rollback_test?sslmode=disable",
		host, port.Int())

	t.Run("Rollback with Foreign Key Constraints", func(t *testing.T) {
		// Open database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Run migrations up
		err = m.Up()
		require.NoError(t, err)

		// Insert test data with relationships
		_, err = db.Exec(`
			INSERT INTO users (email, password, first_name, last_name, is_active, created_at, updated_at)
			VALUES ('rollback@example.com', 'hashed_password', 'Rollback', 'Test', true, NOW(), NOW())
		`)
		require.NoError(t, err)

		// Get the user ID
		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE email = 'rollback@example.com'").Scan(&userID)
		require.NoError(t, err)

		// Insert related data
		_, err = db.Exec(`
			INSERT INTO token_blacklists (token_hash, reason, expires_at, created_at)
			VALUES ('rollback_token_hash', 'test_rollback', NOW() + INTERVAL '1 hour', NOW())
		`)
		require.NoError(t, err)

		// Test rollback - should handle foreign key constraints properly
		err = m.Down()
		require.NoError(t, err)

		// Verify all tables are removed
		var tableCount int
		err = db.QueryRow(`
			SELECT COUNT(*) FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name IN ('users', 'token_blacklists')
		`).Scan(&tableCount)
		require.NoError(t, err)
		assert.Equal(t, 0, tableCount)
	})

	t.Run("Partial Rollback", func(t *testing.T) {
		// Open database connection
		db, err := sql.Open("postgres", dsn)
		require.NoError(t, err)
		defer db.Close()

		// Create migration driver
		driver, err := postgres.WithInstance(db, &postgres.Config{})
		require.NoError(t, err)

		// Create migration instance
		m, err := migrate.NewWithDatabaseInstance(
			"file://../../migrations",
			"postgres",
			driver,
		)
		require.NoError(t, err)

		// Run migrations up
		err = m.Up()
		require.NoError(t, err)

		// Get current version
		version, _, err := m.Version()
		require.NoError(t, err)

		// Roll back one step
		err = m.Steps(-1)
		require.NoError(t, err)

		// Verify version decreased
		newVersion, _, err := m.Version()
		require.NoError(t, err)
		assert.Less(t, newVersion, version)

		// Verify some tables still exist
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = 'users'
			)
		`).Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists, "Users table should still exist after partial rollback")
	})
}
