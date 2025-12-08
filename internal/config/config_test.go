package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Save original env vars
	originalEnv := map[string]string{
		"DATABASE_URL":      os.Getenv("DATABASE_URL"),
		"SLACK_BOT_TOKEN":   os.Getenv("SLACK_BOT_TOKEN"),
		"SLACK_SIGNING_KEY": os.Getenv("SLACK_SIGNING_KEY"),
		"PORT":              os.Getenv("PORT"),
		"EVENT_STORE_URL":   os.Getenv("EVENT_STORE_URL"),
		"PROJECTION_DB_URL": os.Getenv("PROJECTION_DB_URL"),
		"API_SECRET":        os.Getenv("API_SECRET"),
		"COMMANDS_URL":      os.Getenv("COMMANDS_URL"),
	}

	// Restore original env vars after test
	defer func() {
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	t.Run("loads default values when env vars not set", func(t *testing.T) {
		// Clear all env vars
		for key := range originalEnv {
			os.Unsetenv(key)
		}

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Port != "8080" {
			t.Errorf("Port = %v, want 8080", cfg.Port)
		}
		if cfg.DatabaseURL != "postgres://localhost:5432/statusapp?sslmode=disable" {
			t.Errorf("DatabaseURL = %v, want default", cfg.DatabaseURL)
		}
		if cfg.CommandsURL != "http://localhost:8081" {
			t.Errorf("CommandsURL = %v, want http://localhost:8081", cfg.CommandsURL)
		}
	})

	t.Run("loads values from environment variables", func(t *testing.T) {
		os.Setenv("PORT", "9090")
		os.Setenv("DATABASE_URL", "postgres://test:5432/testdb")
		os.Setenv("API_SECRET", "test-secret")
		os.Setenv("SLACK_BOT_TOKEN", "xoxb-test")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.Port != "9090" {
			t.Errorf("Port = %v, want 9090", cfg.Port)
		}
		if cfg.DatabaseURL != "postgres://test:5432/testdb" {
			t.Errorf("DatabaseURL = %v, want postgres://test:5432/testdb", cfg.DatabaseURL)
		}
		if cfg.APISecret != "test-secret" {
			t.Errorf("APISecret = %v, want test-secret", cfg.APISecret)
		}
		if cfg.SlackBotToken != "xoxb-test" {
			t.Errorf("SlackBotToken = %v, want xoxb-test", cfg.SlackBotToken)
		}
	})

	t.Run("environment variables override defaults", func(t *testing.T) {
		os.Setenv("EVENT_STORE_URL", "postgres://custom:5432/events")
		os.Unsetenv("PORT") // Should use default

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if cfg.EventStoreURL != "postgres://custom:5432/events" {
			t.Errorf("EventStoreURL = %v, want postgres://custom:5432/events", cfg.EventStoreURL)
		}
		if cfg.Port != "8080" {
			t.Errorf("Port = %v, want default 8080", cfg.Port)
		}
	})
}
