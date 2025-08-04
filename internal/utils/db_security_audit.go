package utils

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SecurityAuditLevel defines the severity level of security findings
type SecurityAuditLevel int

const (
	SecurityAuditInfo SecurityAuditLevel = iota
	SecurityAuditWarning
	SecurityAuditCritical
)

func (s SecurityAuditLevel) String() string {
	switch s {
	case SecurityAuditInfo:
		return "INFO"
	case SecurityAuditWarning:
		return "WARNING"
	case SecurityAuditCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// SecurityFinding represents a security audit finding
type SecurityFinding struct {
	ID          string             `json:"id"`
	Level       SecurityAuditLevel `json:"level"`
	Category    string             `json:"category"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Query       string             `json:"query,omitempty"`
	Table       string             `json:"table,omitempty"`
	Column      string             `json:"column,omitempty"`
	Risk        string             `json:"risk"`
	Remediation string             `json:"remediation"`
	Timestamp   time.Time          `json:"timestamp"`
}

// SecurityAuditReport represents the complete security audit report
type SecurityAuditReport struct {
	ID               string            `json:"id"`
	Timestamp        time.Time         `json:"timestamp"`
	DatabaseVersion  string            `json:"database_version"`
	TotalFindings    int               `json:"total_findings"`
	CriticalFindings int               `json:"critical_findings"`
	WarningFindings  int               `json:"warning_findings"`
	InfoFindings     int               `json:"info_findings"`
	Findings         []SecurityFinding `json:"findings"`
	Summary          string            `json:"summary"`
}

// DatabaseSecurityAuditor performs security audits on database configurations and queries
type DatabaseSecurityAuditor struct {
	db          *gorm.DB
	queryLogger QueryLogger
	findings    []SecurityFinding
}

// NewDatabaseSecurityAuditor creates a new database security auditor
func NewDatabaseSecurityAuditor(db *gorm.DB, queryLogger QueryLogger) *DatabaseSecurityAuditor {
	return &DatabaseSecurityAuditor{
		db:          db,
		queryLogger: queryLogger,
		findings:    make([]SecurityFinding, 0),
	}
}

// RunFullAudit performs a comprehensive security audit
func (dsa *DatabaseSecurityAuditor) RunFullAudit(ctx context.Context) (*SecurityAuditReport, error) {
	dsa.findings = make([]SecurityFinding, 0) // Reset findings

	// Run all audit checks
	if err := dsa.auditDatabaseConfiguration(ctx); err != nil {
		return nil, fmt.Errorf("failed to audit database configuration: %w", err)
	}

	if err := dsa.auditTableStructure(ctx); err != nil {
		return nil, fmt.Errorf("failed to audit table structure: %w", err)
	}

	if err := dsa.auditQueryPatterns(ctx); err != nil {
		return nil, fmt.Errorf("failed to audit query patterns: %w", err)
	}

	if err := dsa.auditPermissions(ctx); err != nil {
		return nil, fmt.Errorf("failed to audit permissions: %w", err)
	}

	if err := dsa.auditDataExposure(ctx); err != nil {
		return nil, fmt.Errorf("failed to audit data exposure: %w", err)
	}

	// Generate report
	report := dsa.generateReport()
	return report, nil
}

// auditDatabaseConfiguration audits database configuration settings
func (dsa *DatabaseSecurityAuditor) auditDatabaseConfiguration(ctx context.Context) error {
	// Check SSL/TLS configuration
	var sslMode string
	err := dsa.db.WithContext(ctx).Raw("SHOW ssl").Scan(&sslMode).Error
	if err == nil && strings.ToLower(sslMode) != "on" {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditCritical,
			Category:    "Configuration",
			Title:       "SSL/TLS Not Enabled",
			Description: "Database connection is not using SSL/TLS encryption",
			Risk:        "Data transmitted between application and database can be intercepted",
			Remediation: "Enable SSL/TLS encryption for database connections",
		})
	}

	// Check log settings
	var logStatement string
	err = dsa.db.WithContext(ctx).Raw("SHOW log_statement").Scan(&logStatement).Error
	if err == nil && logStatement == "all" {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditWarning,
			Category:    "Configuration",
			Title:       "Excessive Query Logging",
			Description: "All SQL statements are being logged, which may include sensitive data",
			Risk:        "Sensitive data may be exposed in database logs",
			Remediation: "Configure log_statement to 'ddl' or 'mod' instead of 'all'",
		})
	}

	// Check password encryption
	var passwordEncryption string
	err = dsa.db.WithContext(ctx).Raw("SHOW password_encryption").Scan(&passwordEncryption).Error
	if err == nil && strings.ToLower(passwordEncryption) != "scram-sha-256" {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditWarning,
			Category:    "Configuration",
			Title:       "Weak Password Encryption",
			Description: fmt.Sprintf("Password encryption method is '%s', not the recommended 'scram-sha-256'", passwordEncryption),
			Risk:        "User passwords may be vulnerable to attacks",
			Remediation: "Set password_encryption to 'scram-sha-256'",
		})
	}

	return nil
}

// auditTableStructure audits table structure for security issues
func (dsa *DatabaseSecurityAuditor) auditTableStructure(ctx context.Context) error {
	// Get all tables
	var tables []string
	err := dsa.db.WithContext(ctx).Raw(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = current_schema()
		AND table_type = 'BASE TABLE'
	`).Scan(&tables).Error

	if err != nil {
		return err
	}

	for _, table := range tables {
		if err := dsa.auditTable(ctx, table); err != nil {
			return err
		}
	}

	return nil
}

// auditTable audits a specific table for security issues
func (dsa *DatabaseSecurityAuditor) auditTable(ctx context.Context, tableName string) error {
	// Check for sensitive columns without encryption
	sensitiveColumns := []string{"password", "ssn", "credit_card", "social_security", "tax_id"}

	var columns []struct {
		ColumnName string `gorm:"column:column_name"`
		DataType   string `gorm:"column:data_type"`
	}

	err := dsa.db.WithContext(ctx).Raw(`
		SELECT column_name, data_type
		FROM information_schema.columns
		WHERE table_name = ? AND table_schema = current_schema()
	`, tableName).Scan(&columns).Error

	if err != nil {
		return err
	}

	for _, col := range columns {
		colNameLower := strings.ToLower(col.ColumnName)

		// Check for sensitive data in plain text
		for _, sensitive := range sensitiveColumns {
			if strings.Contains(colNameLower, sensitive) && col.DataType != "bytea" {
				dsa.addFinding(SecurityFinding{
					Level:       SecurityAuditCritical,
					Category:    "Data Protection",
					Title:       "Sensitive Data Not Encrypted",
					Description: fmt.Sprintf("Column '%s' in table '%s' appears to contain sensitive data but is not encrypted", col.ColumnName, tableName),
					Table:       tableName,
					Column:      col.ColumnName,
					Risk:        "Sensitive data stored in plain text can be easily accessed if database is compromised",
					Remediation: "Encrypt sensitive data before storing in database or use database-level encryption",
				})
			}
		}

		// Check for missing constraints on important fields
		if strings.Contains(colNameLower, "email") {
			// Check if email column has proper constraints
			var constraintCount int64
			dsa.db.WithContext(ctx).Raw(`
				SELECT COUNT(*)
				FROM information_schema.table_constraints tc
				JOIN information_schema.constraint_column_usage ccu ON tc.constraint_name = ccu.constraint_name
				WHERE tc.table_name = ? AND ccu.column_name = ? AND tc.constraint_type IN ('UNIQUE', 'PRIMARY KEY')
			`, tableName, col.ColumnName).Scan(&constraintCount)

			if constraintCount == 0 {
				dsa.addFinding(SecurityFinding{
					Level:       SecurityAuditWarning,
					Category:    "Data Integrity",
					Title:       "Email Column Missing Unique Constraint",
					Description: fmt.Sprintf("Email column '%s' in table '%s' does not have a unique constraint", col.ColumnName, tableName),
					Table:       tableName,
					Column:      col.ColumnName,
					Risk:        "Duplicate email addresses can lead to authentication and authorization issues",
					Remediation: "Add a unique constraint to the email column",
				})
			}
		}
	}

	// Check for tables without primary keys
	var pkCount int64
	err = dsa.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM information_schema.table_constraints
		WHERE table_name = ? AND constraint_type = 'PRIMARY KEY'
	`, tableName).Scan(&pkCount).Error

	if err != nil {
		return err
	}

	if pkCount == 0 {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditWarning,
			Category:    "Data Integrity",
			Title:       "Table Missing Primary Key",
			Description: fmt.Sprintf("Table '%s' does not have a primary key", tableName),
			Table:       tableName,
			Risk:        "Tables without primary keys can lead to data integrity issues and make auditing difficult",
			Remediation: "Add a primary key to the table",
		})
	}

	return nil
}

// auditQueryPatterns audits recent query patterns for security issues
func (dsa *DatabaseSecurityAuditor) auditQueryPatterns(ctx context.Context) error {
	if dsa.queryLogger == nil {
		return nil // Skip if no query logger available
	}

	recentQueries := dsa.queryLogger.GetRecentQueries(1000)

	for _, query := range recentQueries {
		if err := dsa.auditQuery(query); err != nil {
			return err
		}
	}

	return nil
}

// auditQuery audits a single query for security issues
func (dsa *DatabaseSecurityAuditor) auditQuery(query QueryLogEntry) error {
	queryUpper := strings.ToUpper(query.Query)

	// Check for potential SQL injection patterns
	injectionPatterns := []struct {
		pattern string
		risk    string
	}{
		{"'.*OR.*'.*'", "Potential SQL injection with OR condition"},
		{"'.*UNION.*SELECT", "Potential SQL injection with UNION"},
		{"'.*;.*--", "Potential SQL injection with comment"},
		{"'.*DROP.*TABLE", "Potential SQL injection attempting to drop tables"},
		{"'.*INSERT.*INTO", "Potential SQL injection attempting data insertion"},
		{"'.*UPDATE.*SET", "Potential SQL injection attempting data modification"},
		{"'.*DELETE.*FROM", "Potential SQL injection attempting data deletion"},
	}

	for _, pattern := range injectionPatterns {
		matched, _ := regexp.MatchString(pattern.pattern, queryUpper)
		if matched {
			dsa.addFinding(SecurityFinding{
				Level:       SecurityAuditCritical,
				Category:    "SQL Injection",
				Title:       "Potential SQL Injection Detected",
				Description: pattern.risk,
				Query:       query.Query,
				Risk:        "SQL injection can lead to unauthorized data access, modification, or deletion",
				Remediation: "Use parameterized queries and input validation",
			})
		}
	}

	// Check for queries without WHERE clauses on sensitive operations
	if strings.Contains(queryUpper, "DELETE FROM") && !strings.Contains(queryUpper, "WHERE") {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditCritical,
			Category:    "Data Safety",
			Title:       "DELETE Without WHERE Clause",
			Description: "DELETE query executed without WHERE clause",
			Query:       query.Query,
			Risk:        "Can result in accidental deletion of all data in the table",
			Remediation: "Always use WHERE clauses with DELETE statements",
		})
	}

	if strings.Contains(queryUpper, "UPDATE") && !strings.Contains(queryUpper, "WHERE") {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditWarning,
			Category:    "Data Safety",
			Title:       "UPDATE Without WHERE Clause",
			Description: "UPDATE query executed without WHERE clause",
			Query:       query.Query,
			Risk:        "Can result in accidental modification of all data in the table",
			Remediation: "Always use WHERE clauses with UPDATE statements",
		})
	}

	// Check for SELECT * queries
	if strings.Contains(queryUpper, "SELECT *") {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditInfo,
			Category:    "Performance",
			Title:       "SELECT * Query Detected",
			Description: "Query uses SELECT * instead of specific columns",
			Query:       query.Query,
			Risk:        "Can expose sensitive data and impact performance",
			Remediation: "Specify only the columns needed in SELECT statements",
		})
	}

	return nil
}

// auditPermissions audits database permissions and roles
func (dsa *DatabaseSecurityAuditor) auditPermissions(ctx context.Context) error {
	// Check for overly permissive roles
	var roles []struct {
		RoleName    string `gorm:"column:rolname"`
		CanLogin    bool   `gorm:"column:rolcanlogin"`
		IsSuperuser bool   `gorm:"column:rolsuper"`
		CreateDB    bool   `gorm:"column:rolcreatedb"`
		CreateRole  bool   `gorm:"column:rolcreaterole"`
	}

	err := dsa.db.WithContext(ctx).Raw(`
		SELECT rolname, rolcanlogin, rolsuper, rolcreatedb, rolcreaterole
		FROM pg_roles
		WHERE rolcanlogin = true
	`).Scan(&roles).Error

	if err != nil {
		return err
	}

	for _, role := range roles {
		if role.IsSuperuser && role.RoleName != "postgres" {
			dsa.addFinding(SecurityFinding{
				Level:       SecurityAuditCritical,
				Category:    "Permissions",
				Title:       "Non-Admin User with Superuser Privileges",
				Description: fmt.Sprintf("Role '%s' has superuser privileges", role.RoleName),
				Risk:        "Superuser privileges allow unrestricted access to all database operations",
				Remediation: "Remove superuser privileges from application roles",
			})
		}

		if role.CreateDB && !strings.Contains(strings.ToLower(role.RoleName), "admin") {
			dsa.addFinding(SecurityFinding{
				Level:       SecurityAuditWarning,
				Category:    "Permissions",
				Title:       "User Can Create Databases",
				Description: fmt.Sprintf("Role '%s' has database creation privileges", role.RoleName),
				Risk:        "Database creation privileges can be misused to create unauthorized databases",
				Remediation: "Remove database creation privileges from application roles",
			})
		}

		if role.CreateRole && !strings.Contains(strings.ToLower(role.RoleName), "admin") {
			dsa.addFinding(SecurityFinding{
				Level:       SecurityAuditWarning,
				Category:    "Permissions",
				Title:       "User Can Create Roles",
				Description: fmt.Sprintf("Role '%s' has role creation privileges", role.RoleName),
				Risk:        "Role creation privileges can be used to escalate privileges",
				Remediation: "Remove role creation privileges from application roles",
			})
		}
	}

	return nil
}

// auditDataExposure audits for potential data exposure issues
func (dsa *DatabaseSecurityAuditor) auditDataExposure(ctx context.Context) error {
	// Check for default passwords or common weak passwords
	var userCount int64
	err := dsa.db.WithContext(ctx).Raw(`
		SELECT COUNT(*)
		FROM users
		WHERE password IN ('password', '123456', 'admin', 'default', '')
		OR password IS NULL
	`).Scan(&userCount).Error

	if err == nil && userCount > 0 {
		dsa.addFinding(SecurityFinding{
			Level:       SecurityAuditCritical,
			Category:    "Authentication",
			Title:       "Weak or Default Passwords Detected",
			Description: fmt.Sprintf("Found %d users with weak or default passwords", userCount),
			Risk:        "Weak passwords can be easily compromised through brute force attacks",
			Remediation: "Enforce strong password policies and require password changes",
		})
	}

	// Check for unencrypted sensitive data
	tables := []string{"users", "customers", "payments"}
	for _, table := range tables {
		var count int64
		// This is a simplified check - in reality, you'd need to know your schema
		query := fmt.Sprintf("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = '%s'", table)
		err := dsa.db.WithContext(ctx).Raw(query).Scan(&count).Error
		if err == nil && count > 0 {
			// Table exists, check for potential issues
			dsa.addFinding(SecurityFinding{
				Level:       SecurityAuditInfo,
				Category:    "Data Protection",
				Title:       "Sensitive Table Detected",
				Description: fmt.Sprintf("Table '%s' may contain sensitive data", table),
				Table:       table,
				Risk:        "Sensitive data should be properly protected",
				Remediation: "Ensure sensitive data is encrypted and access is properly controlled",
			})
		}
	}

	return nil
}

// addFinding adds a security finding to the audit results
func (dsa *DatabaseSecurityAuditor) addFinding(finding SecurityFinding) {
	finding.ID = fmt.Sprintf("finding_%d", time.Now().UnixNano())
	finding.Timestamp = time.Now()
	dsa.findings = append(dsa.findings, finding)
}

// generateReport generates the final security audit report
func (dsa *DatabaseSecurityAuditor) generateReport() *SecurityAuditReport {
	report := &SecurityAuditReport{
		ID:            fmt.Sprintf("audit_%d", time.Now().UnixNano()),
		Timestamp:     time.Now(),
		TotalFindings: len(dsa.findings),
		Findings:      dsa.findings,
	}

	// Count findings by level
	for _, finding := range dsa.findings {
		switch finding.Level {
		case SecurityAuditCritical:
			report.CriticalFindings++
		case SecurityAuditWarning:
			report.WarningFindings++
		case SecurityAuditInfo:
			report.InfoFindings++
		}
	}

	// Generate summary
	if report.CriticalFindings > 0 {
		report.Summary = fmt.Sprintf("CRITICAL: %d critical security issues found that require immediate attention", report.CriticalFindings)
	} else if report.WarningFindings > 0 {
		report.Summary = fmt.Sprintf("WARNING: %d security warnings found that should be addressed", report.WarningFindings)
	} else if report.InfoFindings > 0 {
		report.Summary = fmt.Sprintf("INFO: %d informational findings - consider reviewing for optimization", report.InfoFindings)
	} else {
		report.Summary = "No security issues detected"
	}

	// Get database version
	var version string
	dsa.db.Raw("SELECT version()").Scan(&version)
	report.DatabaseVersion = version

	return report
}

// GetFindingsByLevel returns findings filtered by security level
func (dsa *DatabaseSecurityAuditor) GetFindingsByLevel(level SecurityAuditLevel) []SecurityFinding {
	var filtered []SecurityFinding
	for _, finding := range dsa.findings {
		if finding.Level == level {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// GetFindingsByCategory returns findings filtered by category
func (dsa *DatabaseSecurityAuditor) GetFindingsByCategory(category string) []SecurityFinding {
	var filtered []SecurityFinding
	for _, finding := range dsa.findings {
		if finding.Category == category {
			filtered = append(filtered, finding)
		}
	}
	return filtered
}

// SecurityTestSuite provides utilities for testing database security
type SecurityTestSuite struct {
	db      *gorm.DB
	auditor *DatabaseSecurityAuditor
}

// NewSecurityTestSuite creates a new security test suite
func NewSecurityTestSuite(db *gorm.DB, auditor *DatabaseSecurityAuditor) *SecurityTestSuite {
	return &SecurityTestSuite{
		db:      db,
		auditor: auditor,
	}
}

// TestSQLInjectionResistance tests resistance to SQL injection attacks
func (sts *SecurityTestSuite) TestSQLInjectionResistance(ctx context.Context) []SecurityFinding {
	var findings []SecurityFinding

	// Test common SQL injection patterns
	injectionTests := []string{
		"'; DROP TABLE users; --",
		"' OR '1'='1",
		"' UNION SELECT * FROM users --",
		"'; INSERT INTO users (email) VALUES ('hacker@evil.com'); --",
	}

	for _, injection := range injectionTests {
		// Test with query builder (should be safe)
		qb := NewQueryBuilder(sts.db)
		var count int64
		err := qb.Model(&struct{ ID uint }{}).Where("email = ?", injection).Count(&count)

		if err != nil && strings.Contains(err.Error(), "syntax error") {
			findings = append(findings, SecurityFinding{
				Level:       SecurityAuditCritical,
				Category:    "SQL Injection Test",
				Title:       "SQL Injection Vulnerability Detected",
				Description: fmt.Sprintf("Injection pattern '%s' caused syntax error, indicating potential vulnerability", injection),
				Risk:        "Application may be vulnerable to SQL injection attacks",
				Remediation: "Ensure all queries use parameterized statements",
			})
		}
	}

	return findings
}

// TestAccessControls tests database access controls
func (sts *SecurityTestSuite) TestAccessControls(ctx context.Context) []SecurityFinding {
	var findings []SecurityFinding

	// Test if application can access system tables (it shouldn't be able to modify them)
	systemTables := []string{"pg_user", "pg_shadow", "information_schema.tables"}

	for _, table := range systemTables {
		var count int64
		err := sts.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count).Error

		if err == nil {
			// Can read system tables - check if can modify
			err = sts.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE 1=0", table)).Error
			if err == nil {
				findings = append(findings, SecurityFinding{
					Level:       SecurityAuditCritical,
					Category:    "Access Control Test",
					Title:       "Excessive System Table Access",
					Description: fmt.Sprintf("Application can modify system table '%s'", table),
					Risk:        "Application has excessive privileges that could be exploited",
					Remediation: "Restrict application database user privileges",
				})
			}
		}
	}

	return findings
}
