-- MVP-020: event_store schema (event sourcing)

-- For gen_random_uuid()
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS public.event_store (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  version integer NOT NULL,
  event_type text NOT NULL,
  payload jsonb NOT NULL,
  metadata jsonb NOT NULL,
  occurred_at timestamptz NOT NULL DEFAULT now()
);

-- Optimistic lock / ordering guarantee within an aggregate
CREATE UNIQUE INDEX IF NOT EXISTS event_store_aggregate_version_uidx
  ON public.event_store (org_id, aggregate_type, aggregate_id, version);

-- Hot paths for loading events and listing recent events
CREATE INDEX IF NOT EXISTS event_store_aggregate_idx
  ON public.event_store (org_id, aggregate_type, aggregate_id);

CREATE INDEX IF NOT EXISTS event_store_occurred_at_idx
  ON public.event_store (org_id, occurred_at);
