package projections

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/tests/testutil"
)

func TestProjector_RebuildProjections(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	// Create event store
	eventStore, err := events.NewPostgresStore(testDB.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer eventStore.Close()

	// Create projector
	projector := NewProjector(eventStore, testDB.DB)

	t.Run("handles team registration and multiple updates", func(t *testing.T) {
		teamID := "team-rebuild-1"
		now := time.Now()

		// Event 1: Team registered
		teamRegData, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Original Engineering",
			SlackChannel: "#engineering-old",
			PollSchedule: "weekly",
		})

		event1 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamRegistered,
			AggregateID: teamID,
			Data:        teamRegData,
			Timestamp:   now,
			Version:     1,
		}

		if err := eventStore.Append(ctx, event1); err != nil {
			t.Fatalf("Failed to append event 1: %v", err)
		}

		// Event 2: Team updated (first time)
		teamUpdate1Data, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Updated Engineering",
			SlackChannel: "#engineering-new",
			PollSchedule: "daily",
		})

		event2 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamUpdated,
			AggregateID: teamID,
			Data:        teamUpdate1Data,
			Timestamp:   now.Add(1 * time.Hour),
			Version:     2,
		}

		if err := eventStore.Append(ctx, event2); err != nil {
			t.Fatalf("Failed to append event 2: %v", err)
		}

		// Event 3: Team updated (second time)
		teamUpdate2Data, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Final Engineering Team",
			SlackChannel: "#engineering-final",
			PollSchedule: "twice-daily",
		})

		event3 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamUpdated,
			AggregateID: teamID,
			Data:        teamUpdate2Data,
			Timestamp:   now.Add(2 * time.Hour),
			Version:     3,
		}

		if err := eventStore.Append(ctx, event3); err != nil {
			t.Fatalf("Failed to append event 3: %v", err)
		}

		// Rebuild projections
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild projections: %v", err)
		}

		// Verify final state - should have the LAST update applied
		repo := NewRepository(testDB.DB)
		team, err := repo.GetTeam(ctx, teamID)
		if err != nil {
			t.Fatalf("Failed to get team: %v", err)
		}

		// Should have the final values from event3
		if team.Name != "Final Engineering Team" {
			t.Errorf("Team name = %v, want 'Final Engineering Team'", team.Name)
		}
		if team.SlackChannel != "#engineering-final" {
			t.Errorf("Team slack_channel = %v, want '#engineering-final'", team.SlackChannel)
		}
		if team.PollSchedule != "twice-daily" {
			t.Errorf("Team poll_schedule = %v, want 'twice-daily'", team.PollSchedule)
		}
	})

	t.Run("handles status updates for multiple teams", func(t *testing.T) {
		now := time.Now()

		// Register two teams FIRST (with earlier timestamps)
		teams := []struct {
			id      string
			name    string
			channel string
		}{
			{"team-multi-1", "Team Alpha", "#alpha"},
			{"team-multi-2", "Team Beta", "#beta"},
		}

		for i, team := range teams {
			teamData, _ := json.Marshal(events.TeamRegisteredData{
				TeamID:       team.id,
				Name:         team.name,
				SlackChannel: team.channel,
				PollSchedule: "weekly",
			})

			event := &events.Event{
				ID:          testutil.GenerateID(),
				Type:        events.TeamRegistered,
				AggregateID: team.id,
				Data:        teamData,
				Timestamp:   now.Add(-10 * time.Minute).Add(time.Duration(i) * time.Minute), // Before updates
				Version:     1,
			}

			if err := eventStore.Append(ctx, event); err != nil {
				t.Fatalf("Failed to append team event: %v", err)
			}
		}

		// Add status updates for both teams AFTER teams are registered
		updates := []struct {
			teamID  string
			content string
			author  string
		}{
			{"team-multi-1", "Alpha update 1", "Alice"},
			{"team-multi-1", "Alpha update 2", "Bob"},
			{"team-multi-2", "Beta update 1", "Charlie"},
			{"team-multi-1", "Alpha update 3", "Alice"},
			{"team-multi-2", "Beta update 2", "Dave"},
		}

		for i, update := range updates {
			updateData, _ := json.Marshal(events.StatusUpdateSubmittedData{
				UpdateID:  testutil.GenerateID(),
				TeamID:    update.teamID,
				Content:   update.content,
				Author:    update.author,
				SlackUser: "U" + update.author,
				Timestamp: now.Add(time.Duration(i) * time.Minute),
			})

			event := &events.Event{
				ID:          testutil.GenerateID(),
				Type:        events.StatusUpdateSubmitted,
				AggregateID: update.teamID,
				Data:        updateData,
				Timestamp:   now.Add(time.Duration(i) * time.Minute),
				Version:     i + 1,
			}

			if err := eventStore.Append(ctx, event); err != nil {
				t.Fatalf("Failed to append status update: %v", err)
			}
		}

		// Rebuild projections - should process teams first, then updates
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild projections: %v", err)
		}

		// Verify team-multi-1 has 3 updates
		repo := NewRepository(testDB.DB)
		team1Updates, err := repo.GetTeamUpdates(ctx, "team-multi-1", 100)
		if err != nil {
			t.Fatalf("Failed to get team 1 updates: %v", err)
		}

		if len(team1Updates) != 3 {
			t.Errorf("Team 1 updates count = %d, want 3", len(team1Updates))
		}

		// Verify team-multi-2 has 2 updates
		team2Updates, err := repo.GetTeamUpdates(ctx, "team-multi-2", 100)
		if err != nil {
			t.Fatalf("Failed to get team 2 updates: %v", err)
		}

		if len(team2Updates) != 2 {
			t.Errorf("Team 2 updates count = %d, want 2", len(team2Updates))
		}

		// Verify team summaries
		summary1, err := repo.GetTeamSummary(ctx, "team-multi-1")
		if err != nil {
			t.Fatalf("Failed to get team 1 summary: %v", err)
		}

		if summary1.TotalUpdates != 3 {
			t.Errorf("Team 1 total updates = %d, want 3", summary1.TotalUpdates)
		}

		// Team 1 has updates from Alice (2) and Bob (1) = 2 unique contributors
		if summary1.UniqueContributos != 2 {
			t.Errorf("Team 1 unique contributors = %d, want 2 (Alice, Bob)", summary1.UniqueContributos)
		}

		summary2, err := repo.GetTeamSummary(ctx, "team-multi-2")
		if err != nil {
			t.Fatalf("Failed to get team 2 summary: %v", err)
		}

		if summary2.TotalUpdates != 2 {
			t.Errorf("Team 2 total updates = %d, want 2", summary2.TotalUpdates)
		}

		// Team 2 has updates from Charlie and Dave = 2 unique contributors
		if summary2.UniqueContributos != 2 {
			t.Errorf("Team 2 unique contributors = %d, want 2 (Charlie, Dave)", summary2.UniqueContributos)
		}
	})

	t.Run("handles idempotent event processing", func(t *testing.T) {
		teamID := "team-idempotent"
		now := time.Now()

		// Register team
		teamData, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Idempotent Team",
			SlackChannel: "#idempotent",
			PollSchedule: "weekly",
		})

		teamEvent := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamRegistered,
			AggregateID: teamID,
			Data:        teamData,
			Timestamp:   now,
			Version:     1,
		}

		if err := eventStore.Append(ctx, teamEvent); err != nil {
			t.Fatalf("Failed to append team event: %v", err)
		}

		// Add a status update
		updateID := testutil.GenerateID()
		updateData, _ := json.Marshal(events.StatusUpdateSubmittedData{
			UpdateID:  updateID,
			TeamID:    teamID,
			Content:   "Test update",
			Author:    "Author",
			SlackUser: "U123",
			Timestamp: now,
		})

		updateEvent := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.StatusUpdateSubmitted,
			AggregateID: teamID,
			Data:        updateData,
			Timestamp:   now,
			Version:     2,
		}

		if err := eventStore.Append(ctx, updateEvent); err != nil {
			t.Fatalf("Failed to append update event: %v", err)
		}

		// Rebuild projections first time
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild projections (first): %v", err)
		}

		// Rebuild projections second time (should be idempotent)
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild projections (second): %v", err)
		}

		// Verify we still only have 1 team and 1 update (not duplicates)
		repo := NewRepository(testDB.DB)
		
		team, err := repo.GetTeam(ctx, teamID)
		if err != nil {
			t.Fatalf("Failed to get team: %v", err)
		}
		if team.Name != "Idempotent Team" {
			t.Errorf("Team name = %v, want 'Idempotent Team'", team.Name)
		}

		updates, err := repo.GetTeamUpdates(ctx, teamID, 100)
		if err != nil {
			t.Fatalf("Failed to get updates: %v", err)
		}

		if len(updates) != 1 {
			t.Errorf("Update count = %d, want 1 (idempotent rebuild)", len(updates))
		}
	})

	t.Run("handles empty event stream", func(t *testing.T) {
		// Create a fresh projector on the same DB (already has events from other tests)
		// This tests that rebuild only processes existing events without errors
		
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild with existing events: %v", err)
		}

		// Should complete without error even if called multiple times
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed second rebuild: %v", err)
		}
	})

	t.Run("verifies event ordering matters for team updates", func(t *testing.T) {
		teamID := "team-ordering"
		now := time.Now()

		// Event sequence: register -> update -> update
		// The LAST update should be what we see in the projection

		// 1. Register
		registerData, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "First Name",
			SlackChannel: "#first",
			PollSchedule: "weekly",
		})

		event1 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamRegistered,
			AggregateID: teamID,
			Data:        registerData,
			Timestamp:   now,
			Version:     1,
		}

		if err := eventStore.Append(ctx, event1); err != nil {
			t.Fatalf("Failed to append register event: %v", err)
		}

		// 2. Update to "Second Name"
		update1Data, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Second Name",
			SlackChannel: "#second",
			PollSchedule: "daily",
		})

		event2 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamUpdated,
			AggregateID: teamID,
			Data:        update1Data,
			Timestamp:   now.Add(1 * time.Hour),
			Version:     2,
		}

		if err := eventStore.Append(ctx, event2); err != nil {
			t.Fatalf("Failed to append first update: %v", err)
		}

		// 3. Update to "Third Name" 
		update2Data, _ := json.Marshal(events.TeamRegisteredData{
			TeamID:       teamID,
			Name:         "Third Name",
			SlackChannel: "#third",
			PollSchedule: "hourly",
		})

		event3 := &events.Event{
			ID:          testutil.GenerateID(),
			Type:        events.TeamUpdated,
			AggregateID: teamID,
			Data:        update2Data,
			Timestamp:   now.Add(2 * time.Hour),
			Version:     3,
		}

		if err := eventStore.Append(ctx, event3); err != nil {
			t.Fatalf("Failed to append second update: %v", err)
		}

		// Rebuild and verify we have the LAST state
		if err := projector.rebuildProjections(ctx); err != nil {
			t.Fatalf("Failed to rebuild: %v", err)
		}

		repo := NewRepository(testDB.DB)
		team, err := repo.GetTeam(ctx, teamID)
		if err != nil {
			t.Fatalf("Failed to get team: %v", err)
		}

		// Verify the projection has the FINAL state (from event3)
		if team.Name != "Third Name" {
			t.Errorf("Team name = %v, want 'Third Name' (last update)", team.Name)
		}
		if team.SlackChannel != "#third" {
			t.Errorf("SlackChannel = %v, want '#third'", team.SlackChannel)
		}
		if team.PollSchedule != "hourly" {
			t.Errorf("PollSchedule = %v, want 'hourly'", team.PollSchedule)
		}
	})
}
