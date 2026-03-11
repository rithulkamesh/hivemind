# Hivemind Registry

Python package registry (API + web UI + Better Auth). All dev commands and config live in this folder.

## Local dev (host)

```bash
cd registry
cp .env.example .env    # set DATABASE_URL; add GITHUB_CLIENT_ID/SECRET etc. for OAuth
just deps               # start Postgres in Docker
just db-migrate         # Go API tables (packages, orgs, …)
just auth-migrate       # Better Auth tables (user, session, jwks, twoFactor, …) — see https://better-auth.com/docs/adapters/postgresql
just dev                # API (air) + Web (Vite + Better Auth); Ctrl+C kills both
```

- **Web**: http://localhost:3000 (app + `/auth`)
- **API**: http://localhost:8080 (proxied at `/api` and `/simple` from the web app)

If you had a root `.env.registry` before, copy it to `registry/.env` (or copy `registry/.env.example` and fill in values).

See [README.docker.md](README.docker.md) for Docker dev and [docs/oauth-setup.md](docs/oauth-setup.md) for GitHub/Google login.
