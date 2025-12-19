package commands

import (
	"errors"
	"time"

	"github.com/yourusername/status-app/internal/domain"
)

type Command interface {
	Validate() error
}

type SubmitStatusUpdate struct {
	TeamID      domain.TeamID
	ChannelName string
	Content     domain.UpdateContent
	Author      domain.Author
	SlackUser   domain.SlackUserID
	Timestamp   time.Time
}

func (c SubmitStatusUpdate) Validate() error {
	if c.TeamID.IsEmpty() {
		return errors.New("team_id is required")
	}
	if c.Content.String() == "" {
		return errors.New("content is required")
	}
	if c.Author.String() == "" {
		return errors.New("author is required")
	}
	if c.SlackUser.String() == "" {
		return errors.New("slack_user is required")
	}
	if c.ChannelName == "" {
		return errors.New("channel_name is required")
	}
	return nil
}

type RegisterTeam struct {
	Name         domain.TeamName
	SlackChannel domain.SlackChannel
}

func (c RegisterTeam) Validate() error {
	if c.Name.String() == "" {
		return errors.New("name is required")
	}
	if c.SlackChannel.String() == "" {
		return errors.New("slack_channel is required")
	}
	return nil
}

type UpdateTeam struct {
	TeamID       domain.TeamID
	Name         domain.TeamName
	SlackChannel domain.SlackChannel
}

func (c UpdateTeam) Validate() error {
	if c.TeamID.IsEmpty() {
		return errors.New("team_id is required")
	}
	if c.Name.String() == "" {
		return errors.New("name is required")
	}
	if c.SlackChannel.String() == "" {
		return errors.New("slack_channel is required")
	}
	return nil
}
