CREATE SCHEMA IF NOT EXISTS projections;

CREATE TABLE IF NOT EXISTS projections.teams (
    team_id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slack_channel VARCHAR(255) NOT NULL,
    poll_schedule VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE IF NOT EXISTS projections.status_updates (
    update_id VARCHAR(255) PRIMARY KEY,
    team_id VARCHAR(255) NOT NULL REFERENCES projections.teams(team_id),
    content TEXT NOT NULL,
    author VARCHAR(255) NOT NULL,
    slack_user VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE INDEX idx_status_updates_team_id ON projections.status_updates(team_id);
CREATE INDEX idx_status_updates_created_at ON projections.status_updates(created_at DESC);
CREATE INDEX idx_status_updates_team_created ON projections.status_updates(team_id, created_at DESC);
