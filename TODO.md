# Status App TODO List

## 🔴 Critical - Core Functionality

### 1. Implement Real-Time Event Projections with PostgreSQL LISTEN/NOTIFY
**Priority**: High  
**Status**: Not Started  
**Location**: `internal/events/postgres_store.go` line 102

**Problem**: 
- Currently, the `Subscribe()` method returns a stub channel that never sends events
- The projections service only updates on restart, not in real-time
- PostgreSQL NOTIFY is sent but nobody is listening

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

### 2. Implement Slack Message Handling in Slackbot
**Priority**: High  
**Status**: Not Started  
**Location**: `cmd/slackbot/main.go` line 75

**Problem**:
- Slackbot receives messages but doesn't process them
- No integration with Commands service
- Can't actually collect status updates from Slack

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
**Status**: Not Started  
**Location**: `cmd/scheduler/main.go` lines 63-64

**Problem**:
- Scheduler checks teams but doesn't actually send reminders
- No logic to determine if team is due for a reminder
- No integration with Slack to send messages

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

### 4. Add Proper Error Handling and Validation
**Priority**: Medium  
**Status**: Partial

**Issues**:
- Some errors are silently ignored (e.g., `_, _ = s.db.ExecContext(...)`)
- No validation on event data before projection
- No retry logic for failed projections

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

### 5. Add Health Check Endpoints
**Priority**: Medium  
**Status**: Not Started

**Problem**:
- No way to check if services are healthy
- Fly.io health checks would fail
- Hard to debug deployment issues

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
**Status**: Minimal

**Current State**:
- Basic log.Printf statements
- No structured logging
- No metrics or tracing

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

### 10. Add Authentication and Authorization
**Priority**: Low  
**Status**: Not Started

**Current State**:
- No authentication on any endpoint
- Anyone can submit updates
- Anyone can query all data

**Implementation Plan**:

1. **Add API key authentication** (~1 hour)
   - Generate API keys for teams
   - Validate on each request
   - Store in database

2. **Add team-based authorization** (~1 hour)
   - Users can only update their team
   - Users can only view their team data
   - Admin role for cross-team access

**Estimated Time**: 2 hours

---

## 🔧 Technical Debt

### 11. Add Integration Tests
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

### 12. Improve Error Messages and Documentation
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

### 13. Database Migration Management
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

**Total Critical Path**: ~4 hours

### Should Do (Better UX):
4. Error handling (1 hour)
5. Health checks (45 min)
6. Logging (2.5 hours)

**Total Important**: ~4 hours

### Nice to Have (Features):
7-10. Various features (8.5 hours)

### Tech Debt:
11-13. Testing, docs, migrations (6.5 hours)

---

## Next Steps

**Recommended Priority Order**:

1. **Week 1**: Implement items #1-3 (critical functionality)
2. **Week 2**: Implement items #4-5 (stability)
3. **Week 3**: Implement item #6 (observability)
4. **Week 4+**: Pick features based on user feedback

---

**Last Updated**: 2025-12-07
