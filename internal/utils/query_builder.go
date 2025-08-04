package utils

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// QueryBuilder provides a secure way to build database queries with automatic parameterization
type QueryBuilder struct {
	db    *gorm.DB
	query *gorm.DB
}

// NewQueryBuilder creates a new query builder instance
func NewQueryBuilder(db *gorm.DB) *QueryBuilder {
	return &QueryBuilder{
		db:    db,
		query: db,
	}
}

// Model sets the model for the query
func (qb *QueryBuilder) Model(model interface{}) *QueryBuilder {
	qb.query = qb.query.Model(model)
	return qb
}

// Table sets the table name for the query
func (qb *QueryBuilder) Table(name string) *QueryBuilder {
	// Sanitize table name to prevent injection
	sanitizedName := sanitizeIdentifier(name)
	qb.query = qb.query.Table(sanitizedName)
	return qb
}

// Select adds SELECT clause with field validation
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	sanitizedFields := make([]string, len(fields))
	for i, field := range fields {
		sanitizedFields[i] = sanitizeIdentifier(field)
	}
	qb.query = qb.query.Select(sanitizedFields)
	return qb
}

// Where adds WHERE clause with automatic parameterization
func (qb *QueryBuilder) Where(condition string, args ...interface{}) *QueryBuilder {
	// Validate condition to ensure it doesn't contain dangerous patterns
	if err := validateWhereCondition(condition); err != nil {
		// Log the error and return empty result set for security
		qb.query = qb.query.Where("1 = 0") // Always false condition
		return qb
	}

	qb.query = qb.query.Where(condition, args...)
	return qb
}

// WhereIn adds WHERE IN clause with automatic parameterization
func (qb *QueryBuilder) WhereIn(column string, values interface{}) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s IN ?", sanitizedColumn), values)
	return qb
}

// WhereBetween adds WHERE BETWEEN clause with automatic parameterization
func (qb *QueryBuilder) WhereBetween(column string, start, end interface{}) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s BETWEEN ? AND ?", sanitizedColumn), start, end)
	return qb
}

// WhereNull adds WHERE IS NULL clause
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s IS NULL", sanitizedColumn))
	return qb
}

// WhereNotNull adds WHERE IS NOT NULL clause
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s IS NOT NULL", sanitizedColumn))
	return qb
}

// WhereLike adds WHERE LIKE clause with automatic parameterization
func (qb *QueryBuilder) WhereLike(column string, pattern string) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s LIKE ?", sanitizedColumn), pattern)
	return qb
}

// WhereILike adds WHERE ILIKE clause (case-insensitive) with automatic parameterization
func (qb *QueryBuilder) WhereILike(column string, pattern string) *QueryBuilder {
	sanitizedColumn := sanitizeIdentifier(column)
	qb.query = qb.query.Where(fmt.Sprintf("%s ILIKE ?", sanitizedColumn), pattern)
	return qb
}

// Join adds JOIN clause with table and condition validation
func (qb *QueryBuilder) Join(table string, condition string) *QueryBuilder {
	sanitizedTable := sanitizeIdentifier(table)
	if err := validateJoinCondition(condition); err != nil {
		// Log error and return empty result set
		qb.query = qb.query.Where("1 = 0")
		return qb
	}
	qb.query = qb.query.Joins(fmt.Sprintf("JOIN %s ON %s", sanitizedTable, condition))
	return qb
}

// LeftJoin adds LEFT JOIN clause with table and condition validation
func (qb *QueryBuilder) LeftJoin(table string, condition string) *QueryBuilder {
	sanitizedTable := sanitizeIdentifier(table)
	if err := validateJoinCondition(condition); err != nil {
		qb.query = qb.query.Where("1 = 0")
		return qb
	}
	qb.query = qb.query.Joins(fmt.Sprintf("LEFT JOIN %s ON %s", sanitizedTable, condition))
	return qb
}

