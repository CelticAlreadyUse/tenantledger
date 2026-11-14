-- name: CreateTenant :one
INSERT INTO tenancy.tenants (id, name)
VALUES (sqlc.arg(id), sqlc.arg(name))
RETURNING id, name, created_at;

-- name: GetTenant :one
SELECT id, name, created_at
FROM tenancy.tenants
WHERE id = sqlc.arg(id);
