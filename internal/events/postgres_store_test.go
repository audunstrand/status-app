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

func TestPostgresStore_Subscribe_Integration(t *testing.T) {
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

testDB := testutil.SetupTestDB(t)
defer testDB.Cleanup()

store, err := NewPostgresStore(testDB.ConnectionString())
if err != nil {
t.Fatalf("Failed to create store: %v", err)
}
defer store.Close()

// Subscribe to events
eventsCh, err := store.Subscribe(ctx, []string{})
if err != nil {
t.Fatalf("Failed to subscribe: %v", err)
}

// Give subscriber time to connect
time.Sleep(100 * time.Millisecond)

// Append a new event
event := newTeamRegisteredEvent(t, "team-subscribe-test", "Engineering", "#engineering", "", time.Now())

err = store.Append(ctx, event)
if err != nil {
t.Fatalf("Failed to append event: %v", err)
}

// Wait for notification
select {
case receivedEvent := <-eventsCh:
if receivedEvent.ID != event.ID {
t.Errorf("Expected event ID '%s', got '%s'", event.ID, receivedEvent.ID)
}
if receivedEvent.Type != TeamRegistered {
t.Errorf("Expected event type '%s', got '%s'", TeamRegistered, receivedEvent.Type)
}
case <-time.After(2 * time.Second):
t.Fatal("Timeout waiting for event notification")
}
}

func TestPostgresStore_Subscribe_FilterByEventType(t *testing.T) {
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

testDB := testutil.SetupTestDB(t)
defer testDB.Cleanup()

store, err := NewPostgresStore(testDB.ConnectionString())
if err != nil {
t.Fatalf("Failed to create store: %v", err)
}
defer store.Close()

// Subscribe only to TeamRegistered events
eventsCh, err := store.Subscribe(ctx, []string{TeamRegistered})
if err != nil {
t.Fatalf("Failed to subscribe: %v", err)
}

time.Sleep(100 * time.Millisecond)

// Append a StatusUpdateSubmitted event (should be filtered out)
statusEvent := newStatusUpdateEvent(t, "team-filter-test", "Test update", "Alice", "alice", time.Now())
err = store.Append(ctx, statusEvent)
if err != nil {
t.Fatalf("Failed to append status event: %v", err)
}

// Append a TeamRegistered event (should be received)
teamEvent := newTeamRegisteredEvent(t, "team-filter-test", "Product", "#product", "", time.Now())
err = store.Append(ctx, teamEvent)
if err != nil {
t.Fatalf("Failed to append team event: %v", err)
}

// Should only receive the TeamRegistered event
select {
case receivedEvent := <-eventsCh:
if receivedEvent.Type != TeamRegistered {
t.Errorf("Expected TeamRegistered event, got '%s'", receivedEvent.Type)
}
if receivedEvent.ID != teamEvent.ID {
t.Errorf("Expected event ID '%s', got '%s'", teamEvent.ID, receivedEvent.ID)
}
case <-time.After(2 * time.Second):
t.Fatal("Timeout waiting for TeamRegistered event")
}

// Ensure no other events are received
select {
case unexpectedEvent := <-eventsCh:
t.Errorf("Received unexpected event: %s", unexpectedEvent.ID)
case <-time.After(500 * time.Millisecond):
// Good - no additional events
}
}
