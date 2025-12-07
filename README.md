# Status App - Event-Sourced Team Status Updates

An event-sourced application for collecting and managing one-line status updates from teams via Slack.

## ✅ Current Status

**Production Ready**: 85% feature-complete and deployed!

- ✅ **Slack Bot** - Fully functional, processes messages and mentions
- ✅ **Event Sourcing** - Commands, events, and projections architecture
- ✅ **Authentication** - Required for all services (API_SECRET)
- ✅ **HTTP APIs** - Commands and Query APIs working
- ✅ **Deployments** - All services running on Fly.io
- ✅ **Tests** - Unit tests + Docker E2E tests

**Remaining Work**: ~2.5 hours (see [TODO.md](TODO.md))
- Real-time projections (1h)
- Scheduler reminders (1h)
- Minor fixes (30m)

## Architecture

This application follows **Event Sourcing** and **CQRS** patterns:

- **Event Store**: Append-only log of all events (PostgreSQL)
- **Commands**: Actions that generate events
- **Projections**: Read models built from events
- **Loose Coupling**: Components communicate via events

### Components

```
┌─────────────┐      ┌──────────────┐      ┌─────────────┐
│ Slack Bot   │─────▶│   Commands   │─────▶│ Event Store │
└─────────────┘      └──────────────┘      └──────┬──────┘
                                                   │
                     ┌──────────────┐             │
                     │  Scheduler   │             │
                     └──────┬───────┘             │
                            │                     │
                            ▼                     ▼
                     ┌──────────────┐      ┌─────────────┐
                     │   Commands   │      │ Projections │
                     └──────────────┘      └──────┬──────┘
                                                   │
                                            ┌──────▼──────┐
                                            │  Query API  │
                                            └─────────────┘
```

### Services

1. **Commands** (`cmd/commands`) - Processes commands, emits events
2. **Projections** (`cmd/projections`) - Builds read models from events
3. **Slack Bot** (`cmd/slackbot`) - Receives Slack messages, sends commands
4. **Scheduler** (`cmd/scheduler`) - Sends periodic team reminders
5. **API** (`cmd/api`) - Query endpoint for projections

## Getting Started

### Prerequisites

- Go 1.23+
- PostgreSQL 14+
- Slack App with Bot Token
- Docker (for E2E tests)

### Local Development

1. **Install dependencies**:
```bash
go mod download
```

2. **Set environment variables**:
```bash
export API_SECRET=$(openssl rand -hex 32)
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_SIGNING_KEY="xapp-your-key"
export EVENT_STORE_URL="postgres://localhost:5432/eventstore"
export PROJECTION_DB_URL="postgres://localhost:5432/projections"
export COMMANDS_URL="http://localhost:8081"
```

3. **Run migrations**:
```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations
migrate -path migrations -database $EVENT_STORE_URL up
migrate -path migrations -database $PROJECTION_DB_URL up
```

4. **Start services**:
```bash
# Terminal 1: Command service
go run cmd/commands/main.go

# Terminal 2: API service  
go run cmd/api/main.go

# Terminal 3: Projections service
go run cmd/projections/main.go

# Terminal 4: Slackbot service
go run cmd/slackbot/main.go

# Terminal 5: Scheduler service
go run cmd/scheduler/main.go
```

## Testing

### Unit Tests
```bash
go test ./...
```

### Docker E2E Tests
```bash
cd tests/e2e_docker
make test
```

See [docs/TESTING.md](docs/TESTING.md) for more details.

## Deployment

See [docs/FLY_DEPLOYMENT.md](docs/FLY_DEPLOYMENT.md) for Fly.io deployment instructions.

## Documentation

- [Architecture](docs/ARCHITECTURE.md) - Detailed architecture overview
- [Security](docs/SECURITY.md) - Authentication and security setup
- [Testing](docs/TESTING.md) - Testing strategy and guidelines
- [GitHub Actions](docs/GITHUB_ACTIONS.md) - CI/CD pipeline
- [TODO](TODO.md) - Remaining work and priorities

## Project Structure

```
├── cmd/              # Service entry points
│   ├── api/         # Query API service
│   ├── commands/    # Command handler service
│   ├── projections/ # Projection builder
│   ├── scheduler/   # Reminder scheduler
│   └── slackbot/    # Slack integration
├── internal/        # Internal packages
│   ├── auth/       # Authentication middleware
│   ├── commands/   # Command handlers
│   ├── config/     # Configuration
│   ├── events/     # Event store
│   └── projections/# Projection logic
├── migrations/     # Database migrations
├── tests/          # E2E tests
│   ├── e2e/       # Integration tests
│   └── e2e_docker/# Docker E2E tests
├── docs/          # Documentation
└── Dockerfile.*   # Docker images for services
```

## License

MIT

# Terminal 2: Projection service
go run cmd/projections/main.go

# Terminal 3: API service
go run cmd/api/main.go

# Terminal 4: Slack bot
go run cmd/slackbot/main.go

# Terminal 5: Scheduler
go run cmd/scheduler/main.go
```

## Event Types

- `status_update.submitted` - A team member submitted a status update
- `team.registered` - A new team was registered
- `team.updated` - Team information was updated
- `poll.scheduled` - A poll was scheduled for a team
- `reminder.sent` - A reminder was sent to a team

## Commands

- `SubmitStatusUpdate` - Submit a status update
- `RegisterTeam` - Register a new team
- `UpdateTeam` - Update team information
- `SchedulePoll` - Schedule a poll
- `SendReminder` - Send a reminder to a team

## Projections

- `teams` - Current state of all teams
- `status_updates` - All status updates with metadata
- `team_summary` - Aggregate statistics per team

## Development

### Adding New Event Types

1. Define event type constant in `internal/events/events.go`
2. Add event data struct
3. Update command handler to emit new event
4. Add projection handler to build read model

### Adding New Commands

1. Define command in `internal/commands/commands.go`
2. Implement handler in `internal/commands/handler.go`
3. Add HTTP endpoint in `cmd/commands/main.go`

## Future Enhancements

See [TODO.md](TODO.md) for complete list of planned features.

### High Priority
- [x] Service-to-service authentication (Shared secret)
- [ ] Real-time projections with PostgreSQL LISTEN/NOTIFY
- [ ] Slack message handling integration
- [ ] Scheduler reminder logic

### Medium Priority
- [ ] Health check endpoints for all services
- [ ] Structured logging and observability
- [ ] User-facing authentication per team

### Nice to Have
- [ ] Web dashboard for viewing updates
- [ ] Slack slash commands (`/status`, `/team-register`)
- [ ] Analytics and reporting
- [ ] Event versioning and upcasting
- [ ] Snapshot support for projections

## License

MIT
