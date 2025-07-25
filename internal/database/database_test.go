package database

import (
	"testing"

	"go-fiber-template/internal/config"

	"github.com/stretchr/testify/assert"
)

func TestConnect(t *testing.T) {
	// Test with invalid DSN - should fail to connect
	invalidCfg := &config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     5432,
		User:     "test",
		Password: "test",
		Name:     "test",
		SSLMode:  "disable",
	}

	_, err := Connect(invalidCfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestHealthCheck(t *testing.T) {
	// Test with no database connection
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	err := HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is not initialized")
}

func TestGetDB(t *testing.T) {
	// Test with no database connection
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	db := GetDB()
	assert.Nil(t, db)
}

func TestGetConnectionStats(t *testing.T) {
	// Test with no database connection
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	stats, err := GetConnectionStats()
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.Contains(t, err.Error(), "database connection is not initialized")
}

func TestClose(t *testing.T) {
	// Test with no database connection - should not error
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	err := Close()
	assert.NoError(t, err)
}

func TestInitialize(t *testing.T) {
	// Test with invalid config - should fail to connect
	invalidCfg := &config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     5432,
		User:     "test",
		Password: "test",
		Name:     "test",
		SSLMode:  "disable",
	}

	err := Initialize(invalidCfg)
	assert.Error(t, err)
}

func TestDatabaseConfig_GetDSN(t *testing.T) {
	// Test DSN generation
	cfg := &config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "testuser",
		Password: "testpass",
		Name:     "testdb",
		SSLMode:  "disable",
	}

	expectedDSN := "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable"
	actualDSN := cfg.GetDSN()
	assert.Equal(t, expectedDSN, actualDSN)
}
