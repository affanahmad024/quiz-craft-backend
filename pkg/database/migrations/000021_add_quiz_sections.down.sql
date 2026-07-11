-- Remove quiz sections and section-aware question ordering.

DROP INDEX IF EXISTS idx_questions_section_question_order;
DROP INDEX IF EXISTS idx_questions_section_order;
DROP INDEX IF EXISTS idx_questions_section_id;

ALTER TABLE questions DROP CONSTRAINT IF EXISTS chk_questions_section_order_index;
ALTER TABLE questions DROP COLUMN IF EXISTS section_order_index;
ALTER TABLE questions DROP COLUMN IF EXISTS section_id;

DROP TRIGGER IF EXISTS trigger_update_quiz_sections_updated_at ON quiz_sections;
DROP FUNCTION IF EXISTS update_quiz_sections_updated_at();

DROP INDEX IF EXISTS idx_quiz_sections_default;
DROP INDEX IF EXISTS idx_quiz_sections_quiz_order;
DROP INDEX IF EXISTS idx_quiz_sections_order;
DROP INDEX IF EXISTS idx_quiz_sections_quiz_id;
DROP TABLE IF EXISTS quiz_sections;
