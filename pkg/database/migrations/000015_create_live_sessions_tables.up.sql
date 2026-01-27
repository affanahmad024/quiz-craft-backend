-- Create live_sessions table (unified for public and private sessions)
CREATE TABLE IF NOT EXISTS live_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    quiz_id UUID NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Session type determines behavior
    session_type VARCHAR(20) NOT NULL CHECK (session_type IN ('public', 'private')),

    -- Basic info
    title VARCHAR(255) NOT NULL,
    description TEXT,

    -- Timing (flexible)
    scheduled_at TIMESTAMP WITH TIME ZONE,  -- NULL = instant, NOT NULL = scheduled
    started_at TIMESTAMP WITH TIME ZONE,
    ended_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(50) DEFAULT 'waiting' CHECK (status IN ('waiting', 'live', 'completed', 'cancelled')),

    -- Access control
    access_code VARCHAR(10),  -- Generated for private sessions
    is_discoverable BOOLEAN DEFAULT false,  -- Public listing?
    max_participants INT,

    -- Progression mode
    progression_mode VARCHAR(20) NOT NULL CHECK (progression_mode IN ('manual', 'automatic')),
    question_timer_seconds INT DEFAULT 30,  -- For automatic mode
    current_question_index INT DEFAULT 0,
    current_question_started_at TIMESTAMP WITH TIME ZONE,  -- For timer tracking

    -- Config (flexible for future features)
    config JSONB DEFAULT '{}'::jsonb,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Business rules
    CONSTRAINT valid_public_session CHECK (
        session_type = 'private' OR
        (session_type = 'public' AND scheduled_at IS NOT NULL)
    ),
    CONSTRAINT valid_private_session CHECK (
        session_type = 'public' OR
        (session_type = 'private' AND access_code IS NOT NULL)
    ),
    CONSTRAINT valid_automatic_mode CHECK (
        progression_mode = 'manual' OR
        (progression_mode = 'automatic' AND question_timer_seconds > 0)
    )
);

-- Create live_session_participants table
CREATE TABLE IF NOT EXISTS live_session_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,

    current_score INT DEFAULT 0,
    rank INT DEFAULT 0,

    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'disconnected', 'completed')),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (session_id, user_id)
);

-- Create live_session_answers table
CREATE TABLE IF NOT EXISTS live_session_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    participant_id UUID NOT NULL REFERENCES live_session_participants(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES answers(id) ON DELETE SET NULL,  -- NULL if no answer selected

    is_correct BOOLEAN NOT NULL,
    points_earned INT DEFAULT 0,
    time_taken_ms INT NOT NULL,  -- Milliseconds taken to answer

    answered_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (session_id, participant_id, question_id)
);

-- Create live_session_invitations table (for private sessions)
CREATE TABLE IF NOT EXISTS live_session_invitations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL REFERENCES live_sessions(id) ON DELETE CASCADE,
    inviter_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),

    invited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    responded_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    UNIQUE (session_id, invitee_id)
);

-- Create indexes for live_sessions
CREATE INDEX IF NOT EXISTS idx_live_sessions_creator ON live_sessions(creator_id);
CREATE INDEX IF NOT EXISTS idx_live_sessions_quiz ON live_sessions(quiz_id);
CREATE INDEX IF NOT EXISTS idx_live_sessions_status ON live_sessions(status);
CREATE INDEX IF NOT EXISTS idx_live_sessions_scheduled ON live_sessions(scheduled_at) WHERE scheduled_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_live_sessions_discoverable ON live_sessions(is_discoverable, status) WHERE is_discoverable = true;
CREATE INDEX IF NOT EXISTS idx_live_sessions_access_code ON live_sessions(access_code) WHERE access_code IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_live_sessions_type_status ON live_sessions(session_type, status);

-- Create indexes for live_session_participants
CREATE INDEX IF NOT EXISTS idx_live_participants_session ON live_session_participants(session_id);
CREATE INDEX IF NOT EXISTS idx_live_participants_user ON live_session_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_live_participants_score ON live_session_participants(session_id, current_score DESC);
CREATE INDEX IF NOT EXISTS idx_live_participants_status ON live_session_participants(session_id, status);

-- Create indexes for live_session_answers
CREATE INDEX IF NOT EXISTS idx_live_answers_session ON live_session_answers(session_id);
CREATE INDEX IF NOT EXISTS idx_live_answers_participant ON live_session_answers(participant_id);
CREATE INDEX IF NOT EXISTS idx_live_answers_question ON live_session_answers(question_id);

-- Create indexes for live_session_invitations
CREATE INDEX IF NOT EXISTS idx_live_invitations_session ON live_session_invitations(session_id);
CREATE INDEX IF NOT EXISTS idx_live_invitations_invitee ON live_session_invitations(invitee_id, status);
