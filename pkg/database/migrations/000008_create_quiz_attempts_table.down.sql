-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_quiz_attempts_updated_at ON quiz_attempts;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_quiz_attempts_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_quiz_attempts_quiz_id;
DROP INDEX IF EXISTS idx_quiz_attempts_user_id;
DROP INDEX IF EXISTS idx_quiz_attempts_status;
DROP INDEX IF EXISTS idx_quiz_attempts_expires_at;
DROP INDEX IF EXISTS idx_quiz_attempts_user_quiz;
DROP INDEX IF EXISTS idx_quiz_attempts_created_at;

-- Drop table
DROP TABLE IF EXISTS quiz_attempts;
