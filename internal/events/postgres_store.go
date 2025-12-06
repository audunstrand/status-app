package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/lib/pq"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Append(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (id, type, aggregate_id, data, timestamp, metadata, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		event.AggregateID,
		event.Data,
		event.Timestamp,
		event.Metadata,
		event.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	// Notify listeners (PostgreSQL NOTIFY)
	_, _ = s.db.ExecContext(ctx, "NOTIFY events, $1", event.ID)

	return nil
}

func (s *PostgresStore) GetByAggregateID(ctx context.Context, aggregateID string) ([]*Event, error) {
	query := `
		SELECT id, type, aggregate_id, data, timestamp, metadata, version
		FROM events
		WHERE aggregate_id = $1
		ORDER BY timestamp ASC
	`
	rows, err := s.db.QueryContext(ctx, query, aggregateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

func (s *PostgresStore) GetAll(ctx context.Context, eventType string, offset, limit int) ([]*Event, error) {
	var query string
	var args []interface{}

	if eventType != "" {
		query = `
			SELECT id, type, aggregate_id, data, timestamp, metadata, version
			FROM events
			WHERE type = $1
			ORDER BY timestamp ASC
			LIMIT $2 OFFSET $3
		`
		args = []interface{}{eventType, limit, offset}
	} else {
		query = `
			SELECT id, type, aggregate_id, data, timestamp, metadata, version
			FROM events
			ORDER BY timestamp ASC
			LIMIT $1 OFFSET $2
		`
		args = []interface{}{limit, offset}
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	return s.scanEvents(rows)
}

func (s *PostgresStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error) {
	// TODO: Implement LISTEN/NOTIFY for real-time event streaming
	// For now, return a simple implementation
	ch := make(chan *Event)
	go func() {
		<-ctx.Done()
		close(ch)
	}()
	return ch, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) scanEvents(rows *sql.Rows) ([]*Event, error) {
	var events []*Event

	for rows.Next() {
		var event Event
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
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if metadata.Valid {
			event.Metadata = json.RawMessage(metadata.String)
		}

		events = append(events, &event)
	}

	return events, rows.Err()
}
