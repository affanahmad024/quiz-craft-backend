-- Drop existing table if it exists (clean slate for development)
DROP TABLE IF EXISTS user_passwords CASCADE;

-- Create user_passwords table with modern JSONB design for security metadata
CREATE TABLE user_passwords (
    -- Primary identifiers
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Core authentication
    password_hash VARCHAR(255) NOT NULL,
    password_history JSONB DEFAULT '[]'::jsonb,

    -- Token management (queryable columns for fast lookups)
    verification_token VARCHAR(255),
    verification_token_expiry TIMESTAMP WITH TIME ZONE,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP WITH TIME ZONE,

    -- Two-factor authentication
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(255),

    -- Security metadata (JSONB - flexible for brute-force protection & future extensions)
    -- Contains: failed_login_attempts, last_failed_login, account_locked_until,
    --           last_password_change, must_change_password
    meta JSONB DEFAULT '{
        "failed_login_attempts": 0,
        "last_failed_login": null,
        "account_locked_until": null,
        "last_password_change": null,
        "must_change_password": false
    }'::jsonb,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance optimization

-- Standard user_id lookup
CREATE INDEX IF NOT EXISTS idx_user_passwords_user_id ON user_passwords(user_id);

-- Token lookups (for email verification and password reset)
CREATE INDEX IF NOT EXISTS idx_user_passwords_verification_token
    ON user_passwords(verification_token)
    WHERE verification_token IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_user_passwords_reset_token
    ON user_passwords(reset_token)
    WHERE reset_token IS NOT NULL;

-- JSONB indexes for security queries (brute-force protection)
-- GIN index for general JSONB queries
CREATE INDEX IF NOT EXISTS idx_user_passwords_meta_gin
    ON user_passwords USING gin(meta);

-- Specific index for locked accounts (fast lookup during login)
CREATE INDEX IF NOT EXISTS idx_user_passwords_meta_locked
    ON user_passwords ((meta->>'account_locked_until'))
    WHERE (meta->>'account_locked_until') IS NOT NULL;

-- Index for failed login attempts (security monitoring)
CREATE INDEX IF NOT EXISTS idx_user_passwords_meta_failed_attempts
    ON user_passwords (((meta->>'failed_login_attempts')::int))
    WHERE (meta->>'failed_login_attempts')::int > 0;

-- Apply trigger to user_passwords table
CREATE TRIGGER update_user_passwords_updated_at
    BEFORE UPDATE ON user_passwords
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
