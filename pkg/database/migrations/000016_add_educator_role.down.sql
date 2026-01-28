BEGIN;

-- Revert educator role - first demote all educators to users
UPDATE users SET role = 'user' WHERE role = 'educator';

-- Restore original role constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS chk_user_role;
ALTER TABLE users ADD CONSTRAINT chk_user_role
  CHECK (role IN ('user', 'admin'));

COMMIT;
