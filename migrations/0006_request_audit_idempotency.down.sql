-- MVP-032: rollback audit idempotency

DROP INDEX IF EXISTS public.request_audit_trail_org_event_global_position_uidx;

ALTER TABLE public.request_audit_trail
  DROP COLUMN IF EXISTS event_global_position;
