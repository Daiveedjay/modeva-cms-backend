-- Create orders table
CREATE TABLE orders (
    id                   uuid PRIMARY KEY,
    user_id              uuid NOT NULL,
    order_number         varchar(50) NOT NULL UNIQUE,
    payment_method_id    uuid,
    address_id           uuid,
    payment_method_type  varchar(20),
    payment_method_last4 varchar(4),
    address_snapshot     jsonb,
    subtotal             numeric(10,2) NOT NULL,
    tax                  numeric(10,2) DEFAULT 0.00,
    shipping_cost        numeric(10,2) DEFAULT 0.00,
    discount             numeric(10,2) DEFAULT 0.00,
    total_amount         numeric(10,2) NOT NULL,
    status               varchar(50) DEFAULT 'pending',
    customer_notes       text,
    admin_notes          text,
    created_at           timestamp without time zone NOT NULL DEFAULT now(),
    updated_at           timestamp without time zone NOT NULL DEFAULT now(),
    confirmed_at         timestamp without time zone,
    shipped_at           timestamp without time zone,
    delivered_at         timestamp without time zone,
    
    -- Foreign keys
    CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE RESTRICT,
    CONSTRAINT orders_address_id_fkey FOREIGN KEY (address_id) 
        REFERENCES addresses(id) ON DELETE SET NULL,
    CONSTRAINT orders_payment_method_id_fkey FOREIGN KEY (payment_method_id) 
        REFERENCES user_payment_methods(id) ON DELETE SET NULL,
    
    -- Check constraints
    CONSTRAINT orders_subtotal_check CHECK (subtotal >= 0),
    CONSTRAINT orders_tax_check CHECK (tax >= 0),
    CONSTRAINT orders_shipping_cost_check CHECK (shipping_cost >= 0),
    CONSTRAINT orders_discount_check CHECK (discount >= 0),
    CONSTRAINT orders_total_amount_check CHECK (total_amount >= 0),
    CONSTRAINT orders_status_check CHECK (status IN ('pending', 'processing', 'shipped', 'completed', 'cancelled'))
);

-- Create indexes
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at DESC);

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON orders
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- Function to generate order number (ORD-YYYY-NNNNNN)
CREATE OR REPLACE FUNCTION set_order_number()
RETURNS TRIGGER AS $$
DECLARE
    year_prefix TEXT;
    next_number INT;
    new_order_number TEXT;
BEGIN
    -- Get current year
    year_prefix := TO_CHAR(NOW(), 'YYYY');
    
    -- Get the highest number for this year
    SELECT COALESCE(MAX(CAST(SUBSTRING(order_number FROM 10) AS INT)), 0) + 1
    INTO next_number
    FROM orders
    WHERE order_number LIKE 'ORD-' || year_prefix || '-%';
    
    -- Generate new order number: ORD-2025-000001
    new_order_number := 'ORD-' || year_prefix || '-' || LPAD(next_number::TEXT, 6, '0');
    
    NEW.order_number := new_order_number;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_set_order_number
    BEFORE INSERT ON orders
    FOR EACH ROW
    EXECUTE FUNCTION set_order_number();