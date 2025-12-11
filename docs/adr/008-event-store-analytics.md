# ADR 008: Event Store Analytics and Metrics

**Date**: 2025-12-11  
**Status**: Proposed  
**Deciders**: Audun

## Context

We have an event-sourced system storing all state changes as events, but currently have no visibility into:
- Event volume and patterns over time
- Event type distribution
- Event store growth rate
- Projection lag (time between event stored and projected)
- System health from an event sourcing perspective

This makes it difficult to:
- Understand system usage patterns
- Detect anomalies or issues
- Plan capacity
- Debug projection delays
- Monitor event sourcing health

## Decision Options

### Option 1: Simple HTTP Endpoint with In-Memory Stats

**Implementation**: Add `/metrics/events` endpoint that returns basic stats calculated in-memory.

**Pros**:
- Quick to implement (~2 hours)
- No new dependencies
- Lightweight
- Good for development/debugging

**Cons**:
- Stats lost on restart
- No historical data
- Limited to current process
- Manual querying required
- No alerting capability

**Example Response**:
```json
{
  "events_count_total": 1547,
  "events_last_hour": 42,
  "events_by_type": {
    "team.registered": 12,
    "status.updated": 30
  },
  "projection_lag_ms": 45
}
```

---

### Option 2: Prometheus Metrics ✅ **RECOMMENDED**

**Implementation**: Export Prometheus metrics on `/metrics` endpoint using `prometheus/client_golang`.

**Pros**:
- Industry standard format
- Grafana integration (visualization)
- Historical data with retention
- Alerting support
- Rich query language (PromQL)
- Works with Fly.io metrics
- Battle-tested tooling

