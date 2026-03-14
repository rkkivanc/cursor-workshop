-- 001_create_users.up.sql
-- Creates the users table for IAM.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id            UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    email         TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    display_name  TEXT        NOT NULL DEFAULT '',
    avatar_url    TEXT        NOT NULL DEFAULT '',
    bio           TEXT        NOT NULL DEFAULT '',
    status        TEXT        NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active', 'inactive', 'suspended')),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
