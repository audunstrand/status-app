package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/status-app/internal/commands"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/events"
)

func main() {
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

	// Initialize command handler
	cmdHandler := commands.NewHandler(eventStore)

	// Start HTTP server for receiving commands
	mux := http.NewServeMux()
	mux.HandleFunc("/commands/submit-update", handleSubmitUpdate(cmdHandler))
	mux.HandleFunc("/commands/register-team", handleRegisterTeam(cmdHandler))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Command service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}

func handleSubmitUpdate(handler *commands.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Parse request and create command
		w.WriteHeader(http.StatusOK)
	}
}

func handleRegisterTeam(handler *commands.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Parse request and create command
		w.WriteHeader(http.StatusOK)
	}
}
