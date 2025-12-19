package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/yourusername/status-app/internal/commands"
	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/internal/projections"
	"github.com/yourusername/status-app/tests/testutil"
)

// TestStatusUpdateFlow tests the complete flow:
// Submit status update → Event stored → Projection built → Query via repository
func TestStatusUpdateFlow(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	eventStore := newTestEventStore(testDB.DB)
	cmdHandler := commands.NewHandler(eventStore)
	repo := projections.NewRepository(testDB.DB)

	registerCmd := mustRegisterTeam("Engineering", "#engineering")

	err := cmdHandler.Handle(ctx, registerCmd)
	if err != nil {
		t.Fatalf("Failed to register team: %v", err)
	}

	// Get team ID from events
	allEvents, _ := eventStore.GetAll(ctx, "team.registered", 0, 10)
	if len(allEvents) == 0 {
		t.Fatal("No team registered event found")
	}

	var teamData events.TeamRegisteredData
	json.Unmarshal(allEvents[0].Data, &teamData)
	teamID := teamData.TeamID

	// Project the team
	projectTeamRegistered(ctx, testDB.DB, &teamData)

	// Verify team projection
	team, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}

	if team.Name != "Engineering" {
		t.Errorf("Expected team name 'Engineering', got '%s'", team.Name)
	}

	submitCmd := mustSubmitStatusUpdate(
		teamID,
		"",
		"Completed the event sourcing implementation",
		"John Doe",
		"john.doe",
	)

	err = cmdHandler.Handle(ctx, submitCmd)
	if err != nil {
		t.Fatalf("Failed to submit status update: %v", err)
	}

	// Verify event was stored
	teamEvents, err := eventStore.GetByAggregateID(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get team events: %v", err)
	}

	if len(teamEvents) < 2 {
		t.Fatalf("Expected at least 2 events, got %d", len(teamEvents))
	}

	// Project the status update
	statusEvents, _ := eventStore.GetAll(ctx, "status_update.submitted", 0, 10)
	if len(statusEvents) > 0 {
		var updateData events.StatusUpdateSubmittedData
		json.Unmarshal(statusEvents[0].Data, &updateData)
		projectStatusUpdate(ctx, testDB.DB, &updateData)
	}

	// Step 3: Query the projection
	updates, err := repo.GetTeamUpdates(ctx, teamID, 10)
	if err != nil {
		t.Fatalf("Failed to get team updates: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(updates))
	}

	update := updates[0]
	if update.Content != "Completed the event sourcing implementation" {
		t.Errorf("Expected correct content, got: %s", update.Content)
	}

	if update.Author != "John Doe" {
		t.Errorf("Expected author 'John Doe', got '%s'", update.Author)
	}

	t.Logf("✅ E2E Status Update Flow complete")
}
