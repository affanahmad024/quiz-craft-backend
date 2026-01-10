-- Create quizzes table
CREATE TABLE IF NOT EXISTS quizzes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category VARCHAR(100),
    difficulty VARCHAR(20) CHECK (difficulty IN ('easy', 'medium', 'hard')),
    time_limit_minutes INT CHECK (time_limit_minutes > 0),
    passing_score INT DEFAULT 70 CHECK (passing_score >= 0 AND passing_score <= 100),
    is_published BOOLEAN DEFAULT false,
    is_public BOOLEAN DEFAULT true,
    thumbnail_url TEXT,
    tags TEXT[],
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_quizzes_creator_id ON quizzes(creator_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_category ON quizzes(category) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_difficulty ON quizzes(difficulty) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_is_published ON quizzes(is_published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_is_public ON quizzes(is_public) WHERE deleted_at IS NULL AND is_published = true;
CREATE INDEX IF NOT EXISTS idx_quizzes_tags USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_quizzes_created_at ON quizzes(created_at DESC) WHERE deleted_at IS NULL;

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_quizzes_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_quizzes_updated_at
    BEFORE UPDATE ON quizzes
    FOR EACH ROW
    EXECUTE FUNCTION update_quizzes_updated_at();

-- Comments for documentation
COMMENT ON TABLE quizzes IS 'Stores quiz information created by users';
COMMENT ON COLUMN quizzes.time_limit_minutes IS 'Time limit in minutes, NULL means no time limit';
COMMENT ON COLUMN quizzes.passing_score IS 'Minimum score percentage required to pass (0-100)';
COMMENT ON COLUMN quizzes.is_published IS 'Whether quiz is published and available for attempts';
COMMENT ON COLUMN quizzes.is_public IS 'Whether quiz is visible to all users or only creator';
COMMENT ON COLUMN quizzes.tags IS 'Array of tags for categorization and search';
COMMENT ON COLUMN quizzes.meta IS 'Additional metadata stored as JSON';
