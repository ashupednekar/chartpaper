-- Add image and canary tag columns to dependencies table
ALTER TABLE dependencies ADD COLUMN image_tag TEXT DEFAULT 'N/A';
ALTER TABLE dependencies ADD COLUMN canary_tag TEXT DEFAULT 'N/A';