-- Create generation_jobs table for Phase 4: Async Quiz Generation
CREATE TABLE generation_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    file_id UUID,
    text_content TEXT,
    options JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    progress INT DEFAULT 0,
    quiz_id UUID,
    error_message TEXT,
    job_type VARCHAR(50) DEFAULT 'generation',
    reference_images JSONB DEFAULT '[]'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT fk_generation_jobs_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_generation_jobs_file
        FOREIGN KEY (file_id)
        REFERENCES uploaded_files(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_generation_jobs_quiz
        FOREIGN KEY (quiz_id)
        REFERENCES quizzes(id)
        ON DELETE SET NULL,

    CONSTRAINT chk_generation_jobs_progress
        CHECK (progress >= 0 AND progress <= 100),

    CONSTRAINT chk_generation_jobs_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),

    CONSTRAINT chk_generation_jobs_job_type
        CHECK (job_type IN ('generation', 'extraction', 'addition'))
);

-- Create indexes for performance
CREATE INDEX idx_generation_jobs_user_id ON generation_jobs(user_id);
CREATE INDEX idx_generation_jobs_status ON generation_jobs(status);
CREATE INDEX idx_generation_jobs_pending ON generation_jobs(status) WHERE status = 'pending';
CREATE INDEX idx_generation_jobs_created_at ON generation_jobs(created_at);
CREATE INDEX idx_generation_jobs_job_type ON generation_jobs(job_type);

-- Add comments for documentation
COMMENT ON TABLE generation_jobs IS 'Tracks async AI quiz generation jobs with progress';
COMMENT ON COLUMN generation_jobs.file_id IS 'Reference to uploaded file (NULL if generated from text)';
COMMENT ON COLUMN generation_jobs.text_content IS 'Direct text content if not from file';
COMMENT ON COLUMN generation_jobs.options IS 'QuizGenOptions JSON: num_questions, difficulty, category, etc.';
COMMENT ON COLUMN generation_jobs.status IS 'Job status: pending, processing, completed, failed';
COMMENT ON COLUMN generation_jobs.progress IS 'Progress percentage: 0-100';
COMMENT ON COLUMN generation_jobs.quiz_id IS 'Generated quiz ID (set when completed)';
COMMENT ON COLUMN generation_jobs.job_type IS 'Type of job: generation (new quiz), extraction (from image), addition (add to existing quiz)';
COMMENT ON COLUMN generation_jobs.reference_images IS 'Array of file IDs for reference images (stored as JSONB array)';
