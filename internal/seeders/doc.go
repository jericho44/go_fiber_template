// Package seeders provides a comprehensive database seeding system for the Go Fiber Template.
//
// This package implements a flexible and extensible seeding framework that allows you to:
//   - Define custom seeders with dependencies
//   - Run seeders individually or in batches
//   - Track seeder execution history
//   - Rollback seeded data
//   - Manage different seeder sets for different environments
//
// # Basic Usage
//
// To run all seeders:
//
//	go run cmd/seed/main.go run --all
//
// To run specific seeders:
//
//	go run cmd/seed/main.go run user_seeder demo_data_seeder
//
// To run environment-appropriate seeders:
//
//	go run cmd/seed/main.go run --env
//
// To check seeder status:
//
//	go run cmd/seed/main.go status
//
// To rollback seeders:
//
//	go run cmd/seed/main.go rollback user_seeder
//
// # Creating Custom Seeders
//
// To create a custom seeder, implement the Seeder interface or extend BaseSeeder:
//
//	type MySeeder struct {
//		*BaseSeeder
//	}
//
//	func NewMySeeder() *MySeeder {
//		return &MySeeder{
//			BaseSeeder: NewBaseSeeder(
//				"my_seeder",
//				"Description of what this seeder does",
//				"dependency_seeder", // Optional dependencies
//			),
//		}
//	}
//
//	func (s *MySeeder) Seed(ctx context.Context, db *gorm.DB) error {
//		// Your seeding logic here
//		return nil
//	}
//
// Then register it in the Registry:
//
//	func (r *Registry) registerAllSeeders() {
//		r.manager.RegisterSeeder(NewMySeeder())
//		// ... other seeders
//	}
//
// # Seeder Dependencies
//
// Seeders can declare dependencies on other seeders. The system will automatically
// resolve and execute seeders in the correct order:
//
//	func NewMySeeder() *MySeeder {
//		return &MySeeder{
//			BaseSeeder: NewBaseSeeder(
//				"my_seeder",
//				"My custom seeder",
//				"user_seeder", "other_dependency", // Dependencies
//			),
//		}
//	}
//
// # Environment-Specific Seeders
//
// The system supports different seeders for different environments. By default:
//   - user_seeder: Runs in all environments
//   - demo_data_seeder: Runs only in development
//
// You can customize this behavior in the Registry.GetSeedersByEnvironment() method.
//
// # Execution Tracking
//
// The system automatically tracks seeder executions in the seeder_executions table.
// This prevents duplicate executions and provides execution history.
//
// # Error Handling
//
// All seeder operations are wrapped in database transactions. If a seeder fails,
// the transaction is rolled back and the error is recorded in the execution history.
//
// # Best Practices
//
//   - Keep seeders idempotent - they should be safe to run multiple times
//   - Use dependencies to ensure proper execution order
//   - Implement rollback methods for data cleanup
//   - Use environment-specific logic when appropriate
//   - Test your seeders thoroughly before deployment
package seeders