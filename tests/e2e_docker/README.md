# Docker-based E2E Tests

This directory contains end-to-end tests that run against actual Docker containers built from the project's Dockerfiles.

## Overview

Unlike the testcontainer-based tests that use `httptest.Server`, these tests:
- Build and run actual Docker images using `docker-compose`
- Test real HTTP communication between services
- Verify the complete deployment configuration
- Can be used for integration testing before deployment

## Running Tests

### Quick test run
```bash
cd tests/e2e_docker
make test
```

### Manual testing
```bash
# Start all services
make up

# Services will be available at:
# - Commands: http://localhost:8081
# - API: http://localhost:8082

# Test manually with curl
curl -X POST http://localhost:8081/commands/submit-update \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-secret-123" \
  -d '{"team_id":"team-1","content":"Test update","author":"tester"}'

# View logs
make logs

# Stop services
make down
```

### Cleanup
```bash
make clean
```

## Test Structure

- `docker-compose.test.yml` - Docker Compose configuration for test environment
- `e2e_docker_test.go` - HTTP-based tests against running containers
- `Makefile` - Helper commands for running tests

## Tests Included

1. **TestDockerE2E_SubmitStatusUpdate** - Verifies status update submission
2. **TestDockerE2E_AuthenticationRequired** - Validates auth middleware
3. **TestDockerE2E_APIEndpoints** - Tests API query endpoints
4. **TestDockerE2E_EndToEndFlow** - Complete flow from submission to query

## Requirements

- Docker and Docker Compose
- Go 1.23+
- testcontainers-go library

## Environment Variables

The test environment uses:
- `AUTH_TOKEN=test-secret-123` for commands service
- `AUTH_TOKEN=test-secret-456` for API service
- PostgreSQL databases for event store and projections
