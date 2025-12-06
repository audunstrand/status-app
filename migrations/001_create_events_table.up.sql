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

CREATE INDEX idx_events_aggregate_id ON events(aggregate_id);
CREATE INDEX idx_events_type ON events(type);
CREATE INDEX idx_events_timestamp ON events(timestamp);
