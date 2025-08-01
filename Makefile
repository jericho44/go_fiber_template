# Go Fiber Template Makefile
# This Makefile provides common development tasks for the Go Fiber Template project

# Variables
BINARY_NAME=server
BINARY_DIR=bin
MAIN_PATH=cmd/server/main.go
MIGRATE_PATH=cmd/migrate/main.go
SEED_PATH=cmd/seed/main.go

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Docker parameters
DOCKER_IMAGE=go-fiber-template
DOCKER_TAG=latest
DOCKERFILE=docker/Dockerfile

# Database parameters
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_fiber_template

# Default target
.DEFAULT_GOAL := help

# Help target - shows available commands
.PHONY: help
help: ## Show this help message
	@echo "Go Fiber Template - Available Commands:"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
	@echo ""

# Development targets
.PHONY: dev
dev: ## Run the application in development mode
	@echo "Starting development server..."
	$(GOCMD) run $(MAIN_PATH)

.PHONY: dev-watch
dev-watch: ## Run the application with hot reload (requires air)
	@echo "Starting development server with hot reload..."
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not found. Install with: go install github.com/cosmtrek/air@latest"; \
		echo "Falling back to regular dev mode..."; \
		$(MAKE) dev; \
	fi

# Build targets
.PHONY: build
build: clean ## Build the application
	@echo "Building application..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	$(GOBUILD) -o $(BINARY_DIR)/migrate $(MIGRATE_PATH)
	$(GOBUILD) -o $(BINARY_DIR)/seed $(SEED_PATH)

.PHONY: build-linux
build-linux: clean ## Build for Linux (production)
	@echo "Building for Linux..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o $(BINARY_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o $(BINARY_DIR)/migrate-linux $(MIGRATE_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -a -installsuffix cgo -o $(BINARY_DIR)/seed-linux $(SEED_PATH)

.PHONY: build-windows
build-windows: clean ## Build for Windows
	@echo "Building for Windows..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/migrate.exe $(MIGRATE_PATH)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/seed.exe $(SEED_PATH)

.PHONY: build-mac
build-mac: clean ## Build for macOS
	@echo "Building for macOS..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/$(BINARY_NAME)-mac $(MAIN_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/migrate-mac $(MIGRATE_PATH)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINARY_DIR)/seed-mac $(SEED_PATH)

# Test targets
.PHONY: test
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

.PHONY: test-short
test-short: ## Run tests with short flag
	@echo "Running short tests..."
	$(GOTEST) -short -v ./...

.PHONY: test-race
test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...

.PHONY: benchmark
benchmark: ## Run benchmark tests
	@echo "Running benchmark tests..."
	$(GOTEST) -bench=. -benchmem ./...

# Code quality targets
.PHONY: lint
lint: ## Run linter (golangci-lint)
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo "Falling back to go vet..."; \
		$(GOVET) ./...; \
	fi

.PHONY: format
format: ## Format code with gofmt and goimports
	@echo "Formatting code..."
	$(GOFMT) -w .
	@if command -v goimports > /dev/null; then \
		goimports -w .; \
	else \
		echo "goimports not found. Install with: go install golang.org/x/tools/cmd/goimports@latest"; \
	fi

.PHONY: vet
vet: ## Run go vet
	@echo "Running go vet..."
	$(GOVET) ./...

.PHONY: check
check: format vet lint test ## Run all code quality checks

# Database targets
.PHONY: db-migrate-up
db-migrate-up: ## Run database migrations up
	@echo "Running database migrations up..."
	$(GOCMD) run $(MIGRATE_PATH) up

.PHONY: db-migrate-down
db-migrate-down: ## Run database migrations down
	@echo "Running database migrations down..."
	$(GOCMD) run $(MIGRATE_PATH) down

.PHONY: db-migrate-reset
db-migrate-reset: ## Reset database (down then up)
	@echo "Resetting database..."
	$(MAKE) db-migrate-down
	$(MAKE) db-migrate-up

.PHONY: db-seed
db-seed: ## Run database seeders
	@echo "Running database seeders..."
	$(GOCMD) run $(SEED_PATH) run --all

.PHONY: db-seed-users
db-seed-users: ## Run user seeder only
	@echo "Running user seeder..."
	$(GOCMD) run $(SEED_PATH) run --seeder user_seeder

.PHONY: db-seed-demo
db-seed-demo: ## Run demo data seeder
	@echo "Running demo data seeder..."
	$(GOCMD) run $(SEED_PATH) run --seeder demo_data_seeder

