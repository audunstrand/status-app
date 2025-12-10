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
	// Validate command
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("invalid command: %w", err)
	}

	switch c := cmd.(type) {
	case SubmitStatusUpdate:
		return h.handleSubmitStatusUpdate(ctx, c)
	case RegisterTeam:
		return h.handleRegisterTeam(ctx, c)
	case UpdateTeam:
		return h.handleUpdateTeam(ctx, c)
	default:
		return fmt.Errorf("unknown command type: %T", cmd)
	}
}

// createAndAppendEvent is a helper that marshals data and creates an event
func (h *Handler) createAndAppendEvent(
	ctx context.Context,
	eventType string,
	aggregateID string,
	data interface{},
) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	event := &events.Event{
		ID:          uuid.New().String(),
		Type:        eventType,
		AggregateID: aggregateID,
		Data:        dataJSON,
		Timestamp:   time.Now(),
		Version:     1,
	}

	return h.eventStore.Append(ctx, event)
}

func (h *Handler) handleSubmitStatusUpdate(ctx context.Context, cmd SubmitStatusUpdate) error {
	// Check if team exists by looking for any events with this aggregate ID
	existingEvents, err := h.eventStore.GetByAggregateID(ctx, cmd.TeamID)
	if err != nil {
		return fmt.Errorf("failed to check for existing team: %w", err)
	}

	// If no events exist for this team, auto-register it
	if len(existingEvents) == 0 {
		teamName := cmd.ChannelName
		if teamName == "" {
			teamName = cmd.TeamID
		}

		registerData := events.TeamRegisteredData{
			TeamID:       cmd.TeamID,
			Name:         teamName,
			SlackChannel: cmd.TeamID,
		}

		if err := h.createAndAppendEvent(ctx, events.TeamRegistered, cmd.TeamID, registerData); err != nil {
			return fmt.Errorf("failed to auto-register team: %w", err)
		}
	}

	// Submit the status update
	data := events.StatusUpdateSubmittedData{
		UpdateID:  uuid.New().String(),
		TeamID:    cmd.TeamID,
		Content:   cmd.Content,
		Author:    cmd.Author,
		SlackUser: cmd.SlackUser,
		Timestamp: cmd.Timestamp,
	}

	return h.createAndAppendEvent(ctx, events.StatusUpdateSubmitted, cmd.TeamID, data)
}

func (h *Handler) handleRegisterTeam(ctx context.Context, cmd RegisterTeam) error {
	teamID := uuid.New().String()

	data := events.TeamRegisteredData{
		TeamID:       teamID,
		Name:         cmd.Name,
		SlackChannel: cmd.SlackChannel,
	}

	return h.createAndAppendEvent(ctx, events.TeamRegistered, teamID, data)
}

func (h *Handler) handleUpdateTeam(ctx context.Context, cmd UpdateTeam) error {
	data := events.TeamUpdatedData{
		TeamID:       cmd.TeamID,
		Name:         cmd.Name,
		SlackChannel: cmd.SlackChannel,
	}

	return h.createAndAppendEvent(ctx, events.TeamUpdated, cmd.TeamID, data)
}
