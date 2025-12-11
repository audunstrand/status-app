package events

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Event store metrics
	eventsStoredTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "events",
			Name:      "stored_total",
			Help:      "Total number of events stored by type",
		},
		[]string{"event_type"},
	)

	eventsStoredBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "status_app",
			Subsystem: "events",
			Name:      "stored_bytes_total",
			Help:      "Total bytes of event data stored",
		},
	)

	eventsLoadedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "events",
			Name:      "loaded_total",
			Help:      "Total number of events loaded by type",
		},
		[]string{"event_type"},
	)

	eventStoreErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "events",
			Name:      "errors_total",
			Help:      "Total number of event store errors by operation",
		},
		[]string{"operation"},
	)
)
