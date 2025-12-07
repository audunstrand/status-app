# Testing

## Unit Tests

```bash
go test ./...
```

Tests: Command handlers, validation, auth middleware

## Integration Tests

```bash
cd tests/e2e && go test ./...
```

Uses testcontainers for PostgreSQL.

## Docker E2E Tests

```bash
cd tests/e2e_docker
make test
```

Full HTTP tests with real Docker containers. Takes ~90s.

Tests authentication, endpoints, and complete status submission flow.
