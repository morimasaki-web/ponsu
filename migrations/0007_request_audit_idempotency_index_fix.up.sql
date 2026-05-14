-- MVP-032: fix ON CONFLICT compatibility for audit idempotency
--
-- Postgres cannot infer a *partial* unique index for `ON CONFLICT (cols)` unless the
-- same predicate is specified in the ON CONFLICT clause.
--
-- Our projector query uses:
--   ON CONFLICT (org_id, event_global_position) DO NOTHING
-- so we use a non-partial unique index. NULLs are allowed and do not conflict.

DROP INDEX IF EXISTS public.request_audit_trail_org_event_global_position_uidx;

CREATE UNIQUE INDEX IF NOT EXISTS request_audit_trail_org_event_global_position_uidx
  ON public.request_audit_trail (org_id, event_global_position);
