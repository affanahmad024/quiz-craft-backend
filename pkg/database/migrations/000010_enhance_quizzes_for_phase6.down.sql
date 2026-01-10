-- Drop indexes
DROP INDEX IF EXISTS idx_quizzes_share_code;
DROP INDEX IF EXISTS idx_quizzes_is_template;
DROP INDEX IF EXISTS idx_quizzes_based_on_template;
DROP INDEX IF EXISTS idx_quizzes_view_count;
DROP INDEX IF EXISTS idx_quizzes_search;

-- Drop columns
ALTER TABLE quizzes DROP COLUMN IF EXISTS based_on_template_id;
ALTER TABLE quizzes DROP COLUMN IF EXISTS template_usage_count;
ALTER TABLE quizzes DROP COLUMN IF EXISTS is_template;
ALTER TABLE quizzes DROP COLUMN IF EXISTS view_count;
ALTER TABLE quizzes DROP COLUMN IF EXISTS share_count;
ALTER TABLE quizzes DROP COLUMN IF EXISTS share_code;
