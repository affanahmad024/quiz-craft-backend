-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_answers_updated_at ON answers;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_answers_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_answers_question_id;
DROP INDEX IF EXISTS idx_answers_order;
DROP INDEX IF EXISTS idx_answers_is_correct;
DROP INDEX IF EXISTS idx_answers_question_order;

-- Drop table
DROP TABLE IF EXISTS answers;
