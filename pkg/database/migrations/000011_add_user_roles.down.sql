-- Drop index
DROP INDEX IF EXISTS idx_users_role;

-- Drop constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;

-- Drop role column
ALTER TABLE users DROP COLUMN IF EXISTS role;
