# Registry Docker

## Dev (hot reload)

From the **registry** directory (so `.env` is next to the compose file and justfile):

```bash
cd registry
cp .env.example .env   # edit .env: set DATABASE_URL and optionally OAuth vars
just deps              # start Postgres in Docker
just dev               # API (air) + Web (Vite + Better Auth); Ctrl+C kills both
```

Or full stack in Docker:

```bash
cd registry
docker compose -f docker-compose.dev.yml up --build
```

- **Postgres**: localhost:5432 (user `registry`, password `registry`, db `hivemind_registry`)
- **API**: http://localhost:8080 (Go app with Air hot reload)
- **Web**: http://localhost:3000 (Vite dev server; serves `/auth` in-process, proxies `/api` and `/simple` to API)

Edit code in `registry/api` or `registry/web`; API restarts on Go changes, frontend HMR on save.

If you see **CONNRESET** in the browser or terminal, the API may have restarted (e.g. Air rebuild). Refresh the page or retry; the proxy is configured with a 30s timeout. If it persists, ensure only one `docker compose ... up` is running and that the API container is healthy.

### Testing the registry with your local repo (pip)

1. **Use the web UI**: Open http://localhost:3000, register, log in, create a package (e.g. `my-plugin`), then upload a wheel/sdist via the dashboard or API.

2. **Install from the registry** (use the simple index via the web port so `/api` and `/simple` are on the same origin, or use the API port directly):

   ```bash
   # Prefer proxied URL (same origin as UI)
   pip install --index-url http://localhost:3000/simple/ --extra-index-url https://pypi.org/simple/ my-plugin

   # Or point pip only at the local registry (no PyPI)
   pip install --index-url http://localhost:3000/simple/ my-plugin
   ```

   If you use the API port directly (no Vite proxy):

   ```bash
   pip install --index-url http://localhost:8080/simple/ my-plugin
   ```

3. **Publish from a local project** (after creating the package in the UI and getting an API key or logging in):

   - Configure the index in `pyproject.toml`:
     ```toml
     [tool.hivemind]
     index-url = "http://localhost:3000/simple/"
     ```
   - Upload via the registry UI (package page → upload), or use the API with auth (e.g. `POST /api/v1/packages/{name}/upload` with Bearer token or API key).

Trust/certificate warnings for `http://` are expected for local dev; use HTTPS in production.

### Test plugin (hivemind-plugin-demo)

A minimal plugin lives in `registry/test-plugin/`. To build and optionally seed the registry:

1. **Build the wheel**: From repo root, `just test-plugin-build` (or `cd registry/test-plugin && pip install build && python -m build`).
2. **Create the package and upload**: Use the web UI (register, create package `hivemind-plugin-demo`, upload the wheel from `registry/test-plugin/dist/`), or run the seed script when the API has storage (S3) configured:
   ```bash
   ./registry/scripts/seed-dev.sh
   # Or with custom wheel: ./registry/scripts/seed-dev.sh /path/to/wheel.whl
   ```
   Set `REGISTRY_EMAIL`, `REGISTRY_PASSWORD` if needed; default is `seed@localhost` / `seedpass123`.
3. **Install from the registry**: `pip install --index-url http://localhost:3000/simple/ hivemind-plugin-demo` (or `just test-plugin-install` from repo root).

Without S3 configured, the API will not accept uploads (upload and file download return 503). For local dev with uploads, use AWS credentials and set `S3_BUCKET`, or use a local S3-compatible store (e.g. LocalStack).

### OAuth (GitHub / Google)

To enable "Log in with GitHub" or "Log in with Google", create OAuth apps and set the client ID and secret in your env. See [registry/docs/oauth-setup.md](docs/oauth-setup.md) for step-by-step instructions and callback URLs.

## Prod

1. Copy env example and set secrets:

   ```bash
   cp registry/.env.prod.example registry/.env.prod
   # Edit .env.prod: POSTGRES_PASSWORD, JWT_SECRET, etc.
   ```

2. Build and run (from repo root):

   ```bash
   docker compose -f registry/docker-compose.prod.yml --env-file registry/.env.prod build
   docker compose -f registry/docker-compose.prod.yml --env-file registry/.env.prod up -d
   ```

- **Postgres**: internal only (volume `postgres_data`)
- **API**: port 8080
- **Web**: port 80 (nginx serving built frontend)

Set `VITE_API_BASE_URL` and `VITE_REGISTRY_URL` in `.env.prod` before building so the frontend points at your public API URL. For a single-domain setup, put Caddy/nginx in front: route `/api` and `/simple` to the API container, and everything else (including `/` and `/auth`) to the web container (Node server serving static + Better Auth).
