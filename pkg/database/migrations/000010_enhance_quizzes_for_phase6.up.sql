-- Phase 6: Advanced Features - Quiz Enhancements
-- Add sharing, templates, and view tracking capabilities

-- Add sharing columns
ALTER TABLE quizzes ADD COLUMN share_code VARCHAR(20) UNIQUE;
ALTER TABLE quizzes ADD COLUMN share_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN view_count INT DEFAULT 0;

-- Add template columns
ALTER TABLE quizzes ADD COLUMN is_template BOOLEAN DEFAULT false;
ALTER TABLE quizzes ADD COLUMN template_usage_count INT DEFAULT 0;
ALTER TABLE quizzes ADD COLUMN based_on_template_id UUID REFERENCES quizzes(id) ON DELETE SET NULL;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_quizzes_share_code ON quizzes(share_code) WHERE share_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_is_template ON quizzes(is_template) WHERE is_template = true AND deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_based_on_template ON quizzes(based_on_template_id) WHERE based_on_template_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_quizzes_view_count ON quizzes(view_count DESC) WHERE deleted_at IS NULL;

-- Create full-text search index on title and description
CREATE INDEX IF NOT EXISTS idx_quizzes_search ON quizzes
USING GIN(to_tsvector('english', COALESCE(title, '') || ' ' || COALESCE(description, '')))
WHERE deleted_at IS NULL;

-- Comments for documentation
COMMENT ON COLUMN quizzes.share_code IS 'Unique code for sharing quiz via link';
COMMENT ON COLUMN quizzes.share_count IS 'Number of times quiz has been shared';
COMMENT ON COLUMN quizzes.view_count IS 'Number of times quiz has been viewed';
COMMENT ON COLUMN quizzes.is_template IS 'Whether quiz is a reusable template';
COMMENT ON COLUMN quizzes.template_usage_count IS 'Number of times template has been used';
COMMENT ON COLUMN quizzes.based_on_template_id IS 'Reference to template this quiz was created from';
