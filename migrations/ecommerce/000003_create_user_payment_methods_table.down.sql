DROP TRIGGER IF EXISTS trigger_single_default_payment_method ON user_payment_methods;
DROP TRIGGER IF EXISTS trigger_set_updated_at ON user_payment_methods;
DROP FUNCTION IF EXISTS ensure_single_default_payment_method();
DROP TABLE IF EXISTS user_payment_methods CASCADE;