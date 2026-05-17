-- name: GetNotificationTemplate :one
SELECT id, org_id, event_type, template_body, created_at, updated_at
FROM notification_templates
WHERE org_id = $1 AND event_type = $2
LIMIT 1;

-- name: UpsertNotificationTemplate :one
INSERT INTO notification_templates (org_id, event_type, template_body)
VALUES ($1, $2, $3)
ON CONFLICT (org_id, event_type)
DO UPDATE SET template_body = EXCLUDED.template_body, updated_at = NOW()
RETURNING id, org_id, event_type, template_body, created_at, updated_at;
