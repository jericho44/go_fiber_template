# Code Quality Setup

This document describes the code quality tools and processes set up for the Go Fiber Template project.

## Overview

The project uses multiple tools and processes to ensure code quality:

1. **golangci-lint** - Comprehensive Go linter
2. **gofmt** - Go code formatter
3. **goimports** - Go import formatter
4. **go vet** - Go static analysis tool
5. **Git hooks** - Automated quality checks on commit
6. **Pre-commit** - Additional quality checks (optional)

## Tools Configuration

### golangci-lint

Configuration file: `.golangci.yml`

The configuration enables multiple linters including:

- Standard Go linters (errcheck, gosimple, govet, etc.)
- Security linters (gosec)
- Style linters (gofmt, goimports, revive)
- Complexity linters (cyclop, gocyclo, gocognit)
- Performance linters (prealloc)

### Pre-commit Hooks

Configuration file: `.pre-commit-config.yaml`

Includes hooks for:

- Go formatting and linting
- General file checks (trailing whitespace, YAML/JSON validation)
- Security checks (detect-secrets)
- Markdown linting

## Git Hooks

### Pre-commit Hook

Location: `.githooks/pre-commit`

The pre-commit hook runs the following checks:

1. Code formatting (gofmt)
2. Import formatting (goimports)
3. Static analysis (go vet)
4. Linting (golangci-lint)
5. Unit tests
6. Debug statement detection
7. File size warnings

### Commit Message Hook

Location: `.githooks/commit-msg`

Validates commit messages according to conventional commit format:

```
<type>(<scope>): <subject>
```

Supported types:

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test changes
- `build`: Build system changes
- `ci`: CI configuration changes
- `chore`: Other changes
- `revert`: Revert previous commit

## Setup Instructions

### Automatic Setup

Run the setup script to install Git hooks and check for required tools:

**Linux/macOS:**

```bash
./scripts/setup-hooks.sh
```

**Windows:**

```cmd
.\scripts\setup-hooks.bat
```

### Manual Setup

1. **Install development tools:**

   ```bash
   # Using Makefile
   make setup

   # Or manually
   go install golang.org/x/tools/cmd/goimports@latest
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. **Install Git hooks:**

   ```bash
   cp .githooks/* .git/hooks/
   chmod +x .git/hooks/*
   git config commit.template .gitmessage
   ```

3. **Install pre-commit (optional):**
   ```bash
   pip install pre-commit
   pre-commit install
   ```

## Usage

### Running Quality Checks

**Using Makefile:**

```bash
# Format code
make format

# Run linter
make lint

# Run all quality checks
make check
```

**Using batch file (Windows):**

```cmd
# Format code
.\make.bat format

# Run linter
.\make.bat lint
```

**Manual commands:**

```bash
# Format code
gofmt -w .
goimports -w .

# Run linter
golangci-lint run

# Run static analysis
go vet ./...

# Run tests
go test ./...
```

### Bypassing Hooks

If you need to bypass the Git hooks temporarily:

```bash
git commit --no-verify
```

## IDE Integration

### VS Code

Install the following extensions:

- Go (official Go extension)
- golangci-lint
- Error Lens

Add to your VS Code settings:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintFlags": ["--fast"],
  "editor.formatOnSave": true,
  "go.formatTool": "goimports"
}
```

### GoLand/IntelliJ

1. Enable golangci-lint in Settings → Tools → Go Linter
2. Enable goimports in Settings → Tools → File Watchers
3. Configure code style according to gofmt

## Continuous Integration

The quality checks are also run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v3
  with:
    version: latest

- name: Run tests
  run: go test -v ./...
```

## Troubleshooting

### Common Issues

1. **golangci-lint not found:**

   ```bash
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```

2. **goimports not found:**

   ```bash
   go install golang.org/x/tools/cmd/goimports@latest
   ```

3. **Hooks not running:**

   ```bash
   chmod +x .git/hooks/*
   ```

4. **Commit message validation failing:**
   - Check the commit message format
   - Use the template: `git config commit.template .gitmessage`

### Disabling Specific Linters

To disable a specific linter for a file or line:

```go
//nolint:lintername
func problematicFunction() {
    // code here
}
```

To disable for the entire file:

```go
//nolint
package main
```

## Best Practices

1. **Run quality checks before committing:**

   ```bash
   make check
   ```

2. **Fix formatting issues automatically:**

   ```bash
   make format
   ```

3. **Address linter warnings promptly**

4. **Write meaningful commit messages**

5. **Keep functions and files reasonably sized**

6. **Add comments for exported functions and types**

7. **Use meaningful variable and function names**

8. **Handle errors appropriately**

## Configuration Customization

### Modifying golangci-lint Rules

Edit `.golangci.yml` to:

- Enable/disable specific linters
- Adjust complexity thresholds
- Add exclusions for specific files or patterns
- Customize severity levels

### Modifying Git Hooks

Edit files in `.githooks/` to:

- Add new checks
- Modify existing validation rules
- Change error messages
- Adjust thresholds

### Modifying Pre-commit Configuration

Edit `.pre-commit-config.yaml` to:

- Add new hooks
- Update hook versions
- Configure hook-specific settings
- Add exclusions
