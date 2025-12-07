# Architecture

Event-sourced CQRS system for team status updates via Slack.

## Overview

```
Slack → Slackbot → Backend (Commands + Projections + API) → Database
                       ↑                                      (events + projections schemas)
                   Scheduler
```

## Current Architecture (After Phase 1 Consolidation)

**Services**: 5 → 3 (Backend consolidation in progress)
**Databases**: 2 → 1 ✅ Complete

### Components

**Backend** (`cmd/backend`) - *Combines 3 services*
- **Commands**: Receives commands, validates, emits events
- **Projections**: Builds read models from events (background goroutine)
- **API**: Read-only query endpoints
- Port 8080
- Endpoints: `/commands/*` and `/api/*`

**Slackbot** (`cmd/slackbot`)
- Receives Slack messages/mentions
- Sends commands to Backend service

**Scheduler** (`cmd/scheduler`)
- Cron job for team reminders

**Database** (PostgreSQL `status-app-db`)
- **events schema**: Append-only event log (source of truth)
- **projections schema**: Read models (teams, status_updates)

## Data Flow

1. User posts in Slack
2. Slackbot receives message
3. Slackbot sends `SubmitStatusUpdate` to Backend `/commands/submit-update`
4. Backend validates and emits `StatusUpdateSubmitted` event to `events.events`
5. Backend Projections processor polls for new events
6. Projections updates `projections.status_updates` table
7. API queries can read from `projections.*` tables via Backend `/api/*`

## Authentication

All service-to-service calls require `X-API-Key: <API_SECRET>` header

## Migration Path

**Phase 1** ✅: Database consolidation (2 databases → 1 database)
**Phase 2** 🚧: Service consolidation (5 services → 3 services)
- Current: Commands, API, Projections, Slackbot, Scheduler
- Target: Backend, Slackbot, Scheduler
