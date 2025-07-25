package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"go-fiber-template/internal/config"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// MigrationManager handles database migrations
type MigrationManager struct {
	migrate        *migrate.Migrate
	db             *sql.DB
	migrationsPath string
}

// NewMigrationManager creates a new migration manager
func NewMigrationManager(cfg config.DatabaseConfig) (*MigrationManager, error) {
	// Connect to database using sql.DB for migrations
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create postgres driver instance
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// Set migrations path
	migrationsPath := "file://migrations"

	// Create migrate instance
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	return &MigrationManager{
		migrate:        m,
		db:             db,
		migrationsPath: "migrations",
	}, nil
}

// Up runs all pending migrations
func (mm *MigrationManager) UpAll() error {
	log.Println("Running all pending migrations...")

	err := mm.migrate.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No pending migrations found")
		return nil
	}

	log.Println("All migrations completed successfully")
	return nil
}

// Up runs a specific number of migrations
func (mm *MigrationManager) Up(steps int) error {
	log.Printf("Running %d migration(s)...", steps)

	err := mm.migrate.Steps(steps)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run %d migration(s): %w", steps, err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No pending migrations found")
		return nil
	}

	log.Printf("%d migration(s) completed successfully", steps)
	return nil
}

// Down rolls back all migrations
func (mm *MigrationManager) DownAll() error {
	log.Println("Rolling back all migrations...")

	err := mm.migrate.Down()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to rollback")
		return nil
	}

	log.Println("All migrations rolled back successfully")
	return nil
}

// Down rolls back a specific number of migrations
func (mm *MigrationManager) Down(steps int) error {
	log.Printf("Rolling back %d migration(s)...", steps)

	err := mm.migrate.Steps(-steps)
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to rollback %d migration(s): %w", steps, err)
	}

	if err == migrate.ErrNoChange {
		log.Println("No migrations to rollback")
		return nil
	}

	log.Printf("%d migration(s) rolled back successfully", steps)
	return nil
}

// Create creates a new migration file
func (mm *MigrationManager) Create(name string) error {
	if name == "" {
		return fmt.Errorf("migration name cannot be empty")
	}

	// Ensure migrations directory exists
	if err := os.MkdirAll(mm.migrationsPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Generate timestamp
	timestamp := time.Now().Format("20060102150405")

	// Create up migration file
	upFile := filepath.Join(mm.migrationsPath, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	downFile := filepath.Join(mm.migrationsPath, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

	// Create up migration file
	upContent := fmt.Sprintf("-- Migration: %s\n-- Created at: %s\n\n-- Add your up migration SQL here\n", name, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(upFile, []byte(upContent), 0644); err != nil {
		return fmt.Errorf("failed to create up migration file: %w", err)
	}

	// Create down migration file
	downContent := fmt.Sprintf("-- Migration: %s (rollback)\n-- Created at: %s\n\n-- Add your down migration SQL here\n", name, time.Now().Format("2006-01-02 15:04:05"))
	if err := os.WriteFile(downFile, []byte(downContent), 0644); err != nil {
		return fmt.Errorf("failed to create down migration file: %w", err)
	}

	log.Printf("Migration files created:")
	log.Printf("  Up:   %s", upFile)
	log.Printf("  Down: %s", downFile)

	return nil
}

// Status shows the current migration status
func (mm *MigrationManager) Status() error {
	version, dirty, err := mm.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("Migration Status:")
		log.Println("  Current Version: No migrations applied")
		log.Println("  Dirty: false")
	} else {
		log.Println("Migration Status:")
		log.Printf("  Current Version: %d", version)
		log.Printf("  Dirty: %t", dirty)
	}

	// List migration files
	files, err := filepath.Glob(filepath.Join(mm.migrationsPath, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to list migration files: %w", err)
	}

	log.Printf("  Available Migrations: %d", len(files))
	for _, file := range files {
		log.Printf("    - %s", filepath.Base(file))
	}

	return nil
}

// Version shows the current migration version
func (mm *MigrationManager) Version() error {
	version, dirty, err := mm.migrate.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	if err == migrate.ErrNilVersion {
		log.Println("Current Version: No migrations applied")
	} else {
		log.Printf("Current Version: %d", version)
		if dirty {
			log.Println("Warning: Database is in dirty state")
		}
	}

	return nil
}

// Close closes the migration manager and database connection
func (mm *MigrationManager) Close() error {
	var errs []error

	if mm.migrate != nil {
		if sourceErr, dbErr := mm.migrate.Close(); sourceErr != nil || dbErr != nil {
			if sourceErr != nil {
				errs = append(errs, fmt.Errorf("failed to close migrate source: %w", sourceErr))
			}
			if dbErr != nil {
				errs = append(errs, fmt.Errorf("failed to close migrate database: %w", dbErr))
			}
		}
	}

	if mm.db != nil {
		if err := mm.db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing migration manager: %v", errs)
	}

	return nil
}
