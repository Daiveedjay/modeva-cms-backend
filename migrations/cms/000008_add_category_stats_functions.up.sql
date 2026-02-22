-- Create the function that recalculates category stats
CREATE OR REPLACE FUNCTION recalculate_category_stats()
RETURNS void AS $$
BEGIN
  UPDATE category_stats SET
    total_categories          = (SELECT COUNT(*) FROM categories),
    parent_categories         = (SELECT COUNT(*) FROM categories WHERE parent_id IS NULL),
    sub_categories            = (SELECT COUNT(*) FROM categories WHERE parent_id IS NOT NULL),
    active_categories         = (SELECT COUNT(*) FROM categories WHERE status = 'Active'),
    active_parent_categories  = (SELECT COUNT(*) FROM categories WHERE parent_id IS NULL AND status = 'Active'),
    active_sub_categories     = (SELECT COUNT(*) FROM categories WHERE parent_id IS NOT NULL AND status = 'Active'),
    updated_at                = NOW();
END;
$$ LANGUAGE plpgsql;

-- Create the trigger function that calls it
CREATE OR REPLACE FUNCTION update_category_stats_trigger()
RETURNS TRIGGER AS $$
BEGIN
  PERFORM recalculate_category_stats();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Attach the trigger to the categories table
DROP TRIGGER IF EXISTS category_stats_trigger ON categories;
CREATE TRIGGER category_stats_trigger
  AFTER INSERT OR UPDATE OR DELETE ON categories
  FOR EACH STATEMENT
  EXECUTE FUNCTION update_category_stats_trigger();