-- Add quiz sections and section-aware question ordering.

CREATE TABLE IF NOT EXISTS quiz_sections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    instructions TEXT,
    order_index INT NOT NULL CHECK (order_index >= 0),
    time_limit_minutes INT CHECK (time_limit_minutes IS NULL OR time_limit_minutes > 0),
    is_default BOOLEAN NOT NULL DEFAULT false,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_quiz_sections_quiz_id ON quiz_sections(quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_sections_order ON quiz_sections(quiz_id, order_index);
CREATE UNIQUE INDEX IF NOT EXISTS idx_quiz_sections_quiz_order ON quiz_sections(quiz_id, order_index);
CREATE UNIQUE INDEX IF NOT EXISTS idx_quiz_sections_default ON quiz_sections(quiz_id) WHERE is_default = true;

CREATE OR REPLACE FUNCTION update_quiz_sections_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_update_quiz_sections_updated_at ON quiz_sections;
CREATE TRIGGER trigger_update_quiz_sections_updated_at
    BEFORE UPDATE ON quiz_sections
    FOR EACH ROW
    EXECUTE FUNCTION update_quiz_sections_updated_at();

INSERT INTO quiz_sections (id, quiz_id, title, order_index, is_default, meta, created_at, updated_at)
SELECT gen_random_uuid(), q.id, 'General', 0, true, '{}'::jsonb, NOW(), NOW()
FROM quizzes q
WHERE NOT EXISTS (
    SELECT 1 FROM quiz_sections s WHERE s.quiz_id = q.id AND s.is_default = true
);

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS section_id UUID REFERENCES quiz_sections(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS section_order_index INT;

WITH numbered_questions AS (
    SELECT
        q.id,
        s.id AS section_id,
        ROW_NUMBER() OVER (PARTITION BY q.quiz_id ORDER BY q.order_index ASC, q.created_at ASC, q.id ASC) - 1 AS section_order_index
    FROM questions q
    JOIN quiz_sections s ON s.quiz_id = q.quiz_id AND s.is_default = true
    WHERE q.section_id IS NULL
)
UPDATE questions q
SET section_id = numbered_questions.section_id,
    section_order_index = numbered_questions.section_order_index
FROM numbered_questions
WHERE q.id = numbered_questions.id;

UPDATE questions
SET section_order_index = order_index
WHERE section_order_index IS NULL;

ALTER TABLE questions
    ALTER COLUMN section_id SET NOT NULL,
    ALTER COLUMN section_order_index SET NOT NULL;

ALTER TABLE questions
    DROP CONSTRAINT IF EXISTS chk_questions_section_order_index;

ALTER TABLE questions
    ADD CONSTRAINT chk_questions_section_order_index CHECK (section_order_index >= 0);

CREATE INDEX IF NOT EXISTS idx_questions_section_id ON questions(section_id);
CREATE INDEX IF NOT EXISTS idx_questions_section_order ON questions(section_id, section_order_index);
CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_section_question_order ON questions(section_id, section_order_index);

COMMENT ON TABLE quiz_sections IS 'Creator-defined sections used to organize quiz questions';
COMMENT ON COLUMN quiz_sections.is_default IS 'Exactly one default General section per quiz';
COMMENT ON COLUMN quiz_sections.instructions IS 'Optional future section-level instructions';
COMMENT ON COLUMN quiz_sections.time_limit_minutes IS 'Optional future section-level timer';
COMMENT ON COLUMN questions.section_id IS 'Section this question belongs to';
COMMENT ON COLUMN questions.section_order_index IS 'Display order within the assigned section';
