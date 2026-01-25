-- Remove last_login_at column from users table
DROP INDEX IF EXISTS idx_users_last_login_at;
ALTER TABLE users DROP COLUMN IF EXISTS last_login_at;
