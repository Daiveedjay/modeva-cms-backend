-- Migration Down: Drop admin tables

DROP INDEX IF EXISTS idx_admin_invites_expires_at;
DROP INDEX IF EXISTS idx_admin_invites_token_hash;
DROP INDEX IF EXISTS idx_admin_invites_email;
DROP TABLE IF EXISTS admin_invites;

DROP INDEX IF EXISTS idx_admins_joined_at;
DROP INDEX IF EXISTS idx_admins_role;
DROP INDEX IF EXISTS idx_admins_status;
DROP INDEX IF EXISTS idx_admins_email;
DROP TABLE IF EXISTS admins;