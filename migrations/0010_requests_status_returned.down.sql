-- MVP-090: revert 'returned' status allowance

ALTER TABLE public.requests
  DROP CONSTRAINT IF EXISTS requests_status_chk;

ALTER TABLE public.requests
  ADD CONSTRAINT requests_status_chk CHECK (status IN ('draft','submitted','approved','rejected'));
