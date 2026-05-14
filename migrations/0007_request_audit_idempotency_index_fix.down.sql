-- MVP-032: revert audit idempotency index fix

DROP INDEX IF EXISTS public.request_audit_trail_org_event_global_position_uidx;

-- Restore the previous partial unique index.
CREATE UNIQUE INDEX IF NOT EXISTS request_audit_trail_org_event_global_position_uidx
  ON public.request_audit_trail (org_id, event_global_position)
  WHERE event_global_position IS NOT NULL;
