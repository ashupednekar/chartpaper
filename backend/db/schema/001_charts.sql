-- Charts table to store Helm chart information with version history
CREATE TABLE charts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    description TEXT,
    type TEXT NOT NULL DEFAULT 'application',
    chart_url TEXT NOT NULL,
    image_tag TEXT,
    canary_tag TEXT,
    manifest TEXT,
    is_latest BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, version)
);

-- Dependencies table to store chart dependencies
CREATE TABLE dependencies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chart_id INTEGER NOT NULL,
    dependency_name TEXT NOT NULL,
    dependency_version TEXT NOT NULL,
    repository TEXT,
    condition_field TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chart_id) REFERENCES charts (id) ON DELETE CASCADE,
    UNIQUE(chart_id, dependency_name)
);

-- Apps table to store parsed application information
CREATE TABLE apps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    chart_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    image TEXT,
    app_type TEXT,
    ports TEXT, -- JSON array of ports
    configs TEXT, -- JSON object of configs
    mounts TEXT, -- JSON object of mounts
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (chart_id) REFERENCES charts (id) ON DELETE CASCADE
);

-- Indexes for better performance
CREATE INDEX idx_charts_name ON charts(name);
CREATE INDEX idx_dependencies_chart_id ON dependencies(chart_id);
CREATE INDEX idx_apps_chart_id ON apps(chart_id);

-- Chart versions view to easily get version history
CREATE VIEW chart_versions AS
SELECT 
    name,
    version,
    description,
    type,
    chart_url,
    image_tag,
    canary_tag,
    is_latest,
    created_at,
    updated_at
FROM charts
ORDER BY name, created_at DESC;