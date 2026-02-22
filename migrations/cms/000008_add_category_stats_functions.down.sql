DROP TRIGGER IF EXISTS category_stats_trigger ON categories;
DROP FUNCTION IF EXISTS update_category_stats_trigger();
DROP FUNCTION IF EXISTS recalculate_category_stats();