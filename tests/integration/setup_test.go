package integration

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"
	"go-fiber-template/internal/models"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/gorm"
)

// TestContainer holds the test database container and connection
type TestContainer struct {
	Container testcontainers.Container
	DB        *gorm.DB
	Config    *config.DatabaseConfig
}

// SetupTestDatabase creates a PostgreSQL test container and returns the database connection
func SetupTestDatabase(t *testing.T) *TestContainer {
	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgrescontainer.Run(ctx,
		"postgres:15-alpine",
		postgrescontainer.WithDatabase("testdb"),
		postgrescontainer.WithUsername("testuser"),
		postgrescontainer.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %v", err)
	}

	// Get connection details
	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get container host: %v", err)
	}

	port, err := postgresContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("Failed to get container port: %v", err)
	}

	// Create database config
	dbConfig := &config.DatabaseConfig{
		Host:            host,
		Port:            port.Int(),
		User:            "testuser",
		Password:        "testpass",
		Name:            "testdb",
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	// Connect to database
	db, err := database.Connect(dbConfig)
	if err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(dbConfig); err != nil {
		postgresContainer.Terminate(ctx)
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return &TestContainer{
		Container: postgresContainer,
		DB:        db,
		Config:    dbConfig,
	}
}

// TeardownTestDatabase cleans up the test database container
func (tc *TestContainer) TeardownTestDatabase(t *testing.T) {
	ctx := context.Background()

	// Close database connection
	if tc.DB != nil {
		sqlDB, err := tc.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Terminate container
	if tc.Container != nil {
		if err := tc.Container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}
}

// CleanupTestData removes all test data from the database
func (tc *TestContainer) CleanupTestData() error {
	// Clean up in reverse order of dependencies
	tables := []string{
		"token_blacklists",
		"users",
	}

	for _, table := range tables {
		if err := tc.DB.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error; err != nil {
			return fmt.Errorf("failed to clean table %s: %w", table, err)
		}
	}

	return nil
}

// runMigrations runs database migrations
func runMigrations(dbConfig *config.DatabaseConfig) error {
	// Create connection string
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		dbConfig.User,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Name,
		dbConfig.SSLMode,
	)

	// Open database connection for migrations
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database for migrations: %w", err)
	}
	defer db.Close()

	// Create migration driver
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create migration instance
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s/../../migrations", wd),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// CreateTestUser creates a test user in the database
func (tc *TestContainer) CreateTestUser(email, firstName, lastName string) (*models.User, error) {
	user := &models.User{
		Email:     email,
		Password:  "$2a$10$test.hash.for.testing.purposes.only", // Pre-hashed test password
		FirstName: firstName,
		LastName:  lastName,
		IsActive:  true,
	}

	if err := tc.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

// TestMain sets up and tears down test environment
func TestMain(m *testing.M) {
	// Check if Docker is available
	if !isDockerAvailable() {
		log.Println("Docker is not available, skipping integration tests")
		os.Exit(0)
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
}

// isDockerAvailable checks if Docker is available for running containers
func isDockerAvailable() bool {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image: "hello-world",
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          false,
	})

	if err != nil {
		return false
	}

	if container != nil {
		container.Terminate(ctx)
	}

	return true
}
