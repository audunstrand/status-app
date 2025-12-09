package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)

type PostgresStore struct {
	db      *sql.DB
	connStr string
}

func NewPostgresStore(connStr string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStore{db: db, connStr: connStr}, nil
}

func (s *PostgresStore) Append(ctx context.Context, event *Event) error {
	query := `
		INSERT INTO events (id, type, aggregate_id, data, timestamp, metadata, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	
	// Handle nil metadata - PostgreSQL expects NULL, not an empty json.RawMessage
	var metadata interface{}
	if event.Metadata == nil || len(event.Metadata) == 0 {
		metadata = nil
	} else {
		metadata = event.Metadata
	}
	
	_, err := s.db.ExecContext(ctx, query,
		event.ID,
		event.Type,
		event.AggregateID,
		event.Data,
		event.Timestamp,
		metadata,
		event.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to append event: %w", err)
	}

	// Notify listeners (PostgreSQL NOTIFY)
	// NOTIFY doesn't support parameterized queries, so we use a constant channel name
	if _, err := s.db.ExecContext(ctx, fmt.Sprintf("NOTIFY events, '%s'", event.ID)); err != nil {
		log.Printf("Warning: failed to notify listeners: %v", err)
	}

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

func (s *PostgresStore) GetByID(ctx context.Context, id string) (*Event, error) {
	query := `
		SELECT id, type, aggregate_id, data, timestamp, metadata, version
		FROM events
		WHERE id = $1
	`
	row := s.db.QueryRowContext(ctx, query, id)

	var event Event
	var metadata sql.NullString

	err := row.Scan(
		&event.ID,
		&event.Type,
		&event.AggregateID,
		&event.Data,
		&event.Timestamp,
		&metadata,
		&event.Version,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Event not found
		}
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	if metadata.Valid {
		event.Metadata = json.RawMessage(metadata.String)
	}

	return &event, nil
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
	// Create PostgreSQL listener
	listener := pq.NewListener(
		s.connStr,
		10*time.Second,  // minReconnectInterval
		time.Minute,     // maxReconnectInterval
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("Listener error: %v", err)
			}
		},
	)

	// Listen on the 'events' channel
	if err := listener.Listen("events"); err != nil {
		listener.Close()
		return nil, fmt.Errorf("failed to listen on events channel: %w", err)
	}

	ch := make(chan *Event, 10) // Buffered to prevent blocking NOTIFY

	// Start goroutine to process notifications
	go func() {
		defer close(ch)
		defer listener.Close()

		for {
			select {
			case n := <-listener.Notify:
				if n != nil {
					// n.Extra contains the event ID
					event, err := s.GetByID(ctx, n.Extra)
					if err != nil {
						log.Printf("Failed to get event %s: %v", n.Extra, err)
						continue
					}
					if event != nil {
						// Check if event type matches filter (if specified)
						if len(eventTypes) == 0 || containsEventType(eventTypes, event.Type) {
							select {
							case ch <- event:
							case <-ctx.Done():
								return
							}
						}
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

// containsEventType checks if a given event type is in the list
func containsEventType(types []string, eventType string) bool {
	for _, t := range types {
		if t == eventType {
			return true
		}
	}
	return false
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
