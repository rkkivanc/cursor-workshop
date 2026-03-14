-- 004_add_user_role.down.sql
-- Reverts the role column addition.

DROP INDEX IF EXISTS idx_users_role;

ALTER TABLE users DROP COLUMN IF EXISTS role;
