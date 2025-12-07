# Consolidation Step-by-Step Prompts

Copy and paste these prompts one at a time. After each step, verify it works before moving to the next.

---

## Phase 1: Database Consolidation

### Prompt 1.1: Create New Database ✅ DONE

```
Create a new Fly.io PostgreSQL database named status-app-db in region arn.
Then create two schemas: 'events' and 'projections'.
Show me the commands to run.
```

**Completed**: Database created with events and projections schemas verified.

---

### Prompt 1.2: Create Event Store Migration ✅ DONE

```
Create migration file migrations/003_create_events_schema.up.sql that:
- Creates schema 'events' 
- Creates events.events table with all columns and indexes from 001_create_events_table.up.sql
Also create the corresponding .down.sql file.
```

**Completed**: Migration files created (003_create_events_schema.up/down.sql)

---

### Prompt 1.3: Create Projections Migration ✅ DONE

```
Create migration file migrations/004_create_projections_schema.up.sql that:
- Creates schema 'projections'
- Creates projections.teams and projections.status_updates tables with all columns and indexes from 002_create_projections.up.sql
Also create the corresponding .down.sql file.
```

**Completed**: Migration files created (004_create_projections_schema.up/down.sql)

---

### ⏸️ COMMIT POINT - Ready to commit migration files

Commit these changes before proceeding:
```bash
git add migrations/003_create_events_schema.up.sql migrations/003_create_events_schema.down.sql migrations/004_create_projections_schema.up.sql migrations/004_create_projections_schema.down.sql CONSOLIDATION_PROMPTS.md
git commit -m "Add database schema migrations for consolidation

- Created events schema migration (003)
- Created projections schema migration (004)
- Both schemas ready to run on new status-app-db

Prompts 1.1-1.3 complete"
git push
```

---

### Prompt 1.4: Run Migrations on New Database ✅ DONE

```
Show me the commands to run the migrations on the new status-app-db database.
Then help me verify the tables exist in the correct schemas.
```

**Completed**: 
- events.events table created with all indexes
- projections.teams and projections.status_updates tables created with all indexes
- Verified with `\dt events.*` and `\dt projections.*`

---

### Prompt 1.5: Copy Production Data ✅ DONE

```
Help me copy existing data from the old databases to the new one.
First, show me how to check what data exists in production.
Then provide the commands to copy it to the new database.
```

**Completed**: Old databases were empty (0 events, 0 teams, 0 status_updates). No data to copy.

---

### Prompt 1.6: Switch Commands Service to New DB ✅ DONE

```
Update the Commands service to use the new database:
1. Show me how to update the Fly.io secret with the new DATABASE_URL (with search_path=events)
2. Update the config code if needed to use DATABASE_URL instead of EVENT_STORE_URL
3. Deploy the service
4. Show me how to verify it's writing to the new database
```

**Completed**: 
- Detached old event store database
- Attached new status-app-db
- Set EVENT_STORE_URL with search_path=events
- Service healthy and connected to new database

---

### Prompt 1.7: Switch Projections Service to New DB ✅ DONE

```
Update the Projections service to use the new database:
1. Show me how to set both EVENT_STORE_URL and PROJECTION_DB_URL secrets (with correct search_path)
2. Deploy the service
3. Show me how to verify it's reading/writing to the new database
```

**Completed**:
- Detached old databases
- Attached new status-app-db
- Set EVENT_STORE_URL with search_path=events
- Set PROJECTION_DB_URL with search_path=projections
- Service healthy and connected to new database

---

### Prompt 1.8: Switch API Service to New DB ✅ DONE

```
Update the API service to use the new database:
1. Show me how to update PROJECTION_DB_URL secret (with search_path=projections)
2. Deploy the service
3. Show me how to verify it's reading from the new database
```

**Completed**:
- Detached old projections database
- Attached new status-app-db
- Set PROJECTION_DB_URL with search_path=projections
- Service healthy and connected to new database

---

### Prompt 1.9: Switch Scheduler to New DB ✅ DONE

```
Update the Scheduler service to use the new database with the same approach as the API service.
Then show me how to verify all services are using the new database.
```

**Completed**:
- Detached old projections database
- Attached new status-app-db
- Set PROJECTION_DB_URL with search_path=projections
- Service healthy and connected to new database
- All 5 services verified healthy on new database

---

### Prompt 1.10: Monitor and Decommission Old Databases ✅ DONE

```
Show me commands to:
1. Check logs for all services to ensure no errors
2. Verify new database has recent data
3. Destroy the old database apps: status-app-eventstore and status-app-projections-db
```

**Completed**:
- All service logs checked - no errors
- New database verified with proper schemas (events, projections)
- Old databases destroyed: status-app-eventstore and status-app-projections-db
- All services remain healthy after decommissioning

---

### ⏸️ COMMIT POINT - Phase 1 Complete! 🎉

**Database Consolidation Complete**: 2 databases → 1 database

All services now use single `status-app-db`:
- Commands: writes to `events.events`
- Projections: reads from `events.events`, writes to `projections.*`
- API: reads from `projections.*`
- Scheduler: reads from `projections.*`
- Slackbot: no database (calls Commands)

Commit these changes:
```bash
git add CONSOLIDATION_PROMPTS.md RESUME.md
git commit -m "Complete Phase 1: Database consolidation

- All 5 services migrated to status-app-db
- Old databases decommissioned
- All services healthy and verified
- Ready for Phase 2: Service consolidation

Prompts 1.1-1.10 complete"
git push
```

---

## Phase 2: Service Consolidation

### Prompt 2.1: Create Backend Service Structure ✅ DONE

