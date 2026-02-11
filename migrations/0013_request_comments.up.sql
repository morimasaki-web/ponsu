-- LEARN-022F: Create request_comments table for comment read model

CREATE TABLE IF NOT EXISTS request_comments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  org_id UUID NOT NULL,
  request_id UUID NOT NULL,
  user_id UUID NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  FOREIGN KEY (org_id, request_id) REFERENCES requests(org_id, id) ON DELETE CASCADE
);

CREATE INDEX idx_request_comments_request ON request_comments(org_id, request_id, created_at);
