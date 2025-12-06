package commands

import "time"

// Command represents a command in the system
type Command interface {
	CommandType() string
}

// SubmitStatusUpdate command
type SubmitStatusUpdate struct {
	TeamID    string
	Content   string
	Author    string
	SlackUser string
	Timestamp time.Time
}

func (c SubmitStatusUpdate) CommandType() string {
	return "SubmitStatusUpdate"
}

// RegisterTeam command
type RegisterTeam struct {
	Name         string
	SlackChannel string
	PollSchedule string // cron format, e.g., "0 9 * * MON"
}

func (c RegisterTeam) CommandType() string {
	return "RegisterTeam"
}

// UpdateTeam command
type UpdateTeam struct {
	TeamID       string
	Name         string
	SlackChannel string
	PollSchedule string
}

func (c UpdateTeam) CommandType() string {
	return "UpdateTeam"
}

// SchedulePoll command
type SchedulePoll struct {
	TeamID    string
	DueDate   time.Time
	Frequency string
}

func (c SchedulePoll) CommandType() string {
	return "SchedulePoll"
}

// SendReminder command
type SendReminder struct {
	TeamID       string
	SlackChannel string
}

func (c SendReminder) CommandType() string {
	return "SendReminder"
}
