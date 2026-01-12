-- MVP-011: RBAC (organizations/users/memberships)

-- name: UpsertUserFromOIDC :one
INSERT INTO public.users (oidc_issuer, oidc_sub, email, name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (oidc_issuer, oidc_sub) DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name,
    updated_at = now()
RETURNING id, oidc_issuer, oidc_sub, email, name;

-- name: GetAnyMembershipByUserID :one
SELECT org_id, user_id, role
FROM public.memberships
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;

-- name: CreateOrganization :one
INSERT INTO public.organizations (name)
VALUES ($1)
RETURNING id, name;

-- name: UpsertMembership :one
INSERT INTO public.memberships (org_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, user_id) DO UPDATE
SET role = EXCLUDED.role,
    updated_at = now()
RETURNING org_id, user_id, role;
