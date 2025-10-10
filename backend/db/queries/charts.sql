-- name: GetChart :one
SELECT * FROM charts WHERE name = ? AND is_latest = TRUE LIMIT 1;

-- name: GetChartByID :one
SELECT * FROM charts WHERE id = ? LIMIT 1;

-- name: GetChartVersion :one
SELECT * FROM charts WHERE name = ? AND version = ? LIMIT 1;

-- name: ListCharts :many
SELECT * FROM charts WHERE is_latest = TRUE ORDER BY updated_at DESC;

-- name: ListChartVersions :many
SELECT * FROM charts WHERE name = ? ORDER BY created_at DESC;

-- name: CreateChart :one
INSERT INTO charts (
    name, version, description, type, chart_url, image_tag, canary_tag, manifest, is_latest
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: UpdateChart :one
UPDATE charts 
SET version = ?, description = ?, type = ?, chart_url = ?, 
    image_tag = ?, canary_tag = ?, manifest = ?, updated_at = CURRENT_TIMESTAMP
WHERE name = ? AND version = ?
RETURNING *;

-- name: SetLatestVersion :exec
UPDATE charts SET is_latest = FALSE WHERE name = ?;

-- name: SetVersionAsLatest :exec
UPDATE charts SET is_latest = TRUE WHERE name = ? AND version = ?;

-- name: DeleteChart :exec
DELETE FROM charts WHERE name = ?;

-- name: DeleteChartVersion :exec
DELETE FROM charts WHERE name = ? AND version = ?;

-- name: GetChartDependencies :many
SELECT d.*, c.name as chart_name FROM dependencies d
JOIN charts c ON d.chart_id = c.id
WHERE d.chart_id = ?;

-- name: CreateDependency :one
INSERT INTO dependencies (
    chart_id, dependency_name, dependency_version, repository, condition_field
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: DeleteChartDependencies :exec
DELETE FROM dependencies WHERE chart_id = ?;

-- name: GetChartApps :many
SELECT * FROM apps WHERE chart_id = ?;

-- name: CreateApp :one
INSERT INTO apps (
    chart_id, name, image, app_type, ports, configs, mounts
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: DeleteChartApps :exec
DELETE FROM apps WHERE chart_id = ?;

-- name: SearchCharts :many
SELECT * FROM charts 
WHERE name LIKE ? OR description LIKE ?
ORDER BY updated_at DESC;

-- name: FetchChartDependencies :many
SELECT d.dependency_name, d.dependency_version, d.repository 
FROM dependencies d
JOIN charts c ON d.chart_id = c.id
WHERE c.name = ? AND c.is_latest = TRUE;