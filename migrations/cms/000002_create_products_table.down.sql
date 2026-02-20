-- Drop trigger first
DROP TRIGGER IF EXISTS trg_products_set_updated_at ON products;

-- Drop table (CASCADE will drop dependent constraints)
DROP TABLE IF EXISTS products CASCADE;