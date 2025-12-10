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
- **Real-time projections** (PostgreSQL LISTEN/NOTIFY) - Completed 2025-12-10
- **Code quality improvements** (validation, reduced duplication, better tests)

## 🚧 To Implement

### 1. ~~Real-Time Projections~~ ✅ COMPLETE (2025-12-10)

**Implementation completed**:
- PostgreSQL LISTEN/NOTIFY for instant projection updates
- Integration tests for subscription mechanism
- All tests passing
- Deployed to production and verified working

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

**~~Fix ignored NOTIFY error~~**: ✅ Fixed in code quality refactoring (2025-12-10)

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

**Total remaining work**: ~1.5 hours

- Scheduler reminders: ~1h
- Health checks: ~30m

Once complete, the app will be 100% feature-complete.
