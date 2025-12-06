package config

import (
	"os"
)

type Config struct {
	DatabaseURL      string
	SlackBotToken    string
	SlackSigningKey  string
	Port             string
	EventStoreURL    string
	ProjectionDBURL  string
}

func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://localhost:5432/statusapp?sslmode=disable"),
		SlackBotToken:   getEnv("SLACK_BOT_TOKEN", ""),
		SlackSigningKey: getEnv("SLACK_SIGNING_KEY", ""),
		Port:            getEnv("PORT", "8080"),
		EventStoreURL:   getEnv("EVENT_STORE_URL", "postgres://localhost:5432/statusapp_events?sslmode=disable"),
		ProjectionDBURL: getEnv("PROJECTION_DB_URL", "postgres://localhost:5432/statusapp_projections?sslmode=disable"),
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
