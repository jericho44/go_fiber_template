package database

import (
	"context"
	"time"
)

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status       string                 `json:"status"`
	Message      string                 `json:"message"`
	Timestamp    time.Time              `json:"timestamp"`
	Stats        map[string]interface{} `json:"stats,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
}

// CheckHealth performs a comprehensive health check on the database
func CheckHealth(ctx context.Context) *HealthStatus {
	start := time.Now()

	health := &HealthStatus{
		Timestamp: start,
	}

	// Check if database connection exists
	if DB == nil {
		health.Status = "unhealthy"
		health.Message = "Database connection not initialized"
		health.ResponseTime = time.Since(start)
		return health
	}

	// Perform ping with context timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sqlDB, err := DB.DB()
	if err != nil {
		health.Status = "unhealthy"
		health.Message = "Failed to get underlying database connection: " + err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}

	// Ping the database
	if err := sqlDB.PingContext(ctx); err != nil {
		health.Status = "unhealthy"
		health.Message = "Database ping failed: " + err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}

	// Get connection statistics
	stats, err := GetConnectionStats()
	if err != nil {
		health.Status = "degraded"
		health.Message = "Database is reachable but stats unavailable: " + err.Error()
		health.ResponseTime = time.Since(start)
		return health
	}

	health.Status = "healthy"
	health.Message = "Database is healthy and responsive"
	health.Stats = stats
	health.ResponseTime = time.Since(start)

	return health
}

// IsHealthy returns true if the database is healthy
func IsHealthy(ctx context.Context) bool {
	health := CheckHealth(ctx)
	return health.Status == "healthy"
}
