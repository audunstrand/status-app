package domain

import (
	"errors"
	"time"
)

type Team struct {
	id           TeamID
	name         TeamName
	slackChannel SlackChannel
	registered   bool
}

func NewTeam(id TeamID, name TeamName, slackChannel SlackChannel) (*Team, error) {
	if id.IsEmpty() {
		return nil, errors.New("team ID is required")
	}
	if name.String() == "" {
		return nil, errors.New("team name is required")
	}
	if slackChannel.String() == "" {
		return nil, errors.New("slack channel is required")
	}

	return &Team{
		id:           id,
		name:         name,
		slackChannel: slackChannel,
		registered:   true,
	}, nil
}

func (t *Team) ID() TeamID {
	return t.id
}

func (t *Team) Name() TeamName {
	return t.name
}

func (t *Team) SlackChannel() SlackChannel {
	return t.slackChannel
}

func (t *Team) IsRegistered() bool {
	return t.registered
}

func (t *Team) Rename(newName TeamName) error {
	if newName.String() == "" {
		return errors.New("new team name cannot be empty")
	}
	if newName.String() == t.name.String() {
		return errors.New("new team name is the same as current name")
	}
	t.name = newName
	return nil
}

type Update struct {
	id        UpdateID
	teamID    TeamID
	content   UpdateContent
	author    Author
	slackUser SlackUserID
	timestamp time.Time
}

func NewUpdate(id UpdateID, teamID TeamID, content UpdateContent, author Author, slackUser SlackUserID, timestamp time.Time) (*Update, error) {
	if id.String() == "" {
		return nil, errors.New("update ID is required")
	}
	if teamID.IsEmpty() {
		return nil, errors.New("team ID is required")
	}
	if content.String() == "" {
		return nil, errors.New("content is required")
	}
	if author.String() == "" {
		return nil, errors.New("author is required")
	}
	if slackUser.String() == "" {
		return nil, errors.New("slack user is required")
	}
	if timestamp.IsZero() {
		return nil, errors.New("timestamp is required")
	}

	return &Update{
		id:        id,
		teamID:    teamID,
		content:   content,
		author:    author,
		slackUser: slackUser,
		timestamp: timestamp,
	}, nil
}

func (u *Update) ID() UpdateID {
	return u.id
}

func (u *Update) TeamID() TeamID {
	return u.teamID
}

func (u *Update) Content() UpdateContent {
	return u.content
}

func (u *Update) Author() Author {
	return u.author
}

func (u *Update) SlackUser() SlackUserID {
	return u.slackUser
}

func (u *Update) Timestamp() time.Time {
	return u.timestamp
}
