# Testing Guide

## Quick Reference

```bash
# Run all unit tests
go test ./...

# Run Docker E2E tests (requires Docker)
cd tests/e2e_docker && make test

# Run integration tests
cd tests/e2e && go test ./...
```

## Test Types

### Unit Tests ⚡
**Location:** `./internal/...` and `./cmd/...`  
**Dependencies:** None  
**Speed:** Instant (~0.2s)

```bash
go test ./...
```

**What's tested:**
- ✅ Command handlers (3 tests)
- ✅ Request validation (10 tests)  
- ✅ Auth middleware (6 tests)
- ✅ All command types
- ✅ Edge cases and error handling

**Coverage:**
```bash
go test ./... -cover
```

### Integration Tests 🔗
**Location:** `./tests/e2e/...`  
**Dependencies:** Docker (testcontainers)  
**Speed:** ~5-10 seconds

```bash
cd tests/e2e && go test ./...
```

**What's tested:**
- Complete status update flow
- Team management flow
- Event storage and retrieval
- Projection building
- Database queries

### Docker E2E Tests 🐳
**Location:** `./tests/e2e_docker/...`  
**Dependencies:** Docker Desktop  
**Speed:** ~90 seconds (full build + test)

```bash
cd tests/e2e_docker
make test
```

**What's tested:**
- ✅ Real HTTP communication between services
- ✅ Actual Docker containers (not mocks)
- ✅ Authentication flow end-to-end
- ✅ Database migrations
- ✅ Service health checks
- ✅ Complete status submission flow

**Tests:**
1. `TestDockerE2E_SubmitStatusUpdate` - Submit and verify status update
2. `TestDockerE2E_AuthenticationRequired` - Verify auth is enforced
3. `TestDockerE2E_APIEndpoints` - Test query API
4. `TestDockerE2E_EndToEndFlow` - Complete flow with event+projection

**Available commands:**
```bash
make test   # Run full E2E test suite
make up     # Start services manually
make down   # Stop services
make logs   # View logs
make clean  # Full cleanup
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
│   ├── auth/
│   │   └── middleware_test.go      # Auth middleware tests
│   └── commands/
│       └── handler_test.go         # Command handler tests
├── cmd/
│   └── commands/
│       └── validation_test.go      # Validation tests
└── tests/
    ├── testutil/
    │   └── database.go             # Testcontainers setup
    ├── e2e/
    │   ├── helpers.go              # Integration test helpers
    │   ├── status_flow_test.go     # Status update flow
    │   └── team_management_test.go # Team management
    └── e2e_docker/
        ├── docker-compose.test.yml  # Service orchestration
        ├── e2e_docker_test.go       # HTTP E2E tests
        └── Makefile                 # Test commands
```

## CI/CD

GitHub Actions runs all tests on every push:

```yaml
# .github/workflows/ci.yml
- name: Run unit tests
  run: go test ./...
  
- name: Run E2E tests  
  run: cd tests/e2e && go test ./...
```

Docker E2E tests can be added to CI with:
```yaml
- name: Run Docker E2E tests
  run: cd tests/e2e_docker && make test
```
