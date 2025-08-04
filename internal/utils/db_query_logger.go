package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm/logger"
)

// QueryLogLevel defines the logging level for database queries
type QueryLogLevel int

const (
	QueryLogSilent QueryLogLevel = iota
	QueryLogError
	QueryLogWarn
	QueryLogInfo
	QueryLogDebug
)

// QueryLogEntry represents a single database query log entry
type QueryLogEntry struct {
	ID           string        `json:"id"`
	Query        string        `json:"query"`
	Args         []interface{} `json:"args,omitempty"`
	Duration     time.Duration `json:"duration"`
	RowsAffected int64         `json:"rows_affected"`
	Error        string        `json:"error,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
	Source       string        `json:"source,omitempty"`
	UserID       *uint         `json:"user_id,omitempty"`
	RequestID    string        `json:"request_id,omitempty"`
	Level        string        `json:"level"`
}

// QueryLogger interface for database query logging
type QueryLogger interface {
	LogQuery(entry QueryLogEntry)
	GetRecentQueries(limit int) []QueryLogEntry
	GetSlowQueries(threshold time.Duration, limit int) []QueryLogEntry
	GetQueriesByPattern(pattern string, limit int) []QueryLogEntry
	GetQueryStats() QueryStats
	ClearLogs()
}

// InMemoryQueryLogger implements QueryLogger with in-memory storage
type InMemoryQueryLogger struct {
	mu      sync.RWMutex
	entries []QueryLogEntry
	maxSize int
}

// NewInMemoryQueryLogger creates a new in-memory query logger
func NewInMemoryQueryLogger(maxSize int) *InMemoryQueryLogger {
	if maxSize <= 0 {
		maxSize = 1000 // Default max size
	}

	return &InMemoryQueryLogger{
		entries: make([]QueryLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// LogQuery logs a database query
func (l *InMemoryQueryLogger) LogQuery(entry QueryLogEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Add timestamp if not set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Add unique ID if not set
	if entry.ID == "" {
		entry.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Add to entries
	l.entries = append(l.entries, entry)

	// Maintain max size by removing oldest entries
	if len(l.entries) > l.maxSize {
		l.entries = l.entries[len(l.entries)-l.maxSize:]
	}
}

// GetRecentQueries returns the most recent queries
func (l *InMemoryQueryLogger) GetRecentQueries(limit int) []QueryLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if limit <= 0 || limit > len(l.entries) {
		limit = len(l.entries)
	}

	// Return the last 'limit' entries
	start := len(l.entries) - limit
	if start < 0 {
		start = 0
	}

	result := make([]QueryLogEntry, limit)
	copy(result, l.entries[start:])

	// Reverse to get most recent first
	for i := 0; i < len(result)/2; i++ {
		result[i], result[len(result)-1-i] = result[len(result)-1-i], result[i]
	}

	return result
}

// GetSlowQueries returns queries that took longer than the threshold
func (l *InMemoryQueryLogger) GetSlowQueries(threshold time.Duration, limit int) []QueryLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var slowQueries []QueryLogEntry

	for _, entry := range l.entries {
		if entry.Duration >= threshold {
			slowQueries = append(slowQueries, entry)
		}

		if len(slowQueries) >= limit {
			break
		}
	}

	return slowQueries
}

// GetQueriesByPattern returns queries matching a pattern
func (l *InMemoryQueryLogger) GetQueriesByPattern(pattern string, limit int) []QueryLogEntry {
	l.mu.RLock()
	defer l.mu.RUnlock()

	var matchingQueries []QueryLogEntry
	pattern = strings.ToLower(pattern)

	for _, entry := range l.entries {
		if strings.Contains(strings.ToLower(entry.Query), pattern) {
			matchingQueries = append(matchingQueries, entry)
		}

		if len(matchingQueries) >= limit {
			break
		}
	}

	return matchingQueries
}

// GetQueryStats returns aggregated query statistics
func (l *InMemoryQueryLogger) GetQueryStats() QueryStats {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.entries) == 0 {
		return QueryStats{}
	}

	var totalDuration time.Duration
	var errorCount int64
	var slowQueryCount int64
	slowThreshold := 100 * time.Millisecond

	queryTypes := make(map[string]int64)

	for _, entry := range l.entries {
		totalDuration += entry.Duration

		if entry.Error != "" {
			errorCount++
		}

		if entry.Duration >= slowThreshold {
			slowQueryCount++
		}

		// Categorize query type
		queryType := getQueryType(entry.Query)
		queryTypes[queryType]++
	}

	avgDuration := totalDuration / time.Duration(len(l.entries))

	return QueryStats{
		TotalQueries:    int64(len(l.entries)),
		AverageDuration: avgDuration,
		TotalDuration:   totalDuration,
		ErrorCount:      errorCount,
		SlowQueryCount:  slowQueryCount,
		QueryTypes:      queryTypes,
		Timestamp:       time.Now(),
	}
}

// ClearLogs clears all logged queries
func (l *InMemoryQueryLogger) ClearLogs() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.entries = l.entries[:0]
}

// QueryStats represents aggregated query statistics
type QueryStats struct {
	TotalQueries    int64            `json:"total_queries"`
	AverageDuration time.Duration    `json:"average_duration"`
	TotalDuration   time.Duration    `json:"total_duration"`
	ErrorCount      int64            `json:"error_count"`
	SlowQueryCount  int64            `json:"slow_query_count"`
	QueryTypes      map[string]int64 `json:"query_types"`
	Timestamp       time.Time        `json:"timestamp"`
}

// GormQueryLogger implements GORM's logger interface with enhanced logging
type GormQueryLogger struct {
	queryLogger   QueryLogger
	level         QueryLogLevel
	slowThreshold time.Duration
}

// NewGormQueryLogger creates a new GORM query logger
func NewGormQueryLogger(queryLogger QueryLogger, level QueryLogLevel, slowThreshold time.Duration) *GormQueryLogger {
	if slowThreshold <= 0 {
		slowThreshold = 200 * time.Millisecond
	}

	return &GormQueryLogger{
		queryLogger:   queryLogger,
		level:         level,
		slowThreshold: slowThreshold,
	}
}

// LogMode sets the log level
func (l *GormQueryLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	switch level {
	case logger.Silent:
		newLogger.level = QueryLogSilent
	case logger.Error:
		newLogger.level = QueryLogError
	case logger.Warn:
		newLogger.level = QueryLogWarn
	case logger.Info:
		newLogger.level = QueryLogInfo
	default:
		newLogger.level = QueryLogDebug
	}
	return &newLogger
}

// Info logs info messages
func (l *GormQueryLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= QueryLogInfo {
		log.Printf("[INFO] "+msg, data...)
	}
}

// Warn logs warning messages
func (l *GormQueryLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= QueryLogWarn {
		log.Printf("[WARN] "+msg, data...)
	}
}

// Error logs error messages
func (l *GormQueryLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= QueryLogError {
		log.Printf("[ERROR] "+msg, data...)
	}
}

// Trace logs SQL queries
func (l *GormQueryLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= QueryLogSilent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	entry := QueryLogEntry{
		Query:        sql,
		Duration:     elapsed,
		RowsAffected: rows,
		Timestamp:    begin,
	}

	// Extract context values if available
	if requestID := ctx.Value("request_id"); requestID != nil {
		if id, ok := requestID.(string); ok {
			entry.RequestID = id
		}
	}

	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(uint); ok {
			entry.UserID = &id
		}
	}

	// Set error if present
	if err != nil {
		entry.Error = err.Error()
		entry.Level = "error"
	} else if elapsed >= l.slowThreshold {
		entry.Level = "warn" // Slow query
	} else {
		entry.Level = "info"
	}

	// Log the query
	l.queryLogger.LogQuery(entry)

	// Also log to standard logger based on level
	switch {
	case err != nil && l.level >= QueryLogError:
		log.Printf("[ERROR] SQL Error: %v | Duration: %v | SQL: %s", err, elapsed, sql)
	case elapsed >= l.slowThreshold && l.level >= QueryLogWarn:
		log.Printf("[WARN] Slow Query: Duration: %v | Rows: %d | SQL: %s", elapsed, rows, sql)
	case l.level >= QueryLogInfo:
		log.Printf("[INFO] SQL Query: Duration: %v | Rows: %d | SQL: %s", elapsed, rows, sql)
	}
}

// Helper functions

// getQueryType extracts the query type from SQL
func getQueryType(sql string) string {
	sql = strings.TrimSpace(strings.ToUpper(sql))

	if strings.HasPrefix(sql, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(sql, "INSERT") {
		return "INSERT"
	} else if strings.HasPrefix(sql, "UPDATE") {
		return "UPDATE"
	} else if strings.HasPrefix(sql, "DELETE") {
		return "DELETE"
	} else if strings.HasPrefix(sql, "CREATE") {
		return "CREATE"
	} else if strings.HasPrefix(sql, "DROP") {
		return "DROP"
	} else if strings.HasPrefix(sql, "ALTER") {
		return "ALTER"
	}

	return "OTHER"
}

// QueryAnalyzer provides query analysis capabilities
type QueryAnalyzer struct {
	logger QueryLogger
}

// NewQueryAnalyzer creates a new query analyzer
func NewQueryAnalyzer(logger QueryLogger) *QueryAnalyzer {
	return &QueryAnalyzer{
		logger: logger,
	}
}

// AnalyzePerformance analyzes query performance patterns
func (qa *QueryAnalyzer) AnalyzePerformance() PerformanceAnalysis {
	stats := qa.logger.GetQueryStats()
	slowQueries := qa.logger.GetSlowQueries(100*time.Millisecond, 10)

	analysis := PerformanceAnalysis{
		Stats:       stats,
		SlowQueries: slowQueries,
		Timestamp:   time.Now(),
	}

	// Analyze patterns
	analysis.Recommendations = qa.generateRecommendations(stats, slowQueries)

	return analysis
}

// PerformanceAnalysis represents query performance analysis results
type PerformanceAnalysis struct {
	Stats           QueryStats      `json:"stats"`
	SlowQueries     []QueryLogEntry `json:"slow_queries"`
	Recommendations []string        `json:"recommendations"`
	Timestamp       time.Time       `json:"timestamp"`
}

// generateRecommendations generates performance recommendations
func (qa *QueryAnalyzer) generateRecommendations(stats QueryStats, slowQueries []QueryLogEntry) []string {
	var recommendations []string

	// Check error rate
	if stats.TotalQueries > 0 {
		errorRate := float64(stats.ErrorCount) / float64(stats.TotalQueries)
		if errorRate > 0.05 { // More than 5% error rate
			recommendations = append(recommendations,
				fmt.Sprintf("High error rate detected (%.2f%%). Review query logic and error handling.", errorRate*100))
		}
	}

	// Check slow query rate
	if stats.TotalQueries > 0 {
		slowRate := float64(stats.SlowQueryCount) / float64(stats.TotalQueries)
		if slowRate > 0.1 { // More than 10% slow queries
			recommendations = append(recommendations,
				fmt.Sprintf("High slow query rate detected (%.2f%%). Consider adding indexes or optimizing queries.", slowRate*100))
		}
	}

	// Check average duration
	if stats.AverageDuration > 500*time.Millisecond {
		recommendations = append(recommendations,
			fmt.Sprintf("Average query duration is high (%v). Consider query optimization.", stats.AverageDuration))
	}

	// Analyze slow queries for patterns
	if len(slowQueries) > 0 {
		tablePatterns := make(map[string]int)
		for _, query := range slowQueries {
			tables := extractTableNames(query.Query)
			for _, table := range tables {
				tablePatterns[table]++
			}
		}

		for table, count := range tablePatterns {
			if count > 1 {
				recommendations = append(recommendations,
					fmt.Sprintf("Table '%s' appears in %d slow queries. Consider adding indexes.", table, count))
			}
		}
	}

	return recommendations
}

// extractTableNames extracts table names from SQL query (simplified)
func extractTableNames(sql string) []string {
	var tables []string
	sql = strings.ToUpper(sql)

	// Simple pattern matching for FROM and JOIN clauses
	patterns := []string{"FROM ", "JOIN ", "UPDATE ", "INSERT INTO "}

	for _, pattern := range patterns {
		if idx := strings.Index(sql, pattern); idx != -1 {
			remaining := sql[idx+len(pattern):]
			words := strings.Fields(remaining)
			if len(words) > 0 {
				// Remove quotes and get table name
				tableName := strings.Trim(words[0], "`\"'")
				if !contains(tables, tableName) {
					tables = append(tables, tableName)
				}
			}
		}
	}

	return tables
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// QueryLoggerMiddleware creates a middleware for request-scoped query logging
func QueryLoggerMiddleware(logger QueryLogger) func(ctx context.Context, requestID string, userID *uint) context.Context {
	return func(ctx context.Context, requestID string, userID *uint) context.Context {
		ctx = context.WithValue(ctx, "request_id", requestID)
		if userID != nil {
			ctx = context.WithValue(ctx, "user_id", *userID)
		}
		return ctx
	}
}

// ExportQueryLogs exports query logs to JSON
func ExportQueryLogs(logger QueryLogger, limit int) ([]byte, error) {
	queries := logger.GetRecentQueries(limit)
	return json.MarshalIndent(queries, "", "  ")
}

// ImportQueryLogs imports query logs from JSON (for testing/analysis)
func ImportQueryLogs(logger *InMemoryQueryLogger, data []byte) error {
	var entries []QueryLogEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	for _, entry := range entries {
		logger.LogQuery(entry)
	}

	return nil
}
