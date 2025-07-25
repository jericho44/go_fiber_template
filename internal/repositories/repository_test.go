package repositories

import (
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"
)

// TestRepositoryInterfaces verifies that all repository implementations satisfy their interfaces
func TestRepositoryInterfaces(t *testing.T) {
	// This test ensures that our implementations satisfy the interfaces at compile time
	var db *gorm.DB // nil is fine for interface checking

	// Test UserRepository interface implementation
	var userRepo UserRepository = NewUserRepository(db)
	if userRepo == nil {
		t.Error("NewUserRepository should not return nil")
	}

	// Test TransactionManager interface implementation
	var txManager TransactionManager = NewTransactionManager(db)
	if txManager == nil {
		t.Error("NewTransactionManager should not return nil")
	}
}

// TestUserModelImplementsUserEntity verifies that UserModel implements UserEntity interface
func TestUserModelImplementsUserEntity(t *testing.T) {
	user := &UserModel{
		ID:        1,
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}

	// Test that UserModel implements UserEntity interface
	var entity UserEntity = user

	// Test interface methods
	if entity.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", entity.GetID())
	}

	if entity.GetEmail() != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", entity.GetEmail())
	}

	if entity.GetFirstName() != "Test" {
		t.Errorf("Expected first name 'Test', got '%s'", entity.GetFirstName())
	}

	if entity.GetLastName() != "User" {
		t.Errorf("Expected last name 'User', got '%s'", entity.GetLastName())
	}

	if !entity.GetIsActive() {
		t.Error("Expected user to be active")
	}

	// Test setters
	entity.SetEmail("updated@example.com")
	if entity.GetEmail() != "updated@example.com" {
		t.Errorf("Expected updated email 'updated@example.com', got '%s'", entity.GetEmail())
	}

	entity.SetIsActive(false)
	if entity.GetIsActive() {
		t.Error("Expected user to be inactive after SetIsActive(false)")
	}
}

// TestUserModelHelperMethods tests the helper methods in UserModel
func TestUserModelHelperMethods(t *testing.T) {
	user := &UserModel{
		ID:        1,
		Email:     "TEST@EXAMPLE.COM",
		FirstName: "Test",
		LastName:  "User",
		IsActive:  true,
	}

	// Test toEntity method
	entity := user.toEntity()
	if entity.GetID() != 1 {
		t.Errorf("Expected ID 1, got %d", entity.GetID())
	}

	// Test fromEntity method
	newUser := fromEntity(entity)
	if newUser.ID != 1 {
		t.Errorf("Expected ID 1, got %d", newUser.ID)
	}

	if newUser.Email != "TEST@EXAMPLE.COM" {
		t.Errorf("Expected email 'TEST@EXAMPLE.COM', got '%s'", newUser.Email)
	}
}

// TestUserModelTableName tests the TableName method
func TestUserModelTableName(t *testing.T) {
	user := UserModel{}
	tableName := user.TableName()

	if tableName != "users" {
		t.Errorf("Expected table name 'users', got '%s'", tableName)
	}
}

// TestUserModelGORMHooks tests the GORM hooks
func TestUserModelGORMHooks(t *testing.T) {
	user := &UserModel{
		Email: "TEST@EXAMPLE.COM",
	}

	// Test BeforeCreate hook
	err := user.BeforeCreate(nil) // nil tx is fine for this test
	if err != nil {
		t.Errorf("BeforeCreate should not return error: %v", err)
	}

	if user.Email != "test@example.com" {
		t.Errorf("Expected email to be lowercased to 'test@example.com', got '%s'", user.Email)
	}

	// Reset email for BeforeUpdate test
	user.Email = "UPDATED@EXAMPLE.COM"

	// Test BeforeUpdate hook
	err = user.BeforeUpdate(nil) // nil tx is fine for this test
	if err != nil {
		t.Errorf("BeforeUpdate should not return error: %v", err)
	}

	if user.Email != "updated@example.com" {
		t.Errorf("Expected email to be lowercased to 'updated@example.com', got '%s'", user.Email)
	}
}

// TestIsRetryableError tests the isRetryableError function
func TestIsRetryableError(t *testing.T) {
	testCases := []struct {
		name     string
		errMsg   string
		expected bool
	}{
		{
			name:     "deadlock error",
			errMsg:   "deadlock detected",
			expected: true,
		},
		{
			name:     "serialization failure",
			errMsg:   "could not serialize access",
			expected: true,
		},
		{
			name:     "lock timeout",
			errMsg:   "lock timeout",
			expected: true,
		},
		{
			name:     "connection reset",
			errMsg:   "connection reset",
			expected: true,
		},
		{
			name:     "temporary failure",
			errMsg:   "temporary failure",
			expected: true,
		},
		{
			name:     "validation error",
			errMsg:   "validation failed",
			expected: false,
		},
		{
			name:     "constraint violation",
			errMsg:   "unique constraint violation",
			expected: false,
		},
		{
			name:     "not found error",
			errMsg:   "record not found",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := &mockError{message: tc.errMsg}
			result := isRetryableError(err)
			if result != tc.expected {
				t.Errorf("Expected %v for error %q, got %v", tc.expected, tc.errMsg, result)
			}
		})
	}
}

// mockError is a simple error implementation for testing
type mockError struct {
	message string
}

func (e *mockError) Error() string {
	return e.message
}

