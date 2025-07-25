# Development Setup Guide

This guide provides detailed instructions for setting up the Go Fiber Template development environment on different operating systems.

## 📋 Prerequisites

### Required Software

| Software       | Minimum Version | Recommended   | Installation                                                    |
| -------------- | --------------- | ------------- | --------------------------------------------------------------- |
| **Go**         | 1.23+           | Latest stable | [golang.org/dl](https://golang.org/dl/)                         |
| **PostgreSQL** | 12+             | 15+           | [postgresql.org/download](https://www.postgresql.org/download/) |
| **Git**        | 2.20+           | Latest        | [git-scm.com](https://git-scm.com/downloads)                    |

### Optional Tools

| Tool       | Purpose                    | Installation                                        |
| ---------- | -------------------------- | --------------------------------------------------- |
| **Docker** | Containerization & testing | [docker.com](https://www.docker.com/get-started)    |
| **Make**   | Build automation           | Usually pre-installed on Unix systems               |
| **curl**   | API testing                | Usually pre-installed                               |
| **jq**     | JSON processing            | [jqlang.github.io/jq](https://jqlang.github.io/jq/) |

## 🖥️ Platform-Specific Setup

### Windows Setup

#### 1. Install Go

```powershell
# Download and install from https://golang.org/dl/
# Or use Chocolatey
choco install golang

# Verify installation
go version
```

#### 2. Install PostgreSQL

```powershell
# Download from https://www.postgresql.org/download/windows/
# Or use Chocolatey
choco install postgresql

# Start PostgreSQL service
net start postgresql-x64-15
```

#### 3. Setup Environment

```powershell
# Add to PATH (usually done by installer)
$env:PATH += ";C:\Program Files\Go\bin"
$env:PATH += ";C:\Program Files\PostgreSQL\15\bin"

# Set GOPATH (optional, defaults to %USERPROFILE%\go)
$env:GOPATH = "C:\Users\YourUsername\go"
```

### macOS Setup

#### 1. Install Homebrew (if not installed)

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

#### 2. Install Dependencies

```bash
# Install Go
brew install go

# Install PostgreSQL
brew install postgresql@15

# Start PostgreSQL service
brew services start postgresql@15

# Install optional tools
brew install make curl jq docker
```

#### 3. Verify Installation

```bash
go version
psql --version
```

### Linux (Ubuntu/Debian) Setup

#### 1. Update Package Manager

```bash
sudo apt update && sudo apt upgrade -y
```

#### 2. Install Go

```bash
# Method 1: Using package manager (may not be latest)
sudo apt install golang-go

# Method 2: Download latest from golang.org (recommended)
wget https://go.dev/dl/go1.23.11.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.11.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### 3. Install PostgreSQL

```bash
# Install PostgreSQL
sudo apt install postgresql postgresql-contrib

# Start and enable service
sudo systemctl start postgresql
sudo systemctl enable postgresql

# Create database user
sudo -u postgres createuser --interactive
```

#### 4. Install Additional Tools

```bash
sudo apt install make curl jq git
```

### Linux (CentOS/RHEL/Fedora) Setup

#### 1. Install Go

```bash
# Fedora
sudo dnf install golang

# CentOS/RHEL (enable EPEL first)
sudo yum install epel-release
sudo yum install golang
```

#### 2. Install PostgreSQL

```bash
# Fedora
sudo dnf install postgresql postgresql-server postgresql-contrib

# CentOS/RHEL
sudo yum install postgresql postgresql-server postgresql-contrib

# Initialize database
sudo postgresql-setup initdb
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

## 🚀 Project Setup

### 1. Clone Repository

```bash
# Clone the repository
git clone <repository-url>
cd go-fiber-template

# Or if using this as a template
git clone <repository-url> my-new-project
cd my-new-project
```

### 2. Environment Configuration

```bash
# Copy environment template
cp .env.example .env

# Edit configuration (use your preferred editor)
nano .env
# or
code .env
# or
vim .env
```

#### Required Environment Variables

```bash
# Server Configuration
PORT=8080
APP_ENV=development

# Database Configuration - UPDATE THESE!
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username        # Change this
DB_PASSWORD=your_password    # Change this
DB_NAME=go_fiber_template

# JWT Configuration - CHANGE IN PRODUCTION!
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# Optional: Development-friendly settings
SWAGGER_ENABLED=true
CORS_ORIGINS=http://localhost:3000,http://localhost:3001,http://localhost:8080
RATE_LIMIT_MAX=1000
```

### 3. Database Setup

#### Create Database

```bash
# Method 1: Using createdb command
createdb go_fiber_template

# Method 2: Using psql
psql -U postgres
CREATE DATABASE go_fiber_template;
\q
```

#### Create Database User (if needed)

```bash
# Connect to PostgreSQL
psql -U postgres

# Create user and grant permissions
CREATE USER your_username WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE go_fiber_template TO your_username;
ALTER USER your_username CREATEDB;  -- For running tests
\q
```

### 4. Install Dependencies

```bash
# Download and install Go modules
go mod tidy

# Verify dependencies
go mod verify
```

### 5. Run Database Migrations

```bash
# Run all migrations
go run cmd/migrate/main.go up

# Verify migrations
go run cmd/migrate/main.go version
```

### 6. Build and Run

```bash
# Method 1: Run directly
go run cmd/server/main.go

# Method 2: Build then run
go build -o bin/server cmd/server/main.go
./bin/server

# Method 3: Using Make (if Makefile exists)
make run
```

### 7. Verify Installation

Open your browser and check:

- **Health Check**: http://localhost:8080/health
- **Database Health**: http://localhost:8080/health/db
- **Swagger UI**: http://localhost:8080/swagger/
- **API Base**: http://localhost:8080/api/v1

## 🧪 Development Workflow

### Daily Development

```bash
# 1. Pull latest changes
git pull origin main

# 2. Update dependencies (if go.mod changed)
go mod tidy

# 3. Run migrations (if new migrations exist)
go run cmd/migrate/main.go up

# 4. Run tests
go test ./...

# 5. Start development server
go run cmd/server/main.go
```

### Code Generation

```bash
# Generate Swagger documentation
swag init -g cmd/server/main.go -o docs

# Generate mocks (if using mockery)
mockery --dir=internal/repositories --all --output=tests/mocks
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run integration tests (requires Docker)
go test ./tests/integration/...

# Run specific package tests
go test ./internal/services/...

# Run with race detection
go test -race ./...

# Benchmark tests
go test -bench=. ./...
```

## 🐳 Docker Development

### Using Docker for Database

If you prefer not to install PostgreSQL locally:

```bash
# Start PostgreSQL in Docker
docker run --name postgres-dev \
  -e POSTGRES_DB=go_fiber_template \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=password \
  -p 5432:5432 \
  -d postgres:15

# Update .env file
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=password
```

### Full Docker Development

```bash
# Build application image
docker build -f docker/Dockerfile -t go-fiber-template .

# Run with Docker Compose (if available)
docker-compose up -d
```

## 🔧 IDE Setup

### Visual Studio Code

#### Recommended Extensions

```json
{
  "recommendations": [
    "golang.go",
    "ms-vscode.vscode-json",
    "redhat.vscode-yaml",
    "humao.rest-client",
    "ms-vscode.vscode-docker"
  ]
}
```

#### Settings (.vscode/settings.json)

```json
{
  "go.toolsManagement.checkForUpdates": "local",
  "go.useLanguageServer": true,
  "go.formatTool": "goimports",
  "go.lintTool": "golangci-lint",
  "go.testFlags": ["-v"],
  "go.coverOnSave": true,
  "files.exclude": {
    "**/bin": true,
    "**/vendor": true
  }
}
```

### GoLand/IntelliJ IDEA

1. **Install Go plugin** (if using IntelliJ IDEA)
2. **Configure GOROOT**: File → Settings → Go → GOROOT
3. **Configure GOPATH**: File → Settings → Go → GOPATH
4. **Enable Go Modules**: File → Settings → Go → Go Modules

### Vim/Neovim

#### Install vim-go

```vim
" Add to .vimrc
Plug 'fatih/vim-go', { 'do': ':GoUpdateBinaries' }

" Configuration
let g:go_fmt_command = "goimports"
let g:go_auto_type_info = 1
let g:go_highlight_functions = 1
let g:go_highlight_methods = 1
```

## 🛠️ Development Tools

### Essential Go Tools

```bash
# Install development tools
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Optional tools
go install github.com/vektra/mockery/v2@latest
go install github.com/air-verse/air@latest  # Live reload
```

### Live Reload with Air

```bash
# Install Air
go install github.com/air-verse/air@latest

# Create .air.toml configuration
air init

# Run with live reload
air
```

### Database Tools

```bash
# Install migrate CLI (alternative to go run cmd/migrate/main.go)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Use migrate CLI
migrate -path migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up
```

## 🔍 Troubleshooting

### Common Issues

#### 1. Go Module Issues

```bash
# Problem: Module not found
# Solution: Initialize module
go mod init go-fiber-template
go mod tidy

# Problem: Dependency conflicts
# Solution: Clean module cache
go clean -modcache
go mod download
```

#### 2. Database Connection Issues

```bash
# Problem: Connection refused
# Solution: Check PostgreSQL service
sudo systemctl status postgresql  # Linux
brew services list | grep postgres  # macOS
net start postgresql-x64-15  # Windows

# Problem: Authentication failed
# Solution: Check pg_hba.conf or create user
sudo -u postgres psql
CREATE USER your_user WITH PASSWORD 'your_password';
```

#### 3. Port Already in Use

```bash
# Find process using port 8080
lsof -i :8080  # macOS/Linux
netstat -ano | findstr :8080  # Windows

# Kill process
kill -9 <PID>  # macOS/Linux
taskkill /PID <PID> /F  # Windows
```

#### 4. Permission Issues

```bash
# Problem: Permission denied
# Solution: Fix file permissions
chmod +x bin/server

# Problem: Database permission denied
# Solution: Grant database permissions
sudo -u postgres psql
GRANT ALL PRIVILEGES ON DATABASE go_fiber_template TO your_user;
```

### Environment-Specific Issues

#### Windows

```powershell
# Problem: Command not found
# Solution: Add to PATH
$env:PATH += ";C:\path\to\go\bin"

# Problem: Line ending issues
# Solution: Configure Git
git config --global core.autocrlf true
```

#### macOS

```bash
# Problem: Command line tools missing
# Solution: Install Xcode command line tools
xcode-select --install

# Problem: Homebrew permissions
# Solution: Fix Homebrew permissions
sudo chown -R $(whoami) /usr/local/share/zsh
```

#### Linux

```bash
# Problem: Go not in PATH
# Solution: Add to shell profile
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Problem: PostgreSQL authentication
# Solution: Configure pg_hba.conf
sudo nano /etc/postgresql/15/main/pg_hba.conf
# Change 'peer' to 'md5' for local connections
```

## 📚 Next Steps

After successful setup:

1. **Read the main [README.md](../README.md)** for API usage
2. **Explore [API examples](examples/)** for testing workflows
3. **Check [Swagger UI](http://localhost:8080/swagger/)** for interactive docs
4. **Run the test suite** to ensure everything works
5. **Start building your features** following the clean architecture pattern

## 🆘 Getting Help

- **Documentation**: Check [docs/](.) directory
- **Issues**: Common problems are documented above
- **Community**: Open an issue on GitHub
- **Testing**: Use Swagger UI for API exploration

---

**💡 Pro Tip**: Use the Swagger UI at http://localhost:8080/swagger/ to test your API during development!
