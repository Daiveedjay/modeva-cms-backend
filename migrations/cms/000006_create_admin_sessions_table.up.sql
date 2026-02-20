-- Create admin_sessions table for session tracking
CREATE TABLE admin_sessions (
  id UUID PRIMARY KEY,
  admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
  token_hash VARCHAR NOT NULL UNIQUE,
  ip_address VARCHAR,
  user_agent TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  last_activity_at TIMESTAMP NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMP NOT NULL,
  is_active BOOLEAN NOT NULL DEFAULT true
);

-- Create indexes for efficient queries
CREATE INDEX idx_admin_sessions_admin_id ON admin_sessions(admin_id);
CREATE INDEX idx_admin_sessions_token_hash ON admin_sessions(token_hash);
CREATE INDEX idx_admin_sessions_is_active ON admin_sessions(is_active);
CREATE INDEX idx_admin_sessions_last_activity_at ON admin_sessions(last_activity_at);
CREATE INDEX idx_admin_sessions_expires_at ON admin_sessions(expires_at);

-- Create a composite index for the most common query
CREATE INDEX idx_admin_sessions_active_valid ON admin_sessions(is_active, expires_at) 
WHERE is_active = true;