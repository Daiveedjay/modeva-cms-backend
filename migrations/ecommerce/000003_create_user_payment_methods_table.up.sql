-- Create user_payment_methods table (cards only)
CREATE TABLE user_payment_methods (
    id                         uuid PRIMARY KEY,
    user_id                    uuid NOT NULL,
    type                       varchar(20) NOT NULL DEFAULT 'card',
    is_default                 boolean DEFAULT false,
    
    -- Payment provider info (Stripe, etc.)
    provider                   varchar(50),
    provider_payment_method_id varchar(255),
    
    -- Card details (store encrypted in production!)
    card_type                  varchar(10) NOT NULL, -- 'credit' or 'debit'
    card_brand                 varchar(20) NOT NULL, -- 'visa', 'mastercard', 'amex', etc.
    card_number                varchar(255) NOT NULL, -- Full card number (encrypt in production!)
    exp_month                  integer NOT NULL,
    exp_year                   integer NOT NULL,
    cvv                        varchar(4), -- Optional: store if needed (encrypt!)
    cardholder_name            varchar(255) NOT NULL,
    
    status                     varchar(20) DEFAULT 'active',
    created_at                 timestamp without time zone NOT NULL DEFAULT now(),
    updated_at                 timestamp without time zone NOT NULL DEFAULT now(),
    
    -- Foreign key
    CONSTRAINT user_payment_methods_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
    
    -- Constraints
    CONSTRAINT user_payment_methods_type_check CHECK (type = 'card'),
    CONSTRAINT user_payment_methods_card_type_check CHECK (card_type IN ('credit', 'debit')),
    CONSTRAINT user_payment_methods_status_check CHECK (status IN ('active', 'expired', 'deleted')),
    CONSTRAINT user_payment_methods_exp_month_check CHECK (exp_month >= 1 AND exp_month <= 12),
    CONSTRAINT user_payment_methods_exp_year_check CHECK (exp_year >= EXTRACT(YEAR FROM now()))
);

-- Create indexes
CREATE INDEX idx_user_payment_methods_user_id ON user_payment_methods(user_id);
CREATE INDEX idx_user_payment_methods_status ON user_payment_methods(status);
CREATE INDEX idx_user_payment_methods_is_default ON user_payment_methods(is_default) WHERE is_default = true;

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON user_payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Function to ensure only one default payment method per user
CREATE OR REPLACE FUNCTION ensure_single_default_payment_method()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.is_default = true THEN
        UPDATE user_payment_methods 
        SET is_default = false 
        WHERE user_id = NEW.user_id 
          AND id != NEW.id 
          AND is_default = true;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_single_default_payment_method
    BEFORE INSERT OR UPDATE ON user_payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_payment_method();