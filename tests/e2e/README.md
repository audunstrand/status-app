# End-to-End Tests

## Overview
E2E tests verify the complete flow of the event-sourced status app using **Testcontainers** with real PostgreSQL.

## Requirements
- Docker Desktop installed and running
- Go 1.21+

## Running Tests

### 1. Start Docker Desktop
Make sure Docker is running:
```bash
docker ps
```

### 2. Run E2E Tests
```bash
# Run all E2E tests
go test ./tests/e2e -v -timeout=5m

# Run specific test
go test ./tests/e2e -v -run TestStatusUpdateFlow
```

### 3. What Gets Tested

**TestStatusUpdateFlow:**
- Register team → Store event
- Submit status update → Store event  
- Build projections → Query data
- Verify complete flow works end-to-end

**TestTeamManagementFlow:**
- Register team → Query team
- Update team → Query updated team
- Verify event history preserved
- Verify projections stay consistent

## How It Works

### Testcontainers
Tests automatically:
1. Pull `postgres:16-alpine` image
2. Start PostgreSQL container
3. Run migrations
4. Execute tests
5. Clean up container

### Test Structure
```
tests/
├── testutil/
│   └── database.go      # Testcontainers setup
└── e2e/
    ├── helpers.go       # Test event store & projections
    ├── status_flow_test.go
    └── team_management_test.go
```

## Benefits

✅ **Real PostgreSQL** - 100% compatible, no mocks  
✅ **Isolated** - Each test gets fresh database  
✅ **Automatic cleanup** - Containers removed after tests  
✅ **CI/CD Ready** - Works in GitHub Actions  

## Troubleshooting

### Docker not running
```
Error: rootless Docker not found
```
**Solution:** Start Docker Desktop

### Slow first run
First run downloads postgres:16-alpine image (~80MB).  
Subsequent runs are fast (~2-3 seconds per test).

### Port conflicts
Testcontainers uses random ports, so no conflicts.

## CI/CD

GitHub Actions example:
```yaml
- name: Run E2E tests
  run: go test ./tests/e2e -v -timeout=5m
```

Docker is pre-installed in GitHub Actions runners.
