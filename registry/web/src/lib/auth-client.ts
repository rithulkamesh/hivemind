import { createAuthClient } from "better-auth/react";
import { jwtClient, twoFactorClient, oneTapClient, usernameClient, adminClient, organizationClient, lastLoginMethodClient } from "better-auth/client/plugins";
import { passkeyClient } from "@better-auth/passkey/client";
import { apiKeyClient } from "@better-auth/api-key/client";

// Server uses basePath: "/auth"; client defaults to /api/auth so we must pass full URL with path.
const origin = typeof window !== "undefined" ? window.location.origin : import.meta.env.VITE_AUTH_URL ?? "";
const authBaseURL = origin ? `${origin.replace(/\/$/, "")}/auth` : undefined;

export const authClient = createAuthClient({
  baseURL: authBaseURL,
  plugins: [
    jwtClient(),
    twoFactorClient(),
    oneTapClient({
      clientId: import.meta.env.VITE_GOOGLE_CLIENT_ID ?? "",
    }),
    passkeyClient(),
    usernameClient(),
    adminClient(),
    organizationClient(),
    lastLoginMethodClient(),
    apiKeyClient(),
  ],
});

export const { signIn, signUp, signOut, useSession } = authClient;

let cachedToken: string | null = null;
let tokenExpiry = 0;

/** Get JWT for Go API (Bearer). Resolves to null if not signed in. */
export async function getApiToken(forceRefresh = false): Promise<string | null> {
  if (!forceRefresh && cachedToken && Date.now() < tokenExpiry) {
    return cachedToken;
  }

  try {
    const { data } = await authClient.getSession();
    if (!data) return null;

    const res = await fetch("/auth/token", {
      method: "GET",
      headers: { "Content-Type": "application/json" },
    });
    
    if (!res.ok) return null;
    
    const body = await res.json();
    if (body.token) {
      cachedToken = body.token;
      // Cache for 4 minutes (server default is usually 5 mins for access token)
      tokenExpiry = Date.now() + 4 * 60 * 1000;
      return body.token;
    }
  } catch {
    // ignore error
  }
  return null;
}