.PHONY: db-setup
db-setup: db-migrate-up db-seed ## Setup database (migrate + seed)

# Docker targets
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -f $(DOCKERFILE) -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-run-detached
docker-run-detached: ## Run Docker container in detached mode
	@echo "Running Docker container in detached mode..."
	docker run -d -p 8080:8080 --env-file .env --name $(DOCKER_IMAGE) $(DOCKER_IMAGE):$(DOCKER_TAG)

.PHONY: docker-stop
docker-stop: ## Stop Docker container
	@echo "Stopping Docker container..."
	docker stop $(DOCKER_IMAGE)

.PHONY: docker-remove
docker-remove: ## Remove Docker container
	@echo "Removing Docker container..."
	docker rm $(DOCKER_IMAGE)

.PHONY: docker-clean
docker-clean: docker-stop docker-remove ## Stop and remove Docker container
	@echo "Cleaning up Docker container..."

# Documentation targets
.PHONY: docs
docs: ## Generate Swagger documentation
	@echo "Generating Swagger documentation..."
	@if command -v swag > /dev/null; then \
		swag init -g $(MAIN_PATH) -o docs; \
	else \
		echo "swag not found. Install with: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

.PHONY: docs-serve
docs-serve: docs dev ## Generate docs and serve the application
	@echo "Documentation available at: http://localhost:8080/swagger/"

# Dependency management targets
.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	$(GOMOD) tidy
	$(GOGET) -u ./...

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "Vendoring dependencies..."
	$(GOMOD) vendor

# Cleanup targets
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@if [ -d $(BINARY_DIR) ]; then rm -rf $(BINARY_DIR); fi
	@if [ -f coverage.out ]; then rm coverage.out; fi
	@if [ -f coverage.html ]; then rm coverage.html; fi

.PHONY: clean-all
clean-all: clean ## Clean everything including vendor and cache
	@echo "Cleaning everything..."
	@if [ -d vendor ]; then rm -rf vendor; fi
	$(GOCLEAN) -cache
	$(GOCLEAN) -modcache

# Development environment setup
.PHONY: setup
setup: deps docs ## Setup development environment
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@$(GOCMD) install github.com/air-verse/air@latest
	@$(GOCMD) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@$(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@$(GOCMD) install github.com/swaggo/swag/cmd/swag@latest
	@echo "Setting up Git hooks..."
	@if [ -f "scripts/setup-hooks.sh" ]; then \
		chmod +x scripts/setup-hooks.sh && ./scripts/setup-hooks.sh; \
	elif [ -f "scripts/setup-hooks.bat" ]; then \
		scripts/setup-hooks.bat; \
	else \
		echo "No setup script found for Git hooks"; \
	fi
	@echo "Development environment setup complete!"

.PHONY: setup-hooks
setup-hooks: ## Setup Git hooks only
	@echo "Setting up Git hooks..."
	@if [ -f "scripts/setup-hooks.sh" ]; then \
		chmod +x scripts/setup-hooks.sh && ./scripts/setup-hooks.sh; \
	elif [ -f "scripts/setup-hooks.bat" ]; then \
		scripts/setup-hooks.bat; \
	else \
		echo "No setup script found for Git hooks"; \
	fi

# Production targets
.PHONY: prod-build
prod-build: clean test lint build-linux ## Build for production (with tests and linting)

.PHONY: prod-docker
prod-docker: clean test lint docker-build ## Build production Docker image (with tests and linting)

# Quick development workflow
.PHONY: quick
quick: format test ## Quick development check (format + test)

.PHONY: full-check
full-check: clean format vet lint test test-race test-coverage ## Full development check

# Server management
.PHONY: start
start: build ## Build and start the server
	@echo "Starting server..."
	./$(BINARY_DIR)/$(BINARY_NAME)

.PHONY: restart
restart: ## Restart the server (if running with systemd or similar)
	@echo "Restarting server..."
	@echo "Note: This assumes you have a service manager configured"

# Utility targets
.PHONY: version
version: ## Show Go version
	@$(GOCMD) version

.PHONY: env
env: ## Show Go environment
	@$(GOCMD) env

.PHONY: info
info: ## Show project information
	@echo "Go Fiber Template Project Information:"
	@echo "======================================"
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Binary Directory: $(BINARY_DIR)"
	@echo "Main Path: $(MAIN_PATH)"
	@echo "Docker Image: $(DOCKER_IMAGE):$(DOCKER_TAG)"
	@echo ""
	@$(MAKE) version