-- MVP-040: Read Model queries for workflow templates

-- name: CreateWorkflowTemplate :one
INSERT INTO public.workflow_templates (
  org_id,
  name,
  description,
  definition,
  created_by_user_id
)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, org_id, name, description, definition, created_by_user_id, created_at, updated_at;

-- name: ListWorkflowTemplatesByOrg :many
SELECT id, org_id, name, description, definition, created_by_user_id, created_at, updated_at
FROM public.workflow_templates
WHERE org_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3;

-- name: GetWorkflowTemplateByOrgAndID :one
SELECT id, org_id, name, description, definition, created_by_user_id, created_at, updated_at
FROM public.workflow_templates
WHERE org_id = $1 AND id = $2;
