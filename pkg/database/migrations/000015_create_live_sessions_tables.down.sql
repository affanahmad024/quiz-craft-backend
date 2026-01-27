-- Drop indexes for live_session_invitations
DROP INDEX IF EXISTS idx_live_invitations_invitee;
DROP INDEX IF EXISTS idx_live_invitations_session;

-- Drop indexes for live_session_answers
DROP INDEX IF EXISTS idx_live_answers_question;
DROP INDEX IF EXISTS idx_live_answers_participant;
DROP INDEX IF EXISTS idx_live_answers_session;

-- Drop indexes for live_session_participants
DROP INDEX IF EXISTS idx_live_participants_status;
DROP INDEX IF EXISTS idx_live_participants_score;
DROP INDEX IF EXISTS idx_live_participants_user;
DROP INDEX IF EXISTS idx_live_participants_session;

-- Drop indexes for live_sessions
DROP INDEX IF EXISTS idx_live_sessions_type_status;
DROP INDEX IF EXISTS idx_live_sessions_access_code;
DROP INDEX IF EXISTS idx_live_sessions_discoverable;
DROP INDEX IF EXISTS idx_live_sessions_scheduled;
DROP INDEX IF EXISTS idx_live_sessions_status;
DROP INDEX IF EXISTS idx_live_sessions_quiz;
DROP INDEX IF EXISTS idx_live_sessions_creator;

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS live_session_invitations;
DROP TABLE IF EXISTS live_session_answers;
DROP TABLE IF EXISTS live_session_participants;
DROP TABLE IF EXISTS live_sessions;
