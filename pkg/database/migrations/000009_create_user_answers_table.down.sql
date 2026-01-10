-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_user_answers_updated_at ON user_answers;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_user_answers_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_user_answers_attempt_id;
DROP INDEX IF EXISTS idx_user_answers_question_id;
DROP INDEX IF EXISTS idx_user_answers_answer_id;
DROP INDEX IF EXISTS idx_user_answers_answered_at;

-- Drop table
DROP TABLE IF EXISTS user_answers;
