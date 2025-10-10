-- Add columns for manifest metadata
ALTER TABLE charts ADD COLUMN ingress_paths TEXT; -- JSON array of ingress paths
ALTER TABLE charts ADD COLUMN container_images TEXT; -- JSON array of container images
ALTER TABLE charts ADD COLUMN service_ports TEXT; -- JSON array of service ports
ALTER TABLE charts ADD COLUMN manifest_parsed_at DATETIME; -- When manifest was last parsed

-- Create index for faster manifest queries
CREATE INDEX IF NOT EXISTS idx_charts_manifest_parsed ON charts(manifest_parsed_at);