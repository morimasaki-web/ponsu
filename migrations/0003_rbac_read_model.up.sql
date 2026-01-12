-- MVP-011: RBAC read model (organizations/users/memberships)

-- organizations: tenant boundary
CREATE TABLE IF NOT EXISTS public.organizations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

-- users: OIDC identity mapped to internal user id
CREATE TABLE IF NOT EXISTS public.users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  oidc_issuer text NOT NULL,
  oidc_sub text NOT NULL,
  email text NOT NULL DEFAULT '',
  name text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_oidc_uidx
  ON public.users (oidc_issuer, oidc_sub);

CREATE INDEX IF NOT EXISTS users_email_idx
  ON public.users (email);

-- memberships: user belongs to an organization with a role
CREATE TABLE IF NOT EXISTS public.memberships (
  org_id uuid NOT NULL REFERENCES public.organizations(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
  role text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (org_id, user_id),
  CONSTRAINT memberships_role_chk CHECK (role IN ('admin', 'member'))
);

CREATE INDEX IF NOT EXISTS memberships_user_id_idx
  ON public.memberships (user_id);

CREATE INDEX IF NOT EXISTS memberships_org_id_idx
  ON public.memberships (org_id);
