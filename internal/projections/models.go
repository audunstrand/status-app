package projections

import "time"

// Team represents the current state of a team
type Team struct {
	TeamID       string    `json:"team_id"`
	Name         string    `json:"name"`
	SlackChannel string    `json:"slack_channel"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// StatusUpdate represents a status update in the read model
type StatusUpdate struct {
	UpdateID  string    `json:"update_id"`
	TeamID    string    `json:"team_id"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	SlackUser string    `json:"slack_user"`
	CreatedAt time.Time `json:"created_at"`
}

// TeamSummary provides aggregate information about a team
type TeamSummary struct {
	Team               Team      `json:"team"`
	TotalUpdates       int       `json:"total_updates"`
	LastUpdateAt       time.Time `json:"last_update_at"`
	UniqueContributors int       `json:"unique_contributors"`
}
