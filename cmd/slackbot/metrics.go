package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Slack message metrics
	slackMessagesReceivedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "messages_received_total",
			Help:      "Total number of Slack messages received by type",
		},
		[]string{"type"}, // mention, direct_message, slash_command
	)

	slackMessagesSentTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "messages_sent_total",
			Help:      "Total number of Slack messages sent",
		},
	)

	slackAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "api_calls_total",
			Help:      "Total number of Slack API calls by endpoint",
		},
		[]string{"endpoint", "status"}, // endpoint: post_message, open_view, etc; status: success, error
	)

	slackCommandsHandledTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "commands_handled_total",
			Help:      "Total number of slash commands handled",
		},
		[]string{"command"}, // /set-team-name, /updates
	)

	backendAPICallsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "backend_api_calls_total",
			Help:      "Total number of backend API calls",
		},
		[]string{"endpoint", "status"}, // endpoint: submit_update, etc; status: success, error
	)

	slackbotErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "status_app",
			Subsystem: "slackbot",
			Name:      "errors_total",
			Help:      "Total number of slackbot errors by type",
		},
		[]string{"error_type"}, // api_error, parse_error, backend_error
	)
)
