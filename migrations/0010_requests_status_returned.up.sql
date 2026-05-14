-- MVP-090: allow 'returned' status in requests read model

ALTER TABLE public.requests
  DROP CONSTRAINT IF EXISTS requests_status_chk;

ALTER TABLE public.requests
  ADD CONSTRAINT requests_status_chk CHECK (status IN ('draft','submitted','returned','approved','rejected'));
