package events

import (
	"encoding/json"
	"time"
)

// Event represents a domain event in the system
type Event struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	AggregateID string          `json:"aggregate_id"`
	Data        json.RawMessage `json:"data"`
	Timestamp   time.Time       `json:"timestamp"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Version     int             `json:"version"`
}

// Event Types
const (
	StatusUpdateSubmitted = "status_update.submitted"
	TeamRegistered        = "team.registered"
	PollScheduled         = "poll.scheduled"
	ReminderSent          = "reminder.sent"
	TeamUpdated           = "team.updated"
)

// StatusUpdateSubmittedData represents the data for a status update submission
type StatusUpdateSubmittedData struct {
	UpdateID  string    `json:"update_id"`
	TeamID    string    `json:"team_id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	SlackUser string    `json:"slack_user"`
	Timestamp time.Time `json:"timestamp"`
}

// TeamRegisteredData represents the data for team registration
type TeamRegisteredData struct {
	TeamID       string `json:"team_id"`
	Name         string `json:"name"`
	SlackChannel string `json:"slack_channel"`
	PollSchedule string `json:"poll_schedule"` // cron format
}

// PollScheduledData represents scheduled poll information
type PollScheduledData struct {
	PollID    string    `json:"poll_id"`
	TeamID    string    `json:"team_id"`
	DueDate   time.Time `json:"due_date"`
	Frequency string    `json:"frequency"`
}

// ReminderSentData represents a reminder sent to a team
type ReminderSentData struct {
	ReminderID   string    `json:"reminder_id"`
	TeamID       string    `json:"team_id"`
	SlackChannel string    `json:"slack_channel"`
	SentAt       time.Time `json:"sent_at"`
}
