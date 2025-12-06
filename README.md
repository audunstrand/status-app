# Status App - Event-Sourced Team Status Updates

An event-sourced application for collecting and managing one-line status updates from teams via Slack.

## Architecture

This application follows **Event Sourcing** and **CQRS** patterns:

- **Event Store**: Append-only log of all events (source of truth)
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

1. **Event Store** (`cmd/eventstore`) - Stores all events
2. **Commands** (`cmd/commands`) - Processes commands, emits events
3. **Projections** (`cmd/projections`) - Builds read models from events
4. **Slack Bot** (`cmd/slackbot`) - Receives Slack messages, sends commands
5. **Scheduler** (`cmd/scheduler`) - Sends weekly poll reminders
6. **API** (`cmd/api`) - Query endpoint for projections

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 14+
- Slack App with Bot Token

### Setup

1. **Install dependencies**:
```bash
go mod download
```

2. **Set environment variables**:
```bash
export SLACK_BOT_TOKEN="xoxb-your-token"
export SLACK_SIGNING_KEY="xapp-your-key"
export EVENT_STORE_URL="postgres://localhost:5432/statusapp_events"
export PROJECTION_DB_URL="postgres://localhost:5432/statusapp_projections"
```

3. **Run migrations**:
```bash
# Install golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations for event store
migrate -path migrations -database $EVENT_STORE_URL up

# Run migrations for projections
migrate -path migrations -database $PROJECTION_DB_URL up
```

4. **Start services**:
```bash
# Terminal 1: Command service
go run cmd/commands/main.go

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

- [ ] NATS/Kafka for event streaming
- [ ] EventStoreDB for dedicated event store
- [ ] Web dashboard for viewing updates
- [ ] Slack slash commands (`/status`, `/team-register`)
- [ ] Analytics and reporting
- [ ] Event versioning and upcasting
- [ ] Snapshot support for projections

## License

MIT
