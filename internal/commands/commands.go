package commands

import (
	"errors"
	"time"
)

// Command represents a command in the system
type Command interface {
	Validate() error
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

func (c SubmitStatusUpdate) Validate() error {
	if c.TeamID == "" {
		return errors.New("team_id is required")
	}
	if c.Content == "" {
		return errors.New("content is required")
	}
	if c.Author == "" {
		return errors.New("author is required")
	}
	if c.SlackUser == "" {
		return errors.New("slack_user is required")
	}
	return nil
}

// RegisterTeam command
type RegisterTeam struct {
	Name         string
	SlackChannel string
}

func (c RegisterTeam) Validate() error {
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.SlackChannel == "" {
		return errors.New("slack_channel is required")
	}
	return nil
}

// UpdateTeam command
type UpdateTeam struct {
	TeamID       string
	Name         string
	SlackChannel string
}

func (c UpdateTeam) Validate() error {
	if c.TeamID == "" {
		return errors.New("team_id is required")
	}
	if c.Name == "" {
		return errors.New("name is required")
	}
	if c.SlackChannel == "" {
		return errors.New("slack_channel is required")
	}
	return nil
}
