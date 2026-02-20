-- Fix media column default from array to object
ALTER TABLE products 
ALTER COLUMN media SET DEFAULT '{}'::jsonb;

-- Update any existing rows that have array instead of object
UPDATE products 
SET media = '{"primary": {"url": ""}, "other": []}'::jsonb 
WHERE jsonb_typeof(media) = 'array';

-- Optional: Add a comment explaining the structure
COMMENT ON COLUMN products.media IS 'Product media with structure: {"primary": {"url": "...", "order": 1}, "other": [{"url": "...", "order": 2}]}';