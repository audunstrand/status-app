package projections

import (
	"context"
	"testing"

	"github.com/yourusername/status-app/tests/testutil"
)

func setupRepository(t *testing.T) (context.Context, *Repository, *testutil.TestDB) {
	t.Helper()
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	repo := NewRepository(testDB.DB)

	t.Cleanup(func() {
		testDB.Cleanup()
	})

	return ctx, repo, testDB
}

func TestRepository_GetTeam(t *testing.T) {
	ctx, repo, testDB := setupRepository(t)

	testutil.InsertTestTeam(t, testDB.DB, "team-1", "Engineering", "#engineering")

	t.Run("retrieves existing team", func(t *testing.T) {
		team, err := repo.GetTeam(ctx, "team-1")
		testutil.AssertNoError(t, err, "GetTeam")

		testutil.AssertEqual(t, team.TeamID, "team-1", "TeamID")
		testutil.AssertEqual(t, team.Name, "Engineering", "Name")
		testutil.AssertEqual(t, team.SlackChannel, "#engineering", "SlackChannel")
	})

	t.Run("returns error for non-existent team", func(t *testing.T) {
		_, err := repo.GetTeam(ctx, "non-existent")
		if err == nil {
			t.Error("GetTeam() expected error for non-existent team, got nil")
		}
	})
}

func TestRepository_GetAllTeams(t *testing.T) {
	ctx, repo, testDB := setupRepository(t)

	teams := []struct {
		id      string
		name    string
		channel string
	}{
		{"team-1", "Engineering", "#engineering"},
		{"team-2", "Product", "#product"},
		{"team-3", "Design", "#design"},
	}

	for _, team := range teams {
		testutil.InsertTestTeam(t, testDB.DB, team.id, team.name, team.channel)
	}

	allTeams, err := repo.GetAllTeams(ctx)
	testutil.AssertNoError(t, err, "GetAllTeams")

	if len(allTeams) != 3 {
		t.Errorf("GetAllTeams() returned %d teams, want 3", len(allTeams))
	}

	// Verify teams are ordered by name
	if len(allTeams) > 0 {
		testutil.AssertEqual(t, allTeams[0].Name, "Design", "First team name (alphabetical order)")
	}
}

func TestRepository_GetTeamUpdates(t *testing.T) {
	ctx, repo, testDB := setupRepository(t)

	testutil.InsertTestTeam(t, testDB.DB, "team-1", "Engineering", "#engineering")

	// Insert multiple status updates
	for i := 0; i < 5; i++ {
		testutil.InsertTestStatusUpdate(t, testDB.DB, "team-1", "Update "+string(rune('A'+i)), "Author", "U123")
	}

	t.Run("retrieves team updates with limit", func(t *testing.T) {
		updates, err := repo.GetTeamUpdates(ctx, "team-1", 3)
		testutil.AssertNoError(t, err, "GetTeamUpdates")

		if len(updates) != 3 {
			t.Errorf("GetTeamUpdates() returned %d updates, want 3", len(updates))
		}

		// Should be in descending order (most recent first)
		if len(updates) >= 2 && updates[0].CreatedAt.Before(updates[1].CreatedAt) {
			t.Error("Updates not in descending order by created_at")
		}
	})

	t.Run("retrieves all updates when limit is high", func(t *testing.T) {
		updates, err := repo.GetTeamUpdates(ctx, "team-1", 100)
		testutil.AssertNoError(t, err, "GetTeamUpdates")

		if len(updates) != 5 {
			t.Errorf("GetTeamUpdates() returned %d updates, want 5", len(updates))
		}
	})
}

func TestRepository_GetRecentUpdates(t *testing.T) {
	ctx, repo, testDB := setupRepository(t)

	// Insert test teams
	teamIDs := []string{"team-rec-1", "team-rec-2"}
	for i, teamID := range teamIDs {
		testutil.InsertTestTeam(t, testDB.DB, teamID, "Team "+string(rune('A'+i)), "#team")
	}

	// Insert updates across multiple teams
	for _, teamID := range teamIDs {
		for j := 0; j < 3; j++ {
			testutil.InsertTestStatusUpdate(t, testDB.DB, teamID, "Update", "Author", "U123")
		}
	}

	updates, err := repo.GetRecentUpdates(ctx, 4)
	testutil.AssertNoError(t, err, "GetRecentUpdates")

	if len(updates) != 4 {
		t.Errorf("GetRecentUpdates() returned %d updates, want 4", len(updates))
	}

	// Verify descending order
	for i := 0; i < len(updates)-1; i++ {
		if updates[i].CreatedAt.Before(updates[i+1].CreatedAt) {
			t.Error("Updates not in descending order")
		}
	}
}

func TestRepository_GetTeamSummary(t *testing.T) {
	ctx, repo, testDB := setupRepository(t)

	testutil.InsertTestTeam(t, testDB.DB, "team-1", "Engineering", "#engineering")

	// Insert status updates from different users
	users := []string{"U123", "U456", "U123", "U789"}
	for _, user := range users {
		testutil.InsertTestStatusUpdate(t, testDB.DB, "team-1", "Update", "Author", user)
	}

	summary, err := repo.GetTeamSummary(ctx, "team-1")
	testutil.AssertNoError(t, err, "GetTeamSummary")

	testutil.AssertEqual(t, summary.Team.TeamID, "team-1", "TeamID")
	testutil.AssertEqual(t, summary.TotalUpdates, 4, "TotalUpdates")
	testutil.AssertEqual(t, summary.UniqueContributors, 3, "UniqueContributors")

	if summary.LastUpdateAt.IsZero() {
		t.Error("LastUpdateAt should not be zero")
	}
}
