DROP TRIGGER IF EXISTS ensure_single_default_address_trigger ON addresses;
DROP TRIGGER IF EXISTS trigger_set_updated_at ON addresses;
DROP FUNCTION IF EXISTS ensure_single_default_address();
DROP TABLE IF EXISTS addresses CASCADE;