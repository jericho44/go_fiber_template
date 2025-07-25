# Routing System

This document describes the refactored routing system implemented in the Go Fiber Template project.

## Overview

The routing system is organized in a modular way with clear separation of concerns. It connects HTTP routes to their respective controllers while maintaining clean architecture principles. The system now includes enhanced public routes with a Hello World endpoint and better organization.

## Structure

### Files

- `routes.go` - Main routing setup and configuration
- `public_routes.go` - Public routes that don't require authentication
- `auth_routes.go` - Authentication-related routes
- `user_routes.go` - User management routes
- `routes_test.go` - Comprehensive tests for the routing system

### Dependencies

The routing system uses a `Dependencies` struct to inject all required services, repositories, middleware, and controllers:

```go
type Dependencies struct {
    // Repositories
    UserRepo           repositories.UserRepository
    TokenBlacklistRepo repositories.TokenBlacklistRepository
    TxManager          repositories.TransactionManager

    // Services
    AuthService services.AuthService
    UserService services.UserService
    JWTService  services.JWTServiceInterface

    // Utils
    JWTManager *utils.JWTManager

    // Middleware
    AuthMiddleware *middleware.AuthMiddleware

    // Controllers
    AuthController *controllers.AuthController
    UserController *controllers.UserController
}
```

## Route Groups

### Public Routes (No Authentication Required)

- `GET /` - Root endpoint with API information and available endpoints
- `GET /hello` - **Hello World endpoint** with customizable greeting
  - Query parameters:
    - `name` (optional): Custom name for greeting (default: "World")
    - `lang` (optional): Language for greeting (en, es, fr, de, it, pt, ru, ja, ko, zh)
  - Examples:
    - `/hello` → "Hello, World!"
    - `/hello?name=John` → "Hello, John!"
    - `/hello?name=Maria&lang=es` → "Hola, Maria!"
- `GET /health` - Enhanced health check endpoint with system status
- `GET /info` - Comprehensive API information and feature list
- `GET /ping` - Simple connectivity test endpoint

### Authentication Routes (`/api/v1/auth`)

**Public routes (no authentication required):**
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/refresh` - Token refresh

**Protected routes (require JWT authentication):**
- `POST /auth/logout` - User logout

### User Routes (`/api/v1/users`)

**All routes require JWT authentication:**
- `GET /users` - List users with pagination and filtering
- `GET /users/:id` - Get user profile by ID
- `PUT /users/:id` - Update user profile
- `DELETE /users/:id` - Soft delete user

### Documentation Routes

- `GET /swagger/*` - Swagger UI documentation
- `GET /swagger` - Redirects to `/swagger/`
- `GET /docs` - Documentation information endpoint

### API Information Routes

- `GET /api/v1` - API v1 information and available endpoints

## Setup

The routing system is initialized in the main application using:

```go
routes.SetupAllRoutes(app, cfg, dependencies)
```

This function:
1. Sets up public routes (including Hello World)
2. Sets up API versioned routes
3. Sets up documentation routes
4. Configures the 404 handler

## Hello World Feature

The Hello World endpoint (`/hello`) includes several enhanced features:

### Multi-language Support
- English (en): "Hello"
- Spanish (es): "Hola"
- French (fr): "Bonjour"
- German (de): "Hallo"
- Italian (it): "Ciao"
- Portuguese (pt): "Olá"
- Russian (ru): "Привет"
- Japanese (ja): "こんにちは"
- Korean (ko): "안녕하세요"
- Chinese (zh): "你好"

### Usage Examples
```bash
# Basic greeting
curl http://localhost:8080/hello
# → {"message": "Hello, World!", ...}

# Custom name
curl http://localhost:8080/hello?name=Alice
# → {"message": "Hello, Alice!", ...}

# Different language
curl http://localhost:8080/hello?name=Carlos&lang=es
# → {"message": "Hola, Carlos!", ...}
```

## Middleware Integration

The routing system integrates with middleware:

- **Authentication Middleware**: Applied to protected routes using `deps.AuthMiddleware.RequireAuth()`
- **Global Middleware**: Applied in the main application (CORS, rate limiting, logging, etc.)

## Testing

The routing system includes comprehensive tests that verify:

- Route setup doesn't panic
- All public endpoints work correctly
- Hello World endpoint with various parameters
- Documentation routes are properly configured
- Route structure is correctly organized

Run tests with:
```bash
go test ./internal/routes -v
```

## Adding New Routes

To add new routes:

1. Create a new route file (e.g., `product_routes.go`)
2. Add the controller to the `Dependencies` struct
3. Create a setup function (e.g., `setupProductRoutes`)
4. Call the setup function in `SetupAllRoutes`
5. Add tests for the new routes

Example:

```go
// product_routes.go
func setupProductRoutes(api fiber.Router, cfg *config.Config, deps *Dependencies) {
    products := api.Group("/products")
    products.Use(deps.AuthMiddleware.RequireAuth())
    
    products.Get("/", deps.ProductController.ListProducts)
    products.Post("/", deps.ProductController.CreateProduct)
    products.Get("/:id", deps.ProductController.GetProduct)
    products.Put("/:id", deps.ProductController.UpdateProduct)
    products.Delete("/:id", deps.ProductController.DeleteProduct)
}
```

## Benefits

1. **Modular Organization**: Routes are organized by domain/feature
2. **Clean Separation**: Each route group has its own file
3. **Dependency Injection**: Clean dependency management
4. **Enhanced Public Routes**: Rich public endpoints with Hello World feature
5. **Multi-language Support**: Internationalization in Hello World endpoint
6. **Comprehensive Health Checks**: Detailed system status information
7. **Testable**: Easy to test individual route groups
8. **Maintainable**: Easy to add, modify, or remove routes
9. **Consistent**: All routes follow the same patterns
10. **Middleware Integration**: Seamless middleware application