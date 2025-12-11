package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
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

	return &PostgresStore{
		db:      db,
		connStr: connStr,
	}, nil
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
		eventStoreErrors.WithLabelValues("append").Inc()
		return fmt.Errorf("failed to append event: %w", err)
	}

	// Record metrics
	eventsStoredTotal.WithLabelValues(event.Type).Inc()
	eventsStoredBytes.Add(float64(len(event.Data)))

	// Notify listeners (PostgreSQL NOTIFY)
	notifyQuery := fmt.Sprintf("NOTIFY events, '%s'", event.ID)
	if _, err := s.db.ExecContext(ctx, notifyQuery); err != nil {
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
	// Create a new connection for listening (LISTEN requires its own connection)
	listener := pq.NewListener(
		s.connStr,
		10*time.Second,
		time.Minute,
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("pq.Listener error: %v", err)
			}
		},
	)

	if err := listener.Listen("events"); err != nil {
		return nil, fmt.Errorf("failed to listen on events channel: %w", err)
	}

	ch := make(chan *Event, 10) // Buffered channel to avoid blocking

	go func() {
		defer listener.Close()
		defer close(ch)

		for {
			select {
			case <-ctx.Done():
				return
			case notification := <-listener.Notify:
				if notification == nil {
					continue
				}

				// Fetch the event by ID from the notification payload
				eventID := notification.Extra
				event, err := s.getEventByID(ctx, eventID)
				if err != nil {
					log.Printf("failed to fetch event %s: %v", eventID, err)
					continue
				}

				// Filter by event type if specified
				if len(eventTypes) > 0 {
					match := false
					for _, et := range eventTypes {
						if event.Type == et {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}

				select {
				case ch <- event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (s *PostgresStore) getEventByID(ctx context.Context, eventID string) (*Event, error) {
	query := `
		SELECT id, type, aggregate_id, data, timestamp, metadata, version
		FROM events
		WHERE id = $1
	`
	var event Event
	var metadata sql.NullString

	err := s.db.QueryRowContext(ctx, query, eventID).Scan(
		&event.ID,
		&event.Type,
		&event.AggregateID,
		&event.Data,
		&event.Timestamp,
		&metadata,
		&event.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	if metadata.Valid {
		event.Metadata = json.RawMessage(metadata.String)
	}

	return &event, nil
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
			eventStoreErrors.WithLabelValues("scan").Inc()
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		if metadata.Valid {
			event.Metadata = json.RawMessage(metadata.String)
		}

		// Record metric for loaded event
		eventsLoadedTotal.WithLabelValues(event.Type).Inc()

		events = append(events, &event)
	}

	return events, rows.Err()
}
