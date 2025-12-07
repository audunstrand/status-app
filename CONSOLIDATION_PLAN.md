# Architecture Consolidation Plan

**Goal**: Consolidate 5 services + 2 databases → 3 services + 1 database

**Principle**: Each step leaves the system fully working and deployable.

---

## Phase 1: Database Consolidation (Keep All 5 Services)

### Step 1.1: Create Single Database with Schemas
**Goal**: New database with proper schema structure  
**Status**: System still runs on old databases

```bash
# Create new Fly.io Postgres
fly postgres create --name status-app-db --region arn

# Create schemas
fly postgres connect -a status-app-db
CREATE SCHEMA events;
CREATE SCHEMA projections;
```

**Verification**: Database exists, no code changes yet

---

### Step 1.2: Migrate Event Store Schema
**Goal**: Events table in new DB

**Migration file**: `migrations/003_create_events_schema.up.sql`
```sql
CREATE SCHEMA IF NOT EXISTS events;

CREATE TABLE events.events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_events_aggregate_id ON events.events(aggregate_id);
CREATE INDEX idx_events_type ON events.events(type);
CREATE INDEX idx_events_timestamp ON events.events(timestamp);
```

**Run migration**:
```bash
migrate -path migrations -database "postgres://...status-app-db" up
```

**Verification**: `SELECT * FROM events.events` works (empty table)

---

### Step 1.3: Migrate Projections Schema
**Goal**: Teams and status_updates in new DB

**Migration file**: `migrations/004_create_projections_schema.up.sql`
```sql
CREATE SCHEMA IF NOT EXISTS projections;

CREATE TABLE projections.teams (
    team_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slack_channel VARCHAR(255) NOT NULL,
    poll_schedule VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE projections.status_updates (
    update_id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL REFERENCES projections.teams(team_id),
    content TEXT NOT NULL,
    author VARCHAR(255) NOT NULL,
    slack_user VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_status_updates_team_id ON projections.status_updates(team_id);
CREATE INDEX idx_status_updates_created_at ON projections.status_updates(created_at DESC);
CREATE INDEX idx_status_updates_team_created ON projections.status_updates(team_id, created_at DESC);
```

**Run migration**: Same as above

**Verification**: `SELECT * FROM projections.teams` works

---

### Step 1.4: Copy Existing Data
**Goal**: Migrate data from old databases to new

```sql
-- Connect to new database
\c status-app-db

-- Copy events (if any exist in production)
-- Option 1: pg_dump/restore
-- Option 2: INSERT INTO SELECT from dblink
-- Option 3: Manual (if small dataset)

-- Copy teams and status_updates similarly
```

**Verification**: Row counts match between old and new databases

---

### Step 1.5: Switch Commands Service to New DB
**Goal**: Commands writes to new database

**Update Fly.io secrets**:
```bash
fly secrets set DATABASE_URL="postgres://...status-app-db?search_path=events" \
  -a status-app-commands
```

**Code change**: Update config to use `DATABASE_URL` instead of `EVENT_STORE_URL`

**Deploy**: `flyctl deploy --config fly.toml`

**Verification**: 
- Submit status update via Slack
- Check `events.events` table has new row
- Old event store should NOT have new row

**Rollback plan**: Revert secret to old `EVENT_STORE_URL`

---

### Step 1.6: Switch Projections Service to New DB
**Goal**: Projections reads from new event DB, writes to new projection DB

**Update secrets**:
```bash
fly secrets set \
  EVENT_STORE_URL="postgres://...status-app-db?search_path=events" \
  PROJECTION_DB_URL="postgres://...status-app-db?search_path=projections" \
  -a status-app-projections
```

**Deploy**: `flyctl deploy --config fly.projections.toml`

**Verification**:
- Submit status update
- Check projections update in `projections.teams` and `projections.status_updates`

---

### Step 1.7: Switch API Service to New DB
**Goal**: API reads from new projection DB

**Update secrets**:
```bash
fly secrets set \
  PROJECTION_DB_URL="postgres://...status-app-db?search_path=projections" \
  -a status-app-api
```

**Deploy**: `flyctl deploy --config fly.api.toml`

**Verification**: 
- Query API: `curl https://status-app-api.fly.dev/api/teams`
- Should return data from new database

---

### Step 1.8: Update Scheduler
**Goal**: Scheduler uses new projection DB

**Update secrets** + **Deploy** (same as API)

**Verification**: Check scheduler logs, no errors

---

### Step 1.9: Monitor and Decommission Old Databases
**Goal**: Confirm everything works, remove old DBs

**Monitor for 24 hours**:
- All services working
- No errors in logs
- New events being written
- Projections updating

**Then delete old databases**:
```bash
fly apps destroy status-app-eventstore
fly apps destroy status-app-projections-db
```

**Checkpoint**: ✅ All 5 services running on 1 database

---

## Phase 2: Service Consolidation (Backend Service)

### Step 2.1: Create Backend Service Structure
**Goal**: New backend service with all logic

