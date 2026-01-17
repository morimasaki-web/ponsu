-- MVP-070: Read Model for request attachments

CREATE TABLE IF NOT EXISTS public.request_attachments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id uuid NOT NULL,
  request_id uuid NOT NULL,
  filename text NOT NULL DEFAULT '',
  content_type text NOT NULL DEFAULT 'application/octet-stream',
  size_bytes bigint NOT NULL DEFAULT 0,
  sha256 text NOT NULL DEFAULT '',
  storage_key text NOT NULL,
  uploaded_by_user_id uuid NULL REFERENCES public.users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT request_attachments_request_fk
    FOREIGN KEY (org_id, request_id)
    REFERENCES public.requests(org_id, id)
    ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS request_attachments_storage_key_uidx
  ON public.request_attachments (storage_key);

CREATE INDEX IF NOT EXISTS request_attachments_request_id_created_at_idx
  ON public.request_attachments (request_id, created_at DESC);

CREATE INDEX IF NOT EXISTS request_attachments_org_id_created_at_idx
  ON public.request_attachments (org_id, created_at DESC);
