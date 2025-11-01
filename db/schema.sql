-- LogStream Database Schema
-- PostgreSQL 15+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Logs table - main storage for ingested logs
CREATE TABLE IF NOT EXISTS logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMPTZ NOT NULL,
    service VARCHAR(255) NOT NULL,
    level VARCHAR(50) NOT NULL,
    message TEXT NOT NULL,
    metadata JSONB,
    trace_id UUID,
    span_id VARCHAR(32),
    host VARCHAR(255),
    tags TEXT[],
    ingested_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for query performance
CREATE INDEX IF NOT EXISTS idx_logs_timestamp ON logs(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_service_timestamp ON logs(service, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_level ON logs(level);
CREATE INDEX IF NOT EXISTS idx_logs_trace_id ON logs(trace_id) WHERE trace_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_logs_ingested_at ON logs(ingested_at DESC);

-- GIN index for JSONB metadata queries
CREATE INDEX IF NOT EXISTS idx_logs_metadata ON logs USING GIN (metadata);

-- GIN index for tags array
CREATE INDEX IF NOT EXISTS idx_logs_tags ON logs USING GIN (tags);

-- Grant permissions
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO logstream;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO logstream;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO logstream;

-- Sample data for testing
INSERT INTO logs (timestamp, service, level, message, metadata, trace_id, host, tags) VALUES
    (NOW() - INTERVAL '1 hour', 'web-api', 'INFO', 'Server started successfully', '{"port": 8080, "version": "1.0.0"}', uuid_generate_v4(), 'web-01', ARRAY['startup', 'production']),
    (NOW() - INTERVAL '30 minutes', 'web-api', 'ERROR', 'Failed to connect to database', '{"error": "connection timeout", "retry": 3}', uuid_generate_v4(), 'web-01', ARRAY['database', 'error']),
    (NOW() - INTERVAL '15 minutes', 'worker', 'INFO', 'Processing job completed', '{"job_id": "12345", "duration_ms": 1250}', uuid_generate_v4(), 'worker-03', ARRAY['job', 'success']),
    (NOW() - INTERVAL '5 minutes', 'web-api', 'WARN', 'High memory usage detected', '{"memory_percent": 85, "threshold": 80}', uuid_generate_v4(), 'web-02', ARRAY['performance', 'alert'])
ON CONFLICT DO NOTHING;

DO $$
BEGIN
    RAISE NOTICE 'LogStream database schema initialized successfully!';
    RAISE NOTICE 'Sample data inserted for testing.';
END $$;
