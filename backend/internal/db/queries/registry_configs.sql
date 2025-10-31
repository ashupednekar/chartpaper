-- name: GetDefaultRegistryConfig :one
SELECT * FROM registry_configs WHERE is_default = TRUE LIMIT 1;
