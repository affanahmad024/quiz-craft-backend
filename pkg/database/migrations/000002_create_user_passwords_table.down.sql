-- Drop trigger
DROP TRIGGER IF EXISTS update_user_passwords_updated_at ON user_passwords;

-- Drop indexes
DROP INDEX IF EXISTS idx_user_passwords_reset_token;
DROP INDEX IF EXISTS idx_user_passwords_email_token;
DROP INDEX IF EXISTS idx_user_passwords_user_id;

-- Drop table
DROP TABLE IF EXISTS user_passwords;
