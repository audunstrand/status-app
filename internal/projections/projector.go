package projections

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/yourusername/status-app/internal/events"
)

const (
	// maxEventsPerRebuild is the maximum number of events to load during projection rebuild
	// In production, this should be replaced with pagination for unlimited event handling
	maxEventsPerRebuild = 10000
)

// Projector builds read models from events
type Projector struct {
	eventStore events.Store
	db         *sql.DB
}

func NewProjector(eventStore events.Store, db *sql.DB) *Projector {
	return &Projector{
		eventStore: eventStore,
		db:         db,
	}
}

// Start begins processing events and building projections
func (p *Projector) Start(ctx context.Context) error {
	// Initial projection rebuild from all events
	if err := p.rebuildProjections(ctx); err != nil {
		return fmt.Errorf("failed to rebuild projections: %w", err)
	}

	// Subscribe to new events
	eventsCh, err := p.eventStore.Subscribe(ctx, []string{})
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	// Process new events as they arrive
	go func() {
		for {
			select {
			case event, ok := <-eventsCh:
				if !ok {
					log.Println("event channel closed, stopping projection subscription")
					return
				}
				if err := p.processEvent(ctx, event); err != nil {
					log.Printf("failed to process event %s: %v", event.ID, err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (p *Projector) rebuildProjections(ctx context.Context) error {
	// Get all events
	allEvents, err := p.eventStore.GetAll(ctx, "", 0, maxEventsPerRebuild)
	if err != nil {
		return fmt.Errorf("failed to get events: %w", err)
	}

	// Process each event to rebuild projections
	for _, event := range allEvents {
		if err := p.processEvent(ctx, event); err != nil {
			log.Printf("warning: failed to process event %s during rebuild: %v", event.ID, err)
		}
	}

	return nil
}

func (p *Projector) processEvent(ctx context.Context, event *events.Event) error {
	switch event.Type {
	case events.StatusUpdateSubmitted:
		return p.handleStatusUpdateSubmitted(ctx, event)
	case events.TeamRegistered:
		return p.handleTeamRegistered(ctx, event)
	case events.TeamUpdated:
		return p.handleTeamUpdated(ctx, event)
	case events.ReminderSent:
		return p.handleReminderSent(ctx, event)
	default:
		// Unknown event type, skip
		return nil
	}
}

func (p *Projector) handleStatusUpdateSubmitted(ctx context.Context, event *events.Event) error {
	var data events.StatusUpdateSubmittedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	query := `
		INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (update_id) DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query,
		data.UpdateID,
		data.TeamID,
		data.Content,
		data.Author,
		data.SlackUser,
		data.Timestamp,
	)

	return err
}

func (p *Projector) handleTeamRegistered(ctx context.Context, event *events.Event) error {
	var data events.TeamRegisteredData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	query := `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $5)
		ON CONFLICT (team_id) DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query,
		data.TeamID,
		data.Name,
		data.SlackChannel,
		data.PollSchedule,
		event.Timestamp,
	)

	return err
}

func (p *Projector) handleTeamUpdated(ctx context.Context, event *events.Event) error {
	var data events.TeamRegisteredData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return fmt.Errorf("failed to unmarshal event data: %w", err)
	}

	query := `
		UPDATE teams
		SET name = $2, slack_channel = $3, poll_schedule = $4, updated_at = $5
		WHERE team_id = $1
	`
	_, err := p.db.ExecContext(ctx, query,
		data.TeamID,
		data.Name,
		data.SlackChannel,
		data.PollSchedule,
		event.Timestamp,
	)

	return err
}

func (p *Projector) handleReminderSent(ctx context.Context, event *events.Event) error {
	// Could track reminder history in a separate table if needed
	return nil
}
