# Session Resume - 2025-12-09

## What We Accomplished Today

### 1. Removed Unused Code
- **Removed `CommandType()` method** from Command interface
- All commands still work, interface is now a simple marker interface
- Tests passing ✅

### 2. Improved Copilot Instructions
- **Added ADR requirement** for major changes to `.github/copilot-instructions.md`
- For new components, technologies, or major changes:
  - Document 2-3 alternative approaches
  - Ask for user approval before implementing
- This ensures better architectural decisions

### 3. URL Structure Refactoring ✅ COMPLETE
- Changed from CQRS-exposing URLs to RESTful endpoints:
  - `POST /commands/submit-update` → `POST /teams/{id}/updates`
  - `POST /commands/register-team` → `POST /teams`
  - `GET /api/teams` → `GET /teams`
  - `GET /api/updates` → `GET /updates`
- Moved `team_id` from request body to URL path
- Updated all tests (backend, e2e, e2e_docker, auth middleware)
- All tests passing ✅
- Deployed successfully ✅

### 4. Automated Database Migrations ✅ COMPLETE
- **Created ADR 001** documenting migration approaches
- **User chose**: golang-migrate library + Fly.io release commands
- **Implementation**:
  - Added `cmd/migrate` service using `github.com/golang-migrate/migrate/v4`
  - Configured Fly.io release command in `fly.backend.toml`
  - Migrations run automatically before each backend deployment
  - Deployment blocked if migrations fail (safety feature)
  - Updated Dockerfile to include migrate binary
- **Tests**: Environment validation tests added
- All tests passing ✅
- Deployed successfully ✅

## Current Project State

### Architecture
- **3 services**: backend (commands + API + projections), slackbot, scheduler
- **1 database**: PostgreSQL with `events` and `projections` schemas
- **Deployed on**: Fly.io with GitHub Actions CI/CD
- **Migrations**: Automated via golang-migrate + Fly.io release commands

### Test Coverage
- internal/auth: 100%
- internal/commands: 76.6%
- internal/events: 74.5%
- internal/projections: 71.4%
- internal/config: 100%
- E2E tests: 77.6%

### Recent Commits (newest first)
1. `2ec115e` - Update TODO - mark migrations and URL refactoring complete
2. `f1e7c90` - Add automated database migrations with golang-migrate
3. `548c319` - Add ADR for database migration strategy
4. `a38c931` - Add ADR requirement to copilot instructions
5. `9e20f74` - Refactor URL structure to RESTful endpoints
6. `01335e3` - Remove unused CommandType() method from Command interface

## What's Next

### Remaining Tasks (~2.5 hours)

1. **Real-Time Projections** (1h)
   - Replace polling with PostgreSQL LISTEN/NOTIFY
   - Projections update immediately when events written
   - Implement `Subscribe()` method in postgres_store.go

2. **Scheduler Reminders** (1h)
   - Actually send Slack reminders on schedule
   - Parse poll schedule (e.g., "monday", "weekly")
   - Add `last_reminded_at` to teams table
   - Send Slack messages via API

3. **Minor Fixes** (30m)
   - Fix ignored NOTIFY error in postgres_store.go
   - Add Fly.io health checks to all services

## Important Files to Know

### Configuration
- `.github/copilot-instructions.md` - Development workflow (TDD, ADRs, deployment)
- `docs/adr/001-database-migration-strategy.md` - Migration decision record
- `TODO.md` - Remaining tasks

### Key Services
- `cmd/backend/main.go` - Combined API + Commands + Projections
- `cmd/migrate/main.go` - Database migration service
- `cmd/slackbot/main.go` - Slack integration
- `cmd/scheduler/main.go` - Reminder scheduling

### Deployment
- `fly.backend.toml` - Backend config with release_command for migrations
- `fly.slackbot.toml` - Slackbot config
- `fly.scheduler.toml` - Scheduler config
- `Dockerfile` - Builds all services + migrate binary
- `.github/workflows/deploy.yml` - CI/CD pipeline

## Notes for Next Session

- All tests are passing
- All services successfully deployed
- Migration service works automatically on deployment
- User prefers golang-migrate over custom solution
- User is happy with Fly.io release command approach (not app startup)
- Follow ADR process for major changes going forward

## Quick Start Commands

```bash
# Run tests
make test

# Build all services
make build

# Check recent commits
git log --oneline -5

# Check deployment status
gh run list --limit 3
```

## Environment

- **Repository**: Not in a git repository (working in `status-app/`)
- **Git root**: `/Users/audunfauchaldstrand/code/snippets/status-app`
- **GitHub**: `audunstrand/status-app`
- **OS**: macOS (Darwin)
