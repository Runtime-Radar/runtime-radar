-- +goose Up
-- Initial schema for runtime-monitor Postgres

-- The table is named "runtime_monitor_configs" rather than "configs" so that the more generic name
-- stays available for a possible cross-component dynamic config mechanism. See model.Config.
CREATE TABLE IF NOT EXISTS runtime_monitor_configs (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    config jsonb
);

CREATE INDEX IF NOT EXISTS idx_runtime_monitor_configs_created_at ON runtime_monitor_configs (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS runtime_monitor_configs;
