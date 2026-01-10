-- Create questions table
CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question_text TEXT NOT NULL,
    question_type VARCHAR(50) NOT NULL DEFAULT 'multiple_choice' CHECK (question_type IN ('multiple_choice', 'true_false', 'short_answer')),
    points INT DEFAULT 1 CHECK (points > 0 AND points <= 100),
    order_index INT NOT NULL CHECK (order_index >= 0),
    explanation TEXT,
    meta JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_questions_quiz_id ON questions(quiz_id);
CREATE INDEX IF NOT EXISTS idx_questions_order ON questions(quiz_id, order_index);
CREATE INDEX IF NOT EXISTS idx_questions_type ON questions(question_type);

-- Unique constraint: ensure order_index is unique per quiz
CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_quiz_order ON questions(quiz_id, order_index);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_questions_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_questions_updated_at
    BEFORE UPDATE ON questions
    FOR EACH ROW
    EXECUTE FUNCTION update_questions_updated_at();

-- Comments for documentation
COMMENT ON TABLE questions IS 'Stores questions belonging to quizzes';
COMMENT ON COLUMN questions.question_type IS 'Type of question: multiple_choice, true_false, or short_answer';
COMMENT ON COLUMN questions.points IS 'Points awarded for correct answer (1-100)';
COMMENT ON COLUMN questions.order_index IS 'Display order within quiz (0-based)';
COMMENT ON COLUMN questions.explanation IS 'Explanation shown after answering, can be NULL';
COMMENT ON COLUMN questions.meta IS 'Additional metadata stored as JSON';
