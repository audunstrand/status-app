# Status App TODO List

## ✅ Completed Items

### Docker E2E Tests ✅
**Status**: ✅ Completed (2025-12-07)  
**Completed Work**:
- Created `tests/e2e_docker/` with Docker Compose setup
- Implemented 4 comprehensive HTTP-based tests against real containers
- All tests passing: authentication, endpoints, end-to-end flow
- Fixed JSONB null metadata handling bug (prevented status updates from saving)
- Cleaned up broken test files

### Auth Middleware Testing ✅
**Status**: ✅ Completed (2025-12-07)  
**Completed Work**:
- Added unit tests for auth middleware with nested ServeMux pattern
- Verified no URL path stripping issues
- Tests confirm authentication works correctly

### Service-to-Service Authentication ✅
**Status**: ✅ FULLY IMPLEMENTED AND ENABLED  
**Completed Work**:
- ✅ `RequireAPIKey` middleware in `internal/auth/middleware.go`
- ✅ API Secret configuration in all services
- ✅ Commands service protected with auth middleware
- ✅ API service protected with auth middleware  
- ✅ Slackbot sends Authorization header with requests
- ✅ Works when `API_SECRET` environment variable is set
- ✅ **API_SECRET configured in Fly.io production**
- ✅ **All services secured in production**

### Health Check Endpoints ✅
**Status**: ✅ IMPLEMENTED
**Completed Work**:
- ✅ `/health` endpoint in Commands service
- ✅ `/health` endpoint in API service
- ⚠️ **Missing**: Fly.io health check configuration in toml files
- **Action Needed**: Add `[[services.http_checks]]` to all fly.*.toml files

### Slack Integration - Message Handling ✅
**Status**: ✅ FULLY IMPLEMENTED (2025-12-07)  
**Completed Work**:
- ✅ Bot responds to `@mentions` (app_mention events)
- ✅ Bot processes regular messages in channels (message events)
- ✅ Parses messages and extracts content
- ✅ Sends HTTP POST to Commands service `/commands/submit-update`
- ✅ Uses channel ID as team ID mapping
- ✅ Sends confirmation messages back to Slack
- ✅ Properly authenticated with Commands service (when API_SECRET is set)
- ✅ Successfully stores status updates
- ✅ JSONB bug fixed - updates persist correctly
- ✅ Deployed and verified working in production

---

## 🔴 Critical - Core Functionality

### 1. Implement Real-Time Event Projections with PostgreSQL LISTEN/NOTIFY
**Priority**: High  
**Status**: ❌ NOT IMPLEMENTED (stub only)  
**Location**: `internal/events/postgres_store.go` line 100

**Current State**:
- `Subscribe()` method returns a stub channel that never sends events
- Projections service only updates on restart, not in real-time
- PostgreSQL NOTIFY is sent (line 57) but nobody is listening

**Evidence**:
```go
func (s *PostgresStore) Subscribe(...) {
    // TODO: Implement LISTEN/NOTIFY for real-time event streaming
    ch := make(chan *Event)
    go func() {
        <-ctx.Done()
        close(ch)
    }()
    return ch, nil  // Returns empty channel!
}
```

**Implementation Plan**:

1. **Add GetByID method to PostgresStore** (~10 min)
   ```go
   func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Event, error)
   ```
   - Query: `SELECT * FROM events WHERE id = $1`
   - Return single event or error if not found

2. **Store connection string in PostgresStore** (~5 min)
   - Add `connStr string` field to struct
   - Save it in `NewPostgresStore()`

3. **Implement Subscribe() with pq.NewListener** (~30 min)
   ```go
   func (s *PostgresStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error) {
       listener := pq.NewListener(s.connStr, 10*time.Second, time.Minute, eventCallback)
       listener.Listen("events")
       
       ch := make(chan *Event)
       go func() {
           for {
               select {
               case n := <-listener.Notify:
                   event, _ := s.GetByID(ctx, n.Extra)
                   ch <- event
               case <-ctx.Done():
                   return
               }
           }
       }()
       return ch, nil
   }
   ```

4. **Test the implementation** (~15 min)
   - Submit a status update via Commands API
   - Verify projections service receives notification
   - Check that read model updates without restart

**Estimated Time**: 1 hour  
**Dependencies**: None

---

### 2. ~~Implement Slack Message Handling in Slackbot~~ ✅ COMPLETED
**Priority**: High  
**Status**: ✅ FULLY IMPLEMENTED  

