-- Create products table
CREATE TABLE products (
    id              uuid PRIMARY KEY,
    name            text NOT NULL,
    description     text NOT NULL,
    composition     jsonb NOT NULL DEFAULT '[]'::jsonb,
    price           numeric(12,2) NOT NULL CHECK (price >= 0),
    sub_category_id uuid NOT NULL,
    status          text NOT NULL CHECK (status IN ('Active', 'Draft')),
    tags            jsonb NOT NULL DEFAULT '[]'::jsonb,
    media           jsonb NOT NULL DEFAULT '[]'::jsonb,
    variants        jsonb NOT NULL DEFAULT '[]'::jsonb,
    inventory       jsonb NOT NULL DEFAULT '[]'::jsonb,
    seo             jsonb NOT NULL DEFAULT '{}'::jsonb,
    views           integer DEFAULT 0,
    created_at      timestamptz NOT NULL DEFAULT now(),
    updated_at      timestamptz NOT NULL DEFAULT now(),
    
    -- Foreign key to categories
    CONSTRAINT fk_products_sub_category 
        FOREIGN KEY (sub_category_id) 
        REFERENCES categories(id) 
        ON DELETE RESTRICT
);

-- Create indexes for better query performance
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_subcategory ON products(sub_category_id);
CREATE INDEX idx_products_tags_gin ON products USING GIN (tags);
CREATE INDEX idx_products_views ON products(views DESC);

-- Check constraint to ensure media is an object (not array)
ALTER TABLE products 
    ADD CONSTRAINT products_media_object 
    CHECK (jsonb_typeof(media) = 'object');

-- Trigger to auto-update updated_at timestamp
CREATE TRIGGER trg_products_set_updated_at
    BEFORE UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();