package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TransactionExamples demonstrates how to use the advanced transaction features
type TransactionExamples struct {
	tm TransactionManager
}

// NewTransactionExamples creates a new instance of TransactionExamples
func NewTransactionExamples(tm TransactionManager) *TransactionExamples {
	return &TransactionExamples{tm: tm}
}

// ExampleBasicTransaction demonstrates basic transaction usage
func (te *TransactionExamples) ExampleBasicTransaction() error {
	return te.tm.WithTransaction(func(tx *gorm.DB) error {
		// Perform database operations within transaction
		user := &UserModel{
			Email:     "example@test.com",
			FirstName: "Example",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	})
}

// ExampleTransactionWithContext demonstrates context-aware transactions
func (te *TransactionExamples) ExampleTransactionWithContext(ctx context.Context) error {
	return te.tm.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
		// Use context for cancellation and timeouts
		user := &UserModel{
			Email:     "context@test.com",
			FirstName: "Context",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.WithContext(ctx).Create(user).Error
	})
}

// ExampleRetryableTransaction demonstrates automatic retry on deadlocks
func (te *TransactionExamples) ExampleRetryableTransaction() error {
	return te.tm.WithRetryableTransaction(func(tx *gorm.DB) error {
		// This transaction will be retried automatically if it encounters
		// deadlocks, serialization failures, or other retryable errors

		user := &UserModel{
			Email:     "retry@test.com",
			FirstName: "Retry",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	}, 3) // Retry up to 3 times
}

// ExampleDeadlockRetry demonstrates the convenience method for deadlock handling
func (te *TransactionExamples) ExampleDeadlockRetry() error {
	// Cast to implementation to access convenience methods
	tmImpl := te.tm.(*transactionManagerImpl)

	return tmImpl.WithDeadlockRetry(func(tx *gorm.DB) error {
		// This will automatically retry up to 3 times on deadlocks
		user := &UserModel{
			Email:     "deadlock@test.com",
			FirstName: "Deadlock",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	})
}

// ExampleSavepoints demonstrates nested transaction support with savepoints
func (te *TransactionExamples) ExampleSavepoints() error {
	return te.tm.WithTransaction(func(tx *gorm.DB) error {
		// Create first user
		user1 := &UserModel{
			Email:     "savepoint1@test.com",
			FirstName: "Savepoint1",
			LastName:  "User",
			IsActive:  true,
		}

		if err := tx.Create(user1).Error; err != nil {
			return err
		}

		// Use savepoint for nested operations
		return te.tm.WithSavepoint(tx, "user_creation", func(spTx *gorm.DB) error {
			user2 := &UserModel{
				Email:     "savepoint2@test.com",
				FirstName: "Savepoint2",
				LastName:  "User",
				IsActive:  true,
			}

			if err := spTx.Create(user2).Error; err != nil {
				return err // This will rollback to savepoint, but keep user1
			}

			// If we return an error here, only user2 creation will be rolled back
			// user1 will still be committed when the main transaction commits
			return nil
		})
	})
}

// ExampleIsolationLevels demonstrates different transaction isolation levels
func (te *TransactionExamples) ExampleIsolationLevels() error {
	// Example with SERIALIZABLE isolation level
	err := te.tm.WithIsolationLevel("SERIALIZABLE", func(tx *gorm.DB) error {
		// This transaction runs with SERIALIZABLE isolation
		// Provides the highest level of isolation but may have performance impact

		user := &UserModel{
			Email:     "serializable@test.com",
			FirstName: "Serializable",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	})

	if err != nil {
		return err
	}

	// Example with READ COMMITTED isolation level
	return te.tm.WithIsolationLevel("READ COMMITTED", func(tx *gorm.DB) error {
		// This transaction runs with READ COMMITTED isolation
		// Good balance between consistency and performance

		user := &UserModel{
			Email:     "readcommitted@test.com",
			FirstName: "ReadCommitted",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	})
}

// ExampleIsolationLevelConvenienceMethods demonstrates convenience methods for common isolation levels
func (te *TransactionExamples) ExampleIsolationLevelConvenienceMethods() error {
	// Cast to implementation to access convenience methods
	tmImpl := te.tm.(*transactionManagerImpl)

	// SERIALIZABLE transaction
	err := tmImpl.WithSerializableTransaction(func(tx *gorm.DB) error {
		user := &UserModel{
			Email:     "serializable_conv@test.com",
			FirstName: "SerializableConv",
			LastName:  "User",
			IsActive:  true,
		}
		return tx.Create(user).Error
	})

	if err != nil {
		return err
	}

	// READ COMMITTED transaction
	err = tmImpl.WithReadCommittedTransaction(func(tx *gorm.DB) error {
		user := &UserModel{
			Email:     "readcommitted_conv@test.com",
			FirstName: "ReadCommittedConv",
			LastName:  "User",
			IsActive:  true,
		}
		return tx.Create(user).Error
	})

	if err != nil {
		return err
	}

	// REPEATABLE READ transaction
	return tmImpl.WithRepeatableReadTransaction(func(tx *gorm.DB) error {
		user := &UserModel{
			Email:     "repeatableread_conv@test.com",
			FirstName: "RepeatableReadConv",
			LastName:  "User",
			IsActive:  true,
		}
		return tx.Create(user).Error
	})
}

// ExampleTimeoutTransactions demonstrates transaction timeout handling
func (te *TransactionExamples) ExampleTimeoutTransactions() error {
	// Transaction with 5 second timeout
	err := te.tm.WithTimeout(5*time.Second, func(tx *gorm.DB) error {
		// This transaction will be cancelled if it takes longer than 5 seconds
		user := &UserModel{
			Email:     "timeout@test.com",
			FirstName: "Timeout",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	})

	if err != nil {
		return err
	}

	// Transaction with context and timeout
	ctx := context.Background()
	err = te.tm.WithTimeoutContext(ctx, 10*time.Second, func(ctx context.Context, tx *gorm.DB) error {
		// This transaction inherits the parent context and adds a 10 second timeout
		user := &UserModel{
			Email:     "timeout_ctx@test.com",
			FirstName: "TimeoutCtx",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.WithContext(ctx).Create(user).Error
	})

	if err != nil {
		return err
	}

	// Transaction with timeout and retry
	return te.tm.WithTimeoutAndRetry(15*time.Second, func(tx *gorm.DB) error {
		// This transaction will retry on retryable errors within the 15 second timeout
		user := &UserModel{
			Email:     "timeout_retry@test.com",
			FirstName: "TimeoutRetry",
			LastName:  "User",
			IsActive:  true,
		}

		return tx.Create(user).Error
	}, 3) // Retry up to 3 times within the timeout period
}

// ExampleComplexTransaction demonstrates combining multiple advanced features
func (te *TransactionExamples) ExampleComplexTransaction() error {
	ctx := context.Background()

	return te.tm.WithIsolationLevelContext(ctx, "REPEATABLE READ", func(ctx context.Context, tx *gorm.DB) error {
		// Create main user
		mainUser := &UserModel{
			Email:     "complex_main@test.com",
			FirstName: "ComplexMain",
			LastName:  "User",
			IsActive:  true,
		}

		if err := tx.WithContext(ctx).Create(mainUser).Error; err != nil {
			return err
		}

		// Use savepoint for related operations
		return te.tm.WithSavepoint(tx, "related_operations", func(spTx *gorm.DB) error {
			// Create related user
			relatedUser := &UserModel{
				Email:     "complex_related@test.com",
				FirstName: "ComplexRelated",
				LastName:  "User",
				IsActive:  true,
			}

			if err := spTx.WithContext(ctx).Create(relatedUser).Error; err != nil {
				return err
			}

			// Simulate some business logic that might fail
			// If this fails, only the savepoint operations are rolled back
			if mainUser.ID%2 == 0 {
				return errors.New("business rule violation")
			}

			return nil
		})
	})
}

// ExampleManualTransactionManagement demonstrates manual transaction control
func (te *TransactionExamples) ExampleManualTransactionManagement() error {
	// Begin transaction manually
	tx, err := te.tm.BeginTransaction()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Always ensure transaction is handled
	defer func() {
		if r := recover(); r != nil {
			te.tm.RollbackTransaction(tx)
			panic(r)
		}
	}()

	// Perform operations
	user1 := &UserModel{
		Email:     "manual1@test.com",
		FirstName: "Manual1",
		LastName:  "User",
		IsActive:  true,
	}

	if err := tx.Create(user1).Error; err != nil {
		te.tm.RollbackTransaction(tx)
		return err
	}

	user2 := &UserModel{
		Email:     "manual2@test.com",
		FirstName: "Manual2",
		LastName:  "User",
		IsActive:  true,
	}

	if err := tx.Create(user2).Error; err != nil {
		te.tm.RollbackTransaction(tx)
		return err
	}

	// Commit transaction
	return te.tm.CommitTransaction(tx)
}

// ExampleManualTransactionWithContext demonstrates manual transaction control with context
func (te *TransactionExamples) ExampleManualTransactionWithContext(ctx context.Context) error {
	// Begin transaction with context
	tx, err := te.tm.BeginTransactionWithContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Always ensure transaction is handled
	defer func() {
		if r := recover(); r != nil {
			te.tm.RollbackTransaction(tx)
			panic(r)
		}
	}()

	// Check context cancellation
	select {
	case <-ctx.Done():
		te.tm.RollbackTransaction(tx)
		return ctx.Err()
	default:
	}

	// Perform operations with context
	user := &UserModel{
		Email:     "manual_ctx@test.com",
		FirstName: "ManualCtx",
		LastName:  "User",
		IsActive:  true,
	}

	if err := tx.WithContext(ctx).Create(user).Error; err != nil {
		te.tm.RollbackTransaction(tx)
		return err
	}

	// Commit transaction
	return te.tm.CommitTransaction(tx)
}

// ExampleErrorHandlingPatterns demonstrates different error handling patterns
func (te *TransactionExamples) ExampleErrorHandlingPatterns() error {
	// Pattern 1: Simple error handling with automatic rollback
	err := te.tm.WithTransaction(func(tx *gorm.DB) error {
		user := &UserModel{
			Email:     "error1@test.com",
			FirstName: "Error1",
			LastName:  "User",
			IsActive:  true,
		}

		// Any error returned here will cause automatic rollback
		return tx.Create(user).Error
	})

	if err != nil {
		// Handle transaction error
		fmt.Printf("Transaction failed: %v\n", err)
		return err
	}

	// Pattern 2: Retry on specific errors
	err = te.tm.WithRetryableTransaction(func(tx *gorm.DB) error {
		user := &UserModel{
			Email:     "error2@test.com",
			FirstName: "Error2",
			LastName:  "User",
			IsActive:  true,
		}

		// This will be retried automatically on deadlocks, serialization failures, etc.
		return tx.Create(user).Error
	}, 3)

	if err != nil {
		// This error occurred after all retries were exhausted
		fmt.Printf("Transaction failed after retries: %v\n", err)
		return err
	}

	// Pattern 3: Partial rollback with savepoints
	return te.tm.WithTransaction(func(tx *gorm.DB) error {
		// Create main record
		mainUser := &UserModel{
			Email:     "error_main@test.com",
			FirstName: "ErrorMain",
			LastName:  "User",
			IsActive:  true,
		}

		if err := tx.Create(mainUser).Error; err != nil {
			return err // Full rollback
		}

		// Try to create optional record with savepoint
		err := te.tm.WithSavepoint(tx, "optional_record", func(spTx *gorm.DB) error {
			optionalUser := &UserModel{
				Email:     "error_optional@test.com",
				FirstName: "ErrorOptional",
				LastName:  "User",
				IsActive:  true,
			}

			return spTx.Create(optionalUser).Error
		})

		if err != nil {
			// Optional record creation failed, but main record is still there
			fmt.Printf("Optional record creation failed: %v\n", err)
			// Continue with transaction - don't return error
		}

		// Transaction will commit with main record only
		return nil
	})
}
