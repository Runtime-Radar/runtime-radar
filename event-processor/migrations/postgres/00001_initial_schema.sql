-- +goose Up
-- Initial schema for event-processor Postgres

-- The table is named "event_processor_configs" rather than "configs" so that the more generic name
-- stays available for a possible cross-component dynamic config mechanism. See model.Config.
CREATE TABLE IF NOT EXISTS event_processor_configs (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    config jsonb
);

CREATE INDEX IF NOT EXISTS idx_event_processor_configs_created_at ON event_processor_configs (created_at DESC);

CREATE TABLE IF NOT EXISTS detectors (
    id text NOT NULL,
    created_at timestamptz,
    name text,
    description text,
    version bigint NOT NULL,
    author text,
    contact text,
    license text,
    wasm_binary bytea,
    wasm_hash text,
    PRIMARY KEY (id, version)
);

CREATE INDEX IF NOT EXISTS idx_detectors_name ON detectors (name);
CREATE INDEX IF NOT EXISTS idx_detectors_created_at ON detectors (created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS detectors;
DROP TABLE IF EXISTS event_processor_configs;
