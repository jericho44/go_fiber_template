# Go Fiber Template

A comprehensive, production-ready Go web application template with essential features for building modern web APIs. This template provides a solid foundation with authentication, database management, API documentation, and clean architecture patterns.

## ✨ Features

- 🚀 **Go Fiber** - Fast HTTP web framework with Express.js-like API
- 📚 **Swagger Documentation** - Auto-generated interactive API docs
- 🗄️ **Database Migrations** - Systematic schema management with golang-migrate
- 🔒 **JWT Authentication** - Secure user authentication with access/refresh tokens
- 🏗️ **Clean Architecture** - Repository pattern and layered design
- 📝 **Standardized Responses** - Consistent API response formatting
- 🔄 **Transaction Management** - Comprehensive database transaction support
- 🧪 **Testing Ready** - Unit and integration test structure with testcontainers
- 🛡️ **Middleware Stack** - CORS, rate limiting, logging, error handling
- 🐳 **Docker Support** - Containerization ready
- 📊 **Health Checks** - Database and application health monitoring

## 📁 Project Structure

```
go-fiber-template/
├── cmd/
│   ├── migrate/         # Database migration CLI
│   └── server/          # Application entrypoint
├── internal/
│   ├── config/          # Configuration management
│   ├── controllers/     # HTTP handlers and request validation
│   ├── database/        # Database connection and health checks
│   ├── middleware/      # Custom middleware (auth, CORS, logging, etc.)
│   ├── models/          # Data models with GORM tags
│   ├── repositories/    # Data access layer with transaction support
│   ├── services/        # Business logic and orchestration
│   └── utils/           # Utility functions (validation, responses, etc.)
├── migrations/          # Database migration files
├── docs/               # API documentation and examples
│   └── examples/       # API usage examples and guides
├── tests/
│   └── integration/    # Integration tests with testcontainers
├── docker/             # Docker configurations
├── .env.example        # Environment variables template
└── README.md           # This file
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.23+** - [Download Go](https://golang.org/dl/)
- **PostgreSQL 12+** - [Download PostgreSQL](https://www.postgresql.org/download/)
- **Git** - [Download Git](https://git-scm.com/downloads)

### 1. Clone and Setup

```bash
# Clone the repository
git clone <repository-url>
cd go-fiber-template

# Copy environment configuration
cp .env.example .env
```

### 2. Configure Environment

Edit `.env` file with your settings:

```bash
# Server Configuration
PORT=8080
APP_ENV=development

# Database Configuration (Update these!)
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=go_fiber_template

# JWT Configuration (Change in production!)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
```

### 3. Install Dependencies

```bash
go mod tidy
```

### 4. Setup Database

```bash
# Create database (PostgreSQL)
createdb go_fiber_template

# Run migrations
go run cmd/migrate/main.go up
```

### 5. Run the Application

```bash
# Development mode
go run cmd/server/main.go

# Or build and run
go build -o bin/server cmd/server/main.go
./bin/server
```

### 6. Verify Installation

- **API Base**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health
- **Swagger UI**: http://localhost:8080/swagger/
- **Database Health**: http://localhost:8080/health/db

## 📖 API Documentation

### Interactive Documentation (Recommended)

Visit **http://localhost:8080/swagger/** for interactive API documentation where you can:

- 📖 Browse all available endpoints
- 🧪 Test endpoints directly in your browser
- 🔐 Authenticate and test protected routes
- 📋 View request/response schemas
- 💡 See real examples

### Quick API Overview

| Endpoint         | Method | Description            | Auth Required |
| ---------------- | ------ | ---------------------- | ------------- |
| `/auth/register` | POST   | Register new user      | ❌            |
| `/auth/login`    | POST   | User login             | ❌            |
| `/auth/refresh`  | POST   | Refresh access token   | ❌            |
| `/auth/logout`   | POST   | User logout            | ✅            |
| `/users`         | GET    | List users (paginated) | ✅            |
| `/users/{id}`    | GET    | Get user by ID         | ✅            |
| `/users/{id}`    | PUT    | Update user            | ✅            |
| `/users/{id}`    | DELETE | Delete user            | ✅            |

### Example API Requests

#### 1. Register a New User

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

**Response:**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "user": {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-25T10:30:00Z"
    },
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2025-01-25T10:45:00Z"
  }
}
```

#### 2. Login

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john.doe@example.com",
    "password": "SecurePass123!"
  }'
```

#### 3. Access Protected Endpoint

```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**

