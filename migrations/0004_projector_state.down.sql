-- MVP-022: projector state + demo read model

DROP TABLE IF EXISTS public.projector_demo_aggregate_versions;
DROP TABLE IF EXISTS public.projection_checkpoints;

DROP INDEX IF EXISTS public.event_store_org_global_position_idx;
DROP INDEX IF EXISTS public.event_store_global_position_uidx;

ALTER TABLE public.event_store
	DROP COLUMN IF EXISTS global_position;

DROP SEQUENCE IF EXISTS public.event_store_global_position_seq;
