-- Dashboard statistics queries

-- name: CountRequestsByStatus :many
SELECT status, COUNT(*) AS count
FROM public.requests
WHERE org_id = $1
GROUP BY status
ORDER BY status;

-- name: CountRequestsByMonth :many
SELECT 
  DATE_TRUNC('month', created_at)::timestamptz AS month, 
  COUNT(*)::bigint AS count
FROM public.requests
WHERE org_id = $1
  AND (sqlc.narg(start_date)::timestamptz IS NULL OR created_at >= sqlc.narg(start_date)::timestamptz)
  AND (sqlc.narg(end_date)::timestamptz IS NULL OR created_at <= sqlc.narg(end_date)::timestamptz)
GROUP BY month
ORDER BY month DESC;

-- name: AvgTimeToApproval :one
SELECT 
  COALESCE(AVG(EXTRACT(EPOCH FROM (decided_at - submitted_at))), 0)::double precision AS avg_seconds,
  COUNT(*)::bigint AS sample_count
FROM public.requests
WHERE org_id = $1 
  AND status = 'approved'
  AND submitted_at IS NOT NULL
  AND decided_at IS NOT NULL;

-- name: GetDashboardSummary :one
SELECT 
  COUNT(*) FILTER (WHERE status = 'draft')::bigint AS draft_count,
  COUNT(*) FILTER (WHERE status = 'submitted')::bigint AS submitted_count,
  COUNT(*) FILTER (WHERE status = 'approved')::bigint AS approved_count,
  COUNT(*) FILTER (WHERE status = 'rejected')::bigint AS rejected_count,
  COUNT(*)::bigint AS total_count,
  COALESCE(AVG(EXTRACT(EPOCH FROM (decided_at - submitted_at))) FILTER (WHERE status = 'approved' AND submitted_at IS NOT NULL AND decided_at IS NOT NULL), 0)::double precision AS avg_approval_seconds
FROM public.requests
WHERE org_id = $1;
