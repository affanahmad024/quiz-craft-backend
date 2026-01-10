-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_quizzes_updated_at ON quizzes;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_quizzes_updated_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_quizzes_creator_id;
DROP INDEX IF EXISTS idx_quizzes_category;
DROP INDEX IF EXISTS idx_quizzes_difficulty;
DROP INDEX IF EXISTS idx_quizzes_is_published;
DROP INDEX IF EXISTS idx_quizzes_is_public;
DROP INDEX IF EXISTS idx_quizzes_tags;
DROP INDEX IF EXISTS idx_quizzes_created_at;

-- Drop table
DROP TABLE IF EXISTS quizzes;
