-- +goose Up
-- Initial schema for policy-enforcer Postgres

CREATE TABLE IF NOT EXISTS rules (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    rule jsonb,
    scope jsonb,
    deleted_at timestamptz,
    type smallint
);

CREATE INDEX IF NOT EXISTS idx_rules_created_at ON rules (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rules_name ON rules (name);
CREATE INDEX IF NOT EXISTS idx_rules_deleted_at ON rules (deleted_at);
CREATE INDEX IF NOT EXISTS idx_rules_type ON rules (type);

-- Rules are queried by their JSON contents, hence the GIN index.
CREATE INDEX IF NOT EXISTS idx_rules_rule ON rules USING gin (rule);

-- +goose Down
DROP TABLE IF EXISTS rules;
