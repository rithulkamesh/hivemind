-- Fix unique constraint on packages (namespace, name) to treat NULLs as equal.
-- PostgreSQL 15+ supports NULLS NOT DISTINCT.
ALTER TABLE packages DROP CONSTRAINT IF EXISTS packages_namespace_name_key;
ALTER TABLE packages ADD CONSTRAINT packages_namespace_name_key UNIQUE NULLS NOT DISTINCT (namespace, name);
