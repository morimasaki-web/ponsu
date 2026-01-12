-- MVP-021: event_store append / load

-- name: GetAggregateVersion :one
SELECT COALESCE(MAX(version), 0)::int AS version
FROM public.event_store
WHERE org_id = $1
  AND aggregate_type = $2
  AND aggregate_id = $3;

-- name: AppendEvent :one
WITH current AS (
  SELECT COALESCE(MAX(version), 0)::int AS version
  FROM public.event_store
  WHERE org_id = $1
    AND aggregate_type = $2
    AND aggregate_id = $3
)
INSERT INTO public.event_store (
  org_id,
  aggregate_type,
  aggregate_id,
  version,
  event_type,
  payload,
  metadata
)
SELECT
  $1::uuid AS org_id,
  $2::text AS aggregate_type,
  $3::uuid AS aggregate_id,
  $4::int AS version,
  $5::text AS event_type,
  $6::jsonb AS payload,
  $7::jsonb AS metadata
FROM current
WHERE current.version = ($4 - 1)
RETURNING id, org_id, aggregate_type, aggregate_id, version, event_type, payload, metadata, occurred_at;

-- name: ListEventsByAggregate :many
SELECT id, org_id, aggregate_type, aggregate_id, version, event_type, payload, metadata, occurred_at
FROM public.event_store
WHERE org_id = $1
  AND aggregate_type = $2
  AND aggregate_id = $3
  AND version >= $4
ORDER BY version ASC;
