package e2e

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourusername/status-app/internal/commands"
	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/internal/projections"
	"github.com/yourusername/status-app/tests/testutil"
)

// TestAutoRegisterTeamOnFirstUpdate tests that a team is automatically registered
// when the first status update is posted to an unregistered channel
func TestAutoRegisterTeamOnFirstUpdate(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	eventStore := newTestEventStore(testDB.DB)
	cmdHandler := commands.NewHandler(eventStore)
	repo := projections.NewRepository(testDB.DB)

	// Submit a status update to a channel that hasn't been registered
	channelID := "C123456789"
	channelName := "general"
	
	submitCmd := commands.SubmitStatusUpdate{
		TeamID:      channelID,
		ChannelName: channelName,
		Content:     "First update in this channel",
		Author:      "Jane Smith",
		SlackUser:   "jane.smith",
		Timestamp:   time.Now(),
	}

	err := cmdHandler.Handle(ctx, submitCmd)
	if err != nil {
		t.Fatalf("Failed to submit status update: %v", err)
	}

	// Verify team was auto-registered
	teamEvents, err := eventStore.GetAll(ctx, "team.registered", 0, 10)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(teamEvents) == 0 {
		t.Fatal("Expected team to be auto-registered, but no team.registered event found")
	}

	var teamData events.TeamRegisteredData
	json.Unmarshal(teamEvents[0].Data, &teamData)

	if teamData.TeamID != channelID {
		t.Errorf("Expected team ID to be channel ID %s, got %s", channelID, teamData.TeamID)
	}

	if teamData.Name != channelName {
		t.Errorf("Expected team name to be channel name %s, got %s", channelName, teamData.Name)
	}

	if teamData.SlackChannel != channelID {
		t.Errorf("Expected slack channel to be %s, got %s", channelID, teamData.SlackChannel)
	}

	// Project the team
	projectTeamRegistered(ctx, testDB.DB, &teamData)

	// Verify team projection exists
	team, err := repo.GetTeam(ctx, channelID)
	if err != nil {
		t.Fatalf("Failed to get auto-registered team: %v", err)
	}

	if team.Name != channelName {
		t.Errorf("Expected team name %s, got %s", channelName, team.Name)
	}

	// Verify status update was also stored
	statusEvents, _ := eventStore.GetAll(ctx, "status_update.submitted", 0, 10)
	if len(statusEvents) == 0 {
		t.Fatal("Expected status update event to be stored")
	}

	var updateData events.StatusUpdateSubmittedData
	json.Unmarshal(statusEvents[0].Data, &updateData)
	projectStatusUpdate(ctx, testDB.DB, &updateData)

	// Query the status update
	updates, err := repo.GetTeamUpdates(ctx, channelID, 10)
	if err != nil {
		t.Fatalf("Failed to get team updates: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(updates))
	}

	if updates[0].Content != "First update in this channel" {
		t.Errorf("Expected correct content, got: %s", updates[0].Content)
	}

	t.Logf("✅ Auto-register team on first update complete")
}

// TestSubsequentUpdateDoesNotRegisterAgain tests that submitting a second update
// to an already registered channel does not create duplicate team registrations
func TestSubsequentUpdateDoesNotRegisterAgain(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	eventStore := newTestEventStore(testDB.DB)
	cmdHandler := commands.NewHandler(eventStore)

	channelID := "C987654321"
	channelName := "engineering"

	// First update - should auto-register
	firstCmd := commands.SubmitStatusUpdate{
		TeamID:      channelID,
		ChannelName: channelName,
		Content:     "First update",
		Author:      "Alice",
		SlackUser:   "alice",
		Timestamp:   time.Now(),
	}

	err := cmdHandler.Handle(ctx, firstCmd)
	if err != nil {
		t.Fatalf("Failed to submit first update: %v", err)
	}

	// Second update - should NOT register again
	secondCmd := commands.SubmitStatusUpdate{
		TeamID:      channelID,
		ChannelName: channelName,
		Content:     "Second update",
		Author:      "Bob",
		SlackUser:   "bob",
		Timestamp:   time.Now(),
	}

	err = cmdHandler.Handle(ctx, secondCmd)
	if err != nil {
		t.Fatalf("Failed to submit second update: %v", err)
	}

	// Verify only one team.registered event exists
	teamEvents, err := eventStore.GetAll(ctx, "team.registered", 0, 10)
	if err != nil {
		t.Fatalf("Failed to get events: %v", err)
	}

	if len(teamEvents) != 1 {
		t.Errorf("Expected exactly 1 team.registered event, got %d", len(teamEvents))
	}

	// Verify both status updates were stored
	statusEvents, err := eventStore.GetAll(ctx, "status_update.submitted", 0, 10)
	if err != nil {
		t.Fatalf("Failed to get status events: %v", err)
	}

	if len(statusEvents) != 2 {
		t.Errorf("Expected 2 status_update.submitted events, got %d", len(statusEvents))
	}

	t.Logf("✅ Subsequent update does not register again complete")
}