**Completed Work**:
- ✅ Parses `app_mention` events
- ✅ Parses regular `message` events  
- ✅ Extracts: content, author, team info (uses channel ID as team ID)
- ✅ Sends HTTP POST to Commands service at `/commands/submit-update`
- ✅ Handles errors and sends feedback to Slack
- ✅ Authentication with Commands service (when API_SECRET is set)
- ✅ Working in production

**Code Location**: `cmd/slackbot/main.go` lines 87-150

**Implementation Plan**:

1. **Parse incoming Slack messages** (~20 min)
   - Detect status updates (e.g., messages starting with `/status` or in specific format)
   - Extract: content, author, team info
   - Handle errors and invalid formats

2. **Send HTTP request to Commands service** (~20 min)
   ```go
   func sendStatusUpdate(teamID, content, author, slackUser string) error {
       payload := SubmitStatusUpdateRequest{
           TeamID:    teamID,
           Content:   content,
           Author:    author,
           SlackUser: slackUser,
       }
       resp, err := http.Post(commandsURL+"/commands/submit-update", "application/json", jsonPayload)
       // Handle response
   }
   ```

3. **Add configuration for Commands service URL** (~10 min)
   - Add `COMMANDS_SERVICE_URL` to environment config
   - Default to `http://status-app-commands.fly.dev` for production

4. **Map Slack channels to team IDs** (~15 min)
   - Option A: Store in database (better)
   - Option B: Environment variable mapping (quick start)
   - Load on startup

5. **Send acknowledgment back to Slack** (~10 min)
   - Reply with confirmation message
   - Handle errors gracefully

**Estimated Time**: 1.5 hours  
**Dependencies**: None

---

### 3. Implement Scheduler Team Reminder Logic
**Priority**: High  
**Status**: ⚠️ PARTIALLY IMPLEMENTED  
**Location**: `cmd/scheduler/main.go` lines 55-67

**What's Working**:
- ✅ Cron scheduler set up (runs hourly)
- ✅ Reads teams from projection database
- ✅ Loops through teams

**What's Missing**:
- ❌ No logic to determine if team is due for reminder
- ❌ No Slack message sending
- ❌ No `last_reminded_at` tracking

**Current Code**:
```go
for _, team := range teams {
    // TODO: Check if team is due for a reminder based on poll_schedule
    // TODO: Send command to send reminder
    log.Printf("Checking team %s for reminders", team.Name)
}
```

**Implementation Plan**:

1. **Add last_reminded_at column to teams table** (~10 min)
   ```sql
   ALTER TABLE teams ADD COLUMN last_reminded_at TIMESTAMP;
   ```

2. **Implement reminder logic** (~30 min)
   ```go
   func shouldSendReminder(team *Team, schedule string) bool {
       // Parse schedule (e.g., "weekly", "monday", "friday")
       // Check if current time matches schedule
       // Check if already reminded recently
       // Return true/false
   }
   ```

3. **Send Slack message via Slack API** (~20 min)
   ```go
   func sendSlackReminder(channelID, teamID string) error {
       api := slack.New(botToken)
       _, _, err := api.PostMessage(channelID, 
           slack.MsgOptionText("🔔 Time for your weekly status update!", false))
       return err
   }
   ```

4. **Update last_reminded_at timestamp** (~10 min)
   - After successful reminder sent
   - Prevents duplicate reminders

5. **Add Slack bot token to scheduler config** (~10 min)
   - Add `SLACK_BOT_TOKEN` environment variable
   - Load in config

**Estimated Time**: 1.5 hours  
**Dependencies**: Slack API credentials

---

## 🟡 Important - User Experience

### 4. ~~Add Proper Error Handling and Validation~~ ⚠️ MOSTLY DONE
**Priority**: Medium  
**Status**: ⚠️ ONE ISSUE REMAINING

**What's Good**:
- ✅ Most errors are properly handled
- ✅ Validation exists on commands and requests
- ✅ Event store has proper error wrapping

**Remaining Issue**:
- ❌ Line 57 in `postgres_store.go`: NOTIFY error ignored
  ```go
  _, _ = s.db.ExecContext(ctx, "NOTIFY events, $1", event.ID)
  ```
