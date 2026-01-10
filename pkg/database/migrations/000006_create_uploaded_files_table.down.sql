-- Drop uploaded_files table
DROP INDEX IF EXISTS idx_uploaded_files_expires_at;
DROP INDEX IF EXISTS idx_uploaded_files_status;
DROP INDEX IF EXISTS idx_uploaded_files_user_id;
DROP TABLE IF EXISTS uploaded_files;
