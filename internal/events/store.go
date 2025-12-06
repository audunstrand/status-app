package events

import (
	"context"
)

// Store defines the interface for event storage
type Store interface {
	// Append adds a new event to the store
	Append(ctx context.Context, event *Event) error

	// GetByAggregateID retrieves all events for a specific aggregate
	GetByAggregateID(ctx context.Context, aggregateID string) ([]*Event, error)

	// GetAll retrieves all events optionally filtered by type
	GetAll(ctx context.Context, eventType string, offset, limit int) ([]*Event, error)

	// Subscribe creates a subscription for new events
	Subscribe(ctx context.Context, eventTypes []string) (<-chan *Event, error)

	// Close closes the event store connection
	Close() error
}
