-- MVP-022: projector (synchronous projection)

-- name: GetProjectionCheckpoint :one
SELECT org_id, projector_name, last_position, updated_at
FROM public.projection_checkpoints
WHERE org_id = $1 AND projector_name = $2;

-- name: UpsertProjectionCheckpoint :one
INSERT INTO public.projection_checkpoints (org_id, projector_name, last_position)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, projector_name) DO UPDATE
SET last_position = EXCLUDED.last_position,
    updated_at = now()
RETURNING org_id, projector_name, last_position, updated_at;

-- name: ListEventsForProjector :many
SELECT id, org_id, aggregate_type, aggregate_id, version, global_position, event_type, payload, metadata, occurred_at
FROM public.event_store
WHERE org_id = $1
  AND global_position > $2
ORDER BY global_position ASC
LIMIT $3;

-- Demo projection: store last version per aggregate.
-- name: UpsertDemoAggregateVersion :one
INSERT INTO public.projector_demo_aggregate_versions (org_id, aggregate_type, aggregate_id, last_version)
VALUES ($1, $2, $3, $4)
ON CONFLICT (org_id, aggregate_type, aggregate_id) DO UPDATE
SET last_version = EXCLUDED.last_version,
    updated_at = now()
RETURNING org_id, aggregate_type, aggregate_id, last_version, updated_at;
