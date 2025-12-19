package domain

import (
	"errors"
	"fmt"
	"strings"
)

type TeamID struct {
	value string
}

func NewTeamID(s string) (TeamID, error) {
	if s == "" {
		return TeamID{}, errors.New("team ID cannot be empty")
	}
	if len(s) > 100 {
		return TeamID{}, errors.New("team ID must be 100 characters or less")
	}
	return TeamID{value: s}, nil
}

func (t TeamID) String() string {
	return t.value
}

func (t TeamID) IsEmpty() bool {
	return t.value == ""
}

type TeamName struct {
	value string
}

func NewTeamName(s string) (TeamName, error) {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return TeamName{}, errors.New("team name cannot be empty")
	}
	if len(trimmed) > 100 {
		return TeamName{}, errors.New("team name must be 100 characters or less")
	}
	return TeamName{value: trimmed}, nil
}

func (t TeamName) String() string {
	return t.value
}

type SlackChannel struct {
	value string
}

func NewSlackChannel(s string) (SlackChannel, error) {
	if s == "" {
		return SlackChannel{}, errors.New("slack channel cannot be empty")
	}
	return SlackChannel{value: s}, nil
}

func (s SlackChannel) String() string {
	return s.value
}

type UpdateContent struct {
	value string
}

func NewUpdateContent(s string) (UpdateContent, error) {
	if s == "" {
		return UpdateContent{}, errors.New("update content cannot be empty")
	}
	if len(s) > 500 {
		return UpdateContent{}, errors.New("update content must be 500 characters or less")
	}
	return UpdateContent{value: s}, nil
}

func (u UpdateContent) String() string {
	return u.value
}

type Author struct {
	value string
}

func NewAuthor(s string) (Author, error) {
	if s == "" {
		return Author{}, errors.New("author cannot be empty")
	}
	return Author{value: s}, nil
}

func (a Author) String() string {
	return a.value
}

type SlackUserID struct {
	value string
}

func NewSlackUserID(s string) (SlackUserID, error) {
	if s == "" {
		return SlackUserID{}, errors.New("slack user ID cannot be empty")
	}
	return SlackUserID{value: s}, nil
}

func (s SlackUserID) String() string {
	return s.value
}

type UpdateID struct {
	value string
}

func NewUpdateID(s string) (UpdateID, error) {
	if s == "" {
		return UpdateID{}, errors.New("update ID cannot be empty")
	}
	return UpdateID{value: s}, nil
}

func (u UpdateID) String() string {
	return u.value
}

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}