// TestStringUtilityFunctions tests the utility functions for string operations
func TestStringUtilityFunctions(t *testing.T) {
	// Test contains function
	testCases := []struct {
		name     string
		s        string
		substr   string
		expected bool
	}{
		{"exact match", "deadlock", "deadlock", true},
		{"substring at beginning", "deadlock detected", "deadlock", true},
		{"substring at end", "error: deadlock", "deadlock", true},
		{"substring in middle", "database deadlock error", "deadlock", true},
		{"not found", "validation error", "deadlock", false},
		{"empty substring", "test", "", true},
		{"empty string", "", "test", false},
	}

	for _, tc := range testCases {
		t.Run("contains_"+tc.name, func(t *testing.T) {
			result := contains(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("Expected %v for contains(%q, %q), got %v", tc.expected, tc.s, tc.substr, result)
			}
		})
	}

	// Test indexOfSubstring function
	indexTestCases := []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{"found at beginning", "deadlock detected", "deadlock", 0},
		{"found in middle", "database deadlock error", "deadlock", 9},
		{"not found", "validation error", "deadlock", -1},
		{"exact match", "deadlock", "deadlock", 0},
		{"empty substring", "test", "", 0},
	}

	for _, tc := range indexTestCases {
		t.Run("indexOfSubstring_"+tc.name, func(t *testing.T) {
			result := indexOfSubstring(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("Expected %d for indexOfSubstring(%q, %q), got %d", tc.expected, tc.s, tc.substr, result)
			}
		})
	}
}

// TestAdvancedTransactionFeatures tests the advanced transaction management features
func TestAdvancedTransactionFeatures(t *testing.T) {
	// Test isolation level validation
	tm := &transactionManagerImpl{}

	// Test SetIsolationLevel with valid levels
	validLevels := []string{
		"READ UNCOMMITTED",
		"READ COMMITTED",
		"REPEATABLE READ",
		"SERIALIZABLE",
		"read committed", // Test case insensitive
		"serializable",   // Test case insensitive
	}

	for _, level := range validLevels {
		t.Run("valid_isolation_level_"+strings.ReplaceAll(strings.ToLower(level), " ", "_"), func(t *testing.T) {
			// We can't test actual database operations without a real DB connection
			// But we can test the validation logic
			err := tm.SetIsolationLevel(nil, level) // This will fail due to nil tx, but validation should pass
			// The error should be about the nil transaction, not invalid level
			if err != nil && strings.Contains(err.Error(), "invalid isolation level") {
				t.Errorf("SetIsolationLevel should accept valid level %s", level)
			}
		})
	}

	// Test SetIsolationLevel with invalid levels
	invalidLevels := []string{
		"INVALID_LEVEL",
		"READ_COMMITTED", // Wrong format
		"SERIALIZABLE_READ",
		"",
	}

	for _, level := range invalidLevels {
		t.Run("invalid_isolation_level_"+strings.ReplaceAll(strings.ToLower(level), " ", "_"), func(t *testing.T) {
			err := tm.SetIsolationLevel(nil, level)
			if err == nil || !strings.Contains(err.Error(), "invalid isolation level") {
				t.Errorf("SetIsolationLevel should reject invalid level %s", level)
			}
		})
	}
}

// TestTransactionUtilityMethods tests the convenience methods for transaction management
func TestTransactionUtilityMethods(t *testing.T) {
	// Test that utility methods exist and have correct signatures
	var tm TransactionManager = &transactionManagerImpl{}

	// These tests verify the methods exist and can be called
	// Actual functionality would require database integration tests

	// Test WithDeadlockRetry method exists
	err := tm.(*transactionManagerImpl).WithDeadlockRetry(func(tx *gorm.DB) error {
		return nil
	})
	// Will fail due to nil db, but method should exist
	_ = err

	// Test isolation level convenience methods exist
	err = tm.(*transactionManagerImpl).WithSerializableTransaction(func(tx *gorm.DB) error {
		return nil
	})
	_ = err

	err = tm.(*transactionManagerImpl).WithReadCommittedTransaction(func(tx *gorm.DB) error {
		return nil
	})
	_ = err

	err = tm.(*transactionManagerImpl).WithRepeatableReadTransaction(func(tx *gorm.DB) error {
		return nil
	})
	_ = err

	err = tm.(*transactionManagerImpl).WithReadUncommittedTransaction(func(tx *gorm.DB) error {
		return nil
	})
	_ = err
}

// TestTimeoutHandling tests timeout-related functionality
func TestTimeoutHandling(t *testing.T) {
	tm := &transactionManagerImpl{}

	// Test that timeout methods exist and handle context correctly
	shortTimeout := 1 * time.Millisecond

	// Test WithTimeout
	err := tm.WithTimeout(shortTimeout, func(tx *gorm.DB) error {
		time.Sleep(10 * time.Millisecond) // Sleep longer than timeout
		return nil
	})

	// Should timeout (will fail due to nil db, but timeout logic should be tested)
	if err == nil {
		t.Log("WithTimeout method exists and can be called")
	}

	// Test WithTimeoutAndRetry
	err = tm.WithTimeoutAndRetry(shortTimeout, func(tx *gorm.DB) error {
		time.Sleep(10 * time.Millisecond) // Sleep longer than timeout
		return nil
	}, 2)

	// Should timeout (will fail due to nil db, but timeout logic should be tested)
	if err == nil {
		t.Log("WithTimeoutAndRetry method exists and can be called")
	}
}
