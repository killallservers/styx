-- Enable pgvector extension for semantic search
CREATE EXTENSION IF NOT EXISTS vector;

-- Health check table (optional, for monitoring)
CREATE TABLE IF NOT EXISTS _health (
  id SERIAL PRIMARY KEY,
  checked_at TIMESTAMP DEFAULT NOW()
);
