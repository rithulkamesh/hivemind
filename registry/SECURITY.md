# Security Policy — Hivemind Registry

## Reporting Vulnerabilities

If you discover a security vulnerability, please report it responsibly:

1. **Do NOT** open a public GitHub issue
2. Email **security@hivemind.rithul.dev** with:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact assessment
3. You will receive an acknowledgment within **48 hours**
4. We aim to release a fix within **7 days** for critical issues

## Security Architecture

### Authentication

The registry supports three authentication methods:

| Method | Use Case | Token Format |
|--------|----------|-------------|
| **Better Auth JWT (JWKS)** | Web frontend sessions | RS256/ES256 signed JWT |
| **Legacy JWT** | Fallback when JWKS not configured | HS256 signed JWT |
| **API Key** | CLI/Twine publishing | `hm_` prefixed, SHA-256 hashed |

**JWT algorithm pinning**: JWKS verification is pinned to `RS256` and `ES256` only, preventing algorithm confusion attacks.

**API key security**: Keys are hashed with SHA-256 before storage. Comparison uses `crypto/subtle.ConstantTimeCompare` to prevent timing attacks.

### Authorization

- **Scope-based**: API keys carry explicit scopes (`read`, `publish`, `admin`)
- **JWT sessions** bypass scope checks (full access assumed for authenticated web users)
- **API key expiry** is enforced at middleware level after database lookup

### Input Validation

| Input | Validation |
|-------|-----------|
| Package names | Regex: `^[a-z0-9]([a-z0-9._-]*[a-z0-9])?$`, max 128 chars |
| Version strings | Regex: `^[0-9]+\.[0-9]+(\.[0-9]+)?([a-zA-Z0-9.+_-]*)$`, max 64 chars |
| File uploads | Extensions restricted to `.whl` and `.tar.gz` only |
| `.whl` files | Validated as valid ZIP archives before storage |
| Upload size | Enforced via `http.MaxBytesReader` (default: 100 MB) |
| SQL queries | All parameterized via sqlc-generated code; search uses `plainto_tsquery` |

### CORS

CORS is restricted to an explicit origin allowlist derived from `BASE_URL` and `FRONTEND_URL` configuration. Credentials are only sent for matching origins.

### Rate Limiting

| Endpoint | Limit |
|----------|-------|
| Global | Configurable via `RATE_LIMIT_RPS` (default: 10/sec) |
| Device auth (`/api/v1/auth/device/*`) | 5/min per IP |
| Upload (`/api/v1/packages/*/upload`) | 10/min per IP |
| Search (`/api/v1/search`) | 60/min per IP |

The device flow store is capped at 10,000 pending requests to prevent memory exhaustion.

### Storage Security

- **Server-side encryption**: S3 `PutObject` uses AES-256 server-side encryption in production (disabled for MinIO/LocalStack in development)
- **Presigned URL TTL**: Download URLs expire after 15 minutes
- **Upload timeout**: S3 uploads have a 30-second context deadline

### Transport Security (Caddy)

- HSTS with `preload` directive (max-age 2 years)
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `Permissions-Policy` restricts camera, microphone, geolocation, payment
- `Content-Security-Policy` with strict `connect-src`
- Server header removed
- `/internal/*` endpoints blocked from external access

### Secrets Management

| Secret | Source | Notes |
|--------|--------|-------|
| `INTERNAL_SECRET` | Environment variable | >= 32 chars enforced in production |
| `JWT_SECRET` | Environment variable | Used only for legacy JWT |
| `JWKS_URL` | Environment variable | Must use HTTPS in production |
| OAuth client secrets | Environment variable | **Flag: rotate regularly** |

## Known Limitations

1. **Verification pipeline** (`verify.go`): Currently a stub — no automated package verification is performed. Packages are marked "passed" immediately on upload.

2. **User code entropy** (device flow): The 4-letter + 4-digit user code has ~29 bits of entropy with non-uniform distribution. Acceptable per RFC 8628 but could be improved.

3. **Docker Compose `sslmode=disable`**: The production Docker Compose uses `sslmode=disable` for the PostgreSQL connection. This is acceptable for container-to-container networking on the same Docker bridge, but should NOT be used if the database is on a separate host.

4. **OAuth secrets in `.env`**: The working tree contains real GitHub and Google OAuth client secrets in `registry/.env`. While gitignored, these should be rotated if the repository is ever made public or the secrets are compromised.

## Security Audit Log

| Date | Finding | Severity | Status |
|------|---------|----------|--------|
| 2026-03-13 | Timing attack in `VerifyAPIKey` | Critical | Fixed |
| 2026-03-13 | Timing attack in `RequireInternalSecret` | Critical | Fixed |
| 2026-03-13 | CORS reflects any origin with credentials | Critical | Fixed |
| 2026-03-13 | Debug `fmt.Printf` leaks auth context | Critical | Fixed |
| 2026-03-13 | Expired API keys not checked | Critical | Fixed |
| 2026-03-13 | OAuth secrets in working tree | Critical | Flagged for rotation |
| 2026-03-13 | No package name validation | High | Fixed |
| 2026-03-13 | No version string validation | High | Fixed |
| 2026-03-13 | No upload body size enforcement | High | Fixed |
| 2026-03-13 | No file extension allowlist | High | Fixed |
| 2026-03-13 | Presigned URL TTL too long (1hr) | High | Fixed (15min) |
| 2026-03-13 | No S3 server-side encryption | High | Fixed (conditional) |
| 2026-03-13 | No JWT algorithm pinning | High | Fixed (RS256/ES256) |
| 2026-03-13 | No JWKS HTTPS enforcement | High | Fixed (production) |
| 2026-03-13 | Missing `.dockerignore` files | High | Fixed |
| 2026-03-13 | Caddyfile missing security headers | High | Fixed |
| 2026-03-13 | Unbounded device flow store | High | Fixed (10K cap) |
| 2026-03-13 | `INTERNAL_SECRET` no min length | Medium | Fixed (32 chars prod) |
| 2026-03-13 | S3 upload no timeout | Medium | Fixed (30s) |
| 2026-03-13 | No per-endpoint rate limiting | Medium | Fixed |
| 2026-03-13 | Caddyfile allows `/internal/*` | Medium | Fixed |
| 2026-03-13 | No `.whl` zip validation | Medium | Fixed |
| 2026-03-13 | Unconditional XFF trust | Medium | Fixed |
| 2026-03-13 | `hm_` prefix not checked early | Low | Fixed |
