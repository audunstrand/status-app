# Architecture

Event-sourced CQRS system for team status updates via Slack.

## Overview

```
Slack → Slackbot → Commands → Event Store → Projections → API
                       ↑                          ↓
                   Scheduler                  Read Model
```

## Components

**Commands** (`cmd/commands`)
- Receives commands from Slackbot/Scheduler
- Validates and emits events to Event Store
- Port 8081

**Event Store** (PostgreSQL)
- Append-only log of all events
- Source of truth
- Notifies subscribers via PostgreSQL NOTIFY (TODO)

**Projections** (`cmd/projections`)
- Builds read models from events
- Updates projection database
- Subscribes to events (polling for now, LISTEN/NOTIFY TODO)

**API** (`cmd/api`)
- Read-only query endpoints
- Serves data from projection database
- Port 8082

**Slackbot** (`cmd/slackbot`)
- Receives Slack messages/mentions
- Sends commands to Commands service

**Scheduler** (`cmd/scheduler`)
- Cron job for team reminders (TODO: send reminders)

## Data Flow

1. User posts in Slack
2. Slackbot receives message
3. Slackbot sends `SubmitStatusUpdate` to Commands
4. Commands validates and emits `StatusUpdateSubmitted` event
5. Event stored in Event Store
6. Projections polls events and updates read model
7. API serves current state from read model

## Authentication

All service-to-service calls require `Authorization: Bearer <API_SECRET>`

## Databases

- **Event Store DB**: Events table (append-only)
- **Projection DB**: Teams, status_updates tables (read model)
