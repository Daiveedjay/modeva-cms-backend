-- Remove indexes
DROP INDEX IF EXISTS idx_users_banned_at;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_is_banned;

-- Remove columns
ALTER TABLE users DROP COLUMN IF EXISTS deletion_reason;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS banned_at;
ALTER TABLE users DROP COLUMN IF EXISTS is_banned;