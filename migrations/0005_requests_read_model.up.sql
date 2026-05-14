-- MVP-030: Read Model for requests

-- requests: main list/detail
CREATE TABLE IF NOT EXISTS public.requests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
  title text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'draft',
  created_by_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  decided_by_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  submitted_at timestamptz NULL,
  decided_at timestamptz NULL,
  CONSTRAINT requests_status_chk CHECK (status IN ('draft','submitted','approved','rejected'))
);

-- For composite FK usage (audit org boundary)
CREATE UNIQUE INDEX IF NOT EXISTS requests_org_id_id_uidx
  ON public.requests (org_id, id);

CREATE INDEX IF NOT EXISTS requests_org_id_created_at_idx
  ON public.requests (org_id, created_at DESC);

CREATE INDEX IF NOT EXISTS requests_org_id_status_idx
  ON public.requests (org_id, status);

-- request_steps: step list/status (minimal)
CREATE TABLE IF NOT EXISTS public.request_steps (
  request_id uuid NOT NULL REFERENCES public.requests(id) ON DELETE CASCADE,
  step_index integer NOT NULL,
  label text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'pending',
  assigned_to_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (request_id, step_index),
  CONSTRAINT request_steps_status_chk CHECK (status IN ('pending','completed','skipped'))
);

CREATE INDEX IF NOT EXISTS request_steps_request_id_idx
  ON public.request_steps (request_id);

CREATE INDEX IF NOT EXISTS request_steps_assigned_to_user_id_idx
  ON public.request_steps (assigned_to_user_id);

-- request_audit_trail: timeline entries
CREATE TABLE IF NOT EXISTS public.request_audit_trail (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL,
  request_id uuid NOT NULL,
  actor_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  action text NOT NULL,
  data jsonb NOT NULL DEFAULT '{}'::jsonb,
  occurred_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT request_audit_trail_request_fk
    FOREIGN KEY (org_id, request_id)
    REFERENCES public.requests(org_id, id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS request_audit_trail_request_id_occurred_at_idx
  ON public.request_audit_trail (request_id, occurred_at ASC);

CREATE INDEX IF NOT EXISTS request_audit_trail_org_id_occurred_at_idx
  ON public.request_audit_trail (org_id, occurred_at DESC);
