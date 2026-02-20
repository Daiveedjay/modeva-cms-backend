-- Create addresses table
CREATE TABLE addresses (
    id         uuid PRIMARY KEY,
    user_id    uuid NOT NULL,
    label      varchar(50) NOT NULL,
    first_name varchar(100) NOT NULL,
    last_name  varchar(100) NOT NULL,
    street     varchar(255) NOT NULL,
    city       varchar(100) NOT NULL,
    state      varchar(100) NOT NULL,
    zip        varchar(20) NOT NULL,
    country    varchar(100) NOT NULL,
    phone      varchar(20),
    is_default boolean DEFAULT false,
    status     varchar(20) DEFAULT 'active',
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now(),
    
    -- Foreign key
    CONSTRAINT addresses_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT addresses_status_check CHECK (status IN ('active', 'deleted'))
);

-- Create indexes
CREATE INDEX idx_addresses_user_id ON addresses(user_id);
CREATE INDEX idx_addresses_status ON addresses(status);
CREATE INDEX idx_addresses_is_default ON addresses(is_default) WHERE is_default = true;

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON addresses
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Function to ensure only one default address per user
CREATE OR REPLACE FUNCTION ensure_single_default_address()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_default = true THEN
        UPDATE addresses 
        SET is_default = false 
        WHERE user_id = NEW.user_id 
          AND id != NEW.id 
          AND is_default = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER ensure_single_default_address_trigger
    BEFORE INSERT OR UPDATE ON addresses
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_address();