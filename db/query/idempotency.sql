-- name: InsertIdempotencyIfNotExists :execrows
INSERT INTO idempotency (
    org_id,
    actor_user_id,
    action,
    idempotency_key,
    request_hash,
    status_code,
    response_body,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
ON CONFLICT (org_id, actor_user_id, action, idempotency_key) DO NOTHING;

-- name: GetIdempotencyByKey :one
SELECT org_id, actor_user_id, action, idempotency_key, request_hash, status_code, response_body, created_at
FROM idempotency
WHERE org_id = $1 AND actor_user_id = $2 AND action = $3 AND idempotency_key = $4;

-- name: UpdateIdempotencyResponse :exec
UPDATE idempotency
SET request_hash = $5, status_code = $6, response_body = $7
WHERE org_id = $1 AND actor_user_id = $2 AND action = $3 AND idempotency_key = $4;