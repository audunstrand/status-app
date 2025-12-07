# Testing Guide

## Quick Reference

```bash
# Run all tests (unit + e2e)
make test

# Run only unit tests (fast, no Docker)
make test-unit

# Run only E2E tests (requires Docker)
make test-e2e

# Generate coverage report
make test-coverage

# Watch mode for development
make test-watch
```

## Test Types

### Unit Tests ⚡
**Location:** `./internal/...` and `./cmd/...`  
**Dependencies:** None  
**Speed:** Instant (~0.1s)

```bash
make test-unit
```

**What's tested:**
- Command handlers (6 tests)
- Request validation (10 tests)
- All command types
- Edge cases and error handling

### E2E Tests 🐳
**Location:** `./tests/e2e/...`  
**Dependencies:** Docker Desktop  
**Speed:** ~5-10 seconds

```bash
make test-e2e
```

**What's tested:**
- Complete status update flow
- Team management flow
- Event storage and retrieval
- Projection building
- Database queries

## Coverage Report

Generate HTML coverage report:

```bash
make test-coverage
```

Opens `coverage.html` showing line-by-line coverage.

## Watch Mode (Development)

Auto-run unit tests on file changes:

```bash
# Install entr first (macOS)
brew install entr

# Run watch mode
make test-watch
```

Press `Ctrl+C` to stop.

## CI/CD

The `make test` command runs both unit and E2E tests, perfect for CI:

```yaml
# GitHub Actions example
- name: Run tests
  run: make test
```

## Troubleshooting

### E2E tests fail with "Docker not found"
**Solution:** Start Docker Desktop before running `make test-e2e`

### Unit tests cached
Force re-run:
```bash
go clean -testcache && make test-unit
```

### Slow E2E tests
First run downloads postgres:16-alpine (~80MB).  
Subsequent runs reuse the image and are much faster.

## Test Structure

```
├── internal/
│   └── commands/
│       └── handler_test.go          # Command handler unit tests
├── cmd/
│   └── commands/
│       └── validation_test.go       # Validation unit tests
└── tests/
    ├── testutil/
    │   └── database.go              # Testcontainers setup
    └── e2e/
        ├── helpers.go               # Test helpers
        ├── status_flow_test.go      # Status update E2E test
        └── team_management_test.go  # Team management E2E test
```

## Best Practices

1. **Run unit tests frequently** - They're fast and catch most issues
2. **Run E2E tests before pushing** - Catch integration issues early
3. **Check coverage regularly** - Aim for >80% coverage
4. **Use watch mode during development** - Get instant feedback
