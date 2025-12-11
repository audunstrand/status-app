package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Start HTTP server for metrics
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "scheduler",
		})
	})
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    ":8082", // Different port
		Handler: mux,
	}

	go func() {
		log.Println("Metrics server running on :8082")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Metrics server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	c.Stop()
	
	// Shutdown metrics server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}

func checkAndSendReminders(ctx context.Context, repo *projections.Repository, slackAPI *slack.Client) {
	remindersScheduledTotal.Inc()
	
	teams, err := repo.GetAllTeams(ctx)
	if err != nil {
		schedulerErrorsTotal.WithLabelValues("db_error").Inc()
		log.Printf("Failed to get teams: %v", err)
		return
	}

	successCount := 0
	for _, team := range teams {
		log.Printf("Sending reminder to team %s (%s)", team.Name, team.TeamID)
		
		if err := sendSlackReminder(slackAPI, team.SlackChannel, team.Name); err != nil {
			remindersSentTotal.WithLabelValues("error").Inc()
			schedulerErrorsTotal.WithLabelValues("slack_error").Inc()
			log.Printf("Failed to send Slack message to team %s: %v", team.Name, err)
			continue
		}
		
		remindersSentTotal.WithLabelValues("success").Inc()
		successCount++
		log.Printf("Successfully sent reminder to team %s", team.Name)
	}
	
	teamsReminderCount.Set(float64(successCount))
}

func sendSlackReminder(slackAPI *slack.Client, channelID, teamName string) error {
	message := "🔔 Time for your status update!"
	_, _, err := slackAPI.PostMessage(
		channelID,
		slack.MsgOptionText(message, false),
	)
	return err
}
