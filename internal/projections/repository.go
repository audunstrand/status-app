package projections

import (
	"context"
	"database/sql"
	"fmt"
)

// Repository provides read access to projections
type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// scanTeam scans a Team from a row scanner
func (r *Repository) scanTeam(scanner interface {
	Scan(...interface{}) error
}) (*Team, error) {
	var team Team
	err := scanner.Scan(
		&team.TeamID,
		&team.Name,
		&team.SlackChannel,
		&team.PollSchedule,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	return &team, err
}

// scanStatusUpdate scans a StatusUpdate from a row scanner
func (r *Repository) scanStatusUpdate(scanner interface {
	Scan(...interface{}) error
}) (*StatusUpdate, error) {
	var update StatusUpdate
	err := scanner.Scan(
		&update.UpdateID,
		&update.TeamID,
		&update.Content,
		&update.Author,
		&update.SlackUser,
		&update.CreatedAt,
	)
	return &update, err
}

func (r *Repository) GetTeam(ctx context.Context, teamID string) (*Team, error) {
	query := `
		SELECT team_id, name, slack_channel, poll_schedule, created_at, updated_at
		FROM teams
		WHERE team_id = $1
	`
	return r.scanTeam(r.db.QueryRowContext(ctx, query, teamID))
}

func (r *Repository) GetAllTeams(ctx context.Context) ([]*Team, error) {
	query := `
		SELECT team_id, name, slack_channel, poll_schedule, created_at, updated_at
		FROM teams
		ORDER BY name
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var teams []*Team
	for rows.Next() {
		team, err := r.scanTeam(rows)
		if err != nil {
			return nil, err
		}
		teams = append(teams, team)
	}
	return teams, rows.Err()
}

func (r *Repository) GetTeamUpdates(ctx context.Context, teamID string, limit int) ([]*StatusUpdate, error) {
	query := `
		SELECT update_id, team_id, content, author, slack_user, created_at
		FROM status_updates
		WHERE team_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.QueryContext(ctx, query, teamID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatusUpdates(rows)
}

func (r *Repository) GetRecentUpdates(ctx context.Context, limit int) ([]*StatusUpdate, error) {
	query := `
		SELECT update_id, team_id, content, author, slack_user, created_at
		FROM status_updates
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanStatusUpdates(rows)
}

// scanStatusUpdates scans multiple StatusUpdate rows
func (r *Repository) scanStatusUpdates(rows *sql.Rows) ([]*StatusUpdate, error) {
	var updates []*StatusUpdate
	for rows.Next() {
		update, err := r.scanStatusUpdate(rows)
		if err != nil {
			return nil, err
		}
		updates = append(updates, update)
	}
	return updates, rows.Err()
}

func (r *Repository) GetTeamSummary(ctx context.Context, teamID string) (*TeamSummary, error) {
	team, err := r.GetTeam(ctx, teamID)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT 
			COUNT(*) as total_updates,
			MAX(created_at) as last_update_at,
			COUNT(DISTINCT slack_user) as unique_contributors
		FROM status_updates
		WHERE team_id = $1
	`
	var summary TeamSummary
	summary.Team = *team

	err = r.db.QueryRowContext(ctx, query, teamID).Scan(
		&summary.TotalUpdates,
		&summary.LastUpdateAt,
		&summary.UniqueContributors,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get team summary: %w", err)
	}

	return &summary, nil
}
