-- Migration Down: Drop activity_logs table

DROP INDEX IF EXISTS idx_activity_action_text;
DROP INDEX IF EXISTS idx_activity_created_at;
DROP INDEX IF EXISTS idx_activity_resource_id;
DROP INDEX IF EXISTS idx_activity_action;
DROP INDEX IF EXISTS idx_activity_resource_date;
DROP INDEX IF EXISTS idx_activity_admin_date;
DROP TABLE IF EXISTS activity_logs;