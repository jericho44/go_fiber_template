# Environment Configuration Guide

This guide provides comprehensive documentation for configuring the Go Fiber Template application across different environments.

## 📋 Overview

The application uses environment variables for configuration, supporting multiple deployment environments with different settings. Configuration is loaded from:

1. **Environment variables** (highest priority)
2. **.env file** (development default)
3. **Default values** (fallback)

## 🔧 Configuration Structure

### Configuration Categories

| Category          | Purpose               | Variables                |
| ----------------- | --------------------- | ------------------------ |
| **Server**        | HTTP server settings  | `PORT`, `APP_ENV`        |
| **Database**      | PostgreSQL connection | `DB_*` variables         |
| **JWT**           | Authentication tokens | `JWT_*` variables        |
| **CORS**          | Cross-origin requests | `CORS_*` variables       |
| **Rate Limiting** | API rate limiting     | `RATE_LIMIT_*` variables |
| **Swagger**       | API documentation     | `SWAGGER_*` variables    |

## 🌍 Environment-Specific Configurations

### Development Environment

**File**: `.env`

```bash
# Server Configuration
PORT=8080
APP_ENV=development

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=go_fiber_template_dev
DB_SSLMODE=disable

# Database Connection Pool
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=25
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m

# JWT Configuration (Development keys - NOT for production!)
JWT_SECRET=dev-jwt-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# CORS Configuration (Permissive for development)
CORS_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:8080

# Rate Limiting (Generous for development)
RATE_LIMIT_MAX=1000
RATE_LIMIT_WINDOW=1m

# Swagger Configuration
SWAGGER_ENABLED=true
SWAGGER_HOST=localhost:8080
SWAGGER_BASE_PATH=/api/v1
```

### Testing Environment

**File**: `.env.test`

```bash
# Server Configuration
PORT=8081
APP_ENV=test

# Database Configuration (Separate test database)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=go_fiber_template_test
DB_SSLMODE=disable

# Database Connection Pool (Smaller for tests)
DB_MAX_IDLE_CONNS=5
DB_MAX_OPEN_CONNS=10
DB_CONN_MAX_LIFETIME=30m
DB_CONN_MAX_IDLE_TIME=15m

# JWT Configuration (Test-specific)
JWT_SECRET=test-jwt-secret-key
JWT_ACCESS_EXPIRY=5m
JWT_REFRESH_EXPIRY=1h

# CORS Configuration
CORS_ORIGINS=*

# Rate Limiting (Disabled for tests)
RATE_LIMIT_MAX=10000
RATE_LIMIT_WINDOW=1s

# Swagger Configuration
SWAGGER_ENABLED=false
```

### Staging Environment

**File**: `.env.staging`

```bash
# Server Configuration
PORT=80
APP_ENV=staging

# Database Configuration
DB_HOST=staging-db.example.com
DB_PORT=5432
DB_USER=app_user
DB_PASSWORD=${DB_PASSWORD}  # From environment/secrets
DB_NAME=go_fiber_template_staging
DB_SSLMODE=require

# Database Connection Pool (Production-like)
DB_MAX_IDLE_CONNS=25
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=1h
DB_CONN_MAX_IDLE_TIME=30m

# JWT Configuration
JWT_SECRET=${JWT_SECRET}  # From environment/secrets
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# CORS Configuration
CORS_ORIGINS=https://staging.example.com,https://staging-admin.example.com

# Rate Limiting
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW=1m

# Swagger Configuration (Enabled for staging testing)
SWAGGER_ENABLED=true
SWAGGER_HOST=staging.example.com
SWAGGER_BASE_PATH=/api/v1
```

### Production Environment

**File**: `.env.production`

```bash
# Server Configuration
PORT=80
APP_ENV=production

# Database Configuration
DB_HOST=${DB_HOST}  # From environment/secrets
DB_PORT=5432
DB_USER=${DB_USER}  # From environment/secrets
DB_PASSWORD=${DB_PASSWORD}  # From environment/secrets
DB_NAME=go_fiber_template_prod
DB_SSLMODE=require

# Database Connection Pool (Optimized for production)
DB_MAX_IDLE_CONNS=50
DB_MAX_OPEN_CONNS=200
DB_CONN_MAX_LIFETIME=2h
DB_CONN_MAX_IDLE_TIME=1h

# JWT Configuration
JWT_SECRET=${JWT_SECRET}  # From environment/secrets (256-bit key)
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# CORS Configuration (Restrictive)
CORS_ORIGINS=https://example.com,https://admin.example.com

# Rate Limiting (Strict)
RATE_LIMIT_MAX=100
RATE_LIMIT_WINDOW=1m

# Swagger Configuration (Disabled in production)
SWAGGER_ENABLED=false
```

## 📝 Configuration Variables Reference

### Server Configuration

