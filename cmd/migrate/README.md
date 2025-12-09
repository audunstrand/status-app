# Database Migration Service

Automated database migration runner using [golang-migrate/migrate](https://github.com/golang-migrate/migrate).

## Overview

This service runs database migrations automatically before each backend deployment using Fly.io's release command feature.

## How It Works

1. Backend deploys to Fly.io
2. Fly.io runs `/bin/migrate` as a release command (before starting the service)
3. Migrations from `/migrations` are applied in order
4. If migrations succeed, backend starts
5. If migrations fail, deployment is blocked (prevents bad state)

## Environment Variables

- `DATABASE_URL` (required): PostgreSQL connection string
- `MIGRATIONS_PATH` (optional): Path to migrations directory (default: `file:///migrations`)

## Migration Files

Expected format: `NNN_description.up.sql` and `NNN_description.down.sql`

- Numeric prefix determines execution order (e.g., `001`, `002`, `003`)
- `.up.sql` for forward migrations
- `.down.sql` for rollback migrations

Example:
```
migrations/
  001_create_events_table.up.sql
  001_create_events_table.down.sql
  002_create_projections.up.sql
  002_create_projections.down.sql
```

## Local Testing

```bash
export DATABASE_URL="postgres://localhost:5432/statusapp?sslmode=disable"
export MIGRATIONS_PATH="file://migrations"

./bin/migrate
```

## Deployment

Migrations run automatically on backend deployment via Fly.io release command.

See: [fly.backend.toml](../../fly.backend.toml)

## Rollback

To rollback migrations manually:
```bash
# Install golang-migrate CLI
brew install golang-migrate

# Rollback one migration
migrate -path migrations -database "$DATABASE_URL" down 1
```

## References

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [Fly.io release commands](https://fly.io/docs/reference/configuration/#run-one-off-commands-before-releasing-a-deployment)
