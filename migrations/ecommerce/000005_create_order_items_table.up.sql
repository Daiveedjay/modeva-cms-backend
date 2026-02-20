-- Create order_items table
CREATE TABLE order_items (
    id            uuid PRIMARY KEY,
    order_id      uuid NOT NULL,
    user_id       uuid NOT NULL,
    product_id    uuid NOT NULL,
    product_name  varchar(255) NOT NULL,
    variant_size  varchar(50),
    variant_color varchar(50),
    price         numeric(10,2) NOT NULL,
    quantity      integer NOT NULL,
    subtotal      numeric(10,2) NOT NULL,
    status        varchar(50) DEFAULT 'pending',
    created_at    timestamp without time zone NOT NULL DEFAULT now(),
    updated_at    timestamp without time zone NOT NULL DEFAULT now(),
    
    -- Foreign keys
    CONSTRAINT order_items_user_id_fkey FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE RESTRICT,
    -- Note: No FK to orders table to allow flexibility
    
    -- Check constraints
    CONSTRAINT order_items_quantity_check CHECK (quantity > 0),
    CONSTRAINT order_items_status_check CHECK (status IN ('pending', 'confirmed', 'shipped', 'delivered', 'cancelled', 'refunded'))
);

-- Create indexes
CREATE INDEX idx_order_items_order_id ON order_items(order_id);
CREATE INDEX idx_order_items_user_id ON order_items(user_id);
CREATE INDEX idx_order_items_product_id ON order_items(product_id);
CREATE INDEX idx_order_items_status ON order_items(status);

-- Trigger to auto-update updated_at
CREATE TRIGGER trigger_set_updated_at
    BEFORE UPDATE ON order_items
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();