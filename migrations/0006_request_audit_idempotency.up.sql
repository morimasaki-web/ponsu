-- MVP-032: make audit trail idempotent per projector (best-effort)

ALTER TABLE public.request_audit_trail
  ADD COLUMN IF NOT EXISTS event_global_position bigint;

CREATE UNIQUE INDEX IF NOT EXISTS request_audit_trail_org_event_global_position_uidx
  ON public.request_audit_trail (org_id, event_global_position)
  WHERE event_global_position IS NOT NULL;
