-- 002_create_user_settings.up.sql
-- Creates the user_settings table for per-user preferences.

CREATE TABLE IF NOT EXISTS user_settings (
    id               UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id          UUID        NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    notifications_on BOOLEAN     NOT NULL DEFAULT true,
    theme            TEXT        NOT NULL DEFAULT 'system'
                                   CHECK (theme IN ('light', 'dark', 'system')),
    language         TEXT        NOT NULL DEFAULT 'en',
    timezone         TEXT        NOT NULL DEFAULT 'UTC',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings (user_id);
