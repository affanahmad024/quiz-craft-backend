-- Create quiz_attempts table
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL DEFAULT 'in_progress',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    submitted_at TIMESTAMP WITH TIME ZONE,
    time_limit_seconds INT CHECK (time_limit_seconds > 0),
    expires_at TIMESTAMP WITH TIME ZONE,
    score DECIMAL(5,2) CHECK (score >= 0),
    max_score INT CHECK (max_score >= 0),
    percentage DECIMAL(5,2) CHECK (percentage >= 0 AND percentage <= 100),
    passed BOOLEAN,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    CHECK (status IN ('in_progress', 'submitted', 'expired', 'abandoned')),
    CHECK (score IS NULL OR (max_score IS NOT NULL AND score <= max_score))
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_quiz_id ON quiz_attempts(quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_id ON quiz_attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_status ON quiz_attempts(status);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_expires_at ON quiz_attempts(expires_at) WHERE status = 'in_progress' AND expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_quiz ON quiz_attempts(user_id, quiz_id);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_created_at ON quiz_attempts(created_at DESC);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_quiz_attempts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quiz_attempts_updated_at
    BEFORE UPDATE ON quiz_attempts
    FOR EACH ROW
    EXECUTE FUNCTION update_quiz_attempts_updated_at();

-- Comments for documentation
COMMENT ON TABLE quiz_attempts IS 'Stores user attempts to take quizzes';
COMMENT ON COLUMN quiz_attempts.status IS 'Current status: in_progress, submitted, expired, abandoned';
COMMENT ON COLUMN quiz_attempts.time_limit_seconds IS 'Time limit in seconds copied from quiz at attempt start';
COMMENT ON COLUMN quiz_attempts.expires_at IS 'Timestamp when attempt expires, NULL if no time limit';
COMMENT ON COLUMN quiz_attempts.score IS 'Total points earned by user';
COMMENT ON COLUMN quiz_attempts.max_score IS 'Maximum possible score for this quiz';
COMMENT ON COLUMN quiz_attempts.percentage IS 'Score as percentage (score/max_score * 100)';
COMMENT ON COLUMN quiz_attempts.passed IS 'Whether user passed based on passing_score threshold';
COMMENT ON COLUMN quiz_attempts.meta IS 'Additional metadata stored as JSON';
