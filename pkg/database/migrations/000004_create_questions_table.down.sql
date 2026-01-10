-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_questions_updated_at ON questions;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_questions_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_questions_quiz_id;
DROP INDEX IF EXISTS idx_questions_order;
DROP INDEX IF EXISTS idx_questions_type;
DROP INDEX IF EXISTS idx_questions_quiz_order;

-- Drop table
DROP TABLE IF EXISTS questions;
