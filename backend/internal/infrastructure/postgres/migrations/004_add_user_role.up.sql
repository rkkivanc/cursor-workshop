-- 004_add_user_role.up.sql
-- Adds a role column to users for role-based access control.

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'user'
        CHECK (role IN ('admin', 'moderator', 'user'));

CREATE INDEX IF NOT EXISTS idx_users_role ON users (role);
