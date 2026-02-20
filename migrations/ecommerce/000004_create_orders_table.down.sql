DROP TRIGGER IF EXISTS trigger_set_order_number ON orders;
DROP TRIGGER IF EXISTS trigger_set_updated_at ON orders;
DROP FUNCTION IF EXISTS set_order_number();
DROP TABLE IF EXISTS orders CASCADE;