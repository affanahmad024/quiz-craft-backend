BEGIN;

-- Update role constraint to include 'educator'
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;
ALTER TABLE users ADD CONSTRAINT chk_user_role
  CHECK (role IN ('user', 'admin', 'educator'));

COMMENT ON CONSTRAINT chk_user_role ON users IS 'Allowed user roles: user, admin, educator';

COMMIT;
