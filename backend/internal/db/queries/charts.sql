-- name: GetChart :one
SELECT * FROM charts WHERE name = $1 AND is_latest = TRUE LIMIT 1;

-- name: GetChartByID :one
SELECT * FROM charts WHERE id = $1 LIMIT 1;

-- name: GetChartVersion :one
SELECT * FROM charts WHERE name = $1 AND version = $2 LIMIT 1;

-- name: ListCharts :many
SELECT * FROM charts WHERE is_latest = TRUE ORDER BY updated_at DESC;

-- name: ListChartVersions :many
SELECT * FROM charts WHERE name = $1 ORDER BY created_at DESC;

-- name: CreateChart :one
INSERT INTO charts (
    name, version, description, type, chart_url, image_tag, canary_tag, manifest, is_latest
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: UpdateChart :one
UPDATE charts 
SET version = $1, description = $2, type = $3, chart_url = $4, 
    image_tag = $5, canary_tag = $6, manifest = $7, updated_at = CURRENT_TIMESTAMP
WHERE name = $8 AND version = $9
RETURNING *;

-- name: SetLatestVersion :exec
UPDATE charts SET is_latest = FALSE WHERE name = $1;

-- name: SetVersionAsLatest :exec
UPDATE charts SET is_latest = TRUE WHERE name = $1 AND version = $2;

-- name: DeleteChart :exec
DELETE FROM charts WHERE name = $1;

-- name: DeleteChartVersion :exec
DELETE FROM charts WHERE name = $1 AND version = $2;

-- name: GetChartDependencies :many
SELECT d.*, c.name as chart_name FROM dependencies d
JOIN charts c ON d.chart_id = c.id
WHERE d.chart_id = $1;

-- name: CreateDependency :one
INSERT INTO dependencies (
    chart_id, dependency_name, dependency_version, repository, condition_field
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: DeleteChartDependencies :exec
DELETE FROM dependencies WHERE chart_id = $1;

-- name: GetChartApps :many
SELECT * FROM apps WHERE chart_id = $1;

-- name: CreateApp :one
INSERT INTO apps (
    chart_id, name, image, app_type, ports, configs, mounts
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: DeleteChartApps :exec
DELETE FROM apps WHERE chart_id = $1;

-- name: SearchCharts :many
SELECT * FROM charts 
WHERE name LIKE $1 OR description LIKE $2
ORDER BY updated_at DESC;

-- name: FetchChartDependencies :many
SELECT d.dependency_name, d.dependency_version, d.repository 
FROM dependencies d
JOIN charts c ON d.chart_id = c.id
WHERE c.name = $1 AND c.is_latest = TRUE;