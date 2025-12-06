# Architecture Documentation

## Event Sourcing Overview

Event Sourcing stores all changes to application state as a sequence of events. Instead of storing current state, we store the events that led to that state.

### Benefits

1. **Complete Audit Log**: Every change is recorded
2. **Temporal Queries**: Query state at any point in time
3. **Event Replay**: Rebuild state by replaying events
4. **New Projections**: Add new read models without changing write side
5. **Debugging**: Understand exactly how state was reached

## CQRS Pattern

Command Query Responsibility Segregation separates reads and writes:

- **Commands**: Change state (write)
- **Queries**: Read state (read)

### Write Side (Commands)

1. Slack bot receives message
2. Creates `SubmitStatusUpdate` command
3. Command handler validates and emits `StatusUpdateSubmitted` event
4. Event is appended to event store

### Read Side (Queries)

1. Projection service subscribes to events
2. Builds materialized views (projections)
3. API reads from projections (fast queries)

## Data Flow

```
User → Slack → Bot → Command → Event → Event Store
                                           ↓
                                       Projector
                                           ↓
                                   Read Models (DB)
                                           ↓
                                       Query API
```

## Event Store Schema

```sql
events (
  id            UUID PRIMARY KEY,
  type          VARCHAR,      -- Event type
  aggregate_id  VARCHAR,      -- Team ID, User ID, etc.
  data          JSONB,        -- Event payload
  timestamp     TIMESTAMP,
  metadata      JSONB,        -- Context (user, IP, etc.)
  version       INTEGER       -- Event version for evolution
)
```

## Projections Schema

### Teams Table
```sql
teams (
  team_id       UUID PRIMARY KEY,
  name          VARCHAR,
  slack_channel VARCHAR,
  poll_schedule VARCHAR,
  created_at    TIMESTAMP,
  updated_at    TIMESTAMP
)
```

### Status Updates Table
```sql
status_updates (
  update_id   UUID PRIMARY KEY,
  team_id     UUID REFERENCES teams,
  content     TEXT,
  author      VARCHAR,
  slack_user  VARCHAR,
  created_at  TIMESTAMP
)
```

## Event Processing

### Idempotency

Projections must be idempotent - processing same event multiple times produces same result. Use `ON CONFLICT DO NOTHING` or track processed event IDs.

### Eventual Consistency

Read models are eventually consistent with event store. Small delay between command execution and query visibility.

### Event Versioning

Events use version field for schema evolution:
- Version 1: Original format
- Version 2: Add new field (backward compatible)
- Version 3: Restructure (requires migration)

## Scaling Considerations

1. **Event Store**: Partition by aggregate ID
2. **Projections**: Multiple projectors can process events in parallel
3. **Read Models**: Can use read replicas, caching
4. **Event Bus**: NATS/Kafka for distributed event streaming

## Deployment

Each service can be deployed independently:

```
┌───────────┐  ┌─────────────┐  ┌────────┐
│ Commands  │  │ Projections │  │  API   │
│ (3 pods)  │  │  (2 pods)   │  │(5 pods)│
└─────┬─────┘  └──────┬──────┘  └────────┘
      │               │
      └───────┬───────┘
              ▼
      ┌──────────────┐
      │ Event Store  │
      │ (PostgreSQL) │
      └──────────────┘
```

## Error Handling

1. **Command Validation**: Fail fast before emitting events
2. **Event Store Failures**: Retry with exponential backoff
3. **Projection Failures**: Log and continue (rebuild later)
4. **Dead Letter Queue**: For events that consistently fail

## Monitoring

Key metrics to track:
- Events written per second
- Projection lag (time behind event store)
- Command processing time
- Query response time
- Error rates per service
