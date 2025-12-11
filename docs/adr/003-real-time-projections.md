# ADR 003: Real-Time Projections with PostgreSQL LISTEN/NOTIFY

**Date**: 2025-12-10  
**Status**: Accepted  
**Deciders**: Audun

## Context

In an event-sourced system with CQRS, projections (read models) need to be updated when events are stored. The original implementation had a gap between event storage and projection updates, making the system eventually consistent with noticeable delays.

We needed a mechanism to trigger projection updates immediately after events are stored, without:
- Adding external infrastructure (message queues)
- Polling the database
- Complex coordination logic

## Decision

**Use PostgreSQL LISTEN/NOTIFY for real-time projection updates.**

When an event is stored in the events table, the event store sends a NOTIFY signal. The projector service listens for these notifications and immediately processes new events.

## Options Considered

### Option 1: Polling with Intervals
**Implementation**: Check for new events every N seconds

**Pros**:
- Simple to implement
- No special PostgreSQL features needed
- Works with any database

**Cons**:
- Delay between event and projection (up to N seconds)
- Wasted queries when no new events
- Database load from constant polling
- Not truly real-time

### Option 2: External Message Queue (Redis Pub/Sub, RabbitMQ, etc.)
**Implementation**: Publish event notifications to external message broker

**Pros**:
- Dedicated messaging infrastructure
- Advanced features (persistence, retries, routing)
- Scales independently

**Cons**:
- Additional infrastructure to manage
- More operational complexity
- Network calls between services
- Added dependencies
- Overkill for our use case

### Option 3: PostgreSQL LISTEN/NOTIFY ✅ **CHOSEN**
**Implementation**: Use PostgreSQL's built-in pub/sub mechanism

**Pros**:
- Built into PostgreSQL (no new infrastructure)
- Instant notification (truly real-time)
- No polling overhead
- Simple implementation
- Connection-based (automatic cleanup)
- Works within transaction boundaries

**Cons**:
- PostgreSQL-specific (couples to database)
- Notifications not persisted (lost if listener down)
- Limited to single database instance

## Implementation

### Event Store Changes
```go
// After storing event, notify listeners
_, err := s.db.ExecContext(ctx, "NOTIFY events, $1", event.ID)
if err != nil {
    log.Printf("Warning: failed to notify listeners: %v", err)
}
```

### Projector Changes
```go
// Create listener
listener := pq.NewListener(dbURL, 10*time.Second, time.Minute, nil)
listener.Listen("events")

// Process notifications
for notification := range listener.Notify {
    eventID := notification.Extra
    processEvent(eventID)
}
```

### Integration Tests
Added comprehensive tests to verify:
- Events trigger notifications
- Projections update immediately
- Multiple events handled correctly
- Error cases handled gracefully

## Consequences

### Positive
- **Real-time updates**: Projections update within milliseconds of event storage
- **No additional infrastructure**: Leverages existing PostgreSQL
- **Simple implementation**: ~50 lines of code
- **Database native**: Works with transaction semantics
- **No polling overhead**: Notifications only when events occur

### Negative
- **PostgreSQL dependency**: Tighter coupling to PostgreSQL
  - *Mitigation*: Already committed to PostgreSQL for event sourcing
- **Notification loss risk**: If projector down, notifications lost
  - *Mitigation*: Projector rebuilds from event stream on startup
- **Single database scaling**: LISTEN/NOTIFY doesn't work across replicas
  - *Mitigation*: Current scale doesn't require read replicas

### Neutral
- **Connection management**: Requires persistent connection
- **Error handling**: Added listener error handling

## Alternatives Not Considered

- **Database triggers**: More complex, harder to test
- **CDC (Change Data Capture)**: Overkill for our scale
- **HTTP webhooks from event store**: Requires more infrastructure

## Related Commits

- `84b4f6d` - Implement real-time projections using PostgreSQL LISTEN/NOTIFY
- `53dc2e0` - feat: implement PostgreSQL LISTEN/NOTIFY for real-time projections
- `5ba3570` - fix: real-time projection subscription now working
- `7cb850d` - refactor: improve code quality and add validation

## Verification

- All unit tests passing
- Integration tests verify real-time behavior
- Deployed to production successfully
- Observed instant projection updates in production

## Notes

This decision significantly improved user experience by eliminating the delay between posting a status update and seeing it reflected in queries. The implementation is battle-tested and has been running in production without issues.