```json
{
  "success": true,
  "message": "Users retrieved successfully",
  "data": [
    {
      "id": 1,
      "email": "john.doe@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "is_active": true,
      "created_at": "2025-01-25T10:30:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

## ⚙️ Environment Configuration

### Required Environment Variables

| Variable      | Description             | Default             | Example           |
| ------------- | ----------------------- | ------------------- | ----------------- |
| `PORT`        | Server port             | `3000`              | `8080`            |
| `APP_ENV`     | Application environment | `development`       | `production`      |
| `DB_HOST`     | Database host           | `localhost`         | `db.example.com`  |
| `DB_PORT`     | Database port           | `5432`              | `5432`            |
| `DB_USER`     | Database username       | `postgres`          | `myuser`          |
| `DB_PASSWORD` | Database password       | `password`          | `mypassword`      |
| `DB_NAME`     | Database name           | `go_fiber_template` | `myapp_db`        |
| `JWT_SECRET`  | JWT signing secret      | ⚠️ **Required**     | `your-secret-key` |

### Optional Environment Variables

| Variable               | Description             | Default         |
| ---------------------- | ----------------------- | --------------- |
| `DB_SSLMODE`           | PostgreSQL SSL mode     | `disable`       |
| `DB_MAX_IDLE_CONNS`    | Max idle connections    | `10`            |
| `DB_MAX_OPEN_CONNS`    | Max open connections    | `100`           |
| `DB_CONN_MAX_LIFETIME` | Connection max lifetime | `1h`            |
| `JWT_ACCESS_EXPIRY`    | Access token expiry     | `15m`           |
| `JWT_REFRESH_EXPIRY`   | Refresh token expiry    | `168h` (7 days) |
| `CORS_ORIGINS`         | Allowed CORS origins    | `*`             |
| `RATE_LIMIT_MAX`       | Rate limit max requests | `100`           |
| `RATE_LIMIT_WINDOW`    | Rate limit time window  | `1m`            |
| `SWAGGER_ENABLED`      | Enable Swagger UI       | `true`          |

### Environment-Specific Configurations

#### Development (.env)

```bash
APP_ENV=development
PORT=8080
DB_HOST=localhost
JWT_SECRET=dev-secret-key
SWAGGER_ENABLED=true
CORS_ORIGINS=http://localhost:3000,http://localhost:3001
```

#### Production (.env.production)

```bash
APP_ENV=production
PORT=80
DB_HOST=prod-db-host
DB_SSLMODE=require
JWT_SECRET=super-secure-production-secret
SWAGGER_ENABLED=false
CORS_ORIGINS=https://yourdomain.com
```

## 🛠️ Development Guide

### Project Architecture

This template follows **Clean Architecture** principles:

```
┌─────────────────┐
│   Controllers   │ ← HTTP handlers, request validation
├─────────────────┤
│    Services     │ ← Business logic, orchestration
├─────────────────┤
│  Repositories   │ ← Data access abstraction
├─────────────────┤
│    Database     │ ← PostgreSQL with GORM
└─────────────────┘
```

### Adding New Features

#### 1. Create a Model

```go
// internal/models/product.go
type Product struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Name        string    `json:"name" gorm:"not null"`
    Price       float64   `json:"price" gorm:"not null"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

#### 2. Create Repository Interface

```go
// internal/repositories/product_repository.go
type ProductRepository interface {
    Create(product *Product) error
    CreateTx(tx *gorm.DB, product *Product) error
    GetByID(id uint) (*Product, error)
    Update(product *Product) error
    Delete(id uint) error
}
```

#### 3. Implement Repository

```go
// internal/repositories/product_repository_impl.go
type productRepositoryImpl struct {
    db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
    return &productRepositoryImpl{db: db}
}

func (r *productRepositoryImpl) Create(product *Product) error {
    return r.db.Create(product).Error
}
// ... implement other methods
```

#### 4. Create Service

```go
// internal/services/product_service.go
type ProductService interface {
    CreateProduct(req CreateProductRequest) (*Product, error)
    GetProduct(id uint) (*Product, error)
}

type productService struct {
    productRepo ProductRepository
    txManager   TransactionManager
}
```

#### 5. Create Controller

```go
// internal/controllers/product_controller.go
// @Summary Create product
// @Description Create a new product
// @Tags products
// @Accept json
// @Produce json
// @Param product body CreateProductRequest true "Product data"
// @Success 201 {object} utils.APIResponse{data=Product}
// @Router /products [post]
// @Security BearerAuth
func (c *ProductController) CreateProduct(ctx *fiber.Ctx) error {
    // Implementation
}
```

### Database Migrations

#### Create Migration

```bash
# Create new migration
go run cmd/migrate/main.go create add_products_table

# This creates:
# migrations/YYYYMMDDHHMMSS_add_products_table.up.sql
# migrations/YYYYMMDDHHMMSS_add_products_table.down.sql
```

#### Migration Commands

```bash
# Run all pending migrations
go run cmd/migrate/main.go up

