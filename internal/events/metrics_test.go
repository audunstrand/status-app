package events

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/yourusername/status-app/tests/testutil"
)

func TestEventStoreMetrics_EventsStoredTotal(t *testing.T) {
	// Reset metrics before test
	eventsStoredTotal.Reset()
	
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	ctx := context.Background()
	
	// Store different event types
	event1 := &Event{
		ID:          uuid.New().String(),
		Type:        "team.registered",
		AggregateID: "team-1",
		Data:        json.RawMessage(`{"name":"Engineering"}`),
		Timestamp:   time.Now(),
		Version:     1,
	}
	
	event2 := &Event{
		ID:          uuid.New().String(),
		Type:        "status.updated",
		AggregateID: "team-1",
		Data:        json.RawMessage(`{"content":"Status update"}`),
		Timestamp:   time.Now(),
		Version:     2,
	}
	
	event3 := &Event{
		ID:          uuid.New().String(),
		Type:        "status.updated",
		AggregateID: "team-1",
		Data:        json.RawMessage(`{"content":"Another update"}`),
		Timestamp:   time.Now(),
		Version:     3,
	}
	
	// Store events
	if err := store.Append(ctx, event1); err != nil {
		t.Fatalf("Failed to append event1: %v", err)
	}
	
	if err := store.Append(ctx, event2); err != nil {
		t.Fatalf("Failed to append event2: %v", err)
	}
	
	if err := store.Append(ctx, event3); err != nil {
		t.Fatalf("Failed to append event3: %v", err)
	}
	
	// Verify metrics
	teamRegisteredCount := getCounterValue(t, eventsStoredTotal, "team.registered")
	statusUpdatedCount := getCounterValue(t, eventsStoredTotal, "status.updated")
	
	if teamRegisteredCount != 1 {
		t.Errorf("Expected 1 team.registered event, got %f", teamRegisteredCount)
	}
	
	if statusUpdatedCount != 2 {
		t.Errorf("Expected 2 status.updated events, got %f", statusUpdatedCount)
	}
}

func TestEventStoreMetrics_EventsStoredBytes(t *testing.T) {
	// Reset metrics before test
	eventsStoredBytes.Set(0)
	
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	ctx := context.Background()
	
	data := json.RawMessage(`{"name":"Engineering Team"}`)
	event := &Event{
		ID:          uuid.New().String(),
		Type:        "team.registered",
		AggregateID: "team-1",
		Data:        data,
		Timestamp:   time.Now(),
		Version:     1,
	}
	
	if err := store.Append(ctx, event); err != nil {
		t.Fatalf("Failed to append event: %v", err)
	}
	
	// Verify bytes counter increased
	bytesCount := getGaugeValue(t, eventsStoredBytes)
	expectedBytes := float64(len(data))
	
	if bytesCount < expectedBytes {
		t.Errorf("Expected at least %f bytes stored, got %f", expectedBytes, bytesCount)
	}
}

func TestEventStoreMetrics_EventsLoadedTotal(t *testing.T) {
	// Reset metrics before test
	eventsLoadedTotal.Reset()
	
	store, cleanup := setupTestStore(t)
	defer cleanup()
	
	ctx := context.Background()
	aggregateID := "team-test-load"
	
	// Store some events first
	for i := 1; i <= 3; i++ {
		event := &Event{
			ID:          uuid.New().String(),
			Type:        "status.updated",
			AggregateID: aggregateID,
			Data:        json.RawMessage(`{"content":"update"}`),
			Timestamp:   time.Now(),
			Version:     i,
		}
		if err := store.Append(ctx, event); err != nil {
			t.Fatalf("Failed to append event %d: %v", i, err)
		}
	}
	
	// Load events - should increment metric
	events, err := store.GetByAggregateID(ctx, aggregateID)
	if err != nil {
		t.Fatalf("Failed to load events: %v", err)
	}
	
	if len(events) != 3 {
		t.Errorf("Expected 3 events loaded, got %d", len(events))
	}
	
	// Verify metric
	loadedCount := getCounterValue(t, eventsLoadedTotal, "status.updated")
	if loadedCount != 3 {
		t.Errorf("Expected 3 status.updated events loaded metric, got %f", loadedCount)
	}
}

// Helper function to get counter value with labels
func getCounterValue(t *testing.T, counter *prometheus.CounterVec, label string) float64 {
	t.Helper()
	
	metric := &dto.Metric{}
	c, err := counter.GetMetricWithLabelValues(label)
	if err != nil {
		t.Fatalf("Failed to get counter metric: %v", err)
	}
	
	if err := c.Write(metric); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}
	
	return metric.Counter.GetValue()
}

// Helper function to get gauge value
func getGaugeValue(t *testing.T, gauge prometheus.Gauge) float64 {
	t.Helper()
	
	metric := &dto.Metric{}
	if err := gauge.Write(metric); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}
	
	return metric.Gauge.GetValue()
}

// setupTestStore creates a test event store with test database
func setupTestStore(t *testing.T) (*PostgresStore, func()) {
	t.Helper()
	
	testDB := testutil.SetupTestDB(t)
	
	store, err := NewPostgresStore(testDB.ConnectionString())
	if err != nil {
		t.Fatalf("Failed to create test store: %v", err)
	}
	
	cleanup := func() {
		store.db.Close()
		testDB.Cleanup()
	}
	
	return store, cleanup
}
