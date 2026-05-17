-- MVP-004: sqlc smoke test queries
--
-- NOTE: Uses the MVP-003 migration table.

-- name: Ping :one
SELECT 1::int AS one;

-- name: UpsertMigrationSmoketest :one
INSERT INTO public._ponsu_migrations_smoketest (id)
VALUES ($1)
ON CONFLICT (id) DO UPDATE
SET created_at = now()
RETURNING id, created_at;

-- name: GetMigrationSmoketest :one
SELECT id, created_at
FROM public._ponsu_migrations_smoketest
WHERE id = $1;
