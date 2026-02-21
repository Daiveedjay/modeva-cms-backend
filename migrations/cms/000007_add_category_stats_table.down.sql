-- Drop triggers
DROP TRIGGER IF EXISTS trigger_category_stats_insert ON categories;
DROP TRIGGER IF EXISTS trigger_category_stats_update ON categories;
DROP TRIGGER IF EXISTS trigger_category_stats_delete ON categories;

-- Drop functions
DROP FUNCTION IF EXISTS update_category_stats_trigger();
DROP FUNCTION IF EXISTS recalculate_category_stats();

-- Drop table
DROP TABLE IF EXISTS category_stats;