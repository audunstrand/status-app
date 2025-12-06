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

	// Initialize projection database for reading team schedules
	db, err := sql.Open("postgres", cfg.ProjectionDBURL)
	if err != nil {
		log.Fatalf("Failed to open projection database: %v", err)
	}
	defer db.Close()

	repo := projections.NewRepository(db)

	// Setup cron scheduler
	c := cron.New()

	// Check for teams to poll every hour
	c.AddFunc("@hourly", func() {
		checkAndSendReminders(ctx, repo)
	})

	c.Start()
	log.Println("Scheduler service running")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	c.Stop()
}

func checkAndSendReminders(ctx context.Context, repo *projections.Repository) {
	teams, err := repo.GetAllTeams(ctx)
	if err != nil {
		log.Printf("Failed to get teams: %v", err)
		return
	}

	for _, team := range teams {
		// TODO: Check if team is due for a reminder based on poll_schedule
		// TODO: Send command to send reminder
		log.Printf("Checking team %s for reminders", team.Name)
	}
}
