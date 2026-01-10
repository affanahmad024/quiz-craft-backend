-- Drop indexes
DROP INDEX IF EXISTS idx_activity_logs_user_action;
DROP INDEX IF EXISTS idx_activity_logs_created_at;
DROP INDEX IF EXISTS idx_activity_logs_resource;
DROP INDEX IF EXISTS idx_activity_logs_action;
DROP INDEX IF EXISTS idx_activity_logs_user_id;

-- Drop table
DROP TABLE IF EXISTS activity_logs;
