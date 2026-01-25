-- Add last_login_at column to users table (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'last_login_at'
    ) THEN
        ALTER TABLE users ADD COLUMN last_login_at TIMESTAMP WITH TIME ZONE;
    END IF;
END $$;

-- Create index for efficient last_login queries (idempotent)
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON users(last_login_at DESC) WHERE deleted_at IS NULL;
