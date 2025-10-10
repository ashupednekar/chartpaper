-- Add version history support to existing charts table
ALTER TABLE charts ADD COLUMN is_latest BOOLEAN DEFAULT TRUE;

-- Remove the unique constraint on name only, allow multiple versions
-- Note: SQLite doesn't support dropping constraints directly, so we'll work with the existing structure

-- Update existing charts to be marked as latest
UPDATE charts SET is_latest = TRUE WHERE is_latest IS NULL;

-- Create index for better performance on version queries
CREATE INDEX IF NOT EXISTS idx_charts_name_version ON charts(name, version);
CREATE INDEX IF NOT EXISTS idx_charts_latest ON charts(is_latest) WHERE is_latest = TRUE;