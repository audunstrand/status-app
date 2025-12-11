# Status App

Event-sourced team status updates via Slack.

## Features

- **Event Sourcing**: All state changes stored as immutable events
- **Real-time Updates**: PostgreSQL LISTEN/NOTIFY for instant projection updates
- **Slack Integration**: Post updates and manage teams via Slack
- **Automated Reminders**: Weekly reminders to submit status updates
- **RESTful API**: Query teams and updates with authentication
- **Auto Migrations**: Database migrations run automatically on deployment

## Architecture

**3 services + 1 database**

- **Backend**: Commands + API + Projections (port 8080)
- **Slackbot**: Slack integration (Socket Mode)
- **Scheduler**: Weekly reminder scheduling (Monday 9 AM)
- **Database**: PostgreSQL with `events` and `projections` schemas

## Slack Commands

- **Message mentions**: Send status update by mentioning the bot
- `/set-team-name`: Set a custom name for your team
- `/updates`: View recent updates from your team

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

## API Endpoints

**Teams**
- `POST /teams` - Register a new team
- `GET /teams` - List all teams
- `GET /teams/{id}` - Get team details
- `GET /teams/{id}/updates` - Get team updates
- `PUT /teams/{id}/name` - Update team name

**Updates**
- `POST /teams/{id}/updates` - Submit status update
- `GET /updates` - Get recent updates across all teams

All endpoints require `X-API-Secret` header for authentication.

## Deployment

Deployed to Fly.io via GitHub Actions on push to `master`.

Services:
- Backend: `status-app-backend.fly.dev`
- Slackbot: `status-app-slackbot.fly.dev`
- Scheduler: `status-app-scheduler.fly.dev`

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md) for details.
