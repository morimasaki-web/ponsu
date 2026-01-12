-- 0001_init.up.sql
--
-- MVP-003 establishes the migration mechanism.
-- This migration is intentionally minimal but creates a tangible, reversible artifact.

CREATE TABLE IF NOT EXISTS public._ponsu_migrations_smoketest (
	id integer PRIMARY KEY,
	created_at timestamptz NOT NULL DEFAULT now()
);
