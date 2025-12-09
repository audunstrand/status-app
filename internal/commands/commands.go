package commands

import "time"

// Command represents a command in the system
type Command interface {
}

// SubmitStatusUpdate command
type SubmitStatusUpdate struct {
	TeamID      string
	ChannelName string
	Content     string
	Author      string
	SlackUser   string
	Timestamp   time.Time
}

// RegisterTeam command
type RegisterTeam struct {
	Name         string
	SlackChannel string
	PollSchedule string // cron format, e.g., "0 9 * * MON"
}

// UpdateTeam command
type UpdateTeam struct {
	TeamID       string
	Name         string
	SlackChannel string
	PollSchedule string
}

// SchedulePoll command
type SchedulePoll struct {
	TeamID    string
	DueDate   time.Time
	Frequency string
}

// SendReminder command
type SendReminder struct {
	TeamID       string
	SlackChannel string
}
