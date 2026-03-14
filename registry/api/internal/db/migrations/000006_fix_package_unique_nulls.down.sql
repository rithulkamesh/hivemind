-- Revert to standard UNIQUE (NULLs treated as distinct).
ALTER TABLE packages DROP CONSTRAINT IF EXISTS packages_namespace_name_key;
ALTER TABLE packages ADD CONSTRAINT packages_namespace_name_key UNIQUE (namespace, name);
