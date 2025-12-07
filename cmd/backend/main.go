package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/yourusername/status-app/internal/auth"
	"github.com/yourusername/status-app/internal/commands"
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/events"
	"github.com/yourusername/status-app/internal/projections"
)

// Request types for commands
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

	// Initialize command handler
	cmdHandler := commands.NewHandler(eventStore)

	// Initialize projection repository
	repo := projections.NewRepository(projectionDB)

	// Create and start projector in background goroutine
	projector := projections.NewProjector(eventStore, projectionDB)
	go func() {
		log.Println("Starting projections processor...")
		if err := projector.Start(ctx); err != nil {
			log.Printf("Projector error: %v", err)
		}
	}()
	log.Println("Projections running in background")

	// Setup HTTP routes
	mux := http.NewServeMux()

	// Health check endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "backend",
		})
	})

	// Protected endpoints
	if cfg.APISecret == "" {
		log.Fatal("API_SECRET environment variable is required")
	}

	log.Println("API authentication enabled")
	protectedMux := http.NewServeMux()

	// Command endpoints
	protectedMux.HandleFunc("/commands/submit-update", handleSubmitUpdate(cmdHandler))
	protectedMux.HandleFunc("/commands/register-team", handleRegisterTeam(cmdHandler))

	// API endpoints
	protectedMux.HandleFunc("/api/teams", handleGetTeams(repo))
	protectedMux.HandleFunc("/api/teams/{id}", handleGetTeam(repo))
	protectedMux.HandleFunc("/api/updates", handleGetRecentUpdates(repo))
	protectedMux.HandleFunc("/api/teams/{id}/updates", handleGetTeamUpdates(repo))

	mux.Handle("/commands/", auth.RequireAPIKey(cfg.APISecret)(protectedMux))
	mux.Handle("/api/", auth.RequireAPIKey(cfg.APISecret)(protectedMux))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("Backend service listening on port %s", cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down...")
	cancel() // Stop projector

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	server.Shutdown(shutdownCtx)
}

// Command handlers
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
			SlackUser: req.Author,
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

// API handlers
func handleGetTeams(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams, err := repo.GetAllTeams(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(teams)
	}
}

func handleGetTeam(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamID := r.PathValue("id")
		if teamID == "" {
			http.Error(w, "team ID is required", http.StatusBadRequest)
			return
		}

		team, err := repo.GetTeam(r.Context(), teamID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "team not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(team)
	}
}

func handleGetRecentUpdates(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updates, err := repo.GetRecentUpdates(r.Context(), 50)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updates)
	}
}

func handleGetTeamUpdates(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teamID := r.PathValue("id")
		if teamID == "" {
			http.Error(w, "team ID is required", http.StatusBadRequest)
			return
		}

		updates, err := repo.GetTeamUpdates(r.Context(), teamID, 50)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updates)
	}
}
