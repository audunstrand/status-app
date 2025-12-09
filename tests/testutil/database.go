package testutil

import (
	"context"
	"database/sql"
	"encoding/json"
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
		updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
		last_reminded_at TIMESTAMP WITH TIME ZONE
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

// MustMarshalJSON marshals v to JSON or fails the test
func MustMarshalJSON(t *testing.T, v interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	return data
}

// InsertTestTeam inserts a team into the database for testing
func InsertTestTeam(t *testing.T, db *sql.DB, teamID, name, channel string) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO teams (team_id, name, slack_channel, poll_schedule, created_at, updated_at)
		VALUES ($1, $2, $3, 'weekly', NOW(), NOW())
	`, teamID, name, channel)
	if err != nil {
		t.Fatalf("Failed to insert test team %s: %v", teamID, err)
	}
}

// InsertTestStatusUpdate inserts a status update into the database for testing
func InsertTestStatusUpdate(t *testing.T, db *sql.DB, teamID, content, author, slackUser string) {
	t.Helper()
	ctx := context.Background()
	_, err := db.ExecContext(ctx, `
		INSERT INTO status_updates (update_id, team_id, content, author, slack_user, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, GenerateID(), teamID, content, author, slackUser)
	if err != nil {
		t.Fatalf("Failed to insert test status update: %v", err)
	}
}

// AssertEqual checks if got == want and fails with a clear message if not
func AssertEqual(t *testing.T, got, want interface{}, field string) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", field, got, want)
	}
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error, operation string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s failed: %v", operation, err)
	}
}
