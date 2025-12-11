package projections

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Projection update metrics
	projectionUpdatesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "projections",
			Name:      "updates_total",
			Help:      "Total number of projection updates by projection type",
		},
		[]string{"projection"},
	)

	projectionLagSeconds = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "status_app",
			Subsystem: "projections",
			Name:      "lag_seconds",
			Help:      "Lag between event storage and projection update in seconds",
		},
		[]string{"projection"},
	)

	projectionErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "projections",
			Name:      "errors_total",
			Help:      "Total number of projection errors by projection type",
		},
		[]string{"projection"},
	)

	projectionProcessingDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "status_app",
			Subsystem: "projections",
			Name:      "processing_duration_seconds",
			Help:      "Time taken to process and apply a projection update",
			Buckets:   prometheus.DefBuckets, // Default: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"projection"},
	)
)

// recordProjectionUpdate records metrics for a successful projection update
func recordProjectionUpdate(projection string, eventTime time.Time, duration time.Duration) {
	projectionUpdatesTotal.WithLabelValues(projection).Inc()
	
	// Calculate lag from event timestamp to now
	lag := time.Since(eventTime)
	projectionLagSeconds.WithLabelValues(projection).Set(lag.Seconds())
	
	// Record processing duration
	projectionProcessingDuration.WithLabelValues(projection).Observe(duration.Seconds())
}

// recordProjectionError records a projection error
func recordProjectionError(projection string) {
	projectionErrorsTotal.WithLabelValues(projection).Inc()
}
