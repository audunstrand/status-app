# Code Cleanup Recommendations

## ✅ Completed Cleanup

### Removed Obsolete Service Directories
- ✅ Removed `cmd/commands` - Replaced by cmd/backend
- ✅ Removed `cmd/api` - Replaced by cmd/backend
- ✅ Removed `cmd/projections` - Replaced by cmd/backend
- ✅ Updated Makefile to build only 3 services (backend, slackbot, scheduler)

### Updated Documentation
- ✅ docs/ARCHITECTURE.md - Shows new 3-service + 1-database architecture
- ✅ README.md - Quick start for consolidated services

## Files to Update

### 1. Documentation (Still Needs Update)

**docs/ARCHITECTURE.md** - Update for consolidated architecture:
- ❌ Mentions 5 separate services
- ❌ References separate Event Store DB and Projection DB
- ✅ Should show: Backend (Commands+API+Projections), Slackbot, Scheduler
- ✅ Should show: Single status-app-db with events/projections schemas

**docs/DEPLOYMENT.md** - Update deployment instructions:
- ❌ Shows deploying each service individually
- ✅ Should show deploying backend + slackbot + scheduler

**README.md** - Update quick start:
- ❌ Shows running 5 separate services
- ❌ References separate databases
- ✅ Should show running 3 services (backend, slackbot, scheduler)
- ✅ Should reference single database with schemas

**TODO.md** - Needs review:
- Migration service task still relevant (automation)
- Real-time projections still relevant
- Scheduler reminders still relevant
- Minor fixes section needs update (health checks already added to backend)

### 2. Migration Files (Potentially Redundant)

**migrations/001_create_events_table.* and 002_create_projections.***
- These were for separate databases (old architecture)
- migrations/003 and 004 are for consolidated database (new architecture)
- Once old databases are completely gone, could mark 001-002 as deprecated
- Keep for now for reference/rollback capability

### 3. Fly Configs (Will be deprecated after Phase 2)

**fly.toml** (commands), **fly.api.toml**, **fly.projections.toml**
- Will be replaced by fly.backend.toml
- Should be removed after Prompt 2.7 (decommission old services)
- Keep for now during transition

## No Cleanup Needed

### Code Quality ✅
- No duplicate code found
- Tests are well organized
- Internal packages are clean
- No dead code in Go files

### Configuration ✅
- .gitignore is appropriate
- Dockerfile supports all services
- GitHub Actions workflow will be updated later

## Recommended Actions

### Now (Before continuing Phase 2):
1. **Update docs/ARCHITECTURE.md** - Show new 3-service + 1-database architecture
2. **Update README.md** - Quick start for consolidated services

### After Phase 2 Complete (Prompt 2.7):
3. **Remove old fly configs** - fly.toml, fly.api.toml, fly.projections.toml
4. **Update docs/DEPLOYMENT.md** - Show deploying 3 services
5. **Update TODO.md** - Remove/update completed items
6. **Mark old migrations as deprecated** - Add note that 001-002 are for legacy architecture

### Optional:
7. **Add CHANGELOG.md** - Document the consolidation work
8. **Update .github/workflows/deploy.yml** - Will happen in Prompt 3.2
