package projections

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/status-app/tests/testutil"
)

func TestRepository_GetTeam(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	repo := NewRepository(testDB.DB)

	// Insert test team
	_, err := testDB.DB.ExecContext(ctx, `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "team-1", "Engineering", "#engineering", "weekly", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test team: %v", err)
	}

	t.Run("retrieves existing team", func(t *testing.T) {
		team, err := repo.GetTeam(ctx, "team-1")
		if err != nil {
			t.Errorf("GetTeam() error = %v", err)
		}

		if team.TeamID != "team-1" {
			t.Errorf("TeamID = %v, want team-1", team.TeamID)
		}
		if team.Name != "Engineering" {
			t.Errorf("Name = %v, want Engineering", team.Name)
		}
		if team.SlackChannel != "#engineering" {
			t.Errorf("SlackChannel = %v, want #engineering", team.SlackChannel)
		}
	})

	t.Run("returns error for non-existent team", func(t *testing.T) {
		_, err := repo.GetTeam(ctx, "non-existent")
		if err == nil {
			t.Error("GetTeam() expected error for non-existent team, got nil")
		}
	})
}

func TestRepository_GetAllTeams(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	repo := NewRepository(testDB.DB)

	// Insert multiple teams
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
		_, err := testDB.DB.ExecContext(ctx, `
			INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, team.id, team.name, team.channel, "weekly", time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to insert team: %v", err)
		}
	}

	allTeams, err := repo.GetAllTeams(ctx)
	if err != nil {
		t.Fatalf("GetAllTeams() error = %v", err)
	}

	if len(allTeams) != 3 {
		t.Errorf("GetAllTeams() returned %d teams, want 3", len(allTeams))
	}

	// Verify teams are ordered by name
	if len(allTeams) > 0 && allTeams[0].Name != "Design" {
		t.Errorf("First team = %v, want Design (alphabetical order)", allTeams[0].Name)
	}
}

func TestRepository_GetTeamUpdates(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	repo := NewRepository(testDB.DB)

	// Insert test team
	_, err := testDB.DB.ExecContext(ctx, `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "team-1", "Engineering", "#engineering", "weekly", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert team: %v", err)
	}

	// Insert multiple status updates
	for i := 0; i < 5; i++ {
		_, err := testDB.DB.ExecContext(ctx, `
			INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, 
			testutil.GenerateID(),
			"team-1",
			"Update "+string(rune('A'+i)),
			"Author",
			"U123",
			time.Now().Add(time.Duration(i)*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to insert status update: %v", err)
		}
	}

	t.Run("retrieves team updates with limit", func(t *testing.T) {
		updates, err := repo.GetTeamUpdates(ctx, "team-1", 3)
		if err != nil {
			t.Errorf("GetTeamUpdates() error = %v", err)
		}

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
		if err != nil {
			t.Errorf("GetTeamUpdates() error = %v", err)
		}

		if len(updates) != 5 {
			t.Errorf("GetTeamUpdates() returned %d updates, want 5", len(updates))
		}
	})
}

func TestRepository_GetRecentUpdates(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	repo := NewRepository(testDB.DB)

	// Insert test teams with proper IDs
	teamIDs := []string{"team-rec-1", "team-rec-2"}
	for i, teamID := range teamIDs {
		_, err := testDB.DB.ExecContext(ctx, `
			INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, 
			teamID,
			"Team "+string(rune('A'+i)),
			"#team",
			"weekly",
			time.Now(),
			time.Now(),
		)
		if err != nil {
			t.Fatalf("Failed to insert team: %v", err)
		}
	}

	// Insert updates across multiple teams
	for i, teamID := range teamIDs {
		for j := 0; j < 3; j++ {
			_, err := testDB.DB.ExecContext(ctx, `
				INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
				VALUES ($1, $2, $3, $4, $5, $6)
			`,
				testutil.GenerateID(),
				teamID,
				"Update",
				"Author",
				"U123",
				time.Now().Add(time.Duration(i*10+j)*time.Second),
			)
			if err != nil {
				t.Fatalf("Failed to insert status update: %v", err)
			}
		}
	}

	updates, err := repo.GetRecentUpdates(ctx, 4)
	if err != nil {
		t.Errorf("GetRecentUpdates() error = %v", err)
	}

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
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	repo := NewRepository(testDB.DB)

	// Insert test team
	_, err := testDB.DB.ExecContext(ctx, `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "team-1", "Engineering", "#engineering", "weekly", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert team: %v", err)
	}

	// Insert status updates from different users
	users := []string{"U123", "U456", "U123", "U789"}
	for i, user := range users {
		_, err := testDB.DB.ExecContext(ctx, `
			INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
			testutil.GenerateID(),
			"team-1",
			"Update",
			"Author",
			user,
			time.Now().Add(time.Duration(i)*time.Second),
		)
		if err != nil {
			t.Fatalf("Failed to insert status update: %v", err)
		}
	}

	summary, err := repo.GetTeamSummary(ctx, "team-1")
	if err != nil {
		t.Fatalf("GetTeamSummary() error = %v", err)
	}

	if summary.Team.TeamID != "team-1" {
		t.Errorf("TeamID = %v, want team-1", summary.Team.TeamID)
	}

	if summary.TotalUpdates != 4 {
		t.Errorf("TotalUpdates = %d, want 4", summary.TotalUpdates)
	}

	if summary.UniqueContributos != 3 {
		t.Errorf("UniqueContributos = %d, want 3 (U123, U456, U789)", summary.UniqueContributos)
	}

	if summary.LastUpdateAt.IsZero() {
		t.Error("LastUpdateAt should not be zero")
	}
}