| Variable  | Description             | Type     | Default       | Example      |
| --------- | ----------------------- | -------- | ------------- | ------------ |
| `PORT`    | HTTP server port        | `int`    | `3000`        | `8080`       |
| `APP_ENV` | Application environment | `string` | `development` | `production` |

**Validation Rules:**

- `PORT`: Must be between 1-65535
- `APP_ENV`: Must be one of: `development`, `test`, `staging`, `production`

### Database Configuration

| Variable      | Description       | Type     | Default             | Example           |
| ------------- | ----------------- | -------- | ------------------- | ----------------- |
| `DB_HOST`     | Database host     | `string` | `localhost`         | `db.example.com`  |
| `DB_PORT`     | Database port     | `int`    | `5432`              | `5432`            |
| `DB_USER`     | Database username | `string` | `postgres`          | `app_user`        |
| `DB_PASSWORD` | Database password | `string` | `password`          | `secure_password` |
| `DB_NAME`     | Database name     | `string` | `go_fiber_template` | `myapp_prod`      |
| `DB_SSLMODE`  | SSL mode          | `string` | `disable`           | `require`         |

**SSL Mode Options:**

- `disable`: No SSL (development only)
- `require`: Require SSL
- `verify-ca`: Verify certificate authority
- `verify-full`: Full certificate verification

### Database Connection Pool

| Variable                | Description              | Type       | Default | Example |
| ----------------------- | ------------------------ | ---------- | ------- | ------- |
| `DB_MAX_IDLE_CONNS`     | Maximum idle connections | `int`      | `10`    | `25`    |
| `DB_MAX_OPEN_CONNS`     | Maximum open connections | `int`      | `100`   | `200`   |
| `DB_CONN_MAX_LIFETIME`  | Connection max lifetime  | `duration` | `1h`    | `2h`    |
| `DB_CONN_MAX_IDLE_TIME` | Connection max idle time | `duration` | `30m`   | `1h`    |

**Duration Format:** `1h30m`, `45m`, `30s`, etc.

### JWT Configuration

| Variable             | Description          | Type       | Default      | Example               |
| -------------------- | -------------------- | ---------- | ------------ | --------------------- |
| `JWT_SECRET`         | JWT signing secret   | `string`   | **Required** | `your-256-bit-secret` |
| `JWT_ACCESS_EXPIRY`  | Access token expiry  | `duration` | `15m`        | `30m`                 |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `duration` | `168h`       | `720h`                |

**Security Requirements:**

- `JWT_SECRET`: Minimum 32 characters, use cryptographically secure random string
- Production: Use 256-bit (32-byte) random key

### CORS Configuration

| Variable       | Description     | Type     | Default | Example                                         |
| -------------- | --------------- | -------- | ------- | ----------------------------------------------- |
| `CORS_ORIGINS` | Allowed origins | `string` | `*`     | `https://example.com,https://admin.example.com` |

**Format:** Comma-separated list of origins or `*` for all origins

### Rate Limiting

| Variable            | Description             | Type       | Default | Example |
| ------------------- | ----------------------- | ---------- | ------- | ------- |
| `RATE_LIMIT_MAX`    | Max requests per window | `int`      | `100`   | `1000`  |
| `RATE_LIMIT_WINDOW` | Time window             | `duration` | `1m`    | `5m`    |

### Swagger Configuration

| Variable            | Description       | Type     | Default          | Example           |
| ------------------- | ----------------- | -------- | ---------------- | ----------------- |
| `SWAGGER_ENABLED`   | Enable Swagger UI | `bool`   | `true`           | `false`           |
| `SWAGGER_HOST`      | Swagger host      | `string` | `localhost:3000` | `api.example.com` |
| `SWAGGER_BASE_PATH` | API base path     | `string` | `/api/v1`        | `/v2`             |

## 🔐 Security Best Practices

### JWT Secret Generation

```bash
# Generate secure JWT secret (32 bytes)
openssl rand -base64 32

# Or using Go
go run -c 'package main; import ("crypto/rand"; "encoding/base64"; "fmt"); func main() { b := make([]byte, 32); rand.Read(b); fmt.Println(base64.StdEncoding.EncodeToString(b)) }'

# Or using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"
```

### Database Password Security

```bash
# Generate secure database password
openssl rand -base64 24

# Or using pwgen (if installed)
pwgen -s 32 1
```

### Environment Variable Security

#### Development

- ✅ Use `.env` file (add to `.gitignore`)
- ✅ Use weak secrets for convenience
- ✅ Enable Swagger for testing

#### Production

- ❌ Never commit secrets to version control
- ✅ Use environment variables or secret management
- ✅ Use strong, randomly generated secrets
- ✅ Disable Swagger UI
- ✅ Enable SSL/TLS

## 🚀 Deployment Configurations

### Docker Environment

**docker-compose.yml**

