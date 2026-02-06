-- Drop constraint first
ALTER TABLE live_sessions DROP CONSTRAINT IF EXISTS check_whole_quiz_timer;

-- Drop columns
ALTER TABLE live_sessions DROP COLUMN IF EXISTS timer_mode;
ALTER TABLE live_sessions DROP COLUMN IF EXISTS whole_quiz_timer_seconds;
ALTER TABLE live_sessions DROP COLUMN IF EXISTS quiz_started_at;
