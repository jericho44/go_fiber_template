@echo off
REM Go Fiber Template Windows Build Script
REM This batch file provides common development tasks for Windows users

if "%1"=="" goto help
if "%1"=="help" goto help
if "%1"=="dev" goto dev
if "%1"=="build" goto build
if "%1"=="test" goto test
if "%1"=="test-coverage" goto test-coverage
if "%1"=="format" goto format
if "%1"=="lint" goto lint
if "%1"=="clean" goto clean
if "%1"=="db-migrate-up" goto db-migrate-up
if "%1"=="db-migrate-down" goto db-migrate-down
if "%1"=="db-seed" goto db-seed
if "%1"=="docs" goto docs
if "%1"=="setup" goto setup
if "%1"=="setup-hooks" goto setup-hooks
goto help

:help
echo Go Fiber Template - Available Commands (Windows):
echo.
echo   dev              Run the application in development mode
echo   build            Build the application
echo   test             Run all tests
echo   test-coverage    Run tests with coverage report
echo   format           Format code with gofmt
echo   lint             Run go vet (basic linting)
echo   clean            Clean build artifacts
echo   db-migrate-up    Run database migrations up
echo   db-migrate-down  Run database migrations down
echo   db-seed          Run database seeders
echo   docs             Generate Swagger documentation
echo   setup            Setup development environment
echo   setup-hooks      Setup Git hooks only
echo.
echo Usage: make.bat [command]
echo.
echo Note: For full functionality, install make for Windows or use WSL
goto end

:dev
echo Starting development server...
go run cmd/server/main.go
goto end

:build
echo Building application...
if not exist bin mkdir bin
go build -o bin/server.exe cmd/server/main.go
go build -o bin/migrate.exe cmd/migrate/main.go
go build -o bin/seed.exe cmd/seed/main.go
echo Build complete!
goto end

:test
echo Running tests...
go test -v ./...
goto end

:test-coverage
echo Running tests with coverage...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo Coverage report generated: coverage.html
goto end

:format
echo Formatting code...
gofmt -w .
echo Code formatted!
goto end

:lint
echo Running go vet...
go vet ./...
goto end

:clean
echo Cleaning build artifacts...
go clean
if exist bin rmdir /s /q bin
if exist coverage.out del coverage.out
if exist coverage.html del coverage.html
echo Clean complete!
goto end

:db-migrate-up
echo Running database migrations up...
go run cmd/migrate/main.go up
goto end

:db-migrate-down
echo Running database migrations down...
go run cmd/migrate/main.go down
goto end

:db-seed
echo Running database seeders...
go run cmd/seed/main.go run --all
goto end

:docs
echo Generating Swagger documentation...
swag init -g cmd/server/main.go -o docs
echo Documentation generated!
goto end

:setup
echo Setting up development environment...
echo Installing development tools...
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/swaggo/swag/cmd/swag@latest
echo Setting up Git hooks...
if exist scripts\setup-hooks.bat (
    call scripts\setup-hooks.bat
) else (
    echo No setup script found for Git hooks
)
echo Development environment setup complete!
goto end

:setup-hooks
echo Setting up Git hooks...
if exist scripts\setup-hooks.bat (
    call scripts\setup-hooks.bat
) else (
    echo No setup script found for Git hooks
)
goto end

:end