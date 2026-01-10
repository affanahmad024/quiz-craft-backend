-- Add role column to users table for admin access control
ALTER TABLE users ADD COLUMN role VARCHAR(50) DEFAULT 'user' NOT NULL;

-- Add check constraint to ensure only valid roles
ALTER TABLE users ADD CONSTRAINT chk_user_role CHECK (role IN ('user', 'admin'));

-- Create index for efficient role-based queries
CREATE INDEX idx_users_role ON users(role) WHERE deleted_at IS NULL;

-- Optional: Set initial admin user (update email to match your admin account)
-- UPDATE users SET role = 'admin' WHERE email = 'admin@example.com';
