-- LEARN-021: パーミッションベースRBAC

-- permissions: システム全体で定義可能な権限
CREATE TABLE IF NOT EXISTS public.permissions (
  id text PRIMARY KEY,
  description text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now()
);

-- role_permissions: ロールとパーミッションの紐付け
CREATE TABLE IF NOT EXISTS public.role_permissions (
  role text NOT NULL,
  permission_id text NOT NULL REFERENCES public.permissions(id) ON DELETE CASCADE,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (role, permission_id),
  CONSTRAINT role_permissions_role_chk CHECK (role IN ('admin', 'member'))
);

-- 初期データ: パーミッション定義
INSERT INTO public.permissions (id, description) VALUES
  ('request.view', '申請を閲覧できる'),
  ('request.create', '申請を作成できる'),
  ('request.approve', '申請を承認できる'),
  ('request.export', '申請データをエクスポートできる（管理者専用）')
ON CONFLICT (id) DO NOTHING;

-- 初期データ: ロール-パーミッションマッピング
-- adminは全権限
INSERT INTO public.role_permissions (role, permission_id) VALUES
  ('admin', 'request.view'),
  ('admin', 'request.create'),
  ('admin', 'request.approve'),
  ('admin', 'request.export')
ON CONFLICT DO NOTHING;

-- memberは基本権限のみ
INSERT INTO public.role_permissions (role, permission_id) VALUES
  ('member', 'request.view'),
  ('member', 'request.create')
ON CONFLICT DO NOTHING;