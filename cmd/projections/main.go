package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/lib/pq"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/internal/projections"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize event store
	eventStore, err := events.NewPostgresStore(cfg.EventStoreURL)
	if err != nil {
		log.Fatalf("Failed to create event store: %v", err)
	}
	defer eventStore.Close()

	// Initialize projection database
	projectionDB, err := sql.Open("postgres", cfg.ProjectionDBURL)
	if err != nil {
		log.Fatalf("Failed to open projection database: %v", err)
	}
	defer projectionDB.Close()

	// Create projector
	projector := projections.NewProjector(eventStore, projectionDB)

	// Start projector
	log.Println("Starting projection service...")
	if err := projector.Start(ctx); err != nil {
		log.Fatalf("Failed to start projector: %v", err)
	}

	log.Println("Projection service running")

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
}
