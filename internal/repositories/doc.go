// Package repositories provides the repository pattern implementation for data access operations.
//
// This package defines interfaces for data access operations and provides a clean abstraction
// layer between the business logic and database operations. The repository pattern helps maintain
// separation of concerns and makes the codebase more testable.
//
// Key Components:
//
// BaseRepository[T]: Generic interface defining common CRUD operations for all entities.
// All repository interfaces should embed this interface to ensure consistency.
//
// UserRepository: Specific interface for user-related database operations, extending
// BaseRepository with user-specific methods like GetByEmail, GetActiveUsers, etc.
//
// TransactionManager: Interface for managing database transactions, including support
// for nested transactions, savepoints, deadlock detection, and retry mechanisms.
//
// Future Entity Repositories: Interfaces for TokenBlacklist, Profile, and AuditLog
// repositories that can be implemented as the application grows.
//
// Usage Example:
//
//	// Inject repository into service
//	type UserService struct {
//		userRepo UserRepository
//		txManager TransactionManager
//	}
//
//	// Use repository in service method
//	func (s *UserService) CreateUser(userData CreateUserRequest) error {
//		return s.txManager.WithTransaction(func(tx *gorm.DB) error {
//			user := &models.User{
//				Email:     userData.Email,
//				FirstName: userData.FirstName,
//				LastName:  userData.LastName,
//			}
//			return s.userRepo.CreateTx(tx, user)
//		})
//	}
//
// All repository interfaces support three variants of operations:
//
// 1. Standard operations: Direct database operations
// 2. Transactional operations: Operations within an existing transaction
// 3. Context-aware operations: Operations with context for timeout/cancellation
//
// This design ensures flexibility and allows for proper transaction management
// across complex business operations.
package repositories
