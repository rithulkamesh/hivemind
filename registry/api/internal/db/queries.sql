-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: CreateUser :one
INSERT INTO users (email, username, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUser :one
UPDATE users SET email = $2, username = $3, password_hash = $4, email_verified = $5,
  totp_secret = $6, totp_enabled = $7, last_login_at = $8, banned = $9
WHERE id = $1
RETURNING *;

-- name: GetOAuthIdentity :one
SELECT * FROM oauth_identities WHERE provider = $1 AND provider_user_id = $2;

-- name: CreateOAuthIdentity :one
INSERT INTO oauth_identities (user_id, provider, provider_user_id, provider_email)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrgByName :one
SELECT * FROM organizations WHERE name = $1;

-- name: GetOrgByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: CreateOrg :one
INSERT INTO organizations (name, display_name, billing_email)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateOrg :one
UPDATE organizations SET display_name = $2, billing_email = $3, sso_enabled = $4,
  saml_metadata_url = $5, oidc_issuer = $6
WHERE id = $1
RETURNING *;

-- name: ListOrgsForUser :many
SELECT o.* FROM organizations o
JOIN org_members m ON m.org_id = o.id
WHERE m.user_id = $1;

-- name: GetOrgMember :one
SELECT * FROM org_members WHERE org_id = $1 AND user_id = $2;

-- name: AddOrgMember :one
INSERT INTO org_members (org_id, user_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateOrgMemberRole :one
UPDATE org_members SET role = $3 WHERE org_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveOrgMember :exec
DELETE FROM org_members WHERE org_id = $1 AND user_id = $2;

-- name: ListOrgMembers :many
SELECT m.org_id, m.user_id, m.role, u.email, u.username
FROM org_members m
JOIN users u ON u.id = m.user_id
WHERE m.org_id = $1;

-- name: CreateAPIKey :one
INSERT INTO api_keys (user_id, org_id, name, key_hash, key_prefix, scopes, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys WHERE key_hash = $1 AND NOT revoked;

-- name: ListAPIKeysForUser :many
SELECT id, user_id, org_id, name, key_prefix, scopes, last_used_at, expires_at, created_at
FROM api_keys WHERE user_id = $1 AND NOT revoked;

-- name: RevokeAPIKey :exec
UPDATE api_keys SET revoked = true WHERE id = $1;

-- name: UpdateAPIKeyLastUsed :exec
UPDATE api_keys SET last_used_at = now() WHERE id = $1;

-- name: GetPackageByNamespaceName :one
SELECT * FROM packages WHERE (namespace IS NOT DISTINCT FROM $1) AND name = $2;

-- name: GetPackageByID :one
SELECT * FROM packages WHERE id = $1;

-- name: ListPackages :many
SELECT * FROM packages
WHERE (sqlc.narg('namespace_filter')::text IS NULL OR namespace = sqlc.narg('namespace_filter'))
ORDER BY name
LIMIT sqlc.arg('limit_val') OFFSET sqlc.arg('offset_val');

-- name: CreatePackage :one
INSERT INTO packages (name, namespace, display_name, description, homepage, repository, license, keywords, owner_user_id, owner_org_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: UpdatePackage :one
UPDATE packages SET display_name = $2, description = $3, homepage = $4, repository = $5, license = $6, keywords = $7, verified = $8, trusted = $9
WHERE id = $1
RETURNING *;

-- name: SearchPackages :many
SELECT * FROM packages
WHERE search_vector @@ plainto_tsquery('english', $1)
ORDER BY ts_rank(search_vector, plainto_tsquery('english', $1)) DESC
LIMIT $2 OFFSET $3;

-- name: GetPackageVersion :one
SELECT pv.* FROM package_versions pv
JOIN packages p ON p.id = pv.package_id
WHERE (p.namespace IS NOT DISTINCT FROM $1) AND p.name = $2 AND pv.version = $3;

-- name: GetVersionByID :one
SELECT * FROM package_versions WHERE id = $1;

-- name: ListVersionsForPackage :many
SELECT * FROM package_versions WHERE package_id = $1 ORDER BY uploaded_at DESC;

-- name: CreatePackageVersion :one
INSERT INTO package_versions (package_id, version, requires_python, requires_hivemind, uploaded_by, verification_status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdatePackageVersionVerification :one
UPDATE package_versions SET verification_status = $2, verification_report = $3, published = $4, tool_count = $5, sigstore_bundle = $6
WHERE id = $1
RETURNING *;

-- name: YankPackageVersion :one
UPDATE package_versions SET yanked = true, yank_reason = $2 WHERE id = $1
RETURNING *;

-- name: DeletePackageVersion :exec
DELETE FROM package_versions WHERE id = $1;

-- name: CreatePackageFile :one
INSERT INTO package_files (version_id, filename, filetype, python_version, abi, platform, size_bytes, sha256, md5, s3_key)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetPackageFileByID :one
SELECT * FROM package_files WHERE id = $1;

-- name: ListFilesForVersion :many
SELECT * FROM package_files WHERE version_id = $1;

-- name: GetPackageFileByVersionAndFilename :one
SELECT * FROM package_files WHERE version_id = $1 AND filename = $2;

-- name: IncrementFileDownloadCount :exec
UPDATE package_files SET download_count = download_count + 1 WHERE id = $1;

-- name: IncrementPackageDownloadCount :exec
UPDATE packages SET total_downloads = total_downloads + 1 WHERE id = $1;

-- name: ListPendingVerifications :many
SELECT pv.*, p.name as package_name, p.namespace
FROM package_versions pv
JOIN packages p ON p.id = pv.package_id
WHERE pv.verification_status = 'pending'
ORDER BY pv.uploaded_at
LIMIT $1;

-- name: SetPackageTrusted :one
UPDATE packages SET trusted = true WHERE id = $1
RETURNING *;

-- name: DeletePackage :exec
DELETE FROM packages WHERE id = $1;

-- name: InsertAuditLog :one
INSERT INTO audit_log (actor_user_id, actor_api_key_id, action, resource_type, resource_id, metadata, ip_hash)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id;

-- name: GetGlobalStats :one
SELECT
  (SELECT count(*)::bigint FROM packages) as total_packages,
  (SELECT coalesce(sum(total_downloads), 0)::bigint FROM packages) as total_downloads;

-- name: ListDockerImagesForPackage :many
SELECT * FROM docker_images WHERE package_id = $1 ORDER BY pushed_at DESC;

-- name: CreateDockerImage :one
INSERT INTO docker_images (package_id, tag, digest, ecr_uri, platform, size_bytes)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (package_id, tag) DO UPDATE SET digest = $3, ecr_uri = $4, platform = $5, size_bytes = $6, pushed_at = now()
RETURNING *;

-- name: RecordDownloadEvent :one
INSERT INTO download_events (file_id, ip_hash, user_agent, country_code, installer)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: ListUsersAdmin :many
SELECT id, email, username, email_verified, created_at, last_login_at, banned
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: SetUserBanned :exec
UPDATE users SET banned = true WHERE id = $1;
