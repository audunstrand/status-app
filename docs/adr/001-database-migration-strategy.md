# ADR 001: Database Migration Strategy

**Date**: 2025-12-09  
**Status**: Proposed

## Context

We currently run database migrations manually via `psql` on Fly.io. This is error-prone and not automated. We need a reliable, automated way to apply database schema changes during deployments.

## Decision

We need to choose an approach for automated database migrations.

## Options Considered

### Option 1: Custom Go Migration Service

**Description**: Build a custom migration service in Go that reads SQL files from `/migrations` and executes them in order.

**Pros**:
- No external dependencies
- Full control over migration logic
- Integrates with existing Go codebase
- Simple deployment as Fly.io one-off job
- Tracks migration state in database
- ~200 lines of code

**Cons**:
- Need to implement ordering, versioning, and rollback logic ourselves
- Need to maintain our own migration runner
- Less battle-tested than existing tools

### Option 2: golang-migrate/migrate Library

**Description**: Use the popular `golang-migrate/migrate` library (https://github.com/golang-migrate/migrate)

**Pros**:
- Battle-tested, widely used (11k+ stars)
- Supports multiple databases
- CLI tool available
- Built-in rollback support
- Active maintenance

**Cons**:
- Additional dependency (though well-maintained)
- More features than we need
- Need to wrap in our own service for Fly.io deployment
- Learning curve for library API

### Option 3: Fly.io Release Command

**Description**: Use Fly.io's release command feature to run migrations before each deployment

**Pros**:
- Built-in Fly.io feature
- Runs automatically before deployment
- Simple configuration in fly.toml

**Cons**:
- Still need a migration runner (Option 1 or 2)
- Blocks deployment if migrations fail (could be pro or con)
- Less control over when migrations run

## Recommendation

**Option 1: Custom Go Migration Service** combined with **Option 3: Fly.io Release Command**

**Rationale**:
- Keeps dependencies minimal
- Simple, understandable implementation
- Our migration needs are straightforward (sequential SQL files)
- Can evolve to Option 2 later if needs grow
- Fly.io release command provides automation

**Implementation**:
1. Build custom migration service in `cmd/migrate`
2. Reads `.up.sql` and `.down.sql` files from `/migrations`
3. Tracks applied migrations in `schema_migrations` table
4. Deploy as Fly.io release command (runs before each deployment)

## Questions for Review

1. Do you prefer the custom solution or would you rather use `golang-migrate/migrate`?
2. Should migrations run automatically as release commands, or manually via `fly deploy -c fly.migrate.toml`?
3. Any concerns about the proposed approach?
