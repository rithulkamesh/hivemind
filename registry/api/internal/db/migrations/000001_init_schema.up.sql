-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE NOT NULL,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT,
    email_verified BOOLEAN DEFAULT false,
    totp_secret TEXT,
    totp_enabled BOOLEAN DEFAULT false,
    created_at TIMESTAMPTZ DEFAULT now(),
    last_login_at TIMESTAMPTZ,
    banned BOOLEAN DEFAULT false
);

CREATE TABLE oauth_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    provider_user_id TEXT NOT NULL,
    provider_email TEXT,
    UNIQUE(provider, provider_user_id)
);
CREATE INDEX idx_oauth_identities_user_id ON oauth_identities(user_id);

CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL,
    display_name TEXT NOT NULL,
    billing_email TEXT,
    sso_enabled BOOLEAN DEFAULT false,
    saml_metadata_url TEXT,
    oidc_issuer TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE org_members (
    org_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL,
    PRIMARY KEY(org_id, user_id)
);
CREATE INDEX idx_org_members_user_id ON org_members(user_id);

CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    org_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    key_prefix TEXT NOT NULL,
    scopes TEXT[] NOT NULL,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now(),
    revoked BOOLEAN DEFAULT false
);
CREATE INDEX idx_api_keys_key_hash_active ON api_keys(key_hash) WHERE NOT revoked;
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

CREATE TABLE packages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    namespace TEXT,
    display_name TEXT NOT NULL,
    description TEXT,
    homepage TEXT,
    repository TEXT,
    license TEXT,
    keywords TEXT[],
    created_at TIMESTAMPTZ DEFAULT now(),
    owner_user_id UUID REFERENCES users(id),
    owner_org_id UUID REFERENCES organizations(id),
    verified BOOLEAN DEFAULT false,
    trusted BOOLEAN DEFAULT false,
    total_downloads BIGINT DEFAULT 0,
    search_vector tsvector GENERATED ALWAYS AS (to_tsvector('english', name || ' ' || coalesce(description, ''))) STORED,
    UNIQUE(namespace, name)
);
CREATE INDEX idx_packages_search ON packages USING GIN(search_vector);
CREATE INDEX idx_packages_namespace_name ON packages(namespace, name);

CREATE TABLE package_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    requires_python TEXT,
    requires_hivemind TEXT,
    tool_count INTEGER DEFAULT 0,
    uploaded_by UUID REFERENCES users(id),
    uploaded_at TIMESTAMPTZ DEFAULT now(),
    published BOOLEAN DEFAULT false,
    yanked BOOLEAN DEFAULT false,
    yank_reason TEXT,
    verification_status TEXT DEFAULT 'pending',
    verification_report JSONB,
    sigstore_bundle JSONB,
    UNIQUE(package_id, version)
);
CREATE INDEX idx_package_versions_package_version ON package_versions(package_id, version);

CREATE TABLE package_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version_id UUID NOT NULL REFERENCES package_versions(id) ON DELETE CASCADE,
    filename TEXT NOT NULL,
    filetype TEXT NOT NULL,
    python_version TEXT,
    abi TEXT,
    platform TEXT,
    size_bytes BIGINT NOT NULL,
    sha256 TEXT NOT NULL,
    md5 TEXT NOT NULL,
    s3_key TEXT NOT NULL,
    download_count BIGINT DEFAULT 0,
    uploaded_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_package_files_version_id ON package_files(version_id);

CREATE TABLE docker_images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID NOT NULL REFERENCES packages(id) ON DELETE CASCADE,
    tag TEXT NOT NULL,
    digest TEXT NOT NULL,
    ecr_uri TEXT NOT NULL,
    platform TEXT[],
    size_bytes BIGINT,
    pushed_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(package_id, tag)
);
CREATE INDEX idx_docker_images_package_id ON docker_images(package_id);

CREATE TABLE download_events (
    id BIGSERIAL,
    file_id UUID NOT NULL REFERENCES package_files(id),
    downloaded_at TIMESTAMPTZ DEFAULT now(),
    ip_hash TEXT,
    user_agent TEXT,
    country_code TEXT,
    installer TEXT
);
CREATE INDEX idx_download_events_file_downloaded ON download_events(file_id, downloaded_at);

CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id UUID REFERENCES users(id),
    actor_api_key_id UUID REFERENCES api_keys(id),
    action TEXT NOT NULL,
    resource_type TEXT,
    resource_id TEXT,
    metadata JSONB,
    ip_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at);
