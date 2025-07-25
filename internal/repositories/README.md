# Repository Pattern Documentation

## Overview

This package implements the Repository pattern to provide a clean abstraction layer between the business logic and data access logic. The repository pattern helps maintain separation of concerns and makes the codebase more testable by allowing easy mocking of data access operations.

## Architecture

The repository layer is structured with the following components:

1. **Base Repository Interface**: Defines common CRUD operations that all repositories should implement
2. **Entity-Specific Repositories**: Define operations specific to each entity (User, TokenBlacklist, etc.)
3. **Transaction Manager**: Handles database transactions, rollbacks, and advanced transaction features
4. **Concrete Implementations**: GORM-based implementations of the repository interfaces

## Interface Contracts

### BaseRepository[T any]

The `BaseRepository` interface defines the contract for basic CRUD operations that all repositories must implement:

#### Standard Operations

- `Create(entity *T) error`: Creates a new entity
- `GetByID(id uint) (*T, error)`: Retrieves an entity by its ID
- `Update(entity *T) error`: Updates an existing entity
- `Delete(id uint) error`: Deletes an entity (soft delete if supported)
- `List(limit, offset int) ([]*T, error)`: Lists entities with pagination
- `Count() (int64, error)`: Counts total entities

#### Transactional Operations

All repositories must provide transactional variants of CRUD operations:

- `CreateTx(tx *gorm.DB, entity *T) error`
- `UpdateTx(tx *gorm.DB, entity *T) error`
- `DeleteTx(tx *gorm.DB, id uint) error`

#### Context-Aware Operations

For operations that need context propagation (timeouts, cancellation):

- `CreateWithContext(ctx context.Context, entity *T) error`
- `GetByIDWithContext(ctx context.Context, id uint) (*T, error)`
- `UpdateWithContext(ctx context.Context, entity *T) error`
- `DeleteWithContext(ctx context.Context, id uint) error`
- `ListWithContext(ctx context.Context, limit, offset int) ([]*T, error)`
- `CountWithContext(ctx context.Context) (int64, error)`

### UserRepository

The `UserRepository` extends `BaseRepository[User]` with user-specific operations:

#### Query Operations

- `GetByEmail(email string) (User, error)`: Find user by email address
- `GetActiveUsers(limit, offset int) ([]User, error)`: Get only active users
- `ExistsByEmail(email string) (bool, error)`: Check if user exists by email

#### Bulk Operations

- `CreateBatch(users []User) error`: Create multiple users in a single operation
- `UpdateBatch(users []User) error`: Update multiple users in a single operation

#### Soft Delete Operations

- `SoftDelete(id uint) error`: Soft delete a user (sets DeletedAt)
- `Restore(id uint) error`: Restore a soft-deleted user
- `GetDeletedUsers(limit, offset int) ([]User, error)`: Get soft-deleted users
- `HardDelete(id uint) error`: Permanently delete a user

#### Search Operations

- `SearchByName(firstName, lastName string, limit, offset int) ([]User, error)`: Search users by name

### TransactionManager

The `TransactionManager` interface provides comprehensive transaction management:

#### Basic Transaction Operations

- `WithTransaction(fn func(*gorm.DB) error) error`: Execute function within a transaction
- `WithTransactionContext(ctx context.Context, fn func(context.Context, *gorm.DB) error) error`: Transaction with context

#### Manual Transaction Control

- `BeginTransaction() (*gorm.DB, error)`: Start a new transaction
- `CommitTransaction(tx *gorm.DB) error`: Commit a transaction
- `RollbackTransaction(tx *gorm.DB) error`: Rollback a transaction

#### Advanced Transaction Features

- `WithSavepoint(tx *gorm.DB, name string, fn func(*gorm.DB) error) error`: Use savepoints for nested transactions
- `WithRetryableTransaction(fn func(*gorm.DB) error, maxRetries int) error`: Retry transactions on deadlock

## Usage Examples

### Basic Repository Usage

```go
// Create a new user
user := &models.User{
    Email:     "user@example.com",
    FirstName: "John",
    LastName:  "Doe",
}
err := userRepo.Create(user)

// Get user by ID
user, err := userRepo.GetByID(1)

// Get user by email
user, err := userRepo.GetByEmail("user@example.com")

// Update user
user.FirstName = "Jane"
err = userRepo.Update(user)

// List users with pagination
users, err := userRepo.List(10, 0) // limit=10, offset=0
```

### Transaction Usage

```go
// Simple transaction
err := txManager.WithTransaction(func(tx *gorm.DB) error {
    // Create user
    if err := userRepo.CreateTx(tx, user); err != nil {
        return err
    }

    // Create user profile
    if err := profileRepo.CreateTx(tx, profile); err != nil {
        return err
    }

    return nil
})

// Transaction with context
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := txManager.WithTransactionContext(ctx, func(ctx context.Context, tx *gorm.DB) error {
    // Operations with timeout context
    return userRepo.CreateTx(tx, user)
})
```

### Retryable Transactions

```go
// Retry transaction on deadlock (up to 3 times)
err := txManager.WithRetryableTransaction(func(tx *gorm.DB) error {
    // Operations that might cause deadlock
    return userRepo.UpdateTx(tx, user)
}, 3)
```

### Savepoints (Nested Transactions)

```go
err := txManager.WithTransaction(func(tx *gorm.DB) error {
    // Create user
    if err := userRepo.CreateTx(tx, user); err != nil {
        return err
    }

    // Use savepoint for risky operation
    return txManager.WithSavepoint(tx, "profile_creation", func(tx *gorm.DB) error {
        // If this fails, only this part will be rolled back
        return profileRepo.CreateTx(tx, profile)
    })
})
```

## Error Handling

All repository methods should return appropriate errors:

- **Validation Errors**: When input data is invalid
- **Not Found Errors**: When requested entity doesn't exist
- **Constraint Errors**: When database constraints are violated
- **Transaction Errors**: When transaction operations fail

## Testing

Repository interfaces are designed to be easily mockable for unit testing:

```go
type MockUserRepository struct {
    users map[uint]*models.User
}

func (m *MockUserRepository) Create(user *models.User) error {
    // Mock implementation
}

func (m *MockUserRepository) GetByID(id uint) (*models.User, error) {
    // Mock implementation
}
```

## Best Practices

1. **Always use transactions** for operations that modify multiple entities
2. **Use context-aware methods** for operations that might take time
3. **Handle errors appropriately** and return meaningful error messages
4. **Use bulk operations** when working with multiple entities
5. **Implement proper logging** in concrete implementations
6. **Use soft deletes** when data retention is important
7. **Implement proper pagination** for list operations
8. **Use savepoints** for complex nested operations

## Future Extensions

The repository interfaces are designed to be extensible. When adding new entities:

1. Create a new interface extending `BaseRepository[T]`
2. Add entity-specific operations
3. Implement both standard and transactional variants
4. Add context-aware methods where appropriate
5. Document the interface contract
6. Create comprehensive unit tests

## Implementation Notes

- All concrete implementations should be in separate files (e.g., `user_repository_impl.go`)
- Use GORM for database operations in concrete implementations
- Implement proper error handling and logging
- Use dependency injection for database connections
- Follow the established naming conventions
- Ensure thread safety in concurrent environments
