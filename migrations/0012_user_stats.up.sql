-- user_stats: 
CREATE TABLE IF NOT EXISTS public.user_stats (
  user_id uuid PRIMARY KEY REFERENCES public.users(id) ON DELETE CASCADE,
  request_count bigint NOT NULL DEFAULT 0,
  last_request_at timestamptz NOT NULL DEFAULT now()
);
