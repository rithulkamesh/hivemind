-- Registry profiles: extends Better Auth user with registry-specific data.
-- user_id is the Better Auth user ID (UUID string).
CREATE TABLE IF NOT EXISTS registry_profiles (
    user_id TEXT PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    bio TEXT,
    website TEXT,
    total_packages INTEGER DEFAULT 0,
    total_downloads BIGINT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_registry_profiles_username ON registry_profiles(username);
