package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Scheduler metrics
	remindersScheduledTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "scheduler",
			Name:      "reminders_scheduled_total",
			Help:      "Total number of reminders scheduled",
		},
	)

	remindersSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "scheduler",
			Name:      "reminders_sent_total",
			Help:      "Total number of reminders sent",
		},
		[]string{"status"}, // success, error
	)

	teamsReminderCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "status_app",
			Subsystem: "scheduler",
			Name:      "teams_reminder_count",
			Help:      "Number of teams that received reminders in last run",
		},
	)

	schedulerErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "scheduler",
			Name:      "errors_total",
			Help:      "Total number of scheduler errors by type",
		},
		[]string{"error_type"}, // db_error, slack_error
	)
)
