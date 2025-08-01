#!/bin/bash
# Setup script for Git hooks
# This script installs the Git hooks for the Go Fiber Template project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

echo "Setting up Git hooks for Go Fiber Template..."

# Check if we're in a Git repository
if [ ! -d ".git" ]; then
    print_error "This is not a Git repository"
    exit 1
fi

# Create .git/hooks directory if it doesn't exist
mkdir -p .git/hooks

# Install commit-msg hook
if [ -f ".githooks/commit-msg" ]; then
    cp .githooks/commit-msg .git/hooks/commit-msg
    chmod +x .git/hooks/commit-msg
    print_status "Installed commit-msg hook"
else
    print_warning "commit-msg hook not found in .githooks/"
fi

# Install pre-commit hook
if [ -f ".githooks/pre-commit" ]; then
    cp .githooks/pre-commit .git/hooks/pre-commit
    chmod +x .git/hooks/pre-commit
    print_status "Installed pre-commit hook"
else
    print_warning "pre-commit hook not found in .githooks/"
fi

# Set up Git commit message template
if [ -f ".gitmessage" ]; then
    git config commit.template .gitmessage
    print_status "Set up Git commit message template"
else
    print_warning ".gitmessage file not found"
fi

# Check if required tools are installed
print_info "Checking for required tools..."

# Check Go
if command -v go > /dev/null 2>&1; then
    print_status "Go is installed ($(go version))"
else
    print_error "Go is not installed or not in PATH"
fi

# Check gofmt (comes with Go)
if command -v gofmt > /dev/null 2>&1; then
    print_status "gofmt is available"
else
    print_warning "gofmt is not available"
fi

# Check goimports
if command -v goimports > /dev/null 2>&1; then
    print_status "goimports is available"
else
    print_warning "goimports is not installed. Install with: go install golang.org/x/tools/cmd/goimports@latest"
fi

# Check golangci-lint
if command -v golangci-lint > /dev/null 2>&1; then
    print_status "golangci-lint is available ($(golangci-lint version))"
else
    print_warning "golangci-lint is not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
fi

# Check pre-commit (optional)
if command -v pre-commit > /dev/null 2>&1; then
    print_status "pre-commit is available ($(pre-commit --version))"
    print_info "You can also use pre-commit by running: pre-commit install"
else
    print_info "pre-commit is not installed. You can install it with: pip install pre-commit"
fi

echo ""
print_status "Git hooks setup complete!"
print_info "The hooks will now run automatically on commit"
print_info "To bypass hooks temporarily, use: git commit --no-verify"

echo ""
print_info "Recommended next steps:"
echo "  1. Install missing tools (goimports, golangci-lint)"
echo "  2. Run 'make setup' to install all development tools"
echo "  3. Test the hooks with a sample commit"