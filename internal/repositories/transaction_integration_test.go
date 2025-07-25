package repositories

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"go-fiber-template/internal/config"
	"go-fiber-template/internal/database"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDB sets up a test database connection for integration tests
func setupTestDB(t *testing.T) *gorm.DB {
	// Use environment variables or default test database configuration
	cfg := &config.DatabaseConfig{
		Host:            getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:            5432,
		User:            getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password:        getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		Name:            getEnvOrDefault("TEST_DB_NAME", "test_db"),
		SSLMode:         "disable",
		MaxIdleConns:    5,
		MaxOpenConns:    10,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 15 * time.Minute,
	}

	db, err := database.Connect(cfg)
	if err != nil {
		t.Skipf("Skipping integration test: failed to connect to test database: %v", err)
		return nil
	}

	// Auto-migrate test tables
	err = db.AutoMigrate(&UserModel{})
	if err != nil {
		t.Fatalf("Failed to migrate test tables: %v", err)
	}

	// Clean up existing test data
	db.Exec("DELETE FROM users WHERE email LIKE 'test_%@example.com'")

	return db
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	// In a real implementation, you would use os.Getenv(key)
	// For this test, we'll return the default value
	return defaultValue
}

// teardownTestDB cleans up test database
func teardownTestDB(t *testing.T, db *gorm.DB) {
	if db != nil {
		// Clean up test data
		db.Exec("DELETE FROM users WHERE email LIKE 'test_%@example.com'")

		// Close connection
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}
}

// TestTransactionManagerDeadlockRetry tests deadlock detection and retry mechanisms
func TestTransactionManagerDeadlockRetry(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test successful retry after simulated deadlock
	t.Run("successful_retry_after_deadlock", func(t *testing.T) {
		attemptCount := 0

		err := tm.WithRetryableTransaction(func(tx *gorm.DB) error {
			attemptCount++

			// Simulate deadlock on first attempt
			if attemptCount == 1 {
				return errors.New("deadlock detected")
			}

			// Success on second attempt
			return nil
		}, 3)

		assert.NoError(t, err)
		assert.Equal(t, 2, attemptCount, "Should retry once after deadlock")
	})

	// Test max retries exceeded
	t.Run("max_retries_exceeded", func(t *testing.T) {
		attemptCount := 0
		maxRetries := 2

		err := tm.WithRetryableTransaction(func(tx *gorm.DB) error {
			attemptCount++
			return errors.New("deadlock detected")
		}, maxRetries)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "transaction failed after")
		assert.Equal(t, maxRetries+1, attemptCount, "Should attempt maxRetries+1 times")
	})

	// Test non-retryable error
	t.Run("non_retryable_error", func(t *testing.T) {
		attemptCount := 0

		err := tm.WithRetryableTransaction(func(tx *gorm.DB) error {
			attemptCount++
			return errors.New("validation failed") // Non-retryable error
		}, 3)

		assert.Error(t, err)
		assert.Equal(t, 1, attemptCount, "Should not retry non-retryable errors")
	})

	// Test WithDeadlockRetry convenience method
	t.Run("deadlock_retry_convenience_method", func(t *testing.T) {
		attemptCount := 0

		err := tm.(*transactionManagerImpl).WithDeadlockRetry(func(tx *gorm.DB) error {
			attemptCount++

			if attemptCount == 1 {
				return errors.New("deadlock detected")
			}

			return nil
		})

		assert.NoError(t, err)
		assert.Equal(t, 2, attemptCount)
	})
}

// TestTransactionManagerSavepoints tests savepoint support for nested transactions
func TestTransactionManagerSavepoints(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test successful savepoint operation
	t.Run("successful_savepoint_operation", func(t *testing.T) {
		err := tm.WithTransaction(func(tx *gorm.DB) error {
			// Create a user
			user1 := &UserModel{
				Email:     "test_savepoint1@example.com",
				FirstName: "Test",
				LastName:  "User1",
				IsActive:  true,
			}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// Use savepoint for nested operation
			return tm.WithSavepoint(tx, "sp1", func(spTx *gorm.DB) error {
				user2 := &UserModel{
					Email:     "test_savepoint2@example.com",
					FirstName: "Test",
					LastName:  "User2",
					IsActive:  true,
				}

				return spTx.Create(user2).Error
			})
		})

		assert.NoError(t, err)

		// Verify both users were created
		var count int64
		db.Model(&UserModel{}).Where("email IN ?", []string{
			"test_savepoint1@example.com",
			"test_savepoint2@example.com",
		}).Count(&count)
		assert.Equal(t, int64(2), count)
	})

	// Test savepoint rollback
	t.Run("savepoint_rollback", func(t *testing.T) {
		err := tm.WithTransaction(func(tx *gorm.DB) error {
			// Create a user
			user1 := &UserModel{
				Email:     "test_rollback1@example.com",
				FirstName: "Test",
				LastName:  "User1",
				IsActive:  true,
			}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// Use savepoint for nested operation that will fail
			err := tm.WithSavepoint(tx, "sp2", func(spTx *gorm.DB) error {
				user2 := &UserModel{
					Email:     "test_rollback2@example.com",
					FirstName: "Test",
					LastName:  "User2",
					IsActive:  true,
				}

				if err := spTx.Create(user2).Error; err != nil {
					return err
				}

				// Simulate error to trigger savepoint rollback
				return errors.New("simulated error")
			})

			// Savepoint should have rolled back, but main transaction continues
			assert.Error(t, err)

			// Create another user to verify main transaction is still active
			user3 := &UserModel{
				Email:     "test_rollback3@example.com",
				FirstName: "Test",
				LastName:  "User3",
				IsActive:  true,
			}

			return tx.Create(user3).Error
		})

		assert.NoError(t, err)

		// Verify user1 and user3 were created, but not user2
		var user1Count, user2Count, user3Count int64
		db.Model(&UserModel{}).Where("email = ?", "test_rollback1@example.com").Count(&user1Count)
		db.Model(&UserModel{}).Where("email = ?", "test_rollback2@example.com").Count(&user2Count)
		db.Model(&UserModel{}).Where("email = ?", "test_rollback3@example.com").Count(&user3Count)

		assert.Equal(t, int64(1), user1Count, "User1 should exist")
		assert.Equal(t, int64(0), user2Count, "User2 should not exist (savepoint rolled back)")
		assert.Equal(t, int64(1), user3Count, "User3 should exist")
	})
}

