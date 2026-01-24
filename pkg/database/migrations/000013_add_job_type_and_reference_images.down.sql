-- Remove job_type column from generation_jobs table
ALTER TABLE generation_jobs DROP COLUMN IF EXISTS job_type;

-- Remove reference_images column from generation_jobs table
ALTER TABLE generation_jobs DROP COLUMN IF EXISTS reference_images;
