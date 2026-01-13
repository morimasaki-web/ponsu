-- MVP-040: Read Model for workflow templates (minimal)

CREATE TABLE IF NOT EXISTS public.workflow_templates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
  name text NOT NULL DEFAULT '',
  description text NOT NULL DEFAULT '',
  definition jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_by_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- Name uniqueness within an org (simple UX guard)
CREATE UNIQUE INDEX IF NOT EXISTS workflow_templates_org_id_name_uidx
  ON public.workflow_templates (org_id, name);

CREATE INDEX IF NOT EXISTS workflow_templates_org_id_created_at_idx
  ON public.workflow_templates (org_id, created_at DESC);
