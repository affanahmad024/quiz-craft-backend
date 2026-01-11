-- Add role column to users table for admin access control (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user' NOT NULL;
    END IF;
END $$;

-- Add check constraint to ensure only valid roles (idempotent)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.constraint_column_usage
        WHERE constraint_name = 'chk_user_role'
    ) THEN
        ALTER TABLE users ADD CONSTRAINT chk_user_role CHECK (role IN ('user', 'admin'));
    END IF;
END $$;

-- Create index for efficient role-based queries (idempotent)
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role) WHERE deleted_at IS NULL;

-- Optional: Set initial admin user (update email to match your admin account)
-- UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
