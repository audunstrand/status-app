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
- **Real-time projections** (PostgreSQL LISTEN/NOTIFY)
- **Code quality improvements** (validation, reduced duplication, better tests)
- **Scheduler reminders** (sends reminders every Monday at 9 AM)
- **Slack commands** (`/set-team-name`, `/updates`)

## 🚧 Optional Improvements

### Add Fly.io Health Checks

Add health check endpoints to all services for better monitoring:

```toml
# All fly.*.toml files
[[services.http_checks]]
  interval = "10s"
  timeout = "2s"
  grace_period = "5s"
  method = "GET"
  path = "/health"
```

### Future Enhancements

- Add `/help` Slack command listing all available commands
- Add metrics/observability (Prometheus, Grafana)
- Add configurable reminder schedules per team
- Add reminder preferences (time, frequency)
