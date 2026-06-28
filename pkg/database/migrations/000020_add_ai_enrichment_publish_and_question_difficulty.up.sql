-- Add async publish/enrichment state and question-level enrichment fields.

ALTER TABLE quizzes
    ADD COLUMN IF NOT EXISTS publish_status VARCHAR(50) DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS ai_enrichment_job_id UUID REFERENCES generation_jobs(id) ON DELETE SET NULL;

UPDATE quizzes
SET publish_status = CASE
    WHEN is_published = true THEN 'published'
    ELSE COALESCE(publish_status, 'draft')
END
WHERE publish_status IS NULL OR publish_status = 'draft';

ALTER TABLE quizzes
    DROP CONSTRAINT IF EXISTS chk_quizzes_publish_status;

ALTER TABLE quizzes
    ADD CONSTRAINT chk_quizzes_publish_status
    CHECK (publish_status IN ('draft', 'enrichment_pending', 'enrichment_processing', 'enrichment_failed', 'published', 'unpublished'));

CREATE INDEX IF NOT EXISTS idx_quizzes_publish_status ON quizzes(publish_status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_ai_enrichment_job_id ON quizzes(ai_enrichment_job_id) WHERE ai_enrichment_job_id IS NOT NULL;

ALTER TABLE questions
    ADD COLUMN IF NOT EXISTS difficulty VARCHAR(20),
    ADD COLUMN IF NOT EXISTS tags TEXT[];

ALTER TABLE questions
    DROP CONSTRAINT IF EXISTS chk_questions_difficulty;

UPDATE questions
SET difficulty = NULL
WHERE difficulty = '';

ALTER TABLE questions
    ADD CONSTRAINT chk_questions_difficulty
    CHECK (difficulty IS NULL OR difficulty IN ('easy', 'medium', 'hard'));

CREATE INDEX IF NOT EXISTS idx_questions_difficulty ON questions(difficulty);
CREATE INDEX IF NOT EXISTS idx_questions_tags ON questions USING GIN(tags);

ALTER TABLE generation_jobs
    DROP CONSTRAINT IF EXISTS chk_generation_jobs_job_type;

ALTER TABLE generation_jobs
    ADD CONSTRAINT chk_generation_jobs_job_type
    CHECK (job_type IN ('generation', 'extraction', 'addition', 'enrichment'));

COMMENT ON COLUMN quizzes.publish_status IS 'Publishing workflow state, including async AI enrichment before public publishing';
COMMENT ON COLUMN quizzes.ai_enrichment_job_id IS 'Latest async AI enrichment job associated with quiz publishing';
COMMENT ON COLUMN questions.difficulty IS 'Question-level difficulty: easy, medium, hard';
COMMENT ON COLUMN questions.tags IS 'Question-level search/enrichment tags';

