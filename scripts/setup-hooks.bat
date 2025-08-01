@echo off
REM Setup script for Git hooks (Windows version)
REM This script installs the Git hooks for the Go Fiber Template project

echo Setting up Git hooks for Go Fiber Template...

REM Check if we're in a Git repository
if not exist ".git" (
    echo Error: This is not a Git repository
    exit /b 1
)

REM Create .git/hooks directory if it doesn't exist
if not exist ".git\hooks" mkdir ".git\hooks"

REM Install commit-msg hook
if exist ".githooks\commit-msg" (
    copy ".githooks\commit-msg" ".git\hooks\commit-msg" >nul
    echo ✓ Installed commit-msg hook
) else (
    echo ⚠ commit-msg hook not found in .githooks/
)

REM Install pre-commit hook
if exist ".githooks\pre-commit" (
    copy ".githooks\pre-commit" ".git\hooks\pre-commit" >nul
    echo ✓ Installed pre-commit hook
) else (
    echo ⚠ pre-commit hook not found in .githooks/
)

REM Set up Git commit message template
if exist ".gitmessage" (
    git config commit.template .gitmessage
    echo ✓ Set up Git commit message template
) else (
    echo ⚠ .gitmessage file not found
)

REM Check if required tools are installed
echo.
echo ℹ Checking for required tools...

REM Check Go
go version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ Go is installed
    go version
) else (
    echo ✗ Go is not installed or not in PATH
)

REM Check gofmt (comes with Go)
gofmt -h >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ gofmt is available
) else (
    echo ⚠ gofmt is not available
)

REM Check goimports
goimports -h >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ goimports is available
) else (
    echo ⚠ goimports is not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest
)

REM Check golangci-lint
golangci-lint version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ golangci-lint is available
    golangci-lint version
) else (
    echo ⚠ golangci-lint is not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
)

REM Check pre-commit (optional)
pre-commit --version >nul 2>&1
if %errorlevel% equ 0 (
    echo ✓ pre-commit is available
    pre-commit --version
    echo ℹ You can also use pre-commit by running: pre-commit install
) else (
    echo ℹ pre-commit is not installed. You can install it with: pip install pre-commit
)

echo.
echo ✓ Git hooks setup complete!
echo ℹ The hooks will now run automatically on commit
echo ℹ To bypass hooks temporarily, use: git commit --no-verify

echo.
echo ℹ Recommended next steps:
echo   1. Install missing tools (goimports, golangci-lint)
echo   2. Run 'make.bat setup' to install all development tools
echo   3. Test the hooks with a sample commit

pause