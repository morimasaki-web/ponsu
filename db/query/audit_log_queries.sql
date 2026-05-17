
-- name: SearchAuditLogs :many
SELECT id, org_id, request_id, actor_user_id, action, data, occurred_at
FROM public.request_audit_trail
WHERE org_id = sqlc.arg(org_id)
  AND (sqlc.narg(request_id)::uuid IS NULL OR request_id = sqlc.narg(request_id)::uuid)
  AND (sqlc.narg(actor_user_id)::uuid IS NULL OR actor_user_id = sqlc.narg(actor_user_id)::uuid)
  AND (sqlc.narg(action)::text IS NULL OR action = sqlc.narg(action)::text)
  AND (sqlc.narg(occurred_at_start)::timestamptz IS NULL OR occurred_at >= sqlc.narg(occurred_at_start)::timestamptz)
  AND (sqlc.narg(occurred_at_end)::timestamptz IS NULL OR occurred_at <= sqlc.narg(occurred_at_end)::timestamptz)
ORDER BY occurred_at DESC
LIMIT sqlc.arg(limit_count)
OFFSET sqlc.arg(offset_count); 
