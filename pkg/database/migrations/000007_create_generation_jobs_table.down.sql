-- Drop generation_jobs table
DROP INDEX IF EXISTS idx_generation_jobs_created_at;
DROP INDEX IF EXISTS idx_generation_jobs_pending;
DROP INDEX IF EXISTS idx_generation_jobs_status;
DROP INDEX IF EXISTS idx_generation_jobs_user_id;
DROP TABLE IF EXISTS generation_jobs;
