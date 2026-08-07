-- +goose Up
-- Initial schema for kube-manager Postgres.

CREATE TABLE IF NOT EXISTS kube_manager_configs (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    config jsonb
);

CREATE INDEX IF NOT EXISTS idx_kube_manager_configs_created_at ON kube_manager_configs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS kube_manager_configs;
