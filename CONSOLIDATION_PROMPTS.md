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

### Prompt 1.5: Copy Production Data

```
Help me copy existing data from the old databases to the new one.
First, show me how to check what data exists in production.
Then provide the commands to copy it to the new database.
```

---

### Prompt 1.6: Switch Commands Service to New DB

```
Update the Commands service to use the new database:
1. Show me how to update the Fly.io secret with the new DATABASE_URL (with search_path=events)
2. Update the config code if needed to use DATABASE_URL instead of EVENT_STORE_URL
3. Deploy the service
4. Show me how to verify it's writing to the new database
```

---

### Prompt 1.7: Switch Projections Service to New DB

```
Update the Projections service to use the new database:
1. Show me how to set both EVENT_STORE_URL and PROJECTION_DB_URL secrets (with correct search_path)
2. Deploy the service
3. Show me how to verify it's reading/writing to the new database
```

---

### Prompt 1.8: Switch API Service to New DB

```
Update the API service to use the new database:
1. Show me how to update PROJECTION_DB_URL secret (with search_path=projections)
2. Deploy the service
3. Show me how to verify it's reading from the new database
```

---

### Prompt 1.9: Switch Scheduler to New DB

```
Update the Scheduler service to use the new database with the same approach as the API service.
Then show me how to verify all services are using the new database.
```

---

### Prompt 1.10: Monitor and Decommission Old Databases

```
Show me commands to:
1. Check logs for all services to ensure no errors
2. Verify new database has recent data
3. Destroy the old database apps: status-app-eventstore and status-app-projections-db
```

---

## Phase 2: Service Consolidation

### Prompt 2.1: Create Backend Service Structure

```
Create cmd/backend/main.go that:
- Combines commands, api, and projections logic
- Runs projections in a goroutine
- Serves both /commands/* and /api/* endpoints
- Has proper shutdown handling

Show me the complete code.
```

---

### Prompt 2.2: Test Backend Locally

```
Help me test the new backend service locally:
1. Show me what environment variables to set
2. How to run it
3. How to test both commands and API endpoints work
```

---

### Prompt 2.3: Create Backend Fly Config

```
Create fly.backend.toml for the new backend service with:
- App name: status-app-backend
- Build args for SERVICE=backend
- Proper http_service configuration
- Health check on /health endpoint

Then show me how to deploy it to Fly.io with the right secrets.
```

---

### Prompt 2.4: Switch Slackbot to Backend

```
Update Slackbot to use the new backend service:
1. Show me how to update COMMANDS_URL to point to status-app-backend
2. How to verify it's working by testing a Slack message
3. How to roll back if there are issues
```

---

### Prompt 2.5: Switch Scheduler to Backend

```
Update Scheduler to use the new backend service the same way as Slackbot.
Show me how to verify both Slackbot and Scheduler are using the backend.
```

---

### Prompt 2.6: Monitor Backend

```
Show me commands to:
1. Check backend logs
2. Monitor error rates
3. Check database connection count
4. Verify all endpoints are responding

Tell me what to look for to confirm it's working properly.
```

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