# Rollback last migration
go run cmd/migrate/main.go down

# Check migration status
go run cmd/migrate/main.go version

# Force migration version (use carefully!)
go run cmd/migrate/main.go force 20250125103000

### Database Seeders

The template includes a comprehensive database seeding system for populating your database with initial or test data.

#### Quick Start

```bash
# Run all seeders
go run cmd/seed/main.go run --all

# Run environment-specific seeders (recommended)
go run cmd/seed/main.go run --env

# Check seeder status
go run cmd/seed/main.go status

# List available seeders
go run cmd/seed/main.go list
```

#### Available Seeders

- **user_seeder**: Creates sample admin and regular users
- **demo_data_seeder**: Creates comprehensive demo data (development only)

#### Seeder Commands

```bash
# Run specific seeders
go run cmd/seed/main.go run user_seeder demo_data_seeder

# Rollback seeders
go run cmd/seed/main.go rollback user_seeder

# Rollback all seeders
go run cmd/seed/main.go rollback --all
```

For detailed information about creating custom seeders and advanced usage, see [Seeder System Documentation](docs/SEEDER_SYSTEM.md).
```

### Testing

#### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests only
go test ./tests/integration/...

# Run specific test
go test -run TestUserRepository ./internal/repositories/
```

#### Writing Tests

```go
// internal/services/user_service_test.go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    mockRepo := &mocks.UserRepository{}
    service := NewUserService(mockRepo, nil)

    // Test implementation
    user, err := service.CreateUser(CreateUserRequest{
        Email: "test@example.com",
        Password: "password123",
    })

    // Assertions
    assert.NoError(t, err)
    assert.NotNil(t, user)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### Code Generation

#### Generate Swagger Documentation

```bash
# Install swag (if not installed)
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/server/main.go -o docs

# The generated files:
# docs/docs.go
# docs/swagger.json
# docs/swagger.yaml
```

#### Generate Mocks (Optional)

```bash
# Install mockery
go install github.com/vektra/mockery/v2@latest

# Generate mocks for interfaces
mockery --dir=internal/repositories --all --output=tests/mocks
```

### Docker Development

#### Build Docker Image

```bash
# Build image
docker build -f docker/Dockerfile -t go-fiber-template .

# Run container
docker run -p 8080:8080 --env-file .env go-fiber-template
```

#### Docker Compose (Optional)

Create `docker-compose.yml`:

```yaml
version: "3.8"
services:
  app:
    build:
      context: .
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: go_fiber_template
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Performance Optimization

#### Database Connection Pooling

```bash
# Optimize for your workload
DB_MAX_IDLE_CONNS=25
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m
```

#### Rate Limiting

```bash
# Adjust based on your needs
RATE_LIMIT_MAX=1000
RATE_LIMIT_WINDOW=1m
```

## 🧪 Testing Guide

### API Testing with Swagger UI

1. **Start the server**: `go run cmd/server/main.go`
2. **Open Swagger UI**: http://localhost:8080/swagger/
3. **Register/Login**: Use auth endpoints to get tokens
4. **Authorize**: Click "Authorize" button, enter `Bearer {token}`
5. **Test endpoints**: All protected routes now work

### Manual Testing with cURL

See complete examples in [docs/examples/api-testing-guide.md](docs/examples/api-testing-guide.md)

### Automated Testing

```bash
# Unit tests
go test ./internal/...

# Integration tests (requires Docker)
go test ./tests/integration/...

# Test with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./...
```

## 🚀 Deployment

### Build for Production

```bash
# Build binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server cmd/server/main.go

# Run migrations
./bin/migrate up

# Start server
./bin/server
```

### Environment Setup

1. **Set production environment variables**
2. **Run database migrations**
3. **Configure reverse proxy (nginx/Apache)**
4. **Set up SSL certificates**
5. **Configure monitoring and logging**

## 📚 Additional Resources

- **[API Testing Guide](docs/examples/api-testing-guide.md)** - Complete testing workflows
- **[Authentication Examples](docs/examples/authentication.md)** - Auth flow examples
- **[User Management Examples](docs/examples/users.md)** - User CRUD operations
- **[Swagger UI Guide](docs/examples/swagger-ui-testing-guide.md)** - Interactive testing

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Documentation**: Check [docs/](docs/) directory
- **Issues**: Open an issue on GitHub
- **API Testing**: Use Swagger UI at http://localhost:8080/swagger/

---

**💡 Quick Tip**: Start with the Swagger UI at http://localhost:8080/swagger/ for the best API exploration experience!