**New file**: `cmd/backend/main.go`
```go
package main

import (
    "context"
    "log"
    "net/http"
    "sync"
    
    // Import handlers from commands and api
)

func main() {
    ctx := context.Background()
    var wg sync.WaitGroup
    
    // Start projection worker in goroutine
    wg.Add(1)
    go func() {
        defer wg.Done()
        runProjections(ctx)
    }()
    
    // Start HTTP server (commands + api endpoints)
    mux := http.NewServeMux()
    
    // Commands endpoints
    mux.HandleFunc("/commands/submit-update", handleSubmitUpdate)
    mux.HandleFunc("/commands/register-team", handleRegisterTeam)
    
    // API endpoints  
    mux.HandleFunc("/api/teams", handleGetTeams)
    mux.HandleFunc("/api/teams/{id}", handleGetTeam)
    mux.HandleFunc("/api/updates", handleGetRecentUpdates)
    
    // Health
    mux.HandleFunc("/health", handleHealth)
    
    log.Fatal(http.ListenAndServe(":8080", mux))
}
```

**Verification**: Code compiles, no runtime test yet

---

### Step 2.2: Test Backend Locally
**Goal**: Verify backend works standalone

```bash
# Set env vars
export DATABASE_URL="postgres://localhost:5432/statusapp?search_path=events"
export PROJECTION_DB_URL="postgres://localhost:5432/statusapp?search_path=projections"
export API_SECRET="test"

# Run backend
go run cmd/backend/main.go

# Test in another terminal
curl -H "Authorization: Bearer test" \
  -X POST http://localhost:8080/commands/submit-update \
  -d '{"team_id":"test","content":"test","author":"test"}'

curl -H "Authorization: Bearer test" \
  http://localhost:8080/api/teams
```

**Verification**: Both commands and queries work from single service

---

### Step 2.3: Deploy Backend Alongside Existing Services
**Goal**: Backend runs in production, old services still active

**Create**: `fly.backend.toml`
```toml
app = "status-app-backend"
primary_region = "arn"

[build]
  dockerfile = "Dockerfile"
  [build.args]
    SERVICE = "backend"

[env]
  PORT = "8080"

[[services]]
  internal_port = 8080
  protocol = "tcp"

  [[services.ports]]
    port = 80
    handlers = ["http"]
  
  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]
```

**Deploy**:
```bash
flyctl deploy --config fly.backend.toml
```

**Set secrets**:
```bash
fly secrets set \
  DATABASE_URL="postgres://...status-app-db?search_path=events" \
  PROJECTION_DB_URL="postgres://...status-app-db?search_path=projections" \
  API_SECRET="..." \
  -a status-app-backend
```

**Verification**: Backend is deployed but not receiving traffic yet

---

### Step 2.4: Switch Slackbot to Use Backend
**Goal**: Slackbot sends commands to backend instead of commands service

**Update slackbot config**:
```bash
fly secrets set \
  COMMANDS_URL="https://status-app-backend.fly.dev" \
  -a status-app-slackbot
```

**Verification**:
- Post in Slack
- Check backend logs: receives request
- Check database: event is written
- Check API: update appears

**Rollback**: Change `COMMANDS_URL` back to commands service

---

### Step 2.5: Switch Scheduler to Use Backend
**Goal**: Scheduler sends to backend

**Update secrets** (same as slackbot)

**Verification**: Scheduler logs show backend URL

---

### Step 2.6: Monitor Backend Under Load
**Goal**: Ensure backend handles production traffic

**Monitor for 24 hours**:
- Backend logs
- Error rates
- Response times
- Database connections

**If issues**: Roll back to old services

---

### Step 2.7: Decommission Old Services
**Goal**: Remove commands, api, projections services

**Delete Fly.io apps**:
```bash
fly apps destroy status-app-commands
fly apps destroy status-app-api
fly apps destroy status-app-projections
```

**Clean up code**:
```bash
git rm -r cmd/commands cmd/api cmd/projections
git rm fly.toml fly.api.toml fly.projections.toml
```

**Update docs**: README, DEPLOYMENT, ARCHITECTURE

**Checkpoint**: ✅ 3 services (backend, slackbot, scheduler) + 1 database

---

## Phase 3: Cleanup and Documentation

### Step 3.1: Update Documentation
- README.md: New architecture diagram
- docs/ARCHITECTURE.md: 3 services instead of 5
- docs/DEPLOYMENT.md: Updated service list
- TODO.md: Update with new structure

### Step 3.2: Update CI/CD
- `.github/workflows/ci.yml`: Deploy 3 services instead of 5

### Step 3.3: Verify Everything Works
- Run tests
- Deploy all services
- Test end-to-end in production

---

## Summary Timeline

**Phase 1 (Database)**: 2-3 hours
- Steps can be done in one session
- Low risk (services unchanged)

**Phase 2 (Services)**: 2-3 hours  
- Create backend: 1h
- Deploy and test: 1h
- Monitor and decommission: 1h

**Phase 3 (Cleanup)**: 30 min

**Total**: ~6 hours of work over 2-3 days (allowing for monitoring periods)

---

## Rollback Points

Every step has a rollback:
- **Phase 1**: Can revert secrets to old database URLs
- **Step 2.4-2.5**: Can revert COMMANDS_URL
- **Step 2.7**: Can redeploy old services if needed

---

## Final State

```
Services (3):
├── backend (commands + api + projections)
├── slackbot
└── scheduler

Database (1):
└── status-app-db
    ├── events schema (events table)
    └── projections schema (teams, status_updates)

Deployments:
├── status-app-backend.fly.dev
├── status-app-slackbot.fly.dev
└── status-app-scheduler.fly.dev
```

**Savings**:
- 2 fewer services to deploy/monitor
- 2 fewer Fly.io apps
- 1 fewer database
- Simpler architecture
- Faster deployments

Ready to start with Phase 1, Step 1.1?
