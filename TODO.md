# TODO

## ✅ Completed

- Docker E2E tests
- Auth middleware (API_SECRET required)
- Slack bot (processes messages and mentions)
- Event sourcing architecture
- All services deployed to Fly.io
- **Database consolidation** (2 databases → 1 database)
- **Service consolidation** (5 services → 3 services)

## 🚧 To Implement

### 1. URL Structure Refactoring (1h)

**Current**: URLs expose internal architecture
- `/commands/submit-update` - reveals CQRS command handling
- `/commands/register-team` - reveals implementation detail
- `/api/teams` - reveals separate API layer
- `/api/updates` - reveals query side

**Goal**: URLs should reflect external domain model, not internal architecture

**Proposed structure**:
```
POST   /teams              - Register a new team
PUT    /teams/{id}         - Update team
POST   /teams/{id}/updates - Submit status update
GET    /teams              - List all teams
GET    /teams/{id}         - Get team details
GET    /teams/{id}/updates - Get team's updates
GET    /updates            - Get recent updates
```

**Benefits**:
- RESTful resource-based URLs
- Hides internal CQRS/event sourcing implementation
- Better API discoverability
- Standard HTTP verbs for operations
- Can change internal architecture without breaking API

**Implementation**:
- Update `cmd/backend/main.go` endpoint routing
- Keep internal command/query handlers unchanged
- Update Slackbot to use new URLs
- Add URL versioning: `/v1/teams` for future compatibility

**Estimated time**: 1 hour (routing changes + tests)

---

### 2. Migration Service (30m)

**Current**: Migrations run manually via psql  
**Goal**: Automated migration runner as Fly.io service

**Implementation**:
- Create `cmd/migrate/main.go` that reads and executes SQL files from `/migrations`
- Add `fly.migrate.toml` config for one-off deployment
- Use golang-migrate library or custom SQL file executor
- Run with: `fly deploy -c fly.migrate.toml`

**Benefits**:
- Version-controlled migrations
- Automated on deploy
- Repeatable and safe

---

### 3. Real-Time Projections (1h)

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

### 4. Scheduler Reminders (1h)

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

### 5. Minor Fixes (30m)

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

**Total remaining work**: ~3 hours

Once complete, the app will be 100% feature-complete.