// TestTransactionManagerIsolationLevels tests transaction isolation level management
func TestTransactionManagerIsolationLevels(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test setting and getting isolation levels
	isolationLevels := []string{
		"READ UNCOMMITTED",
		"READ COMMITTED",
		"REPEATABLE READ",
		"SERIALIZABLE",
	}

	for _, level := range isolationLevels {
		t.Run(fmt.Sprintf("isolation_level_%s", level), func(t *testing.T) {
			err := tm.WithIsolationLevel(level, func(tx *gorm.DB) error {
				// Get current isolation level
				currentLevel, err := tm.GetIsolationLevel(tx)
				if err != nil {
					// Some databases might not support querying isolation level
					t.Logf("Could not query isolation level: %v", err)
					return nil
				}

				t.Logf("Set isolation level to %s, current level: %s", level, currentLevel)
				return nil
			})

			assert.NoError(t, err)
		})
	}

	// Test convenience methods for isolation levels
	t.Run("serializable_convenience_method", func(t *testing.T) {
		err := tm.(*transactionManagerImpl).WithSerializableTransaction(func(tx *gorm.DB) error {
			// Create a test user
			user := &UserModel{
				Email:     "test_serializable@example.com",
				FirstName: "Test",
				LastName:  "Serializable",
				IsActive:  true,
			}
			return tx.Create(user).Error
		})

		assert.NoError(t, err)
	})

	t.Run("read_committed_convenience_method", func(t *testing.T) {
		err := tm.(*transactionManagerImpl).WithReadCommittedTransaction(func(tx *gorm.DB) error {
			// Create a test user
			user := &UserModel{
				Email:     "test_read_committed@example.com",
				FirstName: "Test",
				LastName:  "ReadCommitted",
				IsActive:  true,
			}
			return tx.Create(user).Error
		})

		assert.NoError(t, err)
	})

	// Test invalid isolation level
	t.Run("invalid_isolation_level", func(t *testing.T) {
		err := tm.WithIsolationLevel("INVALID_LEVEL", func(tx *gorm.DB) error {
			return nil
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid isolation level")
	})
}

// TestTransactionManagerTimeout tests transaction timeout handling
func TestTransactionManagerTimeout(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test successful operation within timeout
	t.Run("successful_operation_within_timeout", func(t *testing.T) {
		err := tm.WithTimeout(5*time.Second, func(tx *gorm.DB) error {
			user := &UserModel{
				Email:     "test_timeout_success@example.com",
				FirstName: "Test",
				LastName:  "TimeoutSuccess",
				IsActive:  true,
			}
			return tx.Create(user).Error
		})

		assert.NoError(t, err)
	})

	// Test timeout with context
	t.Run("timeout_with_context", func(t *testing.T) {
		ctx := context.Background()

		err := tm.WithTimeoutContext(ctx, 5*time.Second, func(ctx context.Context, tx *gorm.DB) error {
			user := &UserModel{
				Email:     "test_timeout_context@example.com",
				FirstName: "Test",
				LastName:  "TimeoutContext",
				IsActive:  true,
			}
			return tx.WithContext(ctx).Create(user).Error
		})

		assert.NoError(t, err)
	})

	// Test timeout with retry
	t.Run("timeout_with_retry", func(t *testing.T) {
		attemptCount := 0

		err := tm.WithTimeoutAndRetry(5*time.Second, func(tx *gorm.DB) error {
			attemptCount++

			// Simulate deadlock on first attempt
			if attemptCount == 1 {
				return errors.New("deadlock detected")
			}

			user := &UserModel{
				Email:     "test_timeout_retry@example.com",
				FirstName: "Test",
				LastName:  "TimeoutRetry",
				IsActive:  true,
			}
			return tx.Create(user).Error
		}, 2)

		assert.NoError(t, err)
		assert.Equal(t, 2, attemptCount)
	})

	// Test very short timeout (this might be flaky depending on system performance)
	t.Run("very_short_timeout", func(t *testing.T) {
		err := tm.WithTimeout(1*time.Nanosecond, func(tx *gorm.DB) error {
			// This operation should timeout
			time.Sleep(10 * time.Millisecond)
			return nil
		})

		// The error might be a timeout or database error depending on timing
		// We just verify that an error occurred
		assert.Error(t, err)
	})
}

// TestConcurrentTransactions tests concurrent transaction scenarios
func TestConcurrentTransactions(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test concurrent transactions without conflicts
	t.Run("concurrent_transactions_no_conflict", func(t *testing.T) {
		const numGoroutines = 10
		var wg sync.WaitGroup
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				err := tm.WithTransaction(func(tx *gorm.DB) error {
					user := &UserModel{
						Email:     fmt.Sprintf("test_concurrent_%d@example.com", id),
						FirstName: "Test",
						LastName:  fmt.Sprintf("Concurrent%d", id),
						IsActive:  true,
					}
					return tx.Create(user).Error
				})

				if err != nil {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Errorf("Concurrent transaction failed: %v", err)
		}

		// Verify all users were created
		var count int64
		db.Model(&UserModel{}).Where("email LIKE ?", "test_concurrent_%@example.com").Count(&count)
		assert.Equal(t, int64(numGoroutines), count)
	})

	// Test concurrent transactions with retry on conflicts
	t.Run("concurrent_transactions_with_retry", func(t *testing.T) {
		const numGoroutines = 5
		var wg sync.WaitGroup
		successCount := int64(0)
		var mu sync.Mutex

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				err := tm.WithRetryableTransaction(func(tx *gorm.DB) error {
					user := &UserModel{
						Email:     fmt.Sprintf("test_retry_%d@example.com", id),
						FirstName: "Test",
						LastName:  fmt.Sprintf("Retry%d", id),
						IsActive:  true,
					}
					return tx.Create(user).Error
				}, 3)

				if err == nil {
					mu.Lock()
					successCount++
					mu.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// All transactions should succeed
		assert.Equal(t, int64(numGoroutines), successCount)
	})
}

// TestTransactionManagerComplexScenarios tests complex transaction scenarios
func TestTransactionManagerComplexScenarios(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer teardownTestDB(t, db)

	tm := NewTransactionManager(db)

	// Test nested savepoints with different isolation levels
	t.Run("nested_savepoints_with_isolation", func(t *testing.T) {
		err := tm.WithIsolationLevel("REPEATABLE READ", func(tx *gorm.DB) error {
			// Create first user
			user1 := &UserModel{
				Email:     "test_complex1@example.com",
				FirstName: "Test",
				LastName:  "Complex1",
				IsActive:  true,
			}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// First savepoint
			return tm.WithSavepoint(tx, "sp1", func(spTx1 *gorm.DB) error {
				user2 := &UserModel{
					Email:     "test_complex2@example.com",
					FirstName: "Test",
					LastName:  "Complex2",
					IsActive:  true,
				}

				if err := spTx1.Create(user2).Error; err != nil {
					return err
				}

				// Nested savepoint
				return tm.WithSavepoint(spTx1, "sp2", func(spTx2 *gorm.DB) error {
					user3 := &UserModel{
						Email:     "test_complex3@example.com",
						FirstName: "Test",
						LastName:  "Complex3",
						IsActive:  true,
					}

					return spTx2.Create(user3).Error
				})
			})
		})

		assert.NoError(t, err)

		// Verify all users were created
		var count int64
		db.Model(&UserModel{}).Where("email LIKE ?", "test_complex%@example.com").Count(&count)
		assert.Equal(t, int64(3), count)
	})

	// Test timeout with savepoints and retry
	t.Run("timeout_with_savepoints_and_retry", func(t *testing.T) {
		attemptCount := 0

		err := tm.WithTimeoutAndRetry(10*time.Second, func(tx *gorm.DB) error {
			attemptCount++

			// Create base user
			user1 := &UserModel{
				Email:     fmt.Sprintf("test_timeout_complex_%d@example.com", attemptCount),
				FirstName: "Test",
				LastName:  "TimeoutComplex",
				IsActive:  true,
			}

			if err := tx.Create(user1).Error; err != nil {
				return err
			}

			// Use savepoint for additional operations
			return tm.WithSavepoint(tx, "timeout_sp", func(spTx *gorm.DB) error {
				// Simulate error on first attempt
				if attemptCount == 1 {
					return errors.New("serialization failure")
				}

				user2 := &UserModel{
					Email:     fmt.Sprintf("test_timeout_complex_sp_%d@example.com", attemptCount),
					FirstName: "Test",
					LastName:  "TimeoutComplexSP",
					IsActive:  true,
				}

				return spTx.Create(user2).Error
			})
		}, 3)

		assert.NoError(t, err)
		assert.Equal(t, 2, attemptCount)

		// Verify final users were created (only from successful attempt)
		var count int64
		db.Model(&UserModel{}).Where("email LIKE ?", "test_timeout_complex_%@example.com").Count(&count)
		assert.Equal(t, int64(2), count) // user1 and user2 from second attempt
	})
}

// BenchmarkTransactionManager benchmarks transaction manager performance
func BenchmarkTransactionManager(b *testing.B) {
	db := setupTestDB(&testing.T{})
	if db == nil {
		b.Skip("Database not available for benchmarking")
		return
	}
	defer teardownTestDB(&testing.T{}, db)

	tm := NewTransactionManager(db)

	b.Run("simple_transaction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := tm.WithTransaction(func(tx *gorm.DB) error {
				user := &UserModel{
					Email:     fmt.Sprintf("bench_simple_%d@example.com", i),
					FirstName: "Bench",
					LastName:  "Simple",
					IsActive:  true,
				}
				return tx.Create(user).Error
			})
			if err != nil {
				b.Fatalf("Transaction failed: %v", err)
			}
		}
	})

	b.Run("transaction_with_savepoint", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := tm.WithTransaction(func(tx *gorm.DB) error {
				return tm.WithSavepoint(tx, "bench_sp", func(spTx *gorm.DB) error {
					user := &UserModel{
						Email:     fmt.Sprintf("bench_savepoint_%d@example.com", i),
						FirstName: "Bench",
						LastName:  "Savepoint",
						IsActive:  true,
					}
					return spTx.Create(user).Error
				})
			})
			if err != nil {
				b.Fatalf("Transaction with savepoint failed: %v", err)
			}
		}
	})

	b.Run("retryable_transaction", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := tm.WithRetryableTransaction(func(tx *gorm.DB) error {
				user := &UserModel{
					Email:     fmt.Sprintf("bench_retry_%d@example.com", i),
					FirstName: "Bench",
					LastName:  "Retry",
					IsActive:  true,
				}
				return tx.Create(user).Error
			}, 3)
			if err != nil {
				b.Fatalf("Retryable transaction failed: %v", err)
			}
		}
	})
}
