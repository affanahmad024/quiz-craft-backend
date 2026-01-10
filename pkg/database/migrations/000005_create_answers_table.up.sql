-- Create answers table
CREATE TABLE IF NOT EXISTS answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_text TEXT NOT NULL,
    is_correct BOOLEAN DEFAULT false,
    order_index INT NOT NULL CHECK (order_index >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_answers_question_id ON answers(question_id);
CREATE INDEX IF NOT EXISTS idx_answers_order ON answers(question_id, order_index);
CREATE INDEX IF NOT EXISTS idx_answers_is_correct ON answers(question_id, is_correct);

-- Unique constraint: ensure order_index is unique per question
CREATE UNIQUE INDEX IF NOT EXISTS idx_answers_question_order ON answers(question_id, order_index);

-- Trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_answers_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_answers_updated_at
    BEFORE UPDATE ON answers
    FOR EACH ROW
    EXECUTE FUNCTION update_answers_updated_at();

-- Comments for documentation
COMMENT ON TABLE answers IS 'Stores answer options for quiz questions';
COMMENT ON COLUMN answers.is_correct IS 'Whether this answer is correct (can be NULL for survey-style questions)';
COMMENT ON COLUMN answers.order_index IS 'Display order within question (0-based)';
