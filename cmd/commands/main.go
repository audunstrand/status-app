package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/status-app/internal/auth"
	"github.com/yourusername/status-app/internal/commands"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/events"
)

type SubmitStatusUpdateRequest struct {
	TeamID  string `json:"team_id"`
	Content string `json:"content"`
	Author  string `json:"author"`
}

func (r *SubmitStatusUpdateRequest) Validate() error {
	if r.TeamID == "" {
		return errors.New("team_id is required")
	}
	if r.Content == "" {
		return errors.New("content is required")
	}
	if len(r.Content) > 500 {
		return errors.New("content must be 500 characters or less")
	}
	if r.Author == "" {
		return errors.New("author is required")
	}
	return nil
}

type RegisterTeamRequest struct {
	Name         string `json:"name"`
	SlackChannel string `json:"slack_channel"`
	PollSchedule string `json:"poll_schedule"`
}

func (r *RegisterTeamRequest) Validate() error {
	if r.Name == "" {
		return errors.New("name is required")
	}
	if r.SlackChannel == "" {
		return errors.New("slack_channel is required")
	}
	return nil
}

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
	
	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "commands",
		})
	})
	
	// Protected command endpoints
	if cfg.APISecret == "" {
		log.Fatal("API_SECRET environment variable is required")
	}
	
	log.Println("API authentication enabled")
	protectedMux := http.NewServeMux()
	protectedMux.HandleFunc("/commands/submit-update", handleSubmitUpdate(cmdHandler))
	protectedMux.HandleFunc("/commands/register-team", handleRegisterTeam(cmdHandler))
	mux.Handle("/commands/", auth.RequireAPIKey(cfg.APISecret)(protectedMux))

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
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SubmitStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cmd := commands.SubmitStatusUpdate{
			TeamID:    req.TeamID,
			Content:   req.Content,
			Author:    req.Author,
			SlackUser: req.Author, // Default to author if no slack user
			Timestamp: time.Now(),
		}

		if err := handler.Handle(r.Context(), cmd); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
	}
}

func handleRegisterTeam(handler *commands.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterTeamRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if err := req.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cmd := commands.RegisterTeam{
			Name:         req.Name,
			SlackChannel: req.SlackChannel,
			PollSchedule: req.PollSchedule,
		}

		if err := handler.Handle(r.Context(), cmd); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "success",
		})
	}
}
