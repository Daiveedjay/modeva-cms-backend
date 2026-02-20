-- Migration: Create admin and admin_invites tables
-- Up: Create tables for admin management

CREATE TABLE admins (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    avatar TEXT,
    phone_number VARCHAR(20),
    country VARCHAR(100),
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    last_login_at TIMESTAMP,
    joined_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CHECK (role IN ('super_admin', 'admin')),
    CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE INDEX idx_admins_email ON admins(email);
CREATE INDEX idx_admins_role ON admins(role);
CREATE INDEX idx_admins_status ON admins(status);
CREATE INDEX idx_admins_joined_at ON admins(joined_at DESC);

CREATE TABLE admin_invites (
    id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    used_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_admin_invites_email ON admin_invites(email);
CREATE INDEX idx_admin_invites_token_hash ON admin_invites(token_hash);
CREATE INDEX idx_admin_invites_expires_at ON admin_invites(expires_at);