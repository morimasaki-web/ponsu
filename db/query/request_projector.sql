-- MVP-032: projector queries for requests read model

-- name: InsertRequestIfNotExists :exec
INSERT INTO public.requests (
  id,
  org_id,
  title,
  status,
  created_by_user_id,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (id) DO NOTHING;

-- name: SetRequestSubmitted :execrows
UPDATE public.requests
SET status = 'submitted',
    submitted_at = $3,
    updated_at = $3
WHERE org_id = $1 AND id = $2;

-- name: SetRequestApproved :execrows
UPDATE public.requests
SET status = 'approved',
    decided_at = $3,
    updated_at = $3
WHERE org_id = $1 AND id = $2;

-- name: SetRequestRejected :execrows
UPDATE public.requests
SET status = 'rejected',
    decided_at = $3,
    updated_at = $3
WHERE org_id = $1 AND id = $2;

-- name: InsertRequestStepIfNotExists :exec
INSERT INTO public.request_steps (
  request_id,
  step_index,
  label,
  status,
  assigned_to_user_id,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (request_id, step_index) DO NOTHING;

-- name: InsertRequestAuditIfNotExists :exec
INSERT INTO public.request_audit_trail (
  org_id,
  request_id,
  actor_user_id,
  action,
  data,
  occurred_at,
  event_global_position
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (org_id, event_global_position) DO NOTHING;
