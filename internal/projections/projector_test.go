package projections

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/tests/testutil"
)

type projectorTestEnv struct {
	t         *testing.T
	ctx       context.Context
	testDB    *testutil.TestDB
	store     events.Store
	projector *Projector
	repo      *Repository
}

func setupProjector(t *testing.T) *projectorTestEnv {
	t.Helper()
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)

	store, err := events.NewPostgresStore(testDB.ConnectionString())
	testutil.AssertNoError(t, err, "NewPostgresStore")

	t.Cleanup(func() {
		store.Close()
		testDB.Cleanup()
	})

	return &projectorTestEnv{
		t:         t,
		ctx:       ctx,
		testDB:    testDB,
		store:     store,
		projector: NewProjector(store, testDB.DB),
		repo:      NewRepository(testDB.DB),
	}
}

func (e *projectorTestEnv) appendEvent(event *events.Event) {
	e.t.Helper()
	testutil.AssertNoError(e.t, e.store.Append(e.ctx, event), "Append event")
}

func (e *projectorTestEnv) rebuild() {
	e.t.Helper()
	testutil.AssertNoError(e.t, e.projector.rebuildProjections(e.ctx), "Rebuild projections")
}

// newTestEvent creates a test event with sensible defaults
func newTestEvent(t *testing.T, eventType, aggregateID string, data interface{}, timestamp time.Time) *events.Event {
	t.Helper()
	return &events.Event{
		ID:          testutil.GenerateID(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        testutil.MustMarshalJSON(t, data),
		Timestamp:   timestamp,
		Version:     1,
	}
}

// newTeamRegisteredEvent creates a team.registered event
func newTeamRegisteredEvent(t *testing.T, teamID, name, channel, schedule string, timestamp time.Time) *events.Event {
	t.Helper()
	data := events.TeamRegisteredData{
		TeamID:       teamID,
		Name:         name,
		SlackChannel: channel,
		PollSchedule: schedule,
	}
	return newTestEvent(t, events.TeamRegistered, teamID, data, timestamp)
}

// newTeamUpdatedEvent creates a team.updated event
func newTeamUpdatedEvent(t *testing.T, teamID, name, channel, schedule string, timestamp time.Time) *events.Event {
	t.Helper()
	data := events.TeamRegisteredData{
		TeamID:       teamID,
		Name:         name,
		SlackChannel: channel,
		PollSchedule: schedule,
	}
	return newTestEvent(t, events.TeamUpdated, teamID, data, timestamp)
}

// newStatusUpdateEvent creates a status_update.submitted event
func newStatusUpdateEvent(t *testing.T, teamID, content, author, slackUser string, timestamp time.Time) *events.Event {
	t.Helper()
	data := events.StatusUpdateSubmittedData{
		UpdateID:  testutil.GenerateID(),
		TeamID:    teamID,
		Content:   content,
		Author:    author,
		SlackUser: slackUser,
		Timestamp: timestamp,
	}
	return newTestEvent(t, events.StatusUpdateSubmitted, teamID, data, timestamp)
}

func TestProjector_RebuildProjections(t *testing.T) {
	t.Run("handles team registration and multiple updates", func(t *testing.T) {
		env := setupProjector(t)
		teamID := "team-rebuild-1"
		now := time.Now()

		// Event 1: Team registered
		env.appendEvent(newTeamRegisteredEvent(t, teamID, "Original Engineering", "#engineering-old", "weekly", now))

		// Event 2: Team updated (first time)
		env.appendEvent(newTeamUpdatedEvent(t, teamID, "Updated Engineering", "#engineering-new", "daily", now.Add(1*time.Hour)))

		// Event 3: Team updated (second time)
		env.appendEvent(newTeamUpdatedEvent(t, teamID, "Final Engineering Team", "#engineering-final", "twice-daily", now.Add(2*time.Hour)))

		// Rebuild and verify final state
		env.rebuild()

		team, err := env.repo.GetTeam(env.ctx, teamID)
		testutil.AssertNoError(t, err, "GetTeam")

		// Should have the final values from event3
		testutil.AssertEqual(t, team.Name, "Final Engineering Team", "Team name")
		testutil.AssertEqual(t, team.SlackChannel, "#engineering-final", "SlackChannel")
		testutil.AssertEqual(t, team.PollSchedule, "twice-daily", "PollSchedule")
	})

	t.Run("handles status updates for multiple teams", func(t *testing.T) {
		env := setupProjector(t)
		now := time.Now()

		// Register two teams FIRST (with earlier timestamps)
		teams := []struct {
			id, name, channel string
		}{
			{"team-multi-1", "Team Alpha", "#alpha"},
			{"team-multi-2", "Team Beta", "#beta"},
		}

		for i, team := range teams {
			timestamp := now.Add(-10 * time.Minute).Add(time.Duration(i) * time.Minute)
			env.appendEvent(newTeamRegisteredEvent(t, team.id, team.name, team.channel, "weekly", timestamp))
		}

		// Add status updates for both teams AFTER teams are registered
		updates := []struct {
			teamID, content, author string
		}{
			{"team-multi-1", "Alpha update 1", "Alice"},
			{"team-multi-1", "Alpha update 2", "Bob"},
			{"team-multi-2", "Beta update 1", "Charlie"},
			{"team-multi-1", "Alpha update 3", "Alice"},
			{"team-multi-2", "Beta update 2", "Dave"},
		}

		for i, update := range updates {
			timestamp := now.Add(time.Duration(i) * time.Minute)
			env.appendEvent(newStatusUpdateEvent(t, update.teamID, update.content, update.author, "U"+update.author, timestamp))
		}

		// Rebuild projections
		env.rebuild()

		// Verify team-multi-1 has 3 updates
		team1Updates, err := env.repo.GetTeamUpdates(env.ctx, "team-multi-1", 100)
		testutil.AssertNoError(t, err, "GetTeamUpdates team-1")
		testutil.AssertEqual(t, len(team1Updates), 3, "Team 1 update count")

		// Verify team-multi-2 has 2 updates
		team2Updates, err := env.repo.GetTeamUpdates(env.ctx, "team-multi-2", 100)
		testutil.AssertNoError(t, err, "GetTeamUpdates team-2")
		testutil.AssertEqual(t, len(team2Updates), 2, "Team 2 update count")

		// Verify team summaries
		summary1, err := env.repo.GetTeamSummary(env.ctx, "team-multi-1")
		testutil.AssertNoError(t, err, "GetTeamSummary team-1")
		testutil.AssertEqual(t, summary1.TotalUpdates, 3, "Team 1 total updates")
		testutil.AssertEqual(t, summary1.UniqueContributors, 2, "Team 1 unique contributors (Alice, Bob)")

		summary2, err := env.repo.GetTeamSummary(env.ctx, "team-multi-2")
		testutil.AssertNoError(t, err, "GetTeamSummary team-2")
		testutil.AssertEqual(t, summary2.TotalUpdates, 2, "Team 2 total updates")
		testutil.AssertEqual(t, summary2.UniqueContributors, 2, "Team 2 unique contributors (Charlie, Dave)")
	})

	t.Run("handles idempotent event processing", func(t *testing.T) {
		env := setupProjector(t)
		teamID := "team-idempotent"
		now := time.Now()

		// Register team and add status update
		env.appendEvent(newTeamRegisteredEvent(t, teamID, "Idempotent Team", "#idempotent", "weekly", now))
		env.appendEvent(newStatusUpdateEvent(t, teamID, "Test update", "Author", "U123", now))

		// Rebuild projections twice (should be idempotent)
		env.rebuild()
		env.rebuild()

		// Verify we still only have 1 team and 1 update (not duplicates)
		team, err := env.repo.GetTeam(env.ctx, teamID)
		testutil.AssertNoError(t, err, "GetTeam")
		testutil.AssertEqual(t, team.Name, "Idempotent Team", "Team name")

		updates, err := env.repo.GetTeamUpdates(env.ctx, teamID, 100)
		testutil.AssertNoError(t, err, "GetTeamUpdates")
		testutil.AssertEqual(t, len(updates), 1, "Update count (idempotent rebuild)")
	})

	t.Run("handles empty event stream", func(t *testing.T) {
		env := setupProjector(t)

		// Should complete without error even if called multiple times
		env.rebuild()
		env.rebuild()
	})

	t.Run("verifies event ordering matters for team updates", func(t *testing.T) {
		env := setupProjector(t)
		teamID := "team-ordering"
		now := time.Now()

		// Event sequence: register -> update -> update
		// The LAST update should be what we see in the projection
		env.appendEvent(newTeamRegisteredEvent(t, teamID, "First Name", "#first", "weekly", now))
		env.appendEvent(newTeamUpdatedEvent(t, teamID, "Second Name", "#second", "daily", now.Add(1*time.Hour)))
		env.appendEvent(newTeamUpdatedEvent(t, teamID, "Third Name", "#third", "hourly", now.Add(2*time.Hour)))

		// Rebuild and verify we have the LAST state
		env.rebuild()

		team, err := env.repo.GetTeam(env.ctx, teamID)
		testutil.AssertNoError(t, err, "GetTeam")

		// Verify the projection has the FINAL state (from event3)
		testutil.AssertEqual(t, team.Name, "Third Name", "Team name (last update)")
		testutil.AssertEqual(t, team.SlackChannel, "#third", "SlackChannel")
		testutil.AssertEqual(t, team.PollSchedule, "hourly", "PollSchedule")
	})
}

func TestProjector_RealTimeUpdates(t *testing.T) {
	t.Run("projects events in real-time without polling", func(t *testing.T) {
		env := setupProjector(t)

		// Start the projector with real-time subscription
		ctx, cancel := context.WithTimeout(env.ctx, 30*time.Second)
		defer cancel()

		testutil.AssertNoError(t, env.projector.Start(ctx), "Start projector")

		// Give the projector time to initialize
		time.Sleep(200 * time.Millisecond)

		// Add a team event
		teamID := "team-realtime"
		now := time.Now()
		env.appendEvent(newTeamRegisteredEvent(t, teamID, "Real-time Team", "#realtime", "daily", now))

		// Wait a moment for real-time processing
		time.Sleep(500 * time.Millisecond)

		// Verify team was projected
		team, err := env.repo.GetTeam(env.ctx, teamID)
		testutil.AssertNoError(t, err, "GetTeam")
		testutil.AssertEqual(t, team.Name, "Real-time Team", "Team name")

		// Add a status update event
		env.appendEvent(newStatusUpdateEvent(t, teamID, "First update", "Alice", "U123", now.Add(1*time.Minute)))

		// Wait for real-time processing
		time.Sleep(500 * time.Millisecond)

		// Verify update was projected
		updates, err := env.repo.GetTeamUpdates(env.ctx, teamID, 10)
		testutil.AssertNoError(t, err, "GetTeamUpdates")
		testutil.AssertEqual(t, len(updates), 1, "Update count")
		testutil.AssertEqual(t, updates[0].Content, "First update", "Update content")

		// Add another update
		env.appendEvent(newStatusUpdateEvent(t, teamID, "Second update", "Bob", "U456", now.Add(2*time.Minute)))

		// Wait for real-time processing
		time.Sleep(500 * time.Millisecond)

		// Verify second update was projected
		updates, err = env.repo.GetTeamUpdates(env.ctx, teamID, 10)
		testutil.AssertNoError(t, err, "GetTeamUpdates")
		testutil.AssertEqual(t, len(updates), 2, "Update count after second update")
	})

	t.Run("handles rapid event sequences", func(t *testing.T) {
		env := setupProjector(t)

		ctx, cancel := context.WithTimeout(env.ctx, 30*time.Second)
		defer cancel()

		testutil.AssertNoError(t, env.projector.Start(ctx), "Start projector")
		time.Sleep(200 * time.Millisecond)

		teamID := "team-rapid"
		now := time.Now()

		// Register team
		env.appendEvent(newTeamRegisteredEvent(t, teamID, "Rapid Team", "#rapid", "weekly", now))

		// Rapidly add multiple updates
		for i := 0; i < 5; i++ {
			env.appendEvent(newStatusUpdateEvent(t,
				teamID,
				fmt.Sprintf("Update %d", i+1),
				"Author",
				"U123",
				now.Add(time.Duration(i)*time.Second),
			))
		}

		// Wait for all events to be processed
		time.Sleep(2 * time.Second)

		// Verify all updates were projected
		updates, err := env.repo.GetTeamUpdates(env.ctx, teamID, 10)
		testutil.AssertNoError(t, err, "GetTeamUpdates")
		testutil.AssertEqual(t, len(updates), 5, "All rapid updates should be projected")
	})
}
