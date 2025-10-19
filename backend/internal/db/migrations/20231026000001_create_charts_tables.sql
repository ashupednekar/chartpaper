-- +goose Up
-- +goose StatementBegin
-- Charts table to store Helm chart information with version history
CREATE TABLE charts (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL DEFAULT 'application',
    chart_url TEXT NOT NULL,
    image_tag TEXT,
    canary_tag TEXT,
    manifest TEXT,
    is_latest BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ingress_paths TEXT, -- JSON array of ingress paths
    container_images TEXT, -- JSON array of container images
    service_ports TEXT, -- JSON array of service ports
    manifest_parsed_at TIMESTAMP -- When manifest was last parsed
);

-- Dependencies table to store chart dependencies
CREATE TABLE dependencies (
    id SERIAL PRIMARY KEY,
    chart_id INTEGER NOT NULL,
    dependency_name TEXT NOT NULL,
    dependency_version TEXT NOT NULL,
    repository TEXT,
    condition_field TEXT,
    image_tag TEXT DEFAULT 'N/A',
    canary_tag TEXT DEFAULT 'N/A',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chart_id) REFERENCES charts (id) ON DELETE CASCADE,
    UNIQUE(chart_id, dependency_name)
);

-- Apps table to store parsed application information
CREATE TABLE apps (
    id SERIAL PRIMARY KEY,
    chart_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    image TEXT,
    app_type TEXT,
    ports TEXT, -- JSON array of ports
    configs TEXT, -- JSON object of configs
    mounts TEXT, -- JSON object of mounts
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chart_id) REFERENCES charts (id) ON DELETE CASCADE
);

-- Indexes for better performance
CREATE INDEX idx_charts_name ON charts(name);
CREATE INDEX idx_dependencies_chart_id ON dependencies(chart_id);
CREATE INDEX idx_apps_chart_id ON apps(chart_id);
CREATE INDEX idx_charts_latest ON charts(is_latest) WHERE is_latest = TRUE;
CREATE INDEX idx_charts_manifest_parsed ON charts(manifest_parsed_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_charts_manifest_parsed;
DROP INDEX IF EXISTS idx_charts_latest;
DROP INDEX IF EXISTS idx_apps_chart_id;
DROP INDEX IF EXISTS idx_dependencies_chart_id;
DROP INDEX IF EXISTS idx_charts_name;
DROP TABLE IF EXISTS apps;
DROP TABLE IF EXISTS dependencies;
DROP TABLE IF EXISTS charts;
-- +goose StatementEnd