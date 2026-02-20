-- Migration: Add device_type column to orders
-- Up: Add the column with default value 'unknown'
-- Down: Remove the column

-- Add device_type column
ALTER TABLE orders ADD COLUMN device_type VARCHAR(20) DEFAULT 'desktop';

-- Create index for analytics queries
CREATE INDEX idx_orders_device_type ON orders(device_type);