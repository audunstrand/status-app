# Status App

Event-sourced team status updates via Slack.

## Architecture

**3 services + 1 database** (consolidated from 5 services + 2 databases)

- **Backend**: Commands + API + Projections (combined)
- **Slackbot**: Slack integration
- **Scheduler**: Reminder scheduling
- **Database**: PostgreSQL with `events` and `projections` schemas

## Quick Start

```bash
# Set environment
export API_SECRET=$(openssl rand -hex 32)
export SLACK_BOT_TOKEN="xoxb-..."
export SLACK_SIGNING_KEY="xapp-..."
export EVENT_STORE_URL="postgres://localhost:5432/statusapp?sslmode=disable&search_path=events"
export PROJECTION_DB_URL="postgres://localhost:5432/statusapp?sslmode=disable&search_path=projections"
export COMMANDS_URL="http://localhost:8080"  # Backend service

# Run migrations
migrate -path migrations -database "postgres://localhost:5432/statusapp" up

# Start services
go run cmd/backend/main.go    # Port 8080 (Commands + API + Projections)
go run cmd/slackbot/main.go
go run cmd/scheduler/main.go
```

## Testing

```bash
make test           # All tests
make test-unit      # Unit tests only
make test-e2e       # E2E tests only
```

## Build

```bash
make build          # Builds all services to bin/
```

## Deployment

Deployed to Fly.io. See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

GitHub Actions automatically deploys on push to master.

## Status

✅ **Phase 1 Complete**: Database consolidation (2 → 1 database)
🚧 **Phase 2 In Progress**: Service consolidation (5 → 3 services)

See [CONSOLIDATION_PROMPTS.md](CONSOLIDATION_PROMPTS.md) for progress.
