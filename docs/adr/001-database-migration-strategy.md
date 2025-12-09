# ADR 001: Database Migration Strategy

**Date**: 2025-12-09  
**Status**: Accepted

## Context

We currently run database migrations manually via `psql` on Fly.io. This is error-prone and not automated. We need a reliable, automated way to apply database schema changes during deployments.

## Decision

**Use golang-migrate/migrate library with Fly.io release commands for automatic migration on deployment.**

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

### Option 2: golang-migrate/migrate Library ✅ **CHOSEN**

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

### Option 3: Fly.io Release Command ✅ **CHOSEN**

**Description**: Use Fly.io's release command feature to run migrations before each deployment

**Pros**:
- Built-in Fly.io feature
- Runs automatically before deployment
- Simple configuration in fly.toml

**Cons**:
- Still need a migration runner (Option 1 or 2)
- Blocks deployment if migrations fail (could be pro or con)
- Less control over when migrations run

## Rationale

**Option 2 + Option 3**: golang-migrate library with automatic release commands

- Battle-tested and reliable
- Automatic migration on every deployment
- Handles edge cases we might miss in custom implementation
- Active maintenance and community support
- Blocks bad deployments if migrations fail (safety feature)

## Implementation Plan

1. Add `github.com/golang-migrate/migrate/v4` dependency
2. Create `cmd/migrate` service wrapping the library
3. Configure Fly.io release command in `fly.backend.toml`
4. Migrations run automatically before each backend deployment

## Consequences

- **Positive**: Migrations automated, reliable, and battle-tested
- **Positive**: Deployment blocked if migrations fail (prevents bad state)
- **Negative**: Additional dependency to maintain (mitigated by active project)
- **Negative**: Deployment slightly slower due to migration step