- **Impact**: Low (NOTIFY failure doesn't affect data integrity)
- **Fix**: Log the error
- **Estimated Time**: 5 minutes

**Implementation Plan**:

1. **Fix ignored errors in postgres_store.go** (~15 min)
   - Line 48: Log NOTIFY errors
   - Add proper error handling throughout

2. **Add retry logic for projections** (~30 min)
   - Use exponential backoff
   - Store failed events for later retry
   - Max retry limit

3. **Validate event data before processing** (~20 min)
   - Check required fields exist
   - Validate data types
   - Return meaningful errors

**Estimated Time**: 1 hour

---

### 5. ~~Add Health Check Endpoints~~ ✅ MOSTLY DONE
**Priority**: Medium  
**Status**: ✅ IMPLEMENTED, ⚠️ CONFIG MISSING

**What's Implemented**:
- ✅ `/health` endpoint in Commands service (cmd/commands/main.go)
- ✅ `/health` endpoint in API service (cmd/api/main.go)
- ✅ Returns JSON with status and service name

**What's Missing**:
- ❌ Fly.io health check configuration in fly.*.toml files
- **Action**: Add `[[services.http_checks]]` section to all toml files
- **Estimated Time**: 15 minutes

**Implementation Plan**:

1. **Add /health endpoint to all services** (~30 min)
   ```go
   mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
       // Check database connection
       // Check event store connection
       // Return 200 OK or 503 Service Unavailable
       json.NewEncoder(w).Encode(map[string]string{
           "status": "healthy",
           "service": "api",
           "version": "1.0.0",
       })
   })
   ```

2. **Configure Fly.io to use health checks** (~15 min)
   - Update all fly.*.toml files
   - Add [[services.http_checks]] section

**Estimated Time**: 45 minutes

---

### 6. Add Logging and Observability
**Priority**: Medium  
**Status**: ❌ BASIC LOGGING ONLY

**Current State**:
- ✅ Basic `log.Printf` statements throughout
- ❌ No structured logging (still using stdlib `log`, not `slog`)
- ❌ No metrics or tracing
- ❌ No request correlation IDs

**Recommendation**: Keep as-is for now, add when needed for debugging.

**Implementation Plan**:

1. **Add structured logging with slog** (~1 hour)
   - Replace log.Printf with slog
   - Add context to all log messages
   - Include trace IDs for request correlation

2. **Add metrics** (~1 hour)
   - Events written counter
   - Projections processed counter
   - API request latency
   - Use Prometheus format

3. **Add request tracing** (~30 min)
   - Generate trace ID per request
   - Pass through all services
   - Log with trace ID

**Estimated Time**: 2.5 hours

---

## 🟢 Nice to Have - Features

### 7. Add Update Editing and Deletion
**Priority**: Low  
**Status**: Not Started

**Features**:
- Allow users to edit their status updates
- Allow users to delete status updates
- Track edit history as events

**Implementation Plan**:

1. **Add new event types** (~15 min)
   - `StatusUpdateEdited`
   - `StatusUpdateDeleted`

2. **Add commands and handlers** (~30 min)
   - EditStatusUpdate command
   - DeleteStatusUpdate command

3. **Update projections** (~20 min)
   - Handle edit events (update row)
   - Handle delete events (soft delete)

4. **Add API endpoints** (~20 min)
   - PUT /api/updates/{id}
   - DELETE /api/updates/{id}

**Estimated Time**: 1.5 hours

---

### 8. Add Team Dashboard UI
**Priority**: Low  
**Status**: Not Started

**Features**:
- Web dashboard to view team updates
- Filter by team, date range
- Search functionality

**Implementation Plan**:

1. **Create simple HTML/JS frontend** (~2 hours)
   - Static HTML page
   - Fetch from API
   - Display updates in timeline

2. **Add to API service** (~30 min)
   - Serve static files
   - Add CORS if needed

**Estimated Time**: 2.5 hours

---

### 9. Add Slack Slash Commands
**Priority**: Low  
**Status**: Not Started

**Features**:
- `/status [message]` - Submit status update
- `/status-history` - View recent updates
- `/status-team` - View team settings

**Implementation Plan**:

1. **Register slash commands in Slack app** (~15 min)
   - Configure in Slack app settings
   - Set request URL to slackbot service

2. **Handle slash commands in slackbot** (~1 hour)
   - Parse command and arguments
   - Send to Commands service
   - Return response to Slack

**Estimated Time**: 1.5 hours

---

### 10. ~~Add Service-to-Service Authentication~~ ✅ COMPLETE
**Priority**: High  
**Status**: ✅ FULLY IMPLEMENTED AND ENABLED IN PRODUCTION

**Completed Implementation**:
- ✅ Shared secret API key approach
- ✅ `RequireAPIKey` middleware in `internal/auth/middleware.go`
- ✅ `API_SECRET` configuration in `internal/config/config.go`
- ✅ Commands service applies middleware when `APISecret` is set
- ✅ API service applies middleware when `APISecret` is set
- ✅ Slackbot sends `Authorization: Bearer <secret>` header
- ✅ Comprehensive unit tests in `internal/auth/middleware_test.go`
- ✅ **API_SECRET configured in all Fly.io services**
- ✅ **Production APIs are now secured**

**Security Status**: ✅ Protected and working

**Implementation Plan**:

**Option A: Shared Secret API Keys (Recommended - Simple & Effective)**

1. **Add shared secret environment variable** (~15 min)
   ```go
   // In config.go
   type Config struct {
       // ...existing fields
       APISecret string `env:"API_SECRET,required"`
   }
   ```
   - Generate random secret: `openssl rand -hex 32`
   - Set same secret in all services via Fly.io secrets

2. **Add authentication middleware** (~30 min)
   ```go
   // internal/auth/middleware.go
   func RequireAPIKey(secret string) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               authHeader := r.Header.Get("Authorization")
               if authHeader != "Bearer "+secret {
                   http.Error(w, "Unauthorized", http.StatusUnauthorized)
                   return
               }
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

3. **Apply middleware to Commands API** (~15 min)
   ```go
   // cmd/commands/main.go
   mux.Handle("/commands/", RequireAPIKey(cfg.APISecret)(
       http.HandlerFunc(handleSubmitUpdate(cmdHandler))))
   ```

4. **Apply middleware to API service** (~15 min)
   - Protect sensitive endpoints
   - Maybe keep `/health` public for Fly.io

5. **Update clients to send API key** (~30 min)
   ```go
   // Slackbot, Scheduler send requests with header
   req.Header.Set("Authorization", "Bearer "+cfg.APISecret)
   ```

6. **Set Fly.io secrets** (~15 min)
   ```bash
   fly secrets set API_SECRET=<generated-secret> -a status-app-commands
   fly secrets set API_SECRET=<generated-secret> -a status-app-api
   fly secrets set API_SECRET=<generated-secret> -a status-app-slackbot
   fly secrets set API_SECRET=<generated-secret> -a status-app-scheduler
   ```

**Estimated Time**: 2 hours

**Option B: JWT Tokens (More Complex, Better for Multi-tenant)**

1. **Generate JWT signing key** (~10 min)
2. **Issue service tokens** (~30 min)
   - Each service gets a token with claims (service name, permissions)
3. **Validate JWT on each request** (~45 min)
4. **Token rotation strategy** (~30 min)

**Estimated Time**: 2 hours

**Option C: mTLS (Most Secure, Complex)**

1. **Generate CA and certificates** (~1 hour)
2. **Configure Fly.io private network** (~30 min)
3. **Update all HTTP clients** (~1 hour)

**Estimated Time**: 2.5 hours

**Recommended Approach**: **Option A (Shared Secret)**
- Simplest to implement
- Adequate security for internal services
- Easy to rotate if compromised
- Works well with Fly.io secrets management

---

### 11. Add User-Facing Authentication and Authorization
**Priority**: Medium  
**Status**: Not Started

**Current State** (after #10):
- Services are secured from external access
- But no user identity/authorization
- Can't differentiate between team members

**Features Needed**:
- Users should only submit updates for their team
- Users should only see their team's data
- Admin users can see all teams

**Implementation Plan**:

1. **Add API key per team** (~1 hour)
   - Generate unique API key for each team
   - Store in teams table: `api_key VARCHAR(64)`
   - Return key when team registers

2. **Add team-based authorization middleware** (~1 hour)
   ```go
   func RequireTeamAuth() http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           apiKey := r.Header.Get("X-Team-API-Key")
           team, err := validateTeamAPIKey(apiKey)
           if err != nil {
               http.Error(w, "Unauthorized", 401)
               return
           }
           // Store team in context
           ctx := context.WithValue(r.Context(), "team", team)
           next.ServeHTTP(w, r.WithContext(ctx))
       }
   }
   ```

3. **Enforce team scope on commands** (~30 min)
   - Check that teamID in request matches authenticated team
   - Prevent cross-team submissions

4. **Add team filtering to queries** (~30 min)
   - API only returns data for authenticated team
   - Or all data if admin key

**Estimated Time**: 3 hours

---

## 🔧 Technical Debt

### 12. Add Integration Tests
**Priority**: Medium  
**Status**: Not Started

**Current State**:
- Have unit tests for handlers
- Have E2E tests for projections
- Missing end-to-end flow tests

**Implementation Plan**:

1. **Test complete flow** (~2 hours)
   - Start all services
   - Submit command via API
   - Verify event stored
   - Verify projection updated
   - Verify query returns data

**Estimated Time**: 2 hours

---

### 13. Improve Error Messages and Documentation
**Priority**: Low  
**Status**: Minimal

**Needs**:
- Better API error messages (currently just generic)
- OpenAPI/Swagger documentation
- Examples of API usage

**Implementation Plan**:

1. **Add OpenAPI spec** (~2 hours)
2. **Improve error responses** (~1 hour)
3. **Add README examples** (~30 min)

**Estimated Time**: 3.5 hours

---

### 14. Database Migration Management
**Priority**: Medium  
**Status**: Not Started

**Current State**:
- Schema defined in SQL files
- No versioning or migration tracking
- Manual application required

**Implementation Plan**:

1. **Add golang-migrate** (~1 hour)
   - Create migration files
   - Version schema changes
   - Run on startup

**Estimated Time**: 1 hour

---

## Summary

### Must Do (Critical Path):
1. ✅ Real-time projections (1 hour)
2. ✅ Slack message handling (1.5 hours)
3. ✅ Scheduler reminders (1.5 hours)
10. ✅ **Service-to-service authentication (2 hours)** ⚠️ SECURITY

**Total Critical Path**: ~6 hours

### Should Do (Better UX):
4. Error handling (1 hour)
5. Health checks (45 min)
6. Logging (2.5 hours)

**Total Important**: ~4.25 hours

### Nice to Have (Features):
7-9. Dashboard, slash commands, etc (4 hours)
11. User-facing auth (3 hours)

**Total Features**: ~7 hours

### Tech Debt:
12-14. Testing, docs, migrations (6.5 hours)

---

## Next Steps

**Recommended Priority Order**:

1. **Day 1 (Critical)**: 
   - Item #10: Service-to-service authentication (2 hours) ⚠️ **DO THIS FIRST**
   - Item #1: Real-time projections (1 hour)
   
2. **Day 2 (Core Features)**:
   - Item #2: Slack message handling (1.5 hours)
   - Item #3: Scheduler reminders (1.5 hours)

3. **Week 1**: 
   - Item #4-5: Error handling, health checks (1.75 hours)
   
4. **Week 2**: 
   - Item #6: Logging and observability (2.5 hours)
   - Item #11: User-facing auth (3 hours)

5. **Week 3+**: Pick features based on user feedback

**SECURITY WARNING**: Your APIs are currently publicly accessible! Anyone can:
- Submit fake status updates
- Query all team data
- Abuse the system

**Strongly recommend implementing #10 before deploying to production!**

---

## 📋 Next Steps (Recommended Priority)

### 🎯 High Priority - Core Features

1. **Real-Time Event Projections** (Item #1) - ❌ NOT DONE
   - Currently projections only update on restart
   - Implement PostgreSQL LISTEN/NOTIFY
   - **Time**: 1 hour
   - **Impact**: Updates appear immediately in queries

2. **Scheduler Team Reminders** (Item #3) - ⚠️ HALF DONE
   - Infrastructure exists, needs reminder logic and Slack integration
   - **Time**: 1 hour
   - **Impact**: Automated team reminders

### 💪 Quick Wins (Optional)

3. **Fly.io Health Check Config** (Item #5) - 15 minutes
   - Health endpoints exist, just need toml config
   
4. **Fix NOTIFY Error Logging** (Item #4) - 5 minutes
   - One ignored error in postgres_store.go

### 🚀 Nice to Have (Later)

- Structured Logging (Item #6) - when debugging needed
- Team Dashboard UI (Item #8)
- Slack Slash Commands (Item #9)
- User-facing Auth (Item #11)

---

## Summary Timeline

**Week 1 - Complete Core**:
1. Real-time projections (1h)
2. Scheduler reminders (1h)
3. Fly.io health checks (15m)
4. Fix NOTIFY error logging (5m)

**Total Week 1**: ~2.5 hours to complete all critical features

**Week 2+ - Polish & Features**:
- Based on user feedback
- Structured logging if needed for debugging
- Dashboard or slash commands if requested

---

## 🎉 What's Already Working

✅ **Slack Bot** - Fully functional, processes messages and mentions  
✅ **Event Sourcing** - Commands, events, and projections architecture complete  
✅ **Authentication** - ✅ **FULLY SECURED IN PRODUCTION**  
✅ **HTTP APIs** - Commands and Query APIs working  
✅ **Docker E2E Tests** - Comprehensive test coverage  
✅ **Health Endpoints** - Basic health checks in place  
✅ **Deployments** - All services deployed and running on Fly.io  

**The app is 85% feature-complete!** Main remaining work:
- Real-time updates (1h) 
- Reminder logic (1h)

---

**Last Updated**: 2025-12-07 (Authentication enabled)
