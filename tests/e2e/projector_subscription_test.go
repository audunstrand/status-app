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

// TestProjectorSubscriptionRealtime tests that the projector subscribes to new events
// and updates projections in real-time without needing a restart
func TestProjectorSubscriptionRealtime(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	
	// Use real PostgresStore instead of testEventStore for LISTEN/NOTIFY support
	eventStore, err := events.NewPostgresStore(testDB.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to create event store: %v", err)
	}
	defer eventStore.Close()
	
	cmdHandler := commands.NewHandler(eventStore)
	repo := projections.NewRepository(testDB.DB)

	// Start the projector
	projector := projections.NewProjector(eventStore, testDB.DB)
	err = projector.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start projector: %v", err)
	}

	// Wait a bit for the projector to initialize
	time.Sleep(100 * time.Millisecond)

	registerCmd := mustRegisterTeam("Engineering", "#engineering")

	err = cmdHandler.Handle(ctx, registerCmd)
	if err != nil {
		t.Fatalf("Failed to register team: %v", err)
	}

	// Get team ID from events
	teamEvents, _ := eventStore.GetAll(ctx, "team.registered", 0, 10)
	if len(teamEvents) == 0 {
		t.Fatal("No team registered event found")
	}

	var teamData events.TeamRegisteredData
	json.Unmarshal(teamEvents[0].Data, &teamData)
	teamID := teamData.TeamID

	// Wait for projector to process the event
	time.Sleep(500 * time.Millisecond)

	// Verify team was projected in real-time (without manual projection)
	team, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Team should be projected automatically, but got error: %v", err)
	}

	if team.Name != "Engineering" {
		t.Errorf("Expected team name 'Engineering', got '%s'", team.Name)
	}

	submitCmd := mustSubmitStatusUpdate(
		teamID,
		"engineering",
		"Implemented real-time projections",
		"Alice",
		"alice",
	)

	err = cmdHandler.Handle(ctx, submitCmd)
	if err != nil {
		t.Fatalf("Failed to submit status update: %v", err)
	}

	// Wait for projector to process the event
	time.Sleep(500 * time.Millisecond)

	// Verify status update was projected in real-time
	updates, err := repo.GetTeamUpdates(ctx, teamID, 10)
	if err != nil {
		t.Fatalf("Failed to get team updates: %v", err)
	}

	if len(updates) != 1 {
		t.Fatalf("Expected 1 update to be projected automatically, got %d", len(updates))
	}

	if updates[0].Content != "Implemented real-time projections" {
		t.Errorf("Expected correct content, got: %s", updates[0].Content)
	}

	t.Logf("✅ Real-time projection subscription working")
}
