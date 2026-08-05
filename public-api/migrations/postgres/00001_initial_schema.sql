-- +goose Up
-- Initial schema for public-api Postgres

CREATE TABLE IF NOT EXISTS access_tokens (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    user_id text,
    hash text,
    permissions jsonb,
    expires_at timestamptz,
    invalidated_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_access_tokens_created_at ON access_tokens (created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_access_tokens_hash ON access_tokens (hash);

-- +goose Down
DROP TABLE IF EXISTS access_tokens;
