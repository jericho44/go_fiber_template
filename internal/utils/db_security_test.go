package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test models for security audit tests
type TestUserForAudit struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique;not null"`
	Password string `gorm:"not null"`
	SSN      string // Sensitive field without encryption
}

func setupSecurityTestDB(t *testing.T) (*gorm.DB, *InMemoryQueryLogger) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&TestUserForAudit{})
	require.NoError(t, err)

	// Create query logger
	queryLogger := NewInMemoryQueryLogger(100)

	// Seed some test data
	users := []TestUserForAudit{
		{Email: "user1@example.com", Password: "hashedpassword123", SSN: "123-45-6789"},
		{Email: "user2@example.com", Password: "password", SSN: "987-65-4321"}, // Weak password
		{Email: "user3@example.com", Password: "123456", SSN: "555-55-5555"},   // Weak password
	}
	err = db.Create(&users).Error
	require.NoError(t, err)

	return db, queryLogger
}

func TestNewDatabaseSecurityAuditor(t *testing.T) {
	db, queryLogger := setupSecurityTestDB(t)
	auditor := NewDatabaseSecurityAuditor(db, queryLogger)

	assert.NotNil(t, auditor)
	assert.Equal(t, db, auditor.db)
	assert.Equal(t, queryLogger, auditor.queryLogger)
	assert.Empty(t, auditor.findings)
}

func TestSecurityAuditLevel_String(t *testing.T) {
	tests := []struct {
		level    SecurityAuditLevel
		expected string
	}{
		{SecurityAuditInfo, "INFO"},
		{SecurityAuditWarning, "WARNING"},
		{SecurityAuditCritical, "CRITICAL"},
		{SecurityAuditLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		result := test.level.String()
		assert.Equal(t, test.expected, result)
	}
}

func TestDatabaseSecurityAuditor_AuditQuery(t *testing.T) {
	db, queryLogger := setupSecurityTestDB(t)
	auditor := NewDatabaseSecurityAuditor(db, queryLogger)

	tests := []struct {
		name          string
		query         string
		expectedLevel SecurityAuditLevel
		shouldFind    bool
	}{
		{
			name:       "Safe parameterized query",
			query:      "SELECT * FROM users WHERE id = ?",
			shouldFind: false,
		},
		{
			name:          "DELETE without WHERE",
			query:         "DELETE FROM users",
			expectedLevel: SecurityAuditCritical,
			shouldFind:    true,
		},
		{
			name:          "UPDATE without WHERE",
			query:         "UPDATE users SET email = 'test@example.com'",
			expectedLevel: SecurityAuditWarning,
			shouldFind:    true,
		},
		{
			name:          "SELECT * query",
			query:         "SELECT * FROM users",
			expectedLevel: SecurityAuditInfo,
			shouldFind:    true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Reset findings
			auditor.findings = make([]SecurityFinding, 0)

			queryEntry := QueryLogEntry{
				Query:     test.query,
				Duration:  100 * time.Millisecond,
				Timestamp: time.Now(),
			}

			err := auditor.auditQuery(queryEntry)
			assert.NoError(t, err)

			if test.shouldFind {
				assert.NotEmpty(t, auditor.findings, "Expected to find security issue for query: %s", test.query)
				if len(auditor.findings) > 0 {
					assert.Equal(t, test.expectedLevel, auditor.findings[0].Level)
				}
			} else {
				assert.Empty(t, auditor.findings, "Did not expect to find security issue for query: %s", test.query)
			}
		})
	}
}

func TestDatabaseSecurityAuditor_GenerateReport(t *testing.T) {
	db, queryLogger := setupSecurityTestDB(t)
	auditor := NewDatabaseSecurityAuditor(db, queryLogger)

	// Add some test findings
	auditor.addFinding(SecurityFinding{
		Level:       SecurityAuditCritical,
		Category:    "Test",
		Title:       "Critical Issue",
		Description: "Critical test issue",
	})

	auditor.addFinding(SecurityFinding{
		Level:       SecurityAuditWarning,
		Category:    "Test",
		Title:       "Warning Issue",
		Description: "Warning test issue",
	})

	report := auditor.generateReport()

	assert.NotNil(t, report)
	assert.NotEmpty(t, report.ID)
	assert.False(t, report.Timestamp.IsZero())
	assert.Equal(t, 2, report.TotalFindings)
	assert.Equal(t, 1, report.CriticalFindings)
	assert.Equal(t, 1, report.WarningFindings)
	assert.Len(t, report.Findings, 2)
	assert.Contains(t, report.Summary, "CRITICAL")
}

func TestSecurityTestSuite_TestSQLInjectionResistance(t *testing.T) {
	db, queryLogger := setupSecurityTestDB(t)
	auditor := NewDatabaseSecurityAuditor(db, queryLogger)
	testSuite := NewSecurityTestSuite(db, auditor)

	ctx := context.Background()
	findings := testSuite.TestSQLInjectionResistance(ctx)

	// With proper parameterized queries, we shouldn't find any injection vulnerabilities
	assert.Empty(t, findings, "Query builder should be resistant to SQL injection")
}
