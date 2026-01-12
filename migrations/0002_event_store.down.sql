-- MVP-020: rollback event_store schema

DROP INDEX IF EXISTS public.event_store_occurred_at_idx;
DROP INDEX IF EXISTS public.event_store_aggregate_idx;
DROP INDEX IF EXISTS public.event_store_aggregate_version_uidx;

DROP TABLE IF EXISTS public.event_store;
