package main

import (
	"os"
	"testing"
)

func TestRequiredEnvVars(t *testing.T) {
	// Save original env
	originalDB := os.Getenv("DATABASE_URL")
	defer os.Setenv("DATABASE_URL", originalDB)

	tests := []struct {
		name        string
		databaseURL string
		wantErr     bool
	}{
		{
			name:        "DATABASE_URL is set",
			databaseURL: "postgres://localhost:5432/test",
			wantErr:     false,
		},
		{
			name:        "DATABASE_URL is empty",
			databaseURL: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("DATABASE_URL", tt.databaseURL)
			err := validateEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateEnv() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
