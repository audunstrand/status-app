package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourusername/status-app/tests/testutil"
)

func setupEventStore(t *testing.T) (context.Context, Store, *testutil.TestDB) {
	t.Helper()
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)

	store, err := NewPostgresStore(testDB.ConnectionString())
	testutil.AssertNoError(t, err, "NewPostgresStore")

	t.Cleanup(func() {
		store.Close()
		testDB.Cleanup()
	})

	return ctx, store, testDB
}

// newTestEvent creates a test event with sensible defaults
func newTestEvent(t *testing.T, eventType, aggregateID string, data interface{}, timestamp time.Time) *Event {
	t.Helper()
	return &Event{
		ID:          testutil.GenerateID(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        testutil.MustMarshalJSON(t, data),
		Timestamp:   timestamp,
		Version:     1,
	}
}

// newTeamRegisteredEvent creates a team.registered event
func newTeamRegisteredEvent(t *testing.T, teamID, name, channel, schedule string, timestamp time.Time) *Event {
	t.Helper()
	data := TeamRegisteredData{
		TeamID:       teamID,
		Name:         name,
		SlackChannel: channel,
		PollSchedule: schedule,
	}
	return newTestEvent(t, TeamRegistered, teamID, data, timestamp)
}

// newStatusUpdateEvent creates a status_update.submitted event
func newStatusUpdateEvent(t *testing.T, teamID, content, author, slackUser string, timestamp time.Time) *Event {
	t.Helper()
	data := StatusUpdateSubmittedData{
		UpdateID:  testutil.GenerateID(),
		TeamID:    teamID,
		Content:   content,
		Author:    author,
		SlackUser: slackUser,
		Timestamp: timestamp,
	}
	return newTestEvent(t, StatusUpdateSubmitted, teamID, data, timestamp)
}

func TestPostgresStore_Append(t *testing.T) {
	ctx, store, _ := setupEventStore(t)

	t.Run("appends event successfully", func(t *testing.T) {
		event := newTeamRegisteredEvent(t, "team-1", "Engineering", "#engineering", "weekly", time.Now())

		err := store.Append(ctx, event)
		testutil.AssertNoError(t, err, "Append")
	})

	t.Run("retrieves event by aggregate ID", func(t *testing.T) {
		event := newTeamRegisteredEvent(t, "team-2", "Product", "#product", "daily", time.Now())
		testutil.AssertNoError(t, store.Append(ctx, event), "Append")

		events, err := store.GetByAggregateID(ctx, "team-2")
		testutil.AssertNoError(t, err, "GetByAggregateID")

		if len(events) == 0 {
			t.Fatal("GetByAggregateID() returned no events")
		}

		found := false
		for _, e := range events {
			if e.ID == event.ID {
				found = true
				testutil.AssertEqual(t, e.Type, TeamRegistered, "Event type")
				testutil.AssertEqual(t, e.AggregateID, "team-2", "AggregateID")
			}
		}

		if !found {
			t.Errorf("Event %s not found in results", event.ID)
		}
	})

	t.Run("retrieves all events with pagination", func(t *testing.T) {
		// Append multiple events
		for i := 0; i < 5; i++ {
			event := newStatusUpdateEvent(t,
				"team-3",
				"Update "+string(rune('A'+i)),
				"Author",
				"U123",
				time.Now().Add(time.Duration(i)*time.Second),
			)
			testutil.AssertNoError(t, store.Append(ctx, event), "Append")
		}

		events, err := store.GetAll(ctx, "", 0, 10)
		testutil.AssertNoError(t, err, "GetAll")

		if len(events) < 5 {
			t.Errorf("GetAll() returned %d events, want at least 5", len(events))
		}
	})

	t.Run("filters events by type", func(t *testing.T) {
		events, err := store.GetAll(ctx, TeamRegistered, 0, 100)
		testutil.AssertNoError(t, err, "GetAll with filter")

		for _, e := range events {
			testutil.AssertEqual(t, e.Type, TeamRegistered, "Event type")
		}
	})
}

func TestPostgresStore_GetByAggregateID_Empty(t *testing.T) {
	ctx, store, _ := setupEventStore(t)

	events, err := store.GetByAggregateID(ctx, "non-existent")
	testutil.AssertNoError(t, err, "GetByAggregateID")

	if len(events) != 0 {
		t.Errorf("GetByAggregateID() returned %d events, want 0", len(events))
	}
}

func TestEvent_MarshalUnmarshal(t *testing.T) {
	teamData := TeamRegisteredData{
		TeamID:       "team-test",
		Name:         "Test Team",
		SlackChannel: "#test",
		PollSchedule: "daily",
	}

	data := testutil.MustMarshalJSON(t, teamData)

	var unmarshaled TeamRegisteredData
	testutil.AssertNoError(t, json.Unmarshal(data, &unmarshaled), "json.Unmarshal")

	testutil.AssertEqual(t, unmarshaled.TeamID, teamData.TeamID, "TeamID")
	testutil.AssertEqual(t, unmarshaled.Name, teamData.Name, "Name")
}

func TestPostgresStore_GetByID(t *testing.T) {
	ctx, store, _ := setupEventStore(t)

	t.Run("retrieves event by ID", func(t *testing.T) {
		event := newTeamRegisteredEvent(t, "team-getbyid", "GetByID Team", "#getbyid", "daily", time.Now())
		testutil.AssertNoError(t, store.Append(ctx, event), "Append")

		retrieved, err := store.GetByID(ctx, event.ID)
		testutil.AssertNoError(t, err, "GetByID")

		if retrieved == nil {
			t.Fatal("GetByID() returned nil event")
		}

		testutil.AssertEqual(t, retrieved.ID, event.ID, "Event ID")
		testutil.AssertEqual(t, retrieved.Type, event.Type, "Event type")
		testutil.AssertEqual(t, retrieved.AggregateID, event.AggregateID, "AggregateID")
	})

	t.Run("returns nil for non-existent ID", func(t *testing.T) {
		retrieved, err := store.GetByID(ctx, "non-existent-id")
		testutil.AssertNoError(t, err, "GetByID")

		if retrieved != nil {
			t.Errorf("GetByID() returned event for non-existent ID, want nil")
		}
	})
}

func TestPostgresStore_Subscribe(t *testing.T) {
	ctx, store, _ := setupEventStore(t)

	t.Run("receives events in real-time", func(t *testing.T) {
		// Create a cancellable context for the subscription
		subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Start subscription
		eventsCh, err := store.Subscribe(subCtx, []string{})
		testutil.AssertNoError(t, err, "Subscribe")

		// Give the listener time to connect
		time.Sleep(100 * time.Millisecond)

		// Append an event
		event := newTeamRegisteredEvent(t, "team-subscribe", "Subscribe Team", "#subscribe", "weekly", time.Now())
		testutil.AssertNoError(t, store.Append(ctx, event), "Append")

		// Wait for the event to be received
		select {
		case receivedEvent := <-eventsCh:
			if receivedEvent == nil {
				t.Fatal("Received nil event from subscription")
			}
			testutil.AssertEqual(t, receivedEvent.ID, event.ID, "Event ID")
			testutil.AssertEqual(t, receivedEvent.Type, event.Type, "Event type")
		case <-time.After(5 * time.Second):
			t.Fatal("Timed out waiting for event from subscription")
		}
	})

	t.Run("filters events by type", func(t *testing.T) {
		subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		// Subscribe only to StatusUpdateSubmitted events
		eventsCh, err := store.Subscribe(subCtx, []string{StatusUpdateSubmitted})
		testutil.AssertNoError(t, err, "Subscribe")

		time.Sleep(100 * time.Millisecond)

		// Append a TeamRegistered event (should be filtered out)
		teamEvent := newTeamRegisteredEvent(t, "team-filter", "Filter Team", "#filter", "daily", time.Now())
		testutil.AssertNoError(t, store.Append(ctx, teamEvent), "Append team event")

		// Append a StatusUpdateSubmitted event (should be received)
		updateEvent := newStatusUpdateEvent(t, "team-filter", "Update content", "Author", "U123", time.Now())
		testutil.AssertNoError(t, store.Append(ctx, updateEvent), "Append update event")

		// Should receive only the status update
		select {
		case receivedEvent := <-eventsCh:
			if receivedEvent == nil {
				t.Fatal("Received nil event from subscription")
			}
			testutil.AssertEqual(t, receivedEvent.Type, StatusUpdateSubmitted, "Event type should be StatusUpdateSubmitted")
		case <-time.After(5 * time.Second):
			t.Fatal("Timed out waiting for status update event")
		}

		// Verify no more events are received (team event was filtered)
		select {
		case evt := <-eventsCh:
			t.Fatalf("Should not receive TeamRegistered event, got: %v", evt.Type)
		case <-time.After(500 * time.Millisecond):
			// Good - no additional events received
		}
	})

	t.Run("closes channel on context cancellation", func(t *testing.T) {
		subCtx, cancel := context.WithCancel(ctx)

		eventsCh, err := store.Subscribe(subCtx, []string{})
		testutil.AssertNoError(t, err, "Subscribe")

		// Cancel context
		cancel()

		// Channel should be closed
		select {
		case _, ok := <-eventsCh:
			if ok {
				t.Error("Expected channel to be closed")
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Channel was not closed after context cancellation")
		}
	})
}
