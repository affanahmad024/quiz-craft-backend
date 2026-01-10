-- Create uploaded_files table for Phase 4: File Upload
CREATE TABLE uploaded_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) DEFAULT 'uploaded',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT fk_uploaded_files_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX idx_uploaded_files_user_id ON uploaded_files(user_id);
CREATE INDEX idx_uploaded_files_status ON uploaded_files(status);
CREATE INDEX idx_uploaded_files_expires_at ON uploaded_files(expires_at) WHERE status = 'processed';

-- Add comments for documentation
COMMENT ON TABLE uploaded_files IS 'Stores metadata for user-uploaded files for AI quiz generation';
COMMENT ON COLUMN uploaded_files.filename IS 'Generated unique filename with timestamp';
COMMENT ON COLUMN uploaded_files.original_filename IS 'User provided original filename';
COMMENT ON COLUMN uploaded_files.file_path IS 'Full filesystem path to stored file';
COMMENT ON COLUMN uploaded_files.file_size IS 'File size in bytes';
COMMENT ON COLUMN uploaded_files.status IS 'Status: uploaded, processing, processed, failed';
COMMENT ON COLUMN uploaded_files.expires_at IS 'Expiration timestamp for file cleanup (7 days retention)';
