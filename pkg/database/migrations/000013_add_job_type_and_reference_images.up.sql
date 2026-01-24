-- Add job_type column to generation_jobs table
ALTER TABLE generation_jobs ADD COLUMN job_type VARCHAR(50) DEFAULT 'generation';

-- Add reference_images column to generation_jobs table
ALTER TABLE generation_jobs ADD COLUMN reference_images JSONB DEFAULT '[]'::jsonb;

-- Add check constraint for job_type
ALTER TABLE generation_jobs ADD CONSTRAINT chk_generation_jobs_job_type
    CHECK (job_type IN ('generation', 'extraction', 'addition'));

-- Add index for job_type
CREATE INDEX idx_generation_jobs_job_type ON generation_jobs(job_type);

-- Add comments
COMMENT ON COLUMN generation_jobs.job_type IS 'Type of job: generation (new quiz), extraction (from image), addition (add to existing quiz)';
COMMENT ON COLUMN generation_jobs.reference_images IS 'Array of file IDs for reference images (stored as JSONB array)';
