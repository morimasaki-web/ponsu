-- MVP-022: projector state + demo read model

-- event_store.version は aggregate 単位の連番のため、プロジェクターのチェックポイントとしては使えない。
-- org をまたいでも一意な単調増加カーソル（global_position）を追加する。

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'event_store_global_position_seq') THEN
    CREATE SEQUENCE public.event_store_global_position_seq;
  END IF;
END $$;

ALTER TABLE public.event_store
  ADD COLUMN IF NOT EXISTS global_position bigint;

-- 既存行がある場合の backfill（MVP用途の best-effort）。
WITH ordered AS (
  SELECT id, row_number() OVER (ORDER BY occurred_at ASC, id ASC) AS rn
  FROM public.event_store
  WHERE global_position IS NULL
)
UPDATE public.event_store e
SET global_position = ordered.rn
FROM ordered
WHERE e.id = ordered.id;

DO $$
DECLARE
  m bigint;
BEGIN
  SELECT MAX(global_position) INTO m FROM public.event_store;
  IF m IS NULL THEN
    -- Empty table: prepare so nextval() returns 1.
    PERFORM setval('public.event_store_global_position_seq', 1, false);
  ELSE
    -- Non-empty: continue from current max.
    PERFORM setval('public.event_store_global_position_seq', m, true);
  END IF;
END $$;

ALTER TABLE public.event_store
  ALTER COLUMN global_position SET DEFAULT nextval('public.event_store_global_position_seq');

ALTER TABLE public.event_store
  ALTER COLUMN global_position SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS event_store_global_position_uidx
  ON public.event_store (global_position);

CREATE INDEX IF NOT EXISTS event_store_org_global_position_idx
  ON public.event_store (org_id, global_position);

CREATE TABLE IF NOT EXISTS public.projection_checkpoints (
  org_id uuid NOT NULL,
  projector_name text NOT NULL,
  last_position bigint NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, projector_name)
);

-- Demo read model table to prove projection updates.
-- Stores last projected version per aggregate.
CREATE TABLE IF NOT EXISTS public.projector_demo_aggregate_versions (
  org_id uuid NOT NULL,
  aggregate_type text NOT NULL,
  aggregate_id uuid NOT NULL,
  last_version integer NOT NULL DEFAULT 0,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, aggregate_type, aggregate_id)
);
