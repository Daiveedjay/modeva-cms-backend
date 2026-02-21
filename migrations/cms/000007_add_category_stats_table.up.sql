-- Create category_stats table
CREATE TABLE category_stats (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    total_categories int NOT NULL DEFAULT 0,
    parent_categories int NOT NULL DEFAULT 0,
    sub_categories int NOT NULL DEFAULT 0,
    active_categories int NOT NULL DEFAULT 0,
    active_parent_categories int NOT NULL DEFAULT 0,
    active_sub_categories int NOT NULL DEFAULT 0,
    updated_at timestamp NOT NULL DEFAULT NOW()
);

-- Insert initial row (there will only ever be one row)
INSERT INTO category_stats (id) VALUES (gen_random_uuid());

-- Function to recalculate all stats
CREATE OR REPLACE FUNCTION recalculate_category_stats()
RETURNS void AS $$
BEGIN
    UPDATE category_stats SET
        total_categories = (SELECT COUNT(*) FROM categories),
        parent_categories = (SELECT COUNT(*) FROM categories WHERE parent_id IS NULL),
        sub_categories = (SELECT COUNT(*) FROM categories WHERE parent_id IS NOT NULL),
        active_categories = (SELECT COUNT(*) FROM categories WHERE status = 'Active'),
        active_parent_categories = (SELECT COUNT(*) FROM categories WHERE parent_id IS NULL AND status = 'Active'),
        active_sub_categories = (SELECT COUNT(*) FROM categories WHERE parent_id IS NOT NULL AND status = 'Active'),
        updated_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Trigger function that updates stats whenever categories change
CREATE OR REPLACE FUNCTION update_category_stats_trigger()
RETURNS TRIGGER AS $$
BEGIN
    PERFORM recalculate_category_stats();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger on INSERT
CREATE TRIGGER trigger_category_stats_insert
    AFTER INSERT ON categories
    FOR EACH STATEMENT
    EXECUTE FUNCTION update_category_stats_trigger();

-- Trigger on UPDATE
CREATE TRIGGER trigger_category_stats_update
    AFTER UPDATE ON categories
    FOR EACH STATEMENT
    EXECUTE FUNCTION update_category_stats_trigger();

-- Trigger on DELETE
CREATE TRIGGER trigger_category_stats_delete
    AFTER DELETE ON categories
    FOR EACH STATEMENT
    EXECUTE FUNCTION update_category_stats_trigger();

-- Calculate initial stats
SELECT recalculate_category_stats();