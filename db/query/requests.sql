-- MVP-030: Read Model queries for requests

-- name: ListRequestsByOrg :many
SELECT id, org_id, title, status, created_by_user_id, decided_by_user_id, created_at, updated_at, submitted_at, decided_at
FROM public.requests
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetRequestByOrgAndID :one
SELECT id, org_id, title, status, created_by_user_id, decided_by_user_id, created_at, updated_at, submitted_at, decided_at
FROM public.requests
WHERE org_id = $1 AND id = $2;

-- name: ListRequestSteps :many
SELECT request_id, step_index, label, status, assigned_to_user_id, updated_at
FROM public.request_steps
WHERE request_id = $1
ORDER BY step_index ASC;

-- name: ListRequestAuditTrail :many
SELECT id, org_id, request_id, actor_user_id, action, data, occurred_at
FROM public.request_audit_trail
WHERE org_id = $1 AND request_id = $2
ORDER BY occurred_at ASC;

-- name: SearchRequests :many
SELECT id, org_id, title, status, created_by_user_id, decided_by_user_id, created_at, updated_at, submitted_at, decided_at
FROM public.requests
WHERE org_id = sqlc.arg(org_id)
  AND (sqlc.narg(title)::text IS NULL OR title ILIKE '%' || sqlc.narg(title)::text || '%')
  AND (sqlc.narg(status)::text IS NULL OR status = sqlc.narg(status)::text)
  AND (sqlc.narg(created_at_start)::timestamptz IS NULL OR created_at >= sqlc.narg(created_at_start)::timestamptz)
  AND (sqlc.narg(created_at_end)::timestamptz IS NULL OR created_at <= sqlc.narg(created_at_end)::timestamptz)
ORDER BY created_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count);
