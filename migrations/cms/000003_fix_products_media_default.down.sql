-- Revert media column default back to array (for rollback)
ALTER TABLE products 
ALTER COLUMN media SET DEFAULT '[]'::jsonb;

-- Note: We don't revert existing data as that would be destructive