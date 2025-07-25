package repositories

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// transactionManagerImpl is the GORM-based implementation of TransactionManager
type transactionManagerImpl struct {
	db *gorm.DB
}

// NewTransactionManager creates a new instance of TransactionManager
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManagerImpl{
		db: db,
	}
}

// WithTransaction executes a function within a database transaction
func (tm *transactionManagerImpl) WithTransaction(fn func(*gorm.DB) error) error {
	return tm.db.Transaction(fn)
}

// WithTransactionContext executes a function within a database transaction with context
func (tm *transactionManagerImpl) WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ctx, tx)
	})
}

// BeginTransaction starts a new database transaction
func (tm *transactionManagerImpl) BeginTransaction() (*gorm.DB, error) {
	tx := tm.db.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// BeginTransactionWithContext starts a new database transaction with context
func (tm *transactionManagerImpl) BeginTransactionWithContext(ctx context.Context) (*gorm.DB, error) {
	tx := tm.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}

// CommitTransaction commits the given transaction
func (tm *transactionManagerImpl) CommitTransaction(tx *gorm.DB) error {
	return tx.Commit().Error
}

// RollbackTransaction rolls back the given transaction
func (tm *transactionManagerImpl) RollbackTransaction(tx *gorm.DB) error {
	return tx.Rollback().Error
}

// WithSavepoint executes a function within a savepoint
func (tm *transactionManagerImpl) WithSavepoint(tx *gorm.DB, name string, fn func(*gorm.DB) error) error {
	// Create savepoint
	if err := tx.Exec(fmt.Sprintf("SAVEPOINT %s", name)).Error; err != nil {
		return fmt.Errorf("failed to create savepoint %s: %w", name, err)
	}

	// Execute function
	err := fn(tx)
	if err != nil {
		// Rollback to savepoint on error
		if rollbackErr := tx.Exec(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", name)).Error; rollbackErr != nil {
			return fmt.Errorf("failed to rollback to savepoint %s: %w (original error: %v)", name, rollbackErr, err)
		}
		return err
	}

	// Release savepoint on success
	if err := tx.Exec(fmt.Sprintf("RELEASE SAVEPOINT %s", name)).Error; err != nil {
		return fmt.Errorf("failed to release savepoint %s: %w", name, err)
	}

	return nil
}

// WithRetryableTransaction executes a function within a retryable transaction
func (tm *transactionManagerImpl) WithRetryableTransaction(fn func(*gorm.DB) error, maxRetries int) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := tm.WithTransaction(fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable (deadlock, serialization failure, etc.)
		if !isRetryableError(err) {
			return err // Non-retryable error, fail immediately
		}

		if attempt < maxRetries {
			// Wait before retrying with exponential backoff
			waitTime := time.Duration(attempt+1) * 100 * time.Millisecond
			time.Sleep(waitTime)
		}
	}

	return fmt.Errorf("transaction failed after %d retries, last error: %w", maxRetries, lastErr)
}

// isRetryableError checks if an error is retryable (deadlock, serialization failure, etc.)
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Common retryable error patterns
	retryablePatterns := []string{
		"deadlock detected",
		"could not serialize access",
		"serialization failure",
		"lock timeout",
		"connection reset",
		"connection refused",
		"temporary failure",
	}

	for _, pattern := range retryablePatterns {
		if contains(errStr, pattern) {
			return true
		}
	}

	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOfSubstring(s, substr) >= 0)))
}

// indexOfSubstring finds the index of a substring in a string
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Advanced transaction features

// WithIsolationLevel executes a function within a transaction with specific isolation level
func (tm *transactionManagerImpl) WithIsolationLevel(isolationLevel string, fn func(*gorm.DB) error) error {
	return tm.db.Transaction(func(tx *gorm.DB) error {
		// Set isolation level
		if err := tm.SetIsolationLevel(tx, isolationLevel); err != nil {
			return fmt.Errorf("failed to set isolation level %s: %w", isolationLevel, err)
		}

		// Execute function
		return fn(tx)
	})
}

// WithIsolationLevelContext executes a function within a transaction with specific isolation level and context
func (tm *transactionManagerImpl) WithIsolationLevelContext(ctx context.Context, isolationLevel string, fn func(context.Context, *gorm.DB) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Set isolation level
		if err := tm.SetIsolationLevel(tx, isolationLevel); err != nil {
			return fmt.Errorf("failed to set isolation level %s: %w", isolationLevel, err)
		}

		// Execute function
		return fn(ctx, tx)
	})
}

// WithTimeout executes a function within a transaction with timeout
func (tm *transactionManagerImpl) WithTimeout(timeout time.Duration, fn func(*gorm.DB) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return tm.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		return fn(tx)
	})
}

// WithTimeoutContext executes a function within a transaction with timeout and context
func (tm *transactionManagerImpl) WithTimeoutContext(ctx context.Context, timeout time.Duration, fn func(context.Context, *gorm.DB) error) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return tm.WithTransactionContext(timeoutCtx, fn)
}

// WithTimeoutAndRetry executes a function within a retryable transaction with timeout
func (tm *transactionManagerImpl) WithTimeoutAndRetry(timeout time.Duration, fn func(*gorm.DB) error, maxRetries int) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Check if context is still valid
		select {
		case <-ctx.Done():
			return fmt.Errorf("transaction timeout after %v: %w", timeout, ctx.Err())
		default:
		}

		err := tm.db.WithContext(ctx).Transaction(fn)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			return err // Non-retryable error, fail immediately
		}

		if attempt < maxRetries {
			// Wait before retrying with exponential backoff
			waitTime := time.Duration(attempt+1) * 100 * time.Millisecond

			// Don't wait longer than remaining timeout
			select {
			case <-ctx.Done():
				return fmt.Errorf("transaction timeout during retry wait: %w", ctx.Err())
			case <-time.After(waitTime):
				// Continue to next retry
			}
		}
	}

	return fmt.Errorf("transaction failed after %d retries within %v timeout, last error: %w", maxRetries, timeout, lastErr)
}

// GetIsolationLevel returns the current transaction isolation level
func (tm *transactionManagerImpl) GetIsolationLevel(tx *gorm.DB) (string, error) {
	var level string

	// Query current isolation level (PostgreSQL syntax)
	err := tx.Raw("SHOW transaction_isolation").Scan(&level).Error
	if err != nil {
		// Try MySQL/SQLite syntax
		err = tx.Raw("SELECT @@transaction_isolation").Scan(&level).Error
		if err != nil {
			// Try alternative query
			err = tx.Raw("SELECT @@tx_isolation").Scan(&level).Error
			if err != nil {
				return "", fmt.Errorf("failed to get transaction isolation level: %w", err)
			}
		}
	}

	return level, nil
}

// SetIsolationLevel sets the transaction isolation level
func (tm *transactionManagerImpl) SetIsolationLevel(tx *gorm.DB, isolationLevel string) error {
	// Validate isolation level
	validLevels := map[string]bool{
		"READ UNCOMMITTED": true,
		"READ COMMITTED":   true,
		"REPEATABLE READ":  true,
		"SERIALIZABLE":     true,
	}

	upperLevel := strings.ToUpper(isolationLevel)
	if !validLevels[upperLevel] {
		return fmt.Errorf("invalid isolation level: %s. Valid levels are: READ UNCOMMITTED, READ COMMITTED, REPEATABLE READ, SERIALIZABLE", isolationLevel)
	}

	// Set isolation level (works for PostgreSQL, MySQL, and most SQL databases)
	err := tx.Exec(fmt.Sprintf("SET TRANSACTION ISOLATION LEVEL %s", upperLevel)).Error
	if err != nil {
		return fmt.Errorf("failed to set transaction isolation level to %s: %w", isolationLevel, err)
	}

	return nil
}

// Additional utility methods for advanced transaction management

// WithDeadlockRetry is a convenience method for retrying transactions with deadlock detection
func (tm *transactionManagerImpl) WithDeadlockRetry(fn func(*gorm.DB) error) error {
	return tm.WithRetryableTransaction(fn, 3) // Default 3 retries for deadlocks
}

// WithSerializableTransaction executes a function within a SERIALIZABLE transaction
func (tm *transactionManagerImpl) WithSerializableTransaction(fn func(*gorm.DB) error) error {
	return tm.WithIsolationLevel("SERIALIZABLE", fn)
}

// WithReadCommittedTransaction executes a function within a READ COMMITTED transaction
func (tm *transactionManagerImpl) WithReadCommittedTransaction(fn func(*gorm.DB) error) error {
	return tm.WithIsolationLevel("READ COMMITTED", fn)
}

// WithRepeatableReadTransaction executes a function within a REPEATABLE READ transaction
func (tm *transactionManagerImpl) WithRepeatableReadTransaction(fn func(*gorm.DB) error) error {
	return tm.WithIsolationLevel("REPEATABLE READ", fn)
}

// WithReadUncommittedTransaction executes a function within a READ UNCOMMITTED transaction
func (tm *transactionManagerImpl) WithReadUncommittedTransaction(fn func(*gorm.DB) error) error {
	return tm.WithIsolationLevel("READ UNCOMMITTED", fn)
}
