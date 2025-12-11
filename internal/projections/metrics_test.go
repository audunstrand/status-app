package projections

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestProjectionMetrics_UpdatesTotal(t *testing.T) {
	// Reset metrics
	projectionUpdatesTotal.Reset()
	
	// Record some projection updates
	recordProjectionUpdate("teams", time.Now().Add(-1*time.Second), 50*time.Millisecond)
	recordProjectionUpdate("teams", time.Now().Add(-2*time.Second), 45*time.Millisecond)
	recordProjectionUpdate("status_updates", time.Now().Add(-1*time.Second), 30*time.Millisecond)
	
	// Verify metrics
	teamsCount := getCounterValue(t, projectionUpdatesTotal, "teams")
	statusUpdatesCount := getCounterValue(t, projectionUpdatesTotal, "status_updates")
	
	if teamsCount != 2 {
		t.Errorf("Expected 2 teams projection updates, got %f", teamsCount)
	}
	
	if statusUpdatesCount != 1 {
		t.Errorf("Expected 1 status_updates projection update, got %f", statusUpdatesCount)
	}
}

func TestProjectionMetrics_LagSeconds(t *testing.T) {
	// Reset metrics
	projectionLagSeconds.Reset()
	
	// Record projection update with specific event time
	eventTime := time.Now().Add(-2 * time.Second)
	recordProjectionUpdate("teams", eventTime, 50*time.Millisecond)
	
	// Verify lag is recorded (should be approximately 2 seconds)
	lag := getGaugeValue(t, projectionLagSeconds, "teams")
	
	if lag < 1.9 || lag > 2.2 {
		t.Errorf("Expected lag around 2 seconds, got %f", lag)
	}
}

func TestProjectionMetrics_ProcessingDuration(t *testing.T) {
	// Reset metrics
	projectionProcessingDuration.Reset()
	
	// Record projection updates with different durations
	recordProjectionUpdate("teams", time.Now(), 50*time.Millisecond)
	recordProjectionUpdate("teams", time.Now(), 100*time.Millisecond)
	recordProjectionUpdate("teams", time.Now(), 75*time.Millisecond)
	
	// Verify histogram has recorded values
	histogram := getHistogramValue(t, projectionProcessingDuration, "teams")
	
	if *histogram.SampleCount != 3 {
		t.Errorf("Expected 3 samples, got %d", *histogram.SampleCount)
	}
	
	// Average should be around 75ms = 0.075s
	expectedAvg := 0.075
	actualAvg := *histogram.SampleSum / float64(*histogram.SampleCount)
	
	if actualAvg < expectedAvg-0.02 || actualAvg > expectedAvg+0.02 {
		t.Errorf("Expected average around %f seconds, got %f", expectedAvg, actualAvg)
	}
}

func TestProjectionMetrics_ErrorsTotal(t *testing.T) {
	// Reset metrics
	projectionErrorsTotal.Reset()
	
	// Record some errors
	recordProjectionError("teams")
	recordProjectionError("teams")
	recordProjectionError("status_updates")
	
	// Verify metrics
	teamsErrors := getCounterValue(t, projectionErrorsTotal, "teams")
	statusUpdatesErrors := getCounterValue(t, projectionErrorsTotal, "status_updates")
	
	if teamsErrors != 2 {
		t.Errorf("Expected 2 teams projection errors, got %f", teamsErrors)
	}
	
	if statusUpdatesErrors != 1 {
		t.Errorf("Expected 1 status_updates projection error, got %f", statusUpdatesErrors)
	}
}

// Helper functions

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

func getGaugeValue(t *testing.T, gauge *prometheus.GaugeVec, label string) float64 {
	t.Helper()
	
	metric := &dto.Metric{}
	g, err := gauge.GetMetricWithLabelValues(label)
	if err != nil {
		t.Fatalf("Failed to get gauge metric: %v", err)
	}
	
	if err := g.Write(metric); err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}
	
	return metric.Gauge.GetValue()
}

func getHistogramValue(t *testing.T, histogram *prometheus.HistogramVec, label string) *dto.Histogram {
	t.Helper()
	
	metric := &dto.Metric{}
	observer, err := histogram.GetMetricWithLabelValues(label)
	if err != nil {
		t.Fatalf("Failed to get histogram metric: %v", err)
	}
	
	// Need to cast to prometheus.Metric to use Write
	if m, ok := observer.(prometheus.Metric); ok {
		if err := m.Write(metric); err != nil {
			t.Fatalf("Failed to write metric: %v", err)
		}
	} else {
		t.Fatalf("Failed to cast observer to Metric")
	}
	
	return metric.Histogram
}
