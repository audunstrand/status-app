# TODO

## ✅ Completed

- Docker E2E tests
- Auth middleware (API_SECRET required)
- Slack bot (processes messages and mentions)
- Event sourcing architecture
- All services deployed to Fly.io
- **Database consolidation** (2 databases → 1 database)
- **Service consolidation** (5 services → 3 services)
- **Test code refactoring** (reduced by 40%, improved quality)
- **URL structure refactoring** (RESTful endpoints)
- **Automated database migrations** (golang-migrate + Fly.io release commands)

## 🚧 To Implement

### 1. Restore Test Coverage

**Status**: ✅ COMPLETE - Test coverage goals achieved

**Final coverage** (2025-12-08):
- ✅ internal/auth: 100%
- ✅ internal/commands: 76.6%
- ✅ internal/events: 74.5% ✅ (was 0%)
- ✅ internal/projections: 71.4% ✅ (was 0%)
- ✅ internal/config: 100% ✅ (was 0%)
- ✅ E2E tests: 77.6%
- ✅ cmd/backend: 10.1% (validation tests + covered by E2E)

**What was completed**:
1. ✅ Created comprehensive test suite for internal/events (74.5% coverage)
2. ✅ Created comprehensive test suite for internal/projections (71.4% coverage)
3. ✅ Created config tests for internal/config (100% coverage)
4. ✅ Created validation tests for cmd/backend (10.1% coverage)
5. ✅ Refactored all test code for quality:
   - Added test helper utilities (AssertNoError, AssertEqual, MustMarshalJSON)
   - Reduced test code by 35-55% through better structure
   - Fixed all ignored JSON marshal errors
   - Consistent error handling patterns
   - Better test organization with setup/teardown helpers

**Note on cmd/backend coverage**:
- Validation tests cover 10.1% of statements (request validation logic)
- HTTP handlers are thoroughly covered by E2E tests (77.6% coverage)
- E2E tests exercise the full request→validation→command→event→projection flow
- Combined coverage is sufficient for production use

**Remaining work**:
None - test coverage goals achieved!

---

### 1. Real-Time Projections (1h)

**Current**: Projections poll every few seconds  
**Goal**: Update immediately when events are written

**Implementation**:
```go
// internal/events/postgres_store.go
func (s *PostgresStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error) {
    listener := pq.NewListener(s.connStr, 10*time.Second, time.Minute, nil)
    listener.Listen("events")
    
    ch := make(chan *Event)
    go func() {
        for {
            select {
            case n := <-listener.Notify:
                event, _ := s.GetByID(ctx, n.Extra)
                if event != nil {
                    ch <- event
                }
            case <-ctx.Done():
                listener.Close()
                close(ch)
                return
            }
        }
    }()
    return ch, nil
}
```

**Steps**:
1. Add `GetByID(id string)` to postgres_store.go
2. Store connection string in PostgresStore
3. Replace stub Subscribe() with pq.NewListener
4. Test: submit update, verify projection updates without delay

---

### 2. Scheduler Reminders (1h)

**Current**: Scheduler runs but doesn't send reminders  
**Goal**: Send Slack messages on schedule

**Implementation**:
```go
// cmd/scheduler/main.go
func checkAndSendReminders(ctx context.Context, repo *projections.Repository) {
    teams, _ := repo.GetAllTeams(ctx)
    
    for _, team := range teams {
        if shouldRemind(team) {
            sendSlackReminder(team.SlackChannelID, team.Name)
            updateLastReminded(team.ID)
        }
    }
}

func shouldRemind(team *Team) bool {
    // Parse team.PollSchedule (e.g., "monday", "friday", "weekly")
    // Check if today matches schedule
    // Check if already reminded today
    return true // TODO
}

func sendSlackReminder(channelID, teamName string) error {
    api := slack.New(os.Getenv("SLACK_BOT_TOKEN"))
    _, _, err := api.PostMessage(channelID, 
        slack.MsgOptionText("🔔 Time for your status update!", false))
    return err
}
```

**Steps**:
1. Add `last_reminded_at TIMESTAMP` to teams table
2. Implement `shouldRemind()` schedule logic
3. Add Slack API call to send message
4. Update last_reminded_at after sending

---

### 3. Minor Fixes (30m)

**Fix ignored NOTIFY error**:
```go
// internal/events/postgres_store.go:57
if _, err := s.db.ExecContext(ctx, "NOTIFY events, $1", event.ID); err != nil {
    log.Printf("NOTIFY failed: %v", err)
}
```

**Add Fly.io health checks**:
```toml
# All fly.*.toml files
[[services.http_checks]]
  interval = "10s"
  timeout = "2s"
  grace_period = "5s"
  method = "GET"
  path = "/health"
```

---

## Summary

**Total remaining work**: ~2.5 hours

Once complete, the app will be 100% feature-complete.
