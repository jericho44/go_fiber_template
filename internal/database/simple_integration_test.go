package database

import (
	"context"
	"testing"
	"time"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
)

// TestDatabaseImplementationExists verifies that all required database functions exist
func TestDatabaseImplementationExists(t *testing.T) {
	// Test that all required functions exist and can be called

	// Test DefaultConnectionPoolConfig
	poolCfg := DefaultConnectionPoolConfig()
	assert.NotNil(t, poolCfg)
	assert.Equal(t, 10, poolCfg.MaxIdleConns)
	assert.Equal(t, 100, poolCfg.MaxOpenConns)
	assert.Equal(t, time.Hour, poolCfg.ConnMaxLifetime)
	assert.Equal(t, 30*time.Minute, poolCfg.ConnMaxIdleTime)
}

func TestDatabaseConfigValidation(t *testing.T) {
	// Test database configuration validation
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "testuser",
		Password:        "testpass",
		Name:            "testdb",
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	// Test DSN generation
	expectedDSN := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	actualDSN := cfg.GetDSN()
	assert.Equal(t, expectedDSN, actualDSN)

	// Test validation
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestHealthCheckWithoutConnection(t *testing.T) {
	// Store original DB to restore later
	originalDB := DB
	defer func() { DB = originalDB }()

	// Test health check without connection
	DB = nil
	err := HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is not initialized")

	// Test context-based health check
	ctx := context.Background()
	health := CheckHealth(ctx)
	assert.Equal(t, "unhealthy", health.Status)
	assert.Contains(t, health.Message, "Database connection not initialized")
	assert.True(t, health.ResponseTime >= 0)

	// Test IsHealthy
	healthy := IsHealthy(ctx)
	assert.False(t, healthy)
}

func TestConnectionStatsWithoutConnection(t *testing.T) {
	// Store original DB to restore later
	originalDB := DB
	defer func() { DB = originalDB }()

	// Test connection stats without connection
	DB = nil
	stats, err := GetConnectionStats()
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "database connection is not initialized")
}

func TestGetDBFunction(t *testing.T) {
	// Store original DB to restore later
	originalDB := DB
	defer func() { DB = originalDB }()

	// Test GetDB without connection
	DB = nil
	db := GetDB()
	assert.Nil(t, db)
}

func TestCloseFunction(t *testing.T) {
	// Store original DB to restore later
	originalDB := DB
	defer func() { DB = originalDB }()

	// Test Close without connection - should not error
	DB = nil
	err := Close()
	assert.NoError(t, err)
}

func TestConnectWithInvalidConfig(t *testing.T) {
	// Test connection with invalid configuration
	invalidCfg := &config.DatabaseConfig{
		Host:            "invalid-host-that-does-not-exist",
		Port:            5432,
		User:            "test",
		Password:        "test",
		Name:            "test",
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	// This should fail to connect
	_, err := Connect(invalidCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestInitializeWithInvalidConfig(t *testing.T) {
	// Test Initialize with invalid configuration
	invalidCfg := &config.DatabaseConfig{
		Host:            "invalid-host-that-does-not-exist",
		Port:            5432,
		User:            "test",
		Password:        "test",
		Name:            "test",
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	// This should fail to initialize
	err := Initialize(invalidCfg)
	assert.Error(t, err)
}

func TestConnectionPoolConfig(t *testing.T) {
	// Test connection pool configuration structure
	poolCfg := &ConnectionPoolConfig{
		MaxIdleConns:    5,
		MaxOpenConns:    20,
		ConnMaxLifetime: 45 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
	}

	assert.Equal(t, 5, poolCfg.MaxIdleConns)
	assert.Equal(t, 20, poolCfg.MaxOpenConns)
	assert.Equal(t, 45*time.Minute, poolCfg.ConnMaxLifetime)
	assert.Equal(t, 10*time.Minute, poolCfg.ConnMaxIdleTime)
}
