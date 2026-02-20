-- Enable UUID extension (v7 will be generated in Go, but we need the extension for the type)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create categories table
CREATE TABLE categories (
    id          uuid PRIMARY KEY,
    name        text NOT NULL,
    description text NOT NULL,
    status      varchar(20) NOT NULL DEFAULT 'Inactive',
    parent_id   uuid,
    parent_name text,
    created_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at  timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    
    -- Check constraint for status
    CONSTRAINT status_check CHECK (status IN ('Active', 'Inactive')),
    
    -- Self-referencing foreign key
    CONSTRAINT fk_parent FOREIGN KEY (parent_id) 
        REFERENCES categories(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_categories_parent_id ON categories(parent_id);
CREATE INDEX idx_categories_status ON categories(status);
CREATE INDEX idx_categories_created_at ON categories(created_at);

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at on row update
CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON categories
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Function to update children's parent_name when parent name changes
CREATE OR REPLACE FUNCTION update_children_parent_name()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE categories 
    SET parent_name = NEW.name 
    WHERE parent_id = NEW.id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to update children when parent name changes
CREATE TRIGGER trg_update_parent_name
    AFTER UPDATE ON categories
    FOR EACH ROW
    WHEN (OLD.name IS DISTINCT FROM NEW.name)
    EXECUTE FUNCTION update_children_parent_name();