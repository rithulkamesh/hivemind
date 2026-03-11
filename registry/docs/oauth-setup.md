# OAuth setup (GitHub and Google)

The registry supports login with GitHub and Google. Auth (including OAuth) runs in the **web app** (Better Auth). Configure client IDs and secrets so the web app can redirect users and handle callbacks.

## Environment variables

| Variable | Description |
|----------|-------------|
| `GITHUB_CLIENT_ID` | GitHub OAuth App Client ID |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth App Client Secret |
| `GOOGLE_CLIENT_ID` | Google OAuth 2.0 Client ID |
| `GOOGLE_CLIENT_SECRET` | Google OAuth 2.0 Client Secret |
| `BETTER_AUTH_URL` or `BASE_URL` | Public URL of the **web app** (used for redirect_uri). In dev: `http://localhost:3000`. |

If a provider’s client ID is not set, that provider’s login button will return a clear error (`github_oauth_not_configured` or `google_oauth_not_configured`) and password login still works.

---

## GitHub

1. Go to **GitHub → Settings → Developer settings → OAuth Apps** (or [github.com/settings/developers](https://github.com/settings/developers)).
2. Click **New OAuth App**.
3. Set:
   - **Application name**: e.g. `Hivemind Registry (dev)`.
   - **Homepage URL**: e.g. `http://localhost:3000` (your frontend) or production URL.
   - **Authorization callback URL**: must be the **web app** callback URL (port must match where the app runs; using 5173 when the app is on 3000 causes ECONNREFUSED):
     - **Local dev**: `http://localhost:3000/auth/callback/github`
     - **Production**: `https://registry.<your-domain>/auth/callback/github`
4. Create the app, then copy the **Client ID** and generate a **Client Secret**.
5. Set `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET` in your env (e.g. `registry/.env` or Docker env for the web service).

---

## Google

1. Go to [Google Cloud Console](https://console.cloud.google.com/) → **APIs & Services → Credentials**.
2. Click **Create Credentials → OAuth client ID**.
3. If prompted, configure the **OAuth consent screen** (e.g. External, app name, support email).
4. Choose **Application type**: **Web application**.
5. Under **Authorized redirect URIs** add (port must match where the app runs):
   - **Local dev**: `http://localhost:3000/auth/callback/google`
   - **Production**: `https://registry.<your-domain>/auth/callback/google`
6. Create and copy the **Client ID** and **Client Secret**.
7. Set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` in your env.

---

## Local dev and base URL

- Auth runs in the web app. The frontend and `/auth` are at `http://localhost:3000` (same for `just dev` and Docker dev).
- **Callback URL** (in GitHub/Google settings): `http://localhost:3000/auth/callback/github` and `http://localhost:3000/auth/callback/google`. Use port **3000** (not 5173); the app runs on 3000.
- Set the auth base URL in env to that same origin if needed (see `registry/web/server/auth.ts`). When using `cd registry && just dev`, `registry/.env` is loaded automatically.

In production, set the base URL to your public site (e.g. `https://registry.example.com`).