```yaml
version: "3.8"
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - APP_ENV=production
      - DB_HOST=postgres
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD}
      - DB_NAME=go_fiber_template
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=go_fiber_template
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

### Kubernetes ConfigMap

**k8s-config.yaml**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  PORT: "8080"
  APP_ENV: "production"
  DB_HOST: "postgres-service"
  DB_PORT: "5432"
  DB_NAME: "go_fiber_template"
  DB_SSLMODE: "require"
  CORS_ORIGINS: "https://example.com"
  RATE_LIMIT_MAX: "100"
  RATE_LIMIT_WINDOW: "1m"
  SWAGGER_ENABLED: "false"

---
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  DB_USER: <base64-encoded-username>
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>
```

### AWS ECS Task Definition

```json
{
  "family": "go-fiber-template",
  "taskRoleArn": "arn:aws:iam::account:role/ecsTaskRole",
  "containerDefinitions": [
    {
      "name": "app",
      "image": "your-registry/go-fiber-template:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        { "name": "PORT", "value": "8080" },
        { "name": "APP_ENV", "value": "production" },
        { "name": "DB_HOST", "value": "your-rds-endpoint" },
        { "name": "DB_PORT", "value": "5432" },
        { "name": "DB_NAME", "value": "go_fiber_template" },
        { "name": "DB_SSLMODE", "value": "require" }
      ],
      "secrets": [
        {
          "name": "DB_USER",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:db-credentials:username"
        },
        {
          "name": "DB_PASSWORD",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:db-credentials:password"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:region:account:secret:jwt-secret"
        }
      ]
    }
  ]
}
```

## 🔍 Configuration Validation

The application validates configuration on startup:

### Required Variables

These variables must be set:

- `JWT_SECRET` (minimum 32 characters)
- `DB_PASSWORD` (if not using default)

### Validation Rules

```go
// Example validation logic
func validateConfig(cfg *Config) error {
    if len(cfg.JWT.Secret) < 32 {
        return errors.New("JWT_SECRET must be at least 32 characters")
    }

    if cfg.Server.Port < 1 || cfg.Server.Port > 65535 {
        return errors.New("PORT must be between 1 and 65535")
    }

    if cfg.Database.MaxOpenConns < cfg.Database.MaxIdleConns {
        return errors.New("DB_MAX_OPEN_CONNS must be >= DB_MAX_IDLE_CONNS")
    }

    return nil
}
```

## 🛠️ Configuration Management Tools

### Environment File Management

```bash
# Copy environment template
cp .env.example .env.development
cp .env.example .env.staging
cp .env.example .env.production

# Load specific environment
export $(cat .env.staging | xargs)
go run cmd/server/main.go

# Or use direnv (auto-loads .envrc)
echo "dotenv .env.development" > .envrc
direnv allow
```

### Configuration Testing

```bash
# Test configuration loading
go run cmd/server/main.go --config-test

# Validate environment
go run -c '
package main
import "go-fiber-template/internal/config"
func main() {
    cfg, err := config.Load()
    if err != nil {
        panic(err)
    }
    println("Configuration valid!")
}'
```

## 🔧 Troubleshooting

### Common Configuration Issues

#### 1. Database Connection Failed

```bash
# Check database connectivity
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME

# Test with application
go run cmd/server/main.go --test-db
```

#### 2. JWT Token Issues

```bash
# Verify JWT secret length
echo -n "$JWT_SECRET" | wc -c  # Should be >= 32

# Test JWT generation
go run -c 'package main; import "go-fiber-template/internal/utils"; func main() { token, _ := utils.GenerateJWT(1); println(token) }'
```

#### 3. Port Already in Use

```bash
# Find process using port
lsof -i :$PORT  # macOS/Linux
netstat -ano | findstr :$PORT  # Windows

# Use different port
export PORT=8081
```

#### 4. CORS Issues

```bash
# Test CORS headers
curl -H "Origin: https://example.com" \
     -H "Access-Control-Request-Method: POST" \
     -H "Access-Control-Request-Headers: X-Requested-With" \
     -X OPTIONS \
     http://localhost:8080/api/v1/auth/login
```

### Environment-Specific Issues

#### Development

- **Issue**: Database connection refused
- **Solution**: Start PostgreSQL service, check DB_HOST/DB_PORT

#### Production

- **Issue**: Secrets not loading
- **Solution**: Verify secret management system, check IAM permissions

#### Docker

- **Issue**: Environment variables not passed
- **Solution**: Use `--env-file` or `environment` in docker-compose

## 📚 Additional Resources

- **[Development Setup Guide](DEVELOPMENT_SETUP.md)** - Complete setup instructions
- **[API Documentation](README.md)** - API usage and examples
- **[Security Best Practices](https://owasp.org/www-project-go-secure-coding-practices-guide/)** - OWASP Go security guide

---

**⚠️ Security Warning**: Never commit secrets to version control. Use environment variables or secret management systems in production.
