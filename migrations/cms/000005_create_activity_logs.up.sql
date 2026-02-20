-- Migration: Create activity_logs table
-- Up: Create table for tracking admin actions

CREATE TABLE activity_logs (
    id UUID PRIMARY KEY,
    admin_id UUID NOT NULL,
    admin_email VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id VARCHAR(255) NOT NULL,
    resource_name VARCHAR(255),
    changes JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'success',
    error_message TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for efficient querying
CREATE INDEX idx_activity_admin_date ON activity_logs(admin_id DESC, created_at DESC);
CREATE INDEX idx_activity_resource_date ON activity_logs(resource_type DESC, created_at DESC);
CREATE INDEX idx_activity_action ON activity_logs(action);
CREATE INDEX idx_activity_resource_id ON activity_logs(resource_id);
CREATE INDEX idx_activity_created_at ON activity_logs(created_at DESC);

-- Full-text search on action and resource_type
CREATE INDEX idx_activity_action_text ON activity_logs USING GIN(to_tsvector('english', action || ' ' || resource_type));