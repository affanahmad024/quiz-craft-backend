-- Create user_answers table
CREATE TABLE IF NOT EXISTS user_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    attempt_id UUID NOT NULL REFERENCES quiz_attempts(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,
    answer_text TEXT,
    is_correct BOOLEAN,
    points_earned DECIMAL(5,2) DEFAULT 0 CHECK (points_earned >= 0),
    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Constraints
    UNIQUE (attempt_id, question_id),
    CHECK (
        (answer_id IS NOT NULL AND answer_text IS NULL) OR
        (answer_id IS NULL AND answer_text IS NOT NULL) OR
        (answer_id IS NULL AND answer_text IS NULL)
    )
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_user_answers_attempt_id ON user_answers(attempt_id);
CREATE INDEX IF NOT EXISTS idx_user_answers_question_id ON user_answers(question_id);
CREATE INDEX IF NOT EXISTS idx_user_answers_answer_id ON user_answers(answer_id);
CREATE INDEX IF NOT EXISTS idx_user_answers_answered_at ON user_answers(answered_at DESC);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_user_answers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_user_answers_updated_at
    BEFORE UPDATE ON user_answers
    FOR EACH ROW
    EXECUTE FUNCTION update_user_answers_updated_at();

-- Comments for documentation
COMMENT ON TABLE user_answers IS 'Stores individual answers submitted during quiz attempts';
COMMENT ON COLUMN user_answers.answer_id IS 'Reference to selected answer (for MC/TF), NULL for short answer';
COMMENT ON COLUMN user_answers.answer_text IS 'Text answer for short answer questions, NULL for MC/TF';
COMMENT ON COLUMN user_answers.is_correct IS 'Whether answer is correct (auto-graded for MC/TF)';
COMMENT ON COLUMN user_answers.points_earned IS 'Points earned for this answer';
COMMENT ON COLUMN user_answers.answered_at IS 'Timestamp when answer was submitted';
COMMENT ON COLUMN user_answers.meta IS 'Additional metadata stored as JSON';
