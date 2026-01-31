-- name: UpsertUserStatsOnRequestCreated :exec
INSERT INTO public.user_stats (
  user_id,
  request_count,
  last_request_at
)
VALUES (
  $1, -- user_id
  $2, -- increment_by
  $3  -- occurred_at
)
ON CONFLICT (user_id) DO UPDATE
SET
  request_count = public.user_stats.request_count + EXCLUDED.request_count,
  last_request_at = GREATEST(public.user_stats.last_request_at, EXCLUDED.last_request_at);