package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/yourusername/status-app/internal/events"
)

// testEventStore wraps sql.DB to implement events.Store for testing
type testEventStore struct {
	db *sql.DB
}

func newTestEventStore(db *sql.DB) *testEventStore {
	return &testEventStore{db: db}
}

func (s *testEventStore) Append(ctx context.Context, event *events.Event) error {
	query := `
		INSERT INTO events (id, type, aggregate_id, data, timestamp, metadata, version, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		event.AggregateID,
		event.Data,
		event.Timestamp,
		event.Metadata,
		event.Version,
		time.Now(),
	)
	return err
}

func (s *testEventStore) GetByAggregateID(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	query := `
		SELECT id, type, aggregate_id, data, timestamp, metadata, version
		FROM events
		WHERE aggregate_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *testEventStore) GetAll(ctx context.Context, eventType string, offset, limit int) ([]*events.Event, error) {
	var query string
	var args []interface{}

	if eventType == "" {
		query = `SELECT id, type, aggregate_id, data, timestamp, metadata, version
				 FROM events ORDER BY timestamp ASC LIMIT $1 OFFSET $2`
		args = []interface{}{limit, offset}
	} else {
		query = `SELECT id, type, aggregate_id, data, timestamp, metadata, version
				 FROM events WHERE type = $1 ORDER BY timestamp ASC LIMIT $2 OFFSET $3`
		args = []interface{}{eventType, limit, offset}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanEvents(rows)
}

func (s *testEventStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *events.Event, error) {
	ch := make(chan *events.Event)
	close(ch)
	return ch, nil
}

func (s *testEventStore) Close() error {
	return nil
}

func scanEvents(rows *sql.Rows) ([]*events.Event, error) {
	var result []*events.Event
	for rows.Next() {
		var event events.Event
		var metadata sql.NullString
		err := rows.Scan(
			&event.ID,
			&event.Type,
			&event.AggregateID,
			&event.Data,
			&event.Timestamp,
			&metadata,
			&event.Version,
		)
		if err != nil {
			return nil, err
		}
		if metadata.Valid {
			event.Metadata = json.RawMessage(metadata.String)
		}
		result = append(result, &event)
	}
	return result, rows.Err()
}

// Projection helpers

func projectTeamRegistered(ctx context.Context, db *sql.DB, data *events.TeamRegisteredData) error {
	query := `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (team_id) DO NOTHING
	`
	_, err := db.ExecContext(ctx, query,
		data.TeamID,
		data.Name,
		data.SlackChannel,
		data.PollSchedule,
		time.Now(),
		time.Now(),
	)
	return err
}

func projectTeamUpdated(ctx context.Context, db *sql.DB, data *events.TeamRegisteredData) error {
	query := `
		UPDATE teams
		SET name = $2, slack_channel = $3, poll_schedule = $4, updated_at = $5
		WHERE team_id = $1
	`
	_, err := db.ExecContext(ctx, query,
		data.TeamID,
		data.Name,
		data.SlackChannel,
		data.PollSchedule,
		time.Now(),
	)
	return err
}

func projectStatusUpdate(ctx context.Context, db *sql.DB, data *events.StatusUpdateSubmittedData) error {
	query := `
		INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (update_id) DO NOTHING
	`
	_, err := db.ExecContext(ctx, query,
		data.UpdateID,
		data.TeamID,
		data.Content,
		data.Author,
		data.SlackUser,
		data.Timestamp,
	)
	return err
}
