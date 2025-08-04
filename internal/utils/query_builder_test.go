package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Test model for query builder tests
type TestUserQB struct {
	ID        uint   `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	FirstName string
	LastName  string
	Age       int
	IsActive  bool `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&TestUserQB{})
	require.NoError(t, err)

	// Seed test data
	users := []TestUserQB{
		{Email: "john@example.com", FirstName: "John", LastName: "Doe", Age: 30, IsActive: true},
		{Email: "jane@example.com", FirstName: "Jane", LastName: "Smith", Age: 25, IsActive: true},
		{Email: "bob@example.com", FirstName: "Bob", LastName: "Johnson", Age: 35, IsActive: false},
		{Email: "alice@example.com", FirstName: "Alice", LastName: "Brown", Age: 28, IsActive: true},
	}

	err = db.Create(&users).Error
	require.NoError(t, err)

	return db
}

func TestNewQueryBuilder(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	assert.NotNil(t, qb)
	assert.NotNil(t, qb.db)
	assert.NotNil(t, qb.query)
}

func TestQueryBuilder_Model(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	result := qb.Model(&TestUserQB{})
	assert.Equal(t, qb, result) // Should return self for chaining
}

func TestQueryBuilder_Select(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Select("id", "email", "first_name").
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 4)

	// Verify only selected fields are populated (others should be zero values)
	for _, user := range users {
		assert.NotZero(t, user.ID)
		assert.NotEmpty(t, user.Email)
		assert.NotEmpty(t, user.FirstName)
		// LastName should be empty since it wasn't selected
		assert.Empty(t, user.LastName)
	}
}

func TestQueryBuilder_Where(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Where("is_active = ?", true).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 3) // Only active users

	for _, user := range users {
		assert.True(t, user.IsActive)
	}
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	emails := []string{"john@example.com", "jane@example.com"}
	err := qb.Model(&TestUserQB{}).
		WhereIn("email", emails).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 2)

	emailSet := make(map[string]bool)
	for _, user := range users {
		emailSet[user.Email] = true
	}
	assert.True(t, emailSet["john@example.com"])
	assert.True(t, emailSet["jane@example.com"])
}

func TestQueryBuilder_WhereBetween(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		WhereBetween("age", 25, 30).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 3) // John (30), Jane (25), Alice (28)

	for _, user := range users {
		assert.GreaterOrEqual(t, user.Age, 25)
		assert.LessOrEqual(t, user.Age, 30)
	}
}

func TestQueryBuilder_WhereNull(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	// First, create a user with null last name
	user := TestUserQB{Email: "null@example.com", FirstName: "Null"}
	err := db.Create(&user).Error
	require.NoError(t, err)

	// Update to set last_name to NULL
	err = db.Model(&user).Update("last_name", nil).Error
	require.NoError(t, err)

	var users []TestUserQB
	err = qb.Model(&TestUserQB{}).
		WhereNull("last_name").
		Find(&users)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(users), 1)
}

func TestQueryBuilder_WhereLike(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		WhereLike("first_name", "J%").
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 2) // John and Jane

	for _, user := range users {
		assert.True(t, user.FirstName[0] == 'J')
	}
}

func TestQueryBuilder_OrderBy(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		OrderBy("age", "DESC").
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 4)

	// Should be ordered by age descending: Bob (35), John (30), Alice (28), Jane (25)
	assert.Equal(t, 35, users[0].Age)
	assert.Equal(t, 30, users[1].Age)
	assert.Equal(t, 28, users[2].Age)
	assert.Equal(t, 25, users[3].Age)
}

func TestQueryBuilder_Limit(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Limit(2).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
}

func TestQueryBuilder_Count(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var count int64
	err := qb.Model(&TestUserQB{}).
		Where("is_active = ?", true).
		Count(&count)

	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestQueryBuilder_Create(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	user := TestUserQB{
		Email:     "new@example.com",
		FirstName: "New",
		LastName:  "User",
		Age:       40,
		IsActive:  true,
	}

	err := qb.Create(&user)
	assert.NoError(t, err)
	assert.NotZero(t, user.ID)

	// Verify user was created
	var found TestUserQB
	err = db.Where("email = ?", "new@example.com").First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "New", found.FirstName)
}

func TestQueryBuilder_Update(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	err := qb.Model(&TestUserQB{}).
		Where("email = ?", "john@example.com").
		Update("age", 31)

	assert.NoError(t, err)

	// Verify update
	var user TestUserQB
	err = db.Where("email = ?", "john@example.com").First(&user).Error
	assert.NoError(t, err)
	assert.Equal(t, 31, user.Age)
}

func TestQueryBuilder_Delete(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	err := qb.Model(&TestUserQB{}).
		Where("email = ?", "bob@example.com").
		Delete(&TestUserQB{})

	assert.NoError(t, err)

	// Verify deletion
	var count int64
	err = db.Model(&TestUserQB{}).Where("email = ?", "bob@example.com").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestQueryBuilder_Chaining(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Where("is_active = ?", true).
		Where("age >= ?", 25).
		OrderBy("age").
		Limit(2).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.True(t, users[0].IsActive)
	assert.True(t, users[1].IsActive)
	assert.GreaterOrEqual(t, users[0].Age, 25)
	assert.GreaterOrEqual(t, users[1].Age, 25)
	assert.LessOrEqual(t, users[0].Age, users[1].Age) // Should be ordered
}

func TestQueryBuilder_Reset(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	// Build a query
	qb.Model(&TestUserQB{}).Where("is_active = ?", true)

	// Reset
	qb.Reset()

	// Build a different query
	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 4) // Should get all users, not just active ones
}

// Security tests

func TestSanitizeIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user_id", "user_id"},
		{"user.id", "user.id"},
		{"user-id", "userid"},  // Hyphens removed
		{"user id", "userid"},  // Spaces removed
		{"user@id", "userid"},  // Special chars removed
		{"", "id"},             // Empty becomes default
		{"SELECT", "`SELECT`"}, // SQL keyword gets quoted
		{"drop", "`drop`"},     // SQL keyword gets quoted
	}

	for _, test := range tests {
		result := sanitizeIdentifier(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestSanitizeDirection(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ASC", "ASC"},
		{"DESC", "DESC"},
		{"asc", "ASC"},
		{"desc", "DESC"},
		{"invalid", "ASC"}, // Invalid becomes ASC
		{"", "ASC"},        // Empty becomes ASC
	}

	for _, test := range tests {
		result := sanitizeDirection(test.input)
		assert.Equal(t, test.expected, result, "Input: %s", test.input)
	}
}

func TestValidateWhereCondition(t *testing.T) {
	tests := []struct {
		condition string
		shouldErr bool
	}{
		{"id = ?", false},
		{"name LIKE ?", false},
		{"age BETWEEN ? AND ?", false},
		{"DROP TABLE users", true},
		{"DELETE FROM users", true},
		{"UNION SELECT * FROM passwords", true},
		{"id = 1; DROP TABLE users", true},
		{"id = 1 -- comment", true},
		{"id = 1 /* comment */", true},
	}

	for _, test := range tests {
		err := validateWhereCondition(test.condition)
		if test.shouldErr {
			assert.Error(t, err, "Condition: %s", test.condition)
		} else {
			assert.NoError(t, err, "Condition: %s", test.condition)
		}
	}
}

func TestQueryBuilder_SecurityValidation(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	// Test that dangerous WHERE conditions result in empty results
	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Where("DROP TABLE users").
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 0) // Should return empty result set due to security validation
}

func TestQueryBuilder_Pagination(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	opts := PaginationOptions{
		Limit:     2,
		SortField: "id",
		SortDir:   "ASC",
	}

	var users []TestUserQB
	err := qb.Model(&TestUserQB{}).
		Paginate(opts).
		Find(&users)

	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.LessOrEqual(t, users[0].ID, users[1].ID) // Should be ordered by ID ASC
}

func TestBatchOperation(t *testing.T) {
	db := setupTestDB(t)
	qb := NewQueryBuilder(db)

	batch := qb.NewBatchOperation(2)

	// Add items to batch
	user1 := TestUserQB{Email: "batch1@example.com", FirstName: "Batch1"}
	user2 := TestUserQB{Email: "batch2@example.com", FirstName: "Batch2"}
	user3 := TestUserQB{Email: "batch3@example.com", FirstName: "Batch3"}

	err := batch.Add(user1)
	assert.NoError(t, err)

	err = batch.Add(user2)
	assert.NoError(t, err) // Should trigger execution due to batch size

	err = batch.Add(user3)
	assert.NoError(t, err)

	err = batch.Flush() // Flush remaining items
	assert.NoError(t, err)

	// Verify all users were created
	var count int64
	err = db.Model(&TestUserQB{}).Where("email LIKE ?", "batch%").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestIsSQLKeyword(t *testing.T) {
	tests := []struct {
		word     string
		expected bool
	}{
		{"SELECT", true},
		{"DROP", true},
		{"user_id", false},
		{"name", false},
		{"TABLE", true},
		{"table", false}, // Case sensitive
	}

	for _, test := range tests {
		result := isSQLKeyword(test.word)
		assert.Equal(t, test.expected, result, "Word: %s", test.word)
	}
}

// Benchmark tests
func BenchmarkQueryBuilder_SimpleQuery(b *testing.B) {
	db := setupTestDB(&testing.T{})
	qb := NewQueryBuilder(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []TestUserQB
		qb.Reset().Model(&TestUserQB{}).Where("is_active = ?", true).Find(&users)
	}
}

func BenchmarkQueryBuilder_ComplexQuery(b *testing.B) {
	db := setupTestDB(&testing.T{})
	qb := NewQueryBuilder(db)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []TestUserQB
		qb.Reset().Model(&TestUserQB{}).
			Where("is_active = ?", true).
			Where("age >= ?", 25).
			OrderBy("age", "DESC").
			Limit(10).
			Find(&users)
	}
}
