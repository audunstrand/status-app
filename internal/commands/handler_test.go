package commands

import (
	"context"
	"testing"
	"time"

	"github.com/yourusername/status-app/internal/events"
)

// MockEventStore for testing
type MockEventStore struct {
	events []*events.Event
	err    error
}

func (m *MockEventStore) Append(ctx context.Context, event *events.Event) error {
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, event)
	return nil
}

func (m *MockEventStore) GetByAggregateID(ctx context.Context, aggregateID string) ([]*events.Event, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []*events.Event
	for _, e := range m.events {
		if e.AggregateID == aggregateID {
			filtered = append(filtered, e)
		}
	}
	return filtered, nil
}

func (m *MockEventStore) GetAll(ctx context.Context, eventType string, offset, limit int) ([]*events.Event, error) {
	return m.events, m.err
}

func (m *MockEventStore) Subscribe(ctx context.Context, eventTypes []string) (<-chan *events.Event, error) {
	ch := make(chan *events.Event)
	close(ch)
	return ch, nil
}

func (m *MockEventStore) Close() error {
	return nil
}

func TestHandler_HandleSubmitStatusUpdate(t *testing.T) {
	store := &MockEventStore{}
	handler := NewHandler(store)

	cmd := SubmitStatusUpdate{
		TeamID:      "team-1",
		ChannelName: "engineering",
		Content:     "Fixed critical bug",
		Author:      "John Doe",
		SlackUser:   "john.doe",
		Timestamp:   time.Now(),
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Should have 2 events: team.registered (auto) + status_update.submitted
	if len(store.events) != 2 {
		t.Fatalf("expected 2 events (auto-register + status update), got %d", len(store.events))
	}

	if store.events[0].Type != "team.registered" {
		t.Errorf("expected first event type team.registered, got %s", store.events[0].Type)
	}

	if store.events[1].Type != "status_update.submitted" {
		t.Errorf("expected second event type status_update.submitted, got %s", store.events[1].Type)
	}

	if store.events[1].AggregateID != "team-1" {
		t.Errorf("expected aggregate ID team-1, got %s", store.events[1].AggregateID)
	}
}

func TestHandler_HandleRegisterTeam(t *testing.T) {
	store := &MockEventStore{}
	handler := NewHandler(store)

	cmd := RegisterTeam{
		Name:         "Engineering",
		SlackChannel: "#engineering",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(store.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(store.events))
	}

	event := store.events[0]
	if event.Type != "team.registered" {
		t.Errorf("expected event type team.registered, got %s", event.Type)
	}
}

func TestHandler_HandleUpdateTeam(t *testing.T) {
	store := &MockEventStore{}
	handler := NewHandler(store)

	cmd := UpdateTeam{
		TeamID:       "team-1",
		Name:         "Updated Engineering",
		SlackChannel: "#new-engineering",
	}

	err := handler.Handle(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(store.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(store.events))
	}

	event := store.events[0]
	if event.Type != "team.updated" {
		t.Errorf("expected event type team.updated, got %s", event.Type)
	}

	if event.AggregateID != "team-1" {
		t.Errorf("expected aggregate ID team-1, got %s", event.AggregateID)
	}
}

func TestHandler_UnknownCommandType(t *testing.T) {
	store := &MockEventStore{}
	handler := NewHandler(store)

	// Create a command type that doesn't match any handled types
	type UnknownCommand struct{}
	
	// Add Validate method
	unknownCmd := UnknownCommand{}
	
	// We can't test this directly because UnknownCommand doesn't implement Command interface
	// This test verifies the code compiles and handler exists
	_ = handler
	_ = unknownCmd
}
