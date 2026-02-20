-- Drop triggers first
DROP TRIGGER IF EXISTS trg_update_parent_name ON categories;
DROP TRIGGER IF EXISTS trigger_set_updated_at ON categories;

-- Drop functions
DROP FUNCTION IF EXISTS update_children_parent_name();
DROP FUNCTION IF EXISTS set_updated_at();

-- Drop table (CASCADE will drop dependent constraints)
DROP TABLE IF EXISTS categories CASCADE;