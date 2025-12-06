package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/status-app/internal/events"
)

// Handler processes commands and emits events
type Handler struct {
	eventStore events.Store
}

func NewHandler(eventStore events.Store) *Handler {
	return &Handler{
		eventStore: eventStore,
	}
}

func (h *Handler) Handle(ctx context.Context, cmd Command) error {
	switch c := cmd.(type) {
	case SubmitStatusUpdate:
		return h.handleSubmitStatusUpdate(ctx, c)
	case RegisterTeam:
		return h.handleRegisterTeam(ctx, c)
	case UpdateTeam:
		return h.handleUpdateTeam(ctx, c)
	case SchedulePoll:
		return h.handleSchedulePoll(ctx, c)
	case SendReminder:
		return h.handleSendReminder(ctx, c)
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

func (h *Handler) handleSubmitStatusUpdate(ctx context.Context, cmd SubmitStatusUpdate) error {
	updateID := uuid.New().String()

	data := events.StatusUpdateSubmittedData{
		UpdateID:  updateID,
		TeamID:    cmd.TeamID,
		Content:   cmd.Content,
		Author:    cmd.Author,
		SlackUser: cmd.SlackUser,
		Timestamp: cmd.Timestamp,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        events.StatusUpdateSubmitted,
		AggregateID: cmd.TeamID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}

func (h *Handler) handleRegisterTeam(ctx context.Context, cmd RegisterTeam) error {
	teamID := uuid.New().String()

	data := events.TeamRegisteredData{
		TeamID:       teamID,
		Name:         cmd.Name,
		SlackChannel: cmd.SlackChannel,
		PollSchedule: cmd.PollSchedule,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        events.TeamRegistered,
		AggregateID: teamID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}

func (h *Handler) handleUpdateTeam(ctx context.Context, cmd UpdateTeam) error {
	data := events.TeamRegisteredData{
		TeamID:       cmd.TeamID,
		Name:         cmd.Name,
		SlackChannel: cmd.SlackChannel,
		PollSchedule: cmd.PollSchedule,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        events.TeamUpdated,
		AggregateID: cmd.TeamID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}

func (h *Handler) handleSchedulePoll(ctx context.Context, cmd SchedulePoll) error {
	pollID := uuid.New().String()

	data := events.PollScheduledData{
		PollID:    pollID,
		TeamID:    cmd.TeamID,
		DueDate:   cmd.DueDate,
		Frequency: cmd.Frequency,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        events.PollScheduled,
		AggregateID: cmd.TeamID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}

func (h *Handler) handleSendReminder(ctx context.Context, cmd SendReminder) error {
	reminderID := uuid.New().String()

	data := events.ReminderSentData{
		ReminderID:   reminderID,
		TeamID:       cmd.TeamID,
		SlackChannel: cmd.SlackChannel,
		SentAt:       time.Now(),
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        events.ReminderSent,
		AggregateID: cmd.TeamID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}
