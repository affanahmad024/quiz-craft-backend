ALTER TABLE generation_jobs
    DROP CONSTRAINT IF EXISTS chk_generation_jobs_job_type;

ALTER TABLE generation_jobs
    ADD CONSTRAINT chk_generation_jobs_job_type
    CHECK (job_type IN ('generation', 'extraction', 'addition'));

DROP INDEX IF EXISTS idx_questions_tags;
DROP INDEX IF EXISTS idx_questions_difficulty;
ALTER TABLE questions DROP CONSTRAINT IF EXISTS chk_questions_difficulty;
ALTER TABLE questions DROP COLUMN IF EXISTS tags;
ALTER TABLE questions DROP COLUMN IF EXISTS difficulty;

DROP INDEX IF EXISTS idx_quizzes_ai_enrichment_job_id;
DROP INDEX IF EXISTS idx_quizzes_publish_status;
ALTER TABLE quizzes DROP CONSTRAINT IF EXISTS chk_quizzes_publish_status;
ALTER TABLE quizzes DROP COLUMN IF EXISTS ai_enrichment_job_id;
ALTER TABLE quizzes DROP COLUMN IF EXISTS publish_status;
