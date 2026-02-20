-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table
CREATE TABLE users (
    id               uuid PRIMARY KEY,
    email            varchar(255) NOT NULL UNIQUE,
    name             varchar(255) NOT NULL,
    google_id        varchar(255) NOT NULL UNIQUE,
    provider         varchar(50) DEFAULT 'google',
    phone            varchar(50),
    status           varchar(50) DEFAULT 'active',
    email_verified   boolean DEFAULT true,
    avatar           text,
    created_at       timestamp without time zone NOT NULL DEFAULT now(),
    updated_at       timestamp without time zone NOT NULL DEFAULT now(),
    ban_reason       text,
    suspended_until  timestamp without time zone,
    suspended_reason text,
    
    -- Constraints
    CONSTRAINT users_provider_check CHECK (provider = 'google'),
    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'deleted', 'banned'))
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_google_id ON users(google_id);
CREATE INDEX idx_users_phone ON users(phone) WHERE phone IS NOT NULL;
CREATE INDEX idx_users_provider ON users(provider);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Auto-update updated_at trigger
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();