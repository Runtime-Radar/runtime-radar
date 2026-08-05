-- +goose Up
-- Initial schema for notifier Postgres

CREATE TABLE IF NOT EXISTS webhooks (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    url text,
    login text,
    encrypted_password text,
    insecure boolean,
    ca text,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS emails (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    "from" text,
    server text,
    auth_type smallint,
    username text,
    encrypted_password text,
    use_tls boolean NOT NULL DEFAULT false,
    use_start_tls boolean NOT NULL DEFAULT false,
    insecure boolean NOT NULL DEFAULT false,
    ca text,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS syslogs (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    address text,
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS notifications (
    id uuid PRIMARY KEY,
    created_at timestamptz,
    updated_at timestamptz,
    name text,
    recipients text[],
    deleted_at timestamptz,
    integration_id uuid,
    integration_type text,
    event_type text,
    template text,
    central_cs_url text,
    cs_cluster_id text,
    cs_cluster_name text,
    own_cs_url text,
    email_config jsonb,
    webhook_config jsonb,
    syslog_config jsonb
);

CREATE INDEX IF NOT EXISTS idx_webhooks_created_at ON webhooks (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_webhooks_name ON webhooks (name);
CREATE INDEX IF NOT EXISTS idx_webhooks_deleted_at ON webhooks (deleted_at);

CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_emails_name ON emails (name);
CREATE INDEX IF NOT EXISTS idx_emails_deleted_at ON emails (deleted_at);

CREATE INDEX IF NOT EXISTS idx_syslogs_created_at ON syslogs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_syslogs_name ON syslogs (name);
CREATE INDEX IF NOT EXISTS idx_syslogs_deleted_at ON syslogs (deleted_at);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_name ON notifications (name);
CREATE INDEX IF NOT EXISTS idx_notifications_deleted_at ON notifications (deleted_at);
-- Notifications belong to an integration polymorphically, so the lookup is by the pair of columns.
CREATE INDEX IF NOT EXISTS idx_notifications_integration ON notifications (integration_id, integration_type);

-- +goose Down
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS syslogs;
DROP TABLE IF EXISTS emails;
DROP TABLE IF EXISTS webhooks;
