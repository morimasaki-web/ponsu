-- LEARN-022F: SQL queries for request comments

-- name: ListCommentsByRequest :many
SELECT id, org_id, request_id, user_id, content, created_at
FROM request_comments
WHERE org_id = $1 AND request_id = $2
ORDER BY created_at ASC;

-- name: InsertCommentIfNotExists :exec
INSERT INTO request_comments (id, org_id, request_id, user_id, content, created_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (id) DO NOTHING;

-- name: InsertComment :one
INSERT INTO request_comments (org_id, request_id, user_id, content)
VALUES ($1, $2, $3, $4)
RETURNING id, org_id, request_id, user_id, content, created_at;
