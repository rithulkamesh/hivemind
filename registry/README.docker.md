# Registry Docker

## Dev (hot reload)

From repo root:

```bash
docker compose -f registry/docker-compose.dev.yml up --build
```

- **Postgres**: localhost:5432 (user `registry`, password `registry`, db `hivemind_registry`)
- **API**: http://localhost:8080 (Go app with Air hot reload)
- **Web**: http://localhost:5173 (Vite dev server; proxies `/api`, `/auth`, `/simple` to API)

Edit code in `registry/api` or `registry/web`; API restarts on Go changes, frontend HMR on save.

If you see **CONNRESET** in the browser or terminal, the API may have restarted (e.g. Air rebuild). Refresh the page or retry; the proxy is configured with a 30s timeout. If it persists, ensure only one `docker compose ... up` is running and that the API container is healthy.

### Testing the registry with your local repo (pip)

1. **Use the web UI**: Open http://localhost:5173, register, log in, create a package (e.g. `my-plugin`), then upload a wheel/sdist via the dashboard or API.

2. **Install from the registry** (use the simple index via the web port so `/api` and `/simple` are on the same origin, or use the API port directly):

   ```bash
   # Prefer proxied URL (same origin as UI)
   pip install --index-url http://localhost:5173/simple/ --extra-index-url https://pypi.org/simple/ my-plugin

   # Or point pip only at the local registry (no PyPI)
   pip install --index-url http://localhost:5173/simple/ my-plugin
   ```

   If you use the API port directly (no Vite proxy):

   ```bash
   pip install --index-url http://localhost:8080/simple/ my-plugin
   ```

3. **Publish from a local project** (after creating the package in the UI and getting an API key or logging in):

   - Configure the index in `pyproject.toml`:
     ```toml
     [tool.hivemind]
     index-url = "http://localhost:5173/simple/"
     ```
   - Upload via the registry UI (package page → upload), or use the API with auth (e.g. `POST /api/v1/packages/{name}/upload` with Bearer token or API key).

Trust/certificate warnings for `http://` are expected for local dev; use HTTPS in production.

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

Set `VITE_API_BASE_URL` and `VITE_REGISTRY_URL` in `.env.prod` before building so the frontend points at your public API URL. For a single-domain setup, put Caddy/nginx in front and route `/` to web and `/api`, `/auth`, `/simple` to the API container.
