package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCheckHealth(t *testing.T) {
	// Test with no database connection
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	ctx := context.Background()

	health := CheckHealth(ctx)
	assert.Equal(t, "unhealthy", health.Status)
	assert.Contains(t, health.Message, "Database connection not initialized")
	assert.True(t, health.ResponseTime >= 0, "ResponseTime should be non-negative, got: %v", health.ResponseTime)
	// Check that timestamp is recent (within last 5 seconds)
	timeDiff := time.Since(health.Timestamp)
	assert.True(t, timeDiff < 5*time.Second, "Timestamp should be recent, got diff: %v", timeDiff)
}

func TestIsHealthy(t *testing.T) {
	// Test with no database connection
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	ctx := context.Background()

	healthy := IsHealthy(ctx)
	assert.False(t, healthy)
}

func TestCheckHealth_WithTimeout(t *testing.T) {
	// Test with a very short timeout context
	originalDB := DB
	defer func() { DB = originalDB }() // Restore original DB after test

	DB = nil
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	health := CheckHealth(ctx)
	assert.Equal(t, "unhealthy", health.Status)
	assert.Contains(t, health.Message, "Database connection not initialized")
}
