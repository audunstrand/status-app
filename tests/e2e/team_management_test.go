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

// TestTeamManagementFlow tests team registration and updates:
// Register team → Event stored → Query team → Update team → Query updated team
func TestTeamManagementFlow(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	eventStore := newTestEventStore(testDB.DB)
	cmdHandler := commands.NewHandler(eventStore)
	repo := projections.NewRepository(testDB.DB)

	registerCmd := mustRegisterTeam("Product Team", "#product")

	err := cmdHandler.Handle(ctx, registerCmd)
	if err != nil {
		t.Fatalf("Failed to register team: %v", err)
	}

	// Get team ID from event
	allEvents, _ := eventStore.GetAll(ctx, "team.registered", 0, 10)
	if len(allEvents) == 0 {
		t.Fatal("No team registered event found")
	}

	var teamData events.TeamRegisteredData
	json.Unmarshal(allEvents[0].Data, &teamData)
	teamID := teamData.TeamID

	// Project the team
	projectTeamRegistered(ctx, testDB.DB, &teamData)

	// Step 2: Query the team
	team, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}

	if team.Name != "Product Team" {
		t.Errorf("Expected team name 'Product Team', got '%s'", team.Name)
	}

	if team.SlackChannel != "#product" {
		t.Errorf("Expected slack channel '#product', got '%s'", team.SlackChannel)
	}

	updateCmd := mustUpdateTeam(teamID, "Product Team v2", "#product-new")

	err = cmdHandler.Handle(ctx, updateCmd)
	if err != nil {
		t.Fatalf("Failed to update team: %v", err)
	}

	// Project the update
	updateEvents, _ := eventStore.GetAll(ctx, "team.updated", 0, 10)
	if len(updateEvents) > 0 {
		var updatedData events.TeamRegisteredData
		json.Unmarshal(updateEvents[0].Data, &updatedData)
		projectTeamUpdated(ctx, testDB.DB, &updatedData)
	}

	// Step 4: Query updated team
	updatedTeam, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get updated team: %v", err)
	}

	if updatedTeam.Name != "Product Team v2" {
		t.Errorf("Expected updated team name 'Product Team v2', got '%s'", updatedTeam.Name)
	}

	if updatedTeam.SlackChannel != "#product-new" {
		t.Errorf("Expected updated slack channel '#product-new', got '%s'", updatedTeam.SlackChannel)
	}

	// Step 5: Verify event history
	teamEvents, err := eventStore.GetByAggregateID(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get team events: %v", err)
	}

	if len(teamEvents) != 2 {
		t.Fatalf("Expected 2 events (registered + updated), got %d", len(teamEvents))
	}

	if teamEvents[0].Type != "team.registered" {
		t.Errorf("Expected first event to be team.registered, got %s", teamEvents[0].Type)
	}

	if teamEvents[1].Type != "team.updated" {
		t.Errorf("Expected second event to be team.updated, got %s", teamEvents[1].Type)
	}

	// Step 6: Query all teams
	allTeams, err := repo.GetAllTeams(ctx)
	if err != nil {
		t.Fatalf("Failed to get all teams: %v", err)
	}

	if len(allTeams) != 1 {
		t.Fatalf("Expected 1 team, got %d", len(allTeams))
	}

	t.Logf("✅ E2E Team Management complete")
}
