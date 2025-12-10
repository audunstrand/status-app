# ADR 003: Scheduler Reminder Implementation

**Date**: 2025-12-10  
**Status**: Proposed  
**Deciders**: Audun

## Context

The scheduler service exists but doesn't send actual reminders to Slack. We need to implement:
1. Parsing of `poll_schedule` field (e.g., "monday", "weekly", "friday")
2. Determining if a reminder should be sent today
3. Sending Slack messages to the appropriate channels
4. Tracking when reminders were last sent to avoid duplicates

## Decision Options

### Option 1: Simple Cron-style with Last Reminded Tracking

**Approach**:
- Add `last_reminded_at` timestamp to teams table
- Parse `poll_schedule` as simple keywords: "monday", "tuesday", "daily", "weekly"
- Check current day/time against schedule
- Send reminder if scheduled and not sent today
- Update `last_reminded_at` after sending

**Pros**:
- Simple to implement
- Easy to understand and debug
- No external dependencies
- Works with existing event sourcing model

**Cons**:
- Limited schedule flexibility (no "every 2 weeks" or specific times)
- Requires migration to add column
- Mixing read model state (last_reminded_at) with domain logic

**Implementation**:
```go
// Migration: add last_reminded_at to teams table
// Scheduler checks every hour/day
// Simple day-of-week matching logic
```

---

### Option 2: Full Cron Expression with Event Sourcing

**Approach**:
- Store cron expressions in `poll_schedule` (e.g., "0 9 * * MON")
- Use `github.com/robfig/cron` library
- Track reminders as domain events (`reminder.sent`)
- Query events to determine if reminder already sent today

**Pros**:
- Flexible scheduling (any time, any frequency)
- Pure event sourcing (no read model pollution)
- Professional-grade scheduling
- Can rebuild reminder history from events

**Cons**:
- More complex to implement
- Requires users to understand cron syntax
- External dependency
- Overhead of querying events every check

**Implementation**:
```go
// Use cron library for parsing
// Add ReminderSent event type
// Query events to check if sent today
```

---

### Option 3: Hybrid - Simple Keywords with Event Tracking

**Approach**:
- Keep simple schedule keywords ("monday", "friday", "weekly")
- Track reminders as domain events (`reminder.sent`)
- Query projection for last reminder instead of events
- Add `last_reminded_at` to projections table for efficiency

**Pros**:
- User-friendly schedule format
- Event sourcing principles maintained
- Efficient query performance
- Good balance of simplicity and correctness

**Cons**:
- Still requires migration
- Two sources of truth (events + projection)

---

## Recommendation

**Option 3: Hybrid Approach**

**Rationale**:
1. **User-friendly**: Team members understand "weekly" better than "0 9 * * MON"
2. **Event sourcing**: Reminders are tracked as events, maintaining audit trail
3. **Performance**: Projection table allows fast "was reminder sent today?" checks
4. **Future-proof**: Can add more keywords without breaking changes

**Migration needed**:
```sql
ALTER TABLE teams ADD COLUMN last_reminded_at TIMESTAMP WITH TIME ZONE;
```

**Event type**:
Already defined: `reminder.sent` with `ReminderSentData`

**Schedule format examples**:
- "monday" - Send every Monday
- "friday" - Send every Friday
- "weekly" - Send every Monday
- "daily" - Send every day
- "monday,friday" - Send Monday and Friday

## Implementation Plan

1. **Migration**: Add `last_reminded_at` column
2. **Test**: Create scheduler reminder tests
3. **Parser**: Implement schedule parsing logic
4. **Command**: Use existing `SendReminder` command
5. **Integration**: Wire up in scheduler main loop
6. **Deploy**: Test in production

## Questions for Approval

1. Is Option 3 (hybrid approach) acceptable?
2. Should we support comma-separated days (e.g., "monday,friday")?
3. What time should "daily"/"weekly" reminders be sent? (e.g., 9 AM local time?)
4. Should we add a user-facing command to test reminders (e.g., `/test-reminder`)?
