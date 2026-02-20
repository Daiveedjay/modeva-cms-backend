-- Migration Down: Remove device_type column from orders

-- Drop index first
DROP INDEX IF EXISTS idx_orders_device_type;

-- Remove device_type column
ALTER TABLE orders DROP COLUMN device_type;