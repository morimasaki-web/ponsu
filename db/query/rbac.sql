-- MVP-011: RBAC (organizations/users/memberships)

-- name: UpsertUserFromOIDC :one
INSERT INTO public.users (oidc_issuer, oidc_sub, email, name)
VALUES ($1, $2, $3, $4)
ON CONFLICT (oidc_issuer, oidc_sub) DO UPDATE
SET email = EXCLUDED.email,
    name = EXCLUDED.name,
    updated_at = now()
RETURNING id, oidc_issuer, oidc_sub, email, name;

-- name: GetUserByID :one
SELECT id, oidc_issuer, oidc_sub, email, name, created_at, updated_at
FROM public.users
WHERE id = $1;

-- name: GetAnyMembershipByUserID :one
SELECT org_id, user_id, role
FROM public.memberships
WHERE user_id = $1
ORDER BY created_at ASC
LIMIT 1;

-- name: GetMembershipByOrgAndUserID :one
SELECT org_id, user_id, role
FROM public.memberships
WHERE org_id = $1 AND user_id = $2;

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

-- name: HasPermission :one
SELECT EXISTS(
  SELECT 1
  FROM public.role_permissions rp
  WHERE rp.role = $1
    AND rp.permission_id = $2
) AS has_permission;

-- name: ListPermissionsByRole :many
SELECT p.id, p.description, p.created_at
FROM public.permissions p
INNER JOIN public.role_permissions rp ON rp.permission_id = p.id
WHERE rp.role = $1
ORDER BY p.id;