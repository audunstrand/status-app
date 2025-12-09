package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/robfig/cron/v3"
	"github.com/slack-go/slack"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/projections"
	"github.com/yourusername/status-app/internal/scheduler"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize projection database for reading team schedules
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
		checkAndSendReminders(ctx, repo, slackAPI, cfg)
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

func checkAndSendReminders(ctx context.Context, repo *projections.Repository, slackAPI *slack.Client, cfg *config.Config) {
	teams, err := repo.GetAllTeams(ctx)
	if err != nil {
		log.Printf("Failed to get teams: %v", err)
		return
	}

	now := time.Now()
	for _, team := range teams {
		if scheduler.ShouldRemind(team.LastRemindedAt, now) {
			log.Printf("Sending reminder to team %s (%s)", team.Name, team.TeamID)
			
			if err := sendSlackReminder(slackAPI, team.SlackChannel, team.Name); err != nil {
				log.Printf("Failed to send Slack message to team %s: %v", team.Name, err)
				continue
			}
			
			if err := recordReminder(ctx, cfg, team.TeamID, team.SlackChannel); err != nil {
				log.Printf("Failed to record reminder for team %s: %v", team.Name, err)
				continue
			}
			
			log.Printf("Successfully sent reminder to team %s", team.Name)
		}
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

func recordReminder(ctx context.Context, cfg *config.Config, teamID, slackChannel string) error {
	payload := map[string]string{
		"slack_channel": slackChannel,
	}
	
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	
	url := cfg.CommandsURL + "/teams/" + teamID + "/reminders"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APISecret)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to record reminder: status %d", resp.StatusCode)
	}
	
	return nil
}
