-- 003_create_app_settings.up.sql
-- Creates the app_settings table for global application configuration.

CREATE TABLE IF NOT EXISTS app_settings (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    key         TEXT        NOT NULL UNIQUE,
    value       TEXT        NOT NULL DEFAULT '',
    description TEXT        NOT NULL DEFAULT '',
    is_public   BOOLEAN     NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_app_settings_key       ON app_settings (key);
CREATE INDEX IF NOT EXISTS idx_app_settings_is_public ON app_settings (is_public);

-- Seed some default app settings
INSERT INTO app_settings (key, value, description, is_public) VALUES
    ('app.version',          '1.0.0',  'Current application version',          true),
    ('app.maintenance',      'false',  'Maintenance mode flag',                 true),
    ('app.min_app_version',  '1.0.0',  'Minimum supported mobile app version', true)
ON CONFLICT (key) DO NOTHING;
