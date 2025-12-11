# ADR 005: Simplified Scheduler - Fixed Schedule

**Date**: 2025-12-10  
**Status**: Accepted  
**Deciders**: Audun

## Context

The initial scheduler design included per-team configurable reminder schedules:
- `poll_schedule` column in teams table
- Support for "monday", "friday", "daily", "weekly", etc.
- `SchedulePoll`, `SendReminder` commands
- `PollScheduled`, `ReminderSent` events
- `last_reminded_at` timestamp tracking

This added significant complexity:
- Schedule parsing logic
- Per-team schedule management
- Database columns for tracking reminder state
- Event types for reminder history
- Complex scheduler logic to check each team's schedule

However, **no user had requested configurable schedules**, and the most common use case was a simple "remind all teams at the same time" pattern.

## Decision

**Simplify to a fixed schedule: Send reminders to all teams every Monday at 9 AM.**

Remove all per-team scheduling infrastructure:
- Remove `poll_schedule` column from database
- Remove `SchedulePoll` and `SendReminder` commands
- Remove `PollScheduled` and `ReminderSent` events
- Remove `last_reminded_at` tracking
- Simplify scheduler to single cron job: `0 9 * * 1` (Monday 9 AM)

## Options Considered

### Option 1: Per-Team Configurable Schedules
**Implementation**: Full scheduling system with per-team configurations

**Pros**:
- Maximum flexibility
- Teams can choose their own schedule
- Different time zones supported
- Different frequencies supported

**Cons**:
- Significant complexity (200+ lines of code)
- Database columns for state management
- Additional event types
- No user had requested this feature
- Harder to test and maintain
- YAGNI (You Ain't Gonna Need It)

### Option 2: Fixed Schedule for All Teams ✅ **CHOSEN**
**Implementation**: Single cron job sends reminders to all teams

**Pros**:
- Simple implementation (~30 lines of code)
- No database state needed
- Easy to understand and maintain
- Meets current requirements
- Can be enhanced later if needed

**Cons**:
- Less flexible
- All teams get reminders at same time
- Cannot accommodate different time zones

### Option 3: Configurable Global Schedule
**Implementation**: Single schedule configured via environment variable

**Rejected**: Still more complex than needed, no requirement for changing schedule

## Implementation

### Removed Code
1. **Database**: Dropped `poll_schedule` column (migration 005)
2. **Commands**: Removed `SchedulePoll` and `SendReminder` commands
3. **Events**: Removed `PollScheduled` and `ReminderSent` event types
4. **Scheduler**: Simplified to single function

### New Scheduler
```go
// Setup cron scheduler
c := cron.New()

// Run every Monday at 9 AM
c.AddFunc("0 9 * * 1", func() {
    checkAndSendReminders(ctx, repo, slackAPI)
})
```

### Reminder Function
```go
func checkAndSendReminders(ctx context.Context, repo *Repository, slackAPI *slack.Client) {
    teams, err := repo.GetAllTeams(ctx)
    if err != nil {
        log.Printf("Failed to get teams: %v", err)
        return
    }

    for _, team := range teams {
        sendSlackReminder(slackAPI, team.SlackChannel, team.Name)
    }
}
```

## Consequences

### Positive
- **Simplicity**: Reduced complexity by 90%
- **Maintainability**: Easy to understand and modify
- **Reliability**: Fewer moving parts, less can go wrong
- **Performance**: No per-team state to query
- **YAGNI principle**: Only build what's needed

### Negative
- **Flexibility loss**: Cannot customize per team
  - *Mitigation*: Can be added later if users request it
- **Time zone limitations**: Fixed to single time zone
  - *Mitigation*: Acceptable for current user base

### Neutral
- **Functionality**: Still sends reminders as intended
- **User experience**: No change (feature wasn't exposed to users)

## Future Enhancements

If users request configurable schedules, we can add them back:

**Approach**:
1. Add `reminder_schedule` column (JSON or text)
2. Keep schedule parsing simple (e.g., "monday,friday" format)
3. Add validation on team registration/update
4. Scheduler checks each team's schedule

**Don't add unless requested**: Follow YAGNI principle

## Related Commits

- `051c1e7` - refactor: remove unused scheduling/reminder code
- `12ea00d` - Remove reminder time tracking and simplify scheduler
- `e96d80c` - Simplify scheduler to standard Monday 9 AM schedule
- Migration `005_remove_poll_schedule.up.sql`

## Testing

- All tests updated to remove `poll_schedule` references
- Test database schema cleaned up
- E2E tests passing
- Scheduler tested in production

## Notes

This is a textbook example of the YAGNI (You Ain't Gonna Need It) principle. We built complex scheduling infrastructure before any user requested it. Removing it:
- Reduced codebase by ~200 lines
- Eliminated a database column
- Removed 4 unused event types
- Made the system simpler and more maintainable

**Lesson**: Build the simplest thing that works, enhance when users request features.
