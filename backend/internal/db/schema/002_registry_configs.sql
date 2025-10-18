-- Registry configurations table
CREATE TABLE IF NOT EXISTS registry_configs (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    registry_url TEXT NOT NULL,
    username TEXT,
    password TEXT,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_registry_configs_name ON registry_configs(name);
CREATE INDEX IF NOT EXISTS idx_registry_configs_default ON registry_configs(is_default);

-- Insert a default Docker Hub config
INSERT INTO registry_configs (name, registry_url, is_default) 
VALUES ('docker-hub', 'https://registry-1.docker.io', TRUE)
ON CONFLICT (name) DO NOTHING;