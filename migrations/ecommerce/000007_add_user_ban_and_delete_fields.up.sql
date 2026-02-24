-- Add missing ban and delete tracking fields
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_banned boolean DEFAULT false;
ALTER TABLE users ADD COLUMN IF NOT EXISTS banned_at timestamp;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at timestamp;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deletion_reason text;

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_is_banned ON users(is_banned) WHERE is_banned = true;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_banned_at ON users(banned_at) WHERE banned_at IS NOT NULL;

-- Add comment for clarity
COMMENT ON COLUMN users.is_banned IS 'True if user is permanently banned';
COMMENT ON COLUMN users.suspended_until IS 'Temporary suspension end date';
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp';