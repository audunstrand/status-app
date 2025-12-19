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

// TestUpdateTeamName tests updating a team's name after it has been auto-registered
func TestUpdateTeamName(t *testing.T) {
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	ctx := context.Background()
	eventStore := newTestEventStore(testDB.DB)
	cmdHandler := commands.NewHandler(eventStore)
	repo := projections.NewRepository(testDB.DB)

	channelID := "C123456789"
	registerCmd := mustRegisterTeam("general", channelID)

	err := cmdHandler.Handle(ctx, registerCmd)
	if err != nil {
		t.Fatalf("Failed to register team: %v", err)
	}

	// Get team ID from event and project it
	teamEvents, _ := eventStore.GetAll(ctx, "team.registered", 0, 10)
	var teamData events.TeamRegisteredData
	json.Unmarshal(teamEvents[0].Data, &teamData)
	teamID := teamData.TeamID
	projectTeamRegistered(ctx, testDB.DB, &teamData)

	// Verify initial team name
	team, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get team: %v", err)
	}

	if team.Name != "general" {
		t.Errorf("Expected initial team name 'general', got '%s'", team.Name)
	}

	updateCmd := mustUpdateTeam(teamID, "Product Team", channelID)

	err = cmdHandler.Handle(ctx, updateCmd)
	if err != nil {
		t.Fatalf("Failed to update team name: %v", err)
	}

	// Project the update
	updateEvents, _ := eventStore.GetAll(ctx, "team.updated", 0, 10)
	var updatedData events.TeamRegisteredData
	json.Unmarshal(updateEvents[0].Data, &updatedData)
	projectTeamUpdated(ctx, testDB.DB, &updatedData)

	// Verify updated team name
	updatedTeam, err := repo.GetTeam(ctx, teamID)
	if err != nil {
		t.Fatalf("Failed to get updated team: %v", err)
	}

	if updatedTeam.Name != "Product Team" {
		t.Errorf("Expected updated team name 'Product Team', got '%s'", updatedTeam.Name)
	}

	// Verify slack channel didn't change
	if updatedTeam.SlackChannel != channelID {
		t.Errorf("Expected slack channel %s, got %s", channelID, updatedTeam.SlackChannel)
	}

	t.Logf("✅ Update team name complete")
}
