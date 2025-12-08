package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB manages a test database lifecycle with Testcontainers
type TestDB struct {
	DB        *sql.DB
	container *postgres.PostgresContainer
	t         *testing.T
}

// SetupTestDB creates a new PostgreSQL test database using Testcontainers
func SetupTestDB(t *testing.T) *TestDB {
	t.Helper()

	ctx := context.Background()

	// Create PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("Failed to start postgres container: %v", err)
	}

	// Get connection string
	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("Failed to get connection string: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Verify connection
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	testDB := &TestDB{
		DB:        db,
		container: postgresContainer,
		t:         t,
	}

	// Run migrations
	testDB.RunMigrations()

	return testDB
}

// RunMigrations runs database migrations
func (tdb *TestDB) RunMigrations() {
	tdb.t.Helper()

	// Event store migration
	eventStoreMigration := `
	CREATE TABLE IF NOT EXISTS events (
		id VARCHAR(255) PRIMARY KEY,
		type VARCHAR(255) NOT NULL,
		aggregate_id VARCHAR(255) NOT NULL,
		data JSONB NOT NULL,
		timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
		metadata JSONB,
		version INTEGER NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_events_aggregate_id ON events(aggregate_id);
	CREATE INDEX IF NOT EXISTS idx_events_type ON events(type);
	CREATE INDEX IF NOT EXISTS idx_events_timestamp ON events(timestamp);
	`

	_, err := tdb.DB.Exec(eventStoreMigration)
	if err != nil {
		tdb.t.Fatalf("Failed to run event store migration: %v", err)
	}

	// Projections migration
	projectionsMigration := `
	CREATE TABLE IF NOT EXISTS teams (
		team_id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		slack_channel VARCHAR(255) NOT NULL,
		poll_schedule VARCHAR(100),
		created_at TIMESTAMP WITH TIME ZONE NOT NULL,
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	CREATE TABLE IF NOT EXISTS status_updates (
		update_id VARCHAR(255) PRIMARY KEY,
		team_id VARCHAR(255) NOT NULL REFERENCES teams(team_id),
		content TEXT NOT NULL,
		author VARCHAR(255) NOT NULL,
		slack_user VARCHAR(255) NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_status_updates_team_id ON status_updates(team_id);
	CREATE INDEX IF NOT EXISTS idx_status_updates_created_at ON status_updates(created_at DESC);
	CREATE INDEX IF NOT EXISTS idx_status_updates_team_created ON status_updates(team_id, created_at DESC);
	`

	_, err = tdb.DB.Exec(projectionsMigration)
	if err != nil {
		tdb.t.Fatalf("Failed to run projections migration: %v", err)
	}
}

// Cleanup terminates the test database container
func (tdb *TestDB) Cleanup() {
	tdb.t.Helper()

	if tdb.DB != nil {
		tdb.DB.Close()
	}

	if tdb.container != nil {
		ctx := context.Background()
		if err := tdb.container.Terminate(ctx); err != nil {
			tdb.t.Logf("Warning: failed to terminate container: %v", err)
		}
	}
}

// ConnectionString returns the database connection string
func (tdb *TestDB) ConnectionString() string {
	tdb.t.Helper()

	ctx := context.Background()
	connStr, err := tdb.container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		tdb.t.Fatalf("Failed to get connection string: %v", err)
	}
	return connStr
}

// GenerateID generates a unique ID for testing
func GenerateID() string {
	return fmt.Sprintf("test-%d", time.Now().UnixNano())
}
