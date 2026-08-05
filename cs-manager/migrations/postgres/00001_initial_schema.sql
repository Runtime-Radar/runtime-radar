-- +goose Up
-- Initial schema for cs-manager Postgres

CREATE TABLE IF NOT EXISTS registrations (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    token_hash text,
    status smallint,
    error text
);

CREATE INDEX IF NOT EXISTS idx_registrations_created_at ON registrations (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS registrations;
