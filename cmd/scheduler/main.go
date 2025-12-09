package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/projections"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize projection database for reading teams
	db, err := sql.Open("postgres", cfg.ProjectionDBURL)
	if err != nil {
		log.Fatalf("Failed to open projection database: %v", err)
	}
	defer db.Close()

	repo := projections.NewRepository(db)

	// Initialize Slack client
	slackAPI := slack.New(cfg.SlackBotToken)

	// Setup cron scheduler
	c := cron.New()

	// Run every Monday at 9 AM
	c.AddFunc("0 9 * * 1", func() {
		checkAndSendReminders(ctx, repo, slackAPI)
	})

	c.Start()
	log.Println("Scheduler service running (reminders every Monday at 9 AM)")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	c.Stop()
}

func checkAndSendReminders(ctx context.Context, repo *projections.Repository, slackAPI *slack.Client) {
	teams, err := repo.GetAllTeams(ctx)
	if err != nil {
		log.Printf("Failed to get teams: %v", err)
		return
	}

	for _, team := range teams {
		log.Printf("Sending reminder to team %s (%s)", team.Name, team.TeamID)
		
		if err := sendSlackReminder(slackAPI, team.SlackChannel, team.Name); err != nil {
			log.Printf("Failed to send Slack message to team %s: %v", team.Name, err)
			continue
		}
		
		log.Printf("Successfully sent reminder to team %s", team.Name)
	}
}

func sendSlackReminder(slackAPI *slack.Client, channelID, teamName string) error {
	message := "🔔 Time for your status update!"
	_, _, err := slackAPI.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
	)
	return err
}
