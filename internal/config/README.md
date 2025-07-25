# Configuration Package

This package provides comprehensive configuration management for the Go Fiber Template application.

## Features

- Environment variable loading with defaults
- Configuration validation
- Support for multiple environments (development, staging, production)
- Type-safe configuration structures
- Comprehensive unit tests

## Usage

### Basic Usage

```go
package main

import (
    "log"
    "go-fiber-template/internal/config"
)

func main() {
    // Load configuration from environment variables
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load configuration:", err)
    }

    // Use configuration
    fmt.Printf("Server will run on %s\n", cfg.Server.GetServerAddress())
    fmt.Printf("Database DSN: %s\n", cfg.Database.GetDSN())

    // Check environment
    if cfg.IsDevelopment() {
        fmt.Println("Running in development mode")
    }
}
```

### Environment Variables

The following environment variables are supported:

#### Server Configuration

- `PORT` - Server port (default: 3000)
- `HOST` - Server host (default: localhost)
- `APP_ENV` - Application environment: development, staging, production (default: development)

#### Database Configuration

- `DB_HOST` - Database host (default: localhost)
- `DB_PORT` - Database port (default: 5432)
- `DB_USER` - Database user (default: postgres)
- `DB_PASSWORD` - Database password (default: empty)
- `DB_NAME` - Database name (default: go_fiber_template)
- `DB_SSLMODE` - SSL mode: disable, require, verify-ca, verify-full (default: disable)

#### JWT Configuration

- `JWT_SECRET` - JWT secret key (required, minimum 32 characters)
- `JWT_ACCESS_EXPIRY` - Access token expiry duration (default: 15m)
- `JWT_REFRESH_EXPIRY` - Refresh token expiry duration (default: 168h)

#### CORS Configuration

- `CORS_ORIGINS` - Comma-separated list of allowed origins (default: http://localhost:3000)

#### Rate Limiting Configuration

- `RATE_LIMIT_MAX` - Maximum requests per window (default: 100)
- `RATE_LIMIT_WINDOW` - Rate limiting window duration (default: 1m)

#### Swagger Configuration

- `SWAGGER_ENABLED` - Enable Swagger UI (default: true)
- `SWAGGER_HOST` - Swagger host (default: localhost:3000)
- `SWAGGER_BASE_PATH` - Swagger base path (default: /api/v1)

### Configuration Validation

The configuration package automatically validates all loaded configuration values:

- Server port must be between 1 and 65535
- Database SSL mode must be one of: disable, require, verify-ca, verify-full
- JWT secret must be at least 32 characters long
- JWT access expiry must be less than refresh expiry
- At least one CORS origin must be specified
- Rate limit max must be positive

### Environment Methods

```go
cfg, _ := config.Load()

if cfg.IsDevelopment() {
    // Development-specific logic
}

if cfg.IsProduction() {
    // Production-specific logic
}

if cfg.IsStaging() {
    // Staging-specific logic
}
```

### Utility Methods

```go
cfg, _ := config.Load()

// Get database connection string
dsn := cfg.Database.GetDSN()

// Get server address
addr := cfg.Server.GetServerAddress()
```

## Testing

Run the tests with:

```bash
go test ./internal/config -v
```

The package includes comprehensive unit tests covering:

- Configuration loading with default values
- Configuration loading with custom values
- Error handling for invalid values
- Configuration validation
- Environment detection methods
- Utility methods
