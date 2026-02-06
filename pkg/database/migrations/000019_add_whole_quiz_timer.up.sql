-- Add timer mode column
ALTER TABLE live_sessions
ADD COLUMN timer_mode VARCHAR(20) NOT NULL DEFAULT 'per_question'
    CHECK (timer_mode IN ('per_question', 'whole_quiz'));

-- Add whole quiz timer duration column
ALTER TABLE live_sessions
ADD COLUMN whole_quiz_timer_seconds INT;

-- Add quiz started timestamp for whole-quiz timer tracking
ALTER TABLE live_sessions
ADD COLUMN quiz_started_at TIMESTAMP;

-- Add constraint: if timer_mode is 'whole_quiz', then whole_quiz_timer_seconds must be set
ALTER TABLE live_sessions
ADD CONSTRAINT check_whole_quiz_timer
CHECK (
    (timer_mode = 'per_question' AND question_timer_seconds IS NOT NULL AND question_timer_seconds > 0) OR
    (timer_mode = 'whole_quiz' AND whole_quiz_timer_seconds IS NOT NULL AND whole_quiz_timer_seconds > 0)
);
