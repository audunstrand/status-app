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
	"github.com/yourusername/status-app/internal/config"
	"github.com/yourusername/status-app/internal/projections"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize projection database
	db, err := sql.Open("postgres", cfg.ProjectionDBURL)
	if err != nil {
		log.Fatalf("Failed to open projection database: %v", err)
	}
	defer db.Close()

	repo := projections.NewRepository(db)

	// Setup routes
	mux := http.NewServeMux()
	mux.HandleFunc("/api/teams", handleGetTeams(repo))
	mux.HandleFunc("/api/teams/", handleGetTeam(repo))
	mux.HandleFunc("/api/updates", handleGetRecentUpdates(repo))
	mux.HandleFunc("/api/teams/", handleGetTeamUpdates(repo))

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	go func() {
		log.Printf("API service listening on port %s", cfg.Port)
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

func handleGetTeams(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		teams, err := repo.GetAllTeams(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(teams)
	}
}

func handleGetTeam(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract team ID from URL
		w.WriteHeader(http.StatusOK)
	}
}

func handleGetRecentUpdates(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updates, err := repo.GetRecentUpdates(r.Context(), 50)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(updates)
	}
}

func handleGetTeamUpdates(repo *projections.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Extract team ID from URL and get updates
		w.WriteHeader(http.StatusOK)
	}
}
