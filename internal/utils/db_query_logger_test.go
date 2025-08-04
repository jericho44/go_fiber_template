package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewInMemoryQueryLogger(t *testing.T) {
	logger := NewInMemoryQueryLogger(100)

	assert.NotNil(t, logger)
	assert.Equal(t, 100, logger.maxSize)
	assert.Empty(t, logger.entries)
}

func TestInMemoryQueryLogger_LogQuery(t *testing.T) {
	logger := NewInMemoryQueryLogger(10)

	entry := QueryLogEntry{
		Query:     "SELECT * FROM users WHERE id = ?",
		Duration:  100 * time.Millisecond,
		Timestamp: time.Now(),
	}

	logger.LogQuery(entry)

	assert.Len(t, logger.entries, 1)
	assert.NotEmpty(t, logger.entries[0].ID)
	assert.Equal(t, entry.Query, logger.entries[0].Query)
}

func TestInMemoryQueryLogger_GetRecentQueries(t *testing.T) {
	logger := NewInMemoryQueryLogger(10)

	// Add some test queries
	for i := 0; i < 5; i++ {
		entry := QueryLogEntry{
			Query:     "SELECT * FROM users",
			Duration:  time.Duration(i*10) * time.Millisecond,
			Timestamp: time.Now().Add(time.Duration(i) * time.Second),
		}
		logger.LogQuery(entry)
	}

	recent := logger.GetRecentQueries(3)
	assert.Len(t, recent, 3)

	// Should be in reverse chronological order (most recent first)
	assert.True(t, recent[0].Duration >= recent[1].Duration)
}

func TestInMemoryQueryLogger_GetSlowQueries(t *testing.T) {
	logger := NewInMemoryQueryLogger(10)

	// Add fast and slow queries
	fastQuery := QueryLogEntry{
		Query:     "SELECT id FROM users",
		Duration:  50 * time.Millisecond,
		Timestamp: time.Now(),
	}

	slowQuery := QueryLogEntry{
		Query:     "SELECT * FROM users JOIN orders ON users.id = orders.user_id",
		Duration:  300 * time.Millisecond,
		Timestamp: time.Now(),
	}

	logger.LogQuery(fastQuery)
	logger.LogQuery(slowQuery)

	slowQueries := logger.GetSlowQueries(200*time.Millisecond, 10)
	assert.Len(t, slowQueries, 1)
	assert.Equal(t, slowQuery.Query, slowQueries[0].Query)
}

func TestInMemoryQueryLogger_GetQueryStats(t *testing.T) {
	logger := NewInMemoryQueryLogger(10)

	// Add some test queries
	queries := []QueryLogEntry{
		{Query: "SELECT * FROM users", Duration: 100 * time.Millisecond, Error: ""},
		{Query: "INSERT INTO users VALUES (?)", Duration: 50 * time.Millisecond, Error: ""},
		{Query: "UPDATE users SET name = ?", Duration: 200 * time.Millisecond, Error: "connection error"},
	}

	for _, query := range queries {
		logger.LogQuery(query)
	}

	stats := logger.GetQueryStats()

	assert.Equal(t, int64(3), stats.TotalQueries)
	assert.Equal(t, int64(1), stats.ErrorCount)
	assert.Equal(t, int64(2), stats.SlowQueryCount) // 100ms and 200ms queries are slow (threshold is 100ms)
	assert.True(t, stats.AverageDuration > 0)
	assert.Equal(t, int64(1), stats.QueryTypes["SELECT"])
	assert.Equal(t, int64(1), stats.QueryTypes["INSERT"])
	assert.Equal(t, int64(1), stats.QueryTypes["UPDATE"])
}

func TestInMemoryQueryLogger_MaxSize(t *testing.T) {
	logger := NewInMemoryQueryLogger(3) // Small max size

	// Add more queries than max size
	for i := 0; i < 5; i++ {
		entry := QueryLogEntry{
			Query:     "SELECT * FROM users",
			Duration:  100 * time.Millisecond,
			Timestamp: time.Now(),
		}
		logger.LogQuery(entry)
	}

	// Should only keep the last 3 entries
	assert.Len(t, logger.entries, 3)
}

func TestGetQueryType(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"INSERT INTO users VALUES (?)", "INSERT"},
		{"UPDATE users SET name = ?", "UPDATE"},
		{"DELETE FROM users WHERE id = ?", "DELETE"},
		{"CREATE TABLE test (id INT)", "CREATE"},
		{"DROP TABLE test", "DROP"},
		{"ALTER TABLE users ADD COLUMN age INT", "ALTER"},
		{"EXPLAIN SELECT * FROM users", "OTHER"},
	}

	for _, test := range tests {
		result := getQueryType(test.query)
		assert.Equal(t, test.expected, result, "Query: %s", test.query)
	}
}
