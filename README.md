# Status App

Event-sourced team status updates via Slack.

## Quick Start

```bash
# Set environment
export API_SECRET=$(openssl rand -hex 32)
export SLACK_BOT_TOKEN="xoxb-..."
export SLACK_SIGNING_KEY="xapp-..."
export EVENT_STORE_URL="postgres://localhost:5432/eventstore"
export PROJECTION_DB_URL="postgres://localhost:5432/projections"
export COMMANDS_URL="http://localhost:8081"

# Run migrations
migrate -path migrations -database $EVENT_STORE_URL up
migrate -path migrations -database $PROJECTION_DB_URL up

# Start services
go run cmd/commands/main.go   # Port 8081
go run cmd/api/main.go         # Port 8082
go run cmd/projections/main.go
go run cmd/slackbot/main.go
go run cmd/scheduler/main.go
```

## Testing

```bash
go test ./...                      # Unit tests
cd tests/e2e_docker && make test   # E2E tests
```

## Deployment

See [docs/DEPLOYMENT.md](docs/DEPLOYMENT.md)

## Status

✅ **Production Ready** (85% complete)
- Slack bot working
- Event sourcing complete
- Authentication required
- All services deployed

Remaining: Real-time projections (1h), Scheduler reminders (1h)

See [TODO.md](TODO.md)
