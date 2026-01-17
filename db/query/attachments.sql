-- MVP-070: Read Model queries for request attachments

-- name: CreateRequestAttachment :one
INSERT INTO public.request_attachments (
  org_id,
  request_id,
  filename,
  content_type,
  size_bytes,
  sha256,
  storage_key,
  uploaded_by_user_id
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING id, org_id, request_id, filename, content_type, size_bytes, sha256, storage_key, uploaded_by_user_id, created_at;

-- name: ListRequestAttachmentsByOrgAndRequestID :many
SELECT id, org_id, request_id, filename, content_type, size_bytes, sha256, storage_key, uploaded_by_user_id, created_at
FROM public.request_attachments
WHERE org_id = $1 AND request_id = $2
ORDER BY created_at DESC;

-- name: GetRequestAttachmentByOrgRequestAndID :one
SELECT id, org_id, request_id, filename, content_type, size_bytes, sha256, storage_key, uploaded_by_user_id, created_at
FROM public.request_attachments
WHERE org_id = $1 AND request_id = $2 AND id = $3;
