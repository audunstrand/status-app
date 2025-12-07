CREATE SCHEMA IF NOT EXISTS events;

CREATE TABLE IF NOT EXISTS events.events (
    id VARCHAR(255) PRIMARY KEY,
    type VARCHAR(255) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    metadata JSONB,
    version INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_events_aggregate_id ON events.events(aggregate_id);
CREATE INDEX idx_events_type ON events.events(type);
CREATE INDEX idx_events_timestamp ON events.events(timestamp);