**Cons**:
- Need Prometheus/Grafana setup (or use Fly.io's built-in)
- Learning curve for PromQL
- Additional dependency (~200KB)

**Metrics to Export**:
```go
// Counters
events_stored_total{event_type="team.registered"}
events_stored_bytes_total
projection_updates_total{projection="teams"}

// Gauges
projection_lag_seconds
event_store_size_bytes

// Histograms
event_processing_duration_seconds
```

**Query Examples**:
- Events per second: `rate(events_stored_total[5m])`
- Event type distribution: `sum by (event_type) (events_stored_total)`
- Projection lag alert: `projection_lag_seconds > 1`

---

### Option 3: Custom Metrics Database Table

**Implementation**: Store metrics in PostgreSQL table, query via API.

**Pros**:
- Full control over data structure
- Easy to query with SQL
- Already have PostgreSQL

**Cons**:
- More code to maintain
- Need to build own dashboard
- Metrics table grows indefinitely (need cleanup)
- Reinventing the wheel
- No standard tooling

**Schema**:
```sql
CREATE TABLE metrics (
    timestamp TIMESTAMP,
    metric_name VARCHAR,
    metric_value FLOAT,
    labels JSONB
);
```

---

## Recommendation

**Option 2: Prometheus Metrics** is the clear choice.

**Rationale**:
1. **Industry standard**: Well-established pattern for metrics
2. **Fly.io integration**: Fly.io has built-in Prometheus support
3. **Grafana dashboards**: Rich visualization out of the box
4. **Alerting**: Can set up alerts on metric thresholds
5. **Low maintenance**: No custom storage or querying logic
6. **Future-proof**: Can add more metrics easily

## Implementation Plan

### Step 1: Add Prometheus Library
```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/promauto
```

### Step 2: Define Metrics (TDD approach)

**Test First** - Create `internal/events/metrics_test.go`:
```go
func TestEventStoreMetrics(t *testing.T) {
    store := setupTestEventStore(t)
    
    // Store events
    store.Store(ctx, teamRegisteredEvent)
    store.Store(ctx, statusUpdatedEvent)
    
    // Verify metrics were recorded
    metrics := getPrometheusMetrics()
    assertEqual(t, metrics["events_stored_total{type=team.registered}"], 1)
    assertEqual(t, metrics["events_stored_total{type=status.updated}"], 1)
}
```

**Then Implement** - Create `internal/events/metrics.go`:
```go
var (
    eventsStoredTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "events_stored_total",
            Help: "Total number of events stored",
        },
        []string{"event_type"},
    )
    
    eventsStoredBytes = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "events_stored_bytes_total",
            Help: "Total bytes of events stored",
        },
    )
    
    projectionLag = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "projection_lag_seconds",
            Help: "Lag between event storage and projection update",
        },
    )
)

func (s *PostgresEventStore) Store(ctx context.Context, event Event) error {
    // ... existing store logic ...
    
    // Record metrics
    eventsStoredTotal.WithLabelValues(event.Type).Inc()
    eventsStoredBytes.Add(float64(len(event.Data)))
    
    return nil
}
```

### Step 3: Expose Metrics Endpoint

**Test First** - Add to existing tests:
```go
func TestMetricsEndpoint(t *testing.T) {
    req := httptest.NewRequest("GET", "/metrics", nil)
    w := httptest.NewRecorder()
    
    metricsHandler.ServeHTTP(w, req)
    
    assertEqual(t, w.Code, http.StatusOK)
    assertContains(t, w.Body.String(), "events_stored_total")
}
```

**Then Implement** - Add to `cmd/backend/main.go`:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

http.Handle("/metrics", promhttp.Handler())
```

### Step 4: Track Projection Lag

**Implementation** in `internal/projections/projector.go`:
```go
func (p *Projector) processEvent(event events.Event) {
    start := time.Now()
    
    // ... projection logic ...
    
    // Calculate and record lag
    lag := time.Since(event.Timestamp)
    projectionLag.Set(lag.Seconds())
}
```

### Step 5: Grafana Dashboard (Optional)

Create dashboard showing:
- Events per second (graph)
- Event type breakdown (pie chart)
- Projection lag (gauge)
- Event store size growth (graph)

## Metrics to Implement

### Phase 1 (Essential)
- `events_stored_total{event_type}` - Counter
- `events_stored_bytes_total` - Counter  
- `projection_lag_seconds` - Gauge

### Phase 2 (Enhanced)
- `projection_updates_total{projection}` - Counter
- `event_processing_duration_seconds` - Histogram
- `projection_errors_total` - Counter

## Testing Strategy

1. **Unit tests**: Verify metrics are recorded correctly
2. **Integration tests**: Check metrics endpoint returns valid Prometheus format
3. **Manual verification**: Use `curl /metrics` to see output
4. **Grafana test**: Import dashboard and verify graphs populate

## Consequences

### Positive
- **Visibility**: Clear view of event sourcing health
- **Debugging**: Can correlate events with system behavior
- **Alerting**: Can alert on projection lag or anomalies
- **Standard format**: Uses industry standard (Prometheus)
- **Grafana integration**: Rich dashboards available

### Negative
- **New dependency**: Adds Prometheus client library
  - *Mitigation*: Small, well-maintained library
- **Learning curve**: Need to understand Prometheus/Grafana
  - *Mitigation*: Excellent documentation available

### Neutral
- **Fly.io integration**: Works with Fly.io's metrics (free tier has limits)

## Decision

**Approved**: Option 2 - Prometheus Metrics

**Scope**:
- Implement all Phase 1 and Phase 2 metrics
- Add metrics to all three services (backend, slackbot, scheduler)
- Set up Grafana dashboard
- Use Fly.io's built-in Prometheus + Grafana integration

**Fly.io Integration**:
Fly.io provides free Prometheus metrics collection and Grafana dashboards:
- Metrics automatically scraped from `/metrics` endpoint
- Access dashboards at `https://fly.io/apps/<app-name>/metrics`
- Can create custom Grafana dashboards
- Built-in retention (30 days on free tier)

**Services to Instrument**:
1. **Backend**: Event store, projections, API requests
2. **Slackbot**: Slack API calls, message processing, command handling
3. **Scheduler**: Reminder deliveries, scheduling health

## Implementation Schedule

Following TDD workflow:
1. Write tests for metrics
2. Verify tests fail
3. Implement metrics instrumentation
4. Verify tests pass
5. Deploy and verify in Fly.io Grafana
6. Create custom dashboard