// OrderBy adds ORDER BY clause with field validation
func (qb *QueryBuilder) OrderBy(field string, direction ...string) *QueryBuilder {
	sanitizedField := sanitizeIdentifier(field)
	dir := "ASC"
	if len(direction) > 0 {
		dir = sanitizeDirection(direction[0])
	}
	qb.query = qb.query.Order(fmt.Sprintf("%s %s", sanitizedField, dir))
	return qb
}

// GroupBy adds GROUP BY clause with field validation
func (qb *QueryBuilder) GroupBy(fields ...string) *QueryBuilder {
	sanitizedFields := make([]string, len(fields))
	for i, field := range fields {
		sanitizedFields[i] = sanitizeIdentifier(field)
	}
	qb.query = qb.query.Group(strings.Join(sanitizedFields, ", "))
	return qb
}

// Having adds HAVING clause with automatic parameterization
func (qb *QueryBuilder) Having(condition string, args ...interface{}) *QueryBuilder {
	if err := validateWhereCondition(condition); err != nil {
		qb.query = qb.query.Where("1 = 0")
		return qb
	}
	qb.query = qb.query.Having(condition, args...)
	return qb
}

// Limit adds LIMIT clause with validation
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	if limit < 0 {
		limit = 0
	}
	if limit > 10000 { // Prevent excessive limits
		limit = 10000
	}
	qb.query = qb.query.Limit(limit)
	return qb
}

// Offset adds OFFSET clause with validation
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	if offset < 0 {
		offset = 0
	}
	qb.query = qb.query.Offset(offset)
	return qb
}

// Distinct adds DISTINCT clause
func (qb *QueryBuilder) Distinct(fields ...string) *QueryBuilder {
	if len(fields) > 0 {
		sanitizedFields := make([]string, len(fields))
		for i, field := range fields {
			sanitizedFields[i] = sanitizeIdentifier(field)
		}
		qb.query = qb.query.Distinct(sanitizedFields)
	} else {
		qb.query = qb.query.Distinct()
	}
	return qb
}

// Find executes the query and returns results
func (qb *QueryBuilder) Find(dest interface{}) error {
	return qb.query.Find(dest).Error
}

// First executes the query and returns the first result
func (qb *QueryBuilder) First(dest interface{}) error {
	return qb.query.First(dest).Error
}

// Count executes a count query
func (qb *QueryBuilder) Count(count *int64) error {
	return qb.query.Count(count).Error
}

// Create executes an INSERT query
func (qb *QueryBuilder) Create(value interface{}) error {
	return qb.query.Create(value).Error
}

// Update executes an UPDATE query with validation
func (qb *QueryBuilder) Update(column string, value interface{}) error {
	sanitizedColumn := sanitizeIdentifier(column)
	return qb.query.Update(sanitizedColumn, value).Error
}

// Updates executes an UPDATE query with multiple fields
func (qb *QueryBuilder) Updates(values interface{}) error {
	return qb.query.Updates(values).Error
}

// Delete executes a DELETE query
func (qb *QueryBuilder) Delete(value interface{}) error {
	return qb.query.Delete(value).Error
}

// GetQuery returns the underlying GORM query for advanced usage
func (qb *QueryBuilder) GetQuery() *gorm.DB {
	return qb.query
}

// Reset resets the query builder to its initial state
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.query = qb.db
	return qb
}

// Pagination helper for cursor-based pagination
type PaginationOptions struct {
	Cursor    string
	Limit     int
	SortField string
	SortDir   string
}

// Paginate adds cursor-based pagination
func (qb *QueryBuilder) Paginate(opts PaginationOptions) *QueryBuilder {
	// Validate and sanitize inputs
	limit := opts.Limit
	if limit <= 0 || limit > 1000 {
		limit = 50 // Default limit
	}

	sortField := sanitizeIdentifier(opts.SortField)
	if sortField == "" {
		sortField = "id"
	}

	sortDir := sanitizeDirection(opts.SortDir)

	qb.query = qb.query.Limit(limit)

	if opts.Cursor != "" {
		// Decode cursor (in real implementation, you'd decode a base64 encoded value)
		if sortDir == "ASC" {
			qb.query = qb.query.Where(fmt.Sprintf("%s > ?", sortField), opts.Cursor)
		} else {
			qb.query = qb.query.Where(fmt.Sprintf("%s < ?", sortField), opts.Cursor)
		}
	}

	qb.query = qb.query.Order(fmt.Sprintf("%s %s", sortField, sortDir))
	return qb
}

// Security validation functions

// sanitizeIdentifier sanitizes database identifiers (table names, column names)
func sanitizeIdentifier(identifier string) string {
	// Remove any characters that aren't alphanumeric, underscore, or dot
	var result strings.Builder
	for _, r := range identifier {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}

	sanitized := result.String()

	// Prevent empty identifiers
	if sanitized == "" {
		return "id" // Default safe identifier
	}

	// Prevent SQL keywords as identifiers
	if isSQLKeyword(strings.ToUpper(sanitized)) {
		return fmt.Sprintf("`%s`", sanitized)
	}

	return sanitized
}

// sanitizeDirection sanitizes ORDER BY direction
func sanitizeDirection(direction string) string {
	upper := strings.ToUpper(strings.TrimSpace(direction))
	if upper == "DESC" {
		return "DESC"
	}
	return "ASC"
}

// validateWhereCondition validates WHERE clause conditions
func validateWhereCondition(condition string) error {
	condition = strings.ToUpper(strings.TrimSpace(condition))

	// Check for dangerous patterns
	dangerousPatterns := []string{
		"DROP", "DELETE", "INSERT", "UPDATE", "CREATE", "ALTER",
		"EXEC", "EXECUTE", "UNION", "SCRIPT", "DECLARE",
		"--", "/*", "*/", "XP_", "SP_",
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(condition, pattern) {
			return fmt.Errorf("potentially dangerous pattern detected: %s", pattern)
		}
	}

	return nil
}

// validateJoinCondition validates JOIN conditions
func validateJoinCondition(condition string) error {
	return validateWhereCondition(condition)
}

// isSQLKeyword checks if a string is a SQL keyword
func isSQLKeyword(word string) bool {
	keywords := map[string]bool{
		"SELECT": true, "FROM": true, "WHERE": true, "INSERT": true,
		"UPDATE": true, "DELETE": true, "CREATE": true, "DROP": true,
		"ALTER": true, "INDEX": true, "TABLE": true, "DATABASE": true,
		"SCHEMA": true, "VIEW": true, "PROCEDURE": true, "FUNCTION": true,
		"TRIGGER": true, "USER": true, "ROLE": true, "GRANT": true,
		"REVOKE": true, "COMMIT": true, "ROLLBACK": true, "TRANSACTION": true,
	}

	return keywords[word]
}

// WithStats wraps query execution with performance monitoring
func (qb *QueryBuilder) WithStats(operation string) *QueryBuilder {
	// This would integrate with the query logger
	return qb
}

// Batch operations for bulk inserts/updates
type BatchOperation struct {
	qb        *QueryBuilder
	batchSize int
	items     []interface{}
}

// NewBatchOperation creates a new batch operation
func (qb *QueryBuilder) NewBatchOperation(batchSize int) *BatchOperation {
	if batchSize <= 0 || batchSize > 1000 {
		batchSize = 100 // Default batch size
	}

	return &BatchOperation{
		qb:        qb,
		batchSize: batchSize,
		items:     make([]interface{}, 0, batchSize),
	}
}

// Add adds an item to the batch
func (bo *BatchOperation) Add(item interface{}) error {
	bo.items = append(bo.items, item)

	if len(bo.items) >= bo.batchSize {
		return bo.Execute()
	}

	return nil
}

// Execute executes the batch operation
func (bo *BatchOperation) Execute() error {
	if len(bo.items) == 0 {
		return nil
	}

	err := bo.qb.query.Create(&bo.items).Error
	if err != nil {
		return err
	}

	// Clear the batch
	bo.items = bo.items[:0]
	return nil
}

// Flush executes any remaining items in the batch
func (bo *BatchOperation) Flush() error {
	return bo.Execute()
}