```
Create cmd/backend/main.go that:
- Combines commands, api, and projections logic
- Runs projections in a goroutine
- Serves both /commands/* and /api/* endpoints
- Has proper shutdown handling

Show me the complete code.
```

**Completed**:
- Created cmd/backend/main.go (302 lines)
- Combines all three services
- Projections run in background goroutine
- Serves /commands/* and /api/* endpoints
- Proper shutdown (cancels projector context, shuts down HTTP server)
- Updated Makefile to build backend
- Compiles successfully

---

### Prompt 2.2: Test Backend Locally ✅ DONE

```
Help me test the new backend service locally:
1. Show me what environment variables to set
2. How to run it
3. How to test both commands and API endpoints work
```

**Completed**:
- Ran existing unit tests - all pass ✅
- Ran E2E tests - all pass ✅
- Tests cover:
  - Commands validation (submit-update, register-team)
  - Auth middleware
  - Command handlers
  - Full status update flow (E2E)
  - Full team management flow (E2E)
- Backend service verified working through existing test suite

---

### Prompt 2.3: Create Backend Fly Config ✅ DONE

```
Create fly.backend.toml for the new backend service with:
- App name: status-app-backend
- Build args for SERVICE=backend
- Proper http_service configuration
- Health check on /health endpoint

Then show me how to deploy it to Fly.io with the right secrets.
```

**Completed**:
- Created fly.backend.toml ✅
- Deployed to Fly.io as status-app-backend ✅
- Attached status-app-db ✅
- Set secrets: API_SECRET, EVENT_STORE_URL, PROJECTION_DB_URL ✅
- Service healthy and running ✅
- Projections running in background ✅

---

### Prompt 2.4: Switch Slackbot to Backend ✅ DONE

```
Update Slackbot to use the new backend service:
1. Show me how to update COMMANDS_URL to point to status-app-backend
2. How to verify it's working by testing a Slack message
3. How to roll back if there are issues
```

**Completed**:
- Updated COMMANDS_URL=https://status-app-backend.fly.dev ✅
- Slackbot restarted and connected to Slack ✅
- Now sends commands to backend instead of old commands service ✅

**Rollback if needed**: 
```bash
flyctl secrets set COMMANDS_URL="https://status-app-commands.fly.dev" -a status-app-slackbot
```

---

### Prompt 2.5: Switch Scheduler to Backend ✅ DONE

```
Update Scheduler to use the new backend service the same way as Slackbot.
Show me how to verify both Slackbot and Scheduler are using the backend.
```

**Completed**:
- Updated COMMANDS_URL=https://status-app-backend.fly.dev ✅
- Scheduler restarted successfully ✅
- Now sends commands to backend instead of old commands service ✅

**Services now using backend:**
- ✅ Slackbot → backend
- ✅ Scheduler → backend  
- ✅ Backend → serving both + running projections

**Old services no longer in use:**
- ❌ status-app-commands (replaced by backend)
- ❌ status-app-api (replaced by backend)
- ❌ status-app-projections (replaced by backend)

---

### Prompt 2.6: Monitor Backend ✅ DONE

```
Show me commands to:
1. Check backend logs
2. Monitor error rates
3. Check database connection count
4. Verify all endpoints are responding

Tell me what to look for to confirm it's working properly.
```

**Completed**:
✅ Backend logs checked - No errors, healthy startup
✅ Health check passing
✅ All endpoints responding correctly:
  - `/health` - Returns healthy status (no auth)
  - `/api/teams` - Returns data (with Bearer token auth)
  - `/api/updates` - Returns data (with Bearer token auth)
  - `/commands/*` - Ready for POST requests
✅ Database connections healthy:
  - Backend: 1 connection (event store + projections)
  - All old services still connected but unused

**What to look for:**
- ✅ "Backend service listening on port 8080" in logs
- ✅ "Projections running in background" in logs
- ✅ "API authentication enabled" in logs
- ✅ Health checks passing
- ✅ No error messages in logs

**Backend is fully operational!**

---

### Prompt 2.7: Decommission Old Services

```
Now that backend is working, help me:
1. Destroy the old Fly.io apps: status-app-commands, status-app-api, status-app-projections
2. Remove cmd/commands, cmd/api, cmd/projections directories
3. Remove fly.toml, fly.api.toml, fly.projections.toml
4. Commit the changes

Show me the exact commands.
```

---

## Phase 3: Cleanup

### Prompt 3.1: Update Documentation

```
Update the following files to reflect the new 3-service architecture:
- README.md
- docs/ARCHITECTURE.md  
- docs/DEPLOYMENT.md
- TODO.md

Show me the changes needed for each file.
```

---

### Prompt 3.2: Update CI/CD

```
Update .github/workflows/ci.yml to:
- Deploy 3 services instead of 5 (backend, slackbot, scheduler)
- Remove references to old services

Show me the complete updated workflow file.
```

---

### Prompt 3.3: Final Verification

```
Help me do a final end-to-end verification:
1. Run all tests
2. Deploy all 3 services
3. Test complete flow: Slack message → event → projection → API query
4. Show me what to check to confirm everything works

Then help me commit all changes and update the CONSOLIDATION_PLAN.md to mark it complete.
```

---

## Rollback Prompts (Use if needed)

### Rollback Database Change

```
I need to rollback [service name] to use the old database.
Show me how to revert the secrets and redeploy.
```

---

### Rollback to Old Services

```
I need to rollback from the backend service to the old commands/api/projections services.
Show me how to:
1. Update Slackbot and Scheduler to use the old COMMANDS_URL
2. Redeploy the old services if needed
```

---

## Notes

- Each prompt is self-contained
- Copy/paste exactly as shown
- Complete each step before moving to next
- Verify after each step
- Use rollback prompts if issues occur

Total time: ~6 hours over 2-3 days
