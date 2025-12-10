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
}

// TeamUpdatedData represents the data for team updates
type TeamUpdatedData struct {
	TeamID       string `json:"team_id"`
	Name         string `json:"name"`
	SlackChannel string `json:"slack_channel"`
}
