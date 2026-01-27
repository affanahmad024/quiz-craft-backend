-- Create notifications table for in-app notifications (idempotent)
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Notification type and content
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,

    -- Context data (JSONB for flexibility)
    data JSONB DEFAULT '{}'::jsonb,

    -- Read status
    is_read BOOLEAN DEFAULT false,
    read_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CHECK (type IN (
        'new_quiz',
        'live_session_scheduled',
        'live_session_starting',
        'live_session_invitation',
        'new_follower',
        'system'
    ))
);

-- Create indexes for efficient querying (idempotent)
CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications(user_id, is_read) WHERE is_read = false;
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

-- Create composite index for common filter combinations (idempotent)
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread_created ON notifications(user_id, is_read, created_at DESC);
