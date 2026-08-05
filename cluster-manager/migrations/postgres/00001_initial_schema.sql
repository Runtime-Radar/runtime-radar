-- +goose Up
-- Initial schema for cluster-manager Postgres

CREATE TABLE IF NOT EXISTS clusters (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    token text,
    status smallint,
    config jsonb,
    registered_at timestamptz,
    deleted_at timestamptz
);

CREATE INDEX IF NOT EXISTS idx_clusters_created_at ON clusters (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_clusters_name ON clusters (name);
CREATE UNIQUE INDEX IF NOT EXISTS idx_clusters_token ON clusters (token);
CREATE INDEX IF NOT EXISTS idx_clusters_status ON clusters (status);
CREATE INDEX IF NOT EXISTS idx_clusters_registered_at ON clusters (registered_at);
CREATE INDEX IF NOT EXISTS idx_clusters_deleted_at ON clusters (deleted_at);

-- +goose Down
DROP TABLE IF EXISTS clusters;
