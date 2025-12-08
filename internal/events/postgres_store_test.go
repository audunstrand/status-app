package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/yourusername/status-app/tests/testutil"
)

func TestPostgresStore_Append(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	store, err := NewPostgresStore(testDB.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	t.Run("appends event successfully", func(t *testing.T) {
		data, _ := json.Marshal(TeamRegisteredData{
			TeamID:       "team-1",
			Name:         "Engineering",
			SlackChannel: "#engineering",
			PollSchedule: "weekly",
		})

		event := &Event{
			ID:          "evt-1",
			Type:        TeamRegistered,
			AggregateID: "team-1",
			Data:        data,
			Timestamp:   time.Now(),
			Version:     1,
		}

		err := store.Append(ctx, event)
		if err != nil {
			t.Errorf("Append() error = %v", err)
		}
	})

	t.Run("retrieves event by aggregate ID", func(t *testing.T) {
		// Append event
		data, _ := json.Marshal(TeamRegisteredData{
			TeamID:       "team-2",
			Name:         "Product",
			SlackChannel: "#product",
			PollSchedule: "daily",
		})

		event := &Event{
			ID:          "evt-2",
			Type:        TeamRegistered,
			AggregateID: "team-2",
			Data:        data,
			Timestamp:   time.Now(),
			Version:     1,
		}

		err := store.Append(ctx, event)
		if err != nil {
			t.Fatalf("Append() error = %v", err)
		}

		// Retrieve by aggregate ID
		events, err := store.GetByAggregateID(ctx, "team-2")
		if err != nil {
			t.Errorf("GetByAggregateID() error = %v", err)
		}

		if len(events) == 0 {
			t.Error("GetByAggregateID() returned no events")
		}

		found := false
		for _, e := range events {
			if e.ID == "evt-2" {
				found = true
				if e.Type != TeamRegistered {
					t.Errorf("Event type = %v, want %v", e.Type, TeamRegistered)
				}
				if e.AggregateID != "team-2" {
					t.Errorf("AggregateID = %v, want team-2", e.AggregateID)
				}
			}
		}

		if !found {
			t.Error("Event evt-2 not found in results")
		}
	})

	t.Run("retrieves all events with pagination", func(t *testing.T) {
		// Append multiple events
		for i := 0; i < 5; i++ {
			data, _ := json.Marshal(StatusUpdateSubmittedData{
				UpdateID:  "update-" + string(rune('A'+i)),
				TeamID:    "team-3",
				Content:   "Update " + string(rune('A'+i)),
				Author:    "Author",
				SlackUser: "U123",
				Timestamp: time.Now(),
			})

			event := &Event{
				ID:          "evt-3-" + string(rune('A'+i)),
				Type:        StatusUpdateSubmitted,
				AggregateID: "team-3",
				Data:        data,
				Timestamp:   time.Now(),
				Version:     i + 1,
			}

			err := store.Append(ctx, event)
			if err != nil {
				t.Fatalf("Append() error = %v", err)
			}
		}

		// Get all events
		events, err := store.GetAll(ctx, "", 0, 10)
		if err != nil {
			t.Errorf("GetAll() error = %v", err)
		}

		if len(events) < 5 {
			t.Errorf("GetAll() returned %d events, want at least 5", len(events))
		}
	})

	t.Run("filters events by type", func(t *testing.T) {
		// Get only TeamRegistered events
		events, err := store.GetAll(ctx, TeamRegistered, 0, 100)
		if err != nil {
			t.Errorf("GetAll() with filter error = %v", err)
		}

		for _, e := range events {
			if e.Type != TeamRegistered {
				t.Errorf("Event type = %v, want %v", e.Type, TeamRegistered)
			}
		}
	})
}

func TestPostgresStore_GetByAggregateID_Empty(t *testing.T) {
	ctx := context.Background()
	testDB := testutil.SetupTestDB(t)
	defer testDB.Cleanup()

	store, err := NewPostgresStore(testDB.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	events, err := store.GetByAggregateID(ctx, "non-existent")
	if err != nil {
		t.Errorf("GetByAggregateID() error = %v", err)
	}

	if len(events) != 0 {
		t.Errorf("GetByAggregateID() returned %d events, want 0", len(events))
	}
}

func TestEvent_MarshalUnmarshal(t *testing.T) {
	t.Run("marshals and unmarshals event data", func(t *testing.T) {
		teamData := TeamRegisteredData{
			TeamID:       "team-test",
			Name:         "Test Team",
			SlackChannel: "#test",
			PollSchedule: "daily",
		}

		data, err := json.Marshal(teamData)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var unmarshaled TeamRegisteredData
		err = json.Unmarshal(data, &unmarshaled)
		if err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if unmarshaled.TeamID != teamData.TeamID {
			t.Errorf("TeamID = %v, want %v", unmarshaled.TeamID, teamData.TeamID)
		}
		if unmarshaled.Name != teamData.Name {
			t.Errorf("Name = %v, want %v", unmarshaled.Name, teamData.Name)
		}
	})
}
