import { authClient, getApiToken } from "@/lib/auth-client";

export const apiBase = import.meta.env.VITE_API_BASE_URL || (typeof window !== "undefined" ? window.location.origin : "http://localhost:8080");
const BASE = apiBase;

/** Parse API auth error response (JSON { error } or plain text). */
export async function parseAuthError(res: Response): Promise<string> {
  const ct = res.headers.get("content-type");
  if (ct?.includes("application/json")) {
    try {
      const body = (await res.json()) as { error?: string };
      if (typeof body.error === "string") return body.error;
    } catch {
      // ignore
    }
  }
  const text = await res.text();
  return text || "";
}

export async function api<T>(
  path: string,
  opts?: RequestInit & { params?: Record<string, string> }
): Promise<T> {
  const url = new URL(path, BASE);
  if (opts?.params) {
    Object.entries(opts.params).forEach(([k, v]) => url.searchParams.set(k, v));
  }

  const headers = new Headers(opts?.headers);
  headers.set("Content-Type", "application/json");

  // Get current token from Better Auth bearer plugin (manually fetched)
  const token = await getApiToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  let res = await fetch(url.toString(), { ...opts, headers });

  if (res.status === 401) {
    // Token expired — refresh and retry once
    const newToken = await getApiToken(true);
    if (newToken) {
      headers.set("Authorization", `Bearer ${newToken}`);
      res = await fetch(url.toString(), { ...opts, headers });
    } else {
      // Refresh failed — sign out
      await authClient.signOut();
      window.location.href = "/login";
    }
  }

  if (!res.ok) {
    const text = await parseAuthError(res);
    throw new Error(text || `HTTP ${res.status}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const apiRoutes = {
  me: () => "/api/v1/me",
  packages: (params?: { namespace?: string; page?: number }) => {
    const p = new URLSearchParams();
    if (params?.namespace) p.set("namespace", params.namespace);
    if (params?.page) p.set("page", String(params.page));
    return `/api/v1/packages?${p}`;
  },
  package: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}`,
  createPackage: () => "/api/v1/packages",
  deletePackage: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}`,
  version: (name: string, version: string) =>
    `/api/v1/packages/${encodeURIComponent(name)}/${encodeURIComponent(version)}`,
  search: (q: string, page = 1) => `/api/v1/search?q=${encodeURIComponent(q)}&page=${page}`,
  stats: () => "/api/v1/stats",
  packageImages: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}/images`,
  packageDownloads: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}/downloads`,
  apiKeys: () => "/api/v1/me/api-keys",
  revokeApiKey: (id: string) => `/api/v1/me/api-keys/${id}`,
  orgs: () => "/api/v1/orgs",
  org: (slug: string) => `/api/v1/orgs/${slug}`,
  orgMembers: (slug: string) => `/api/v1/orgs/${slug}/members`,
  orgPackages: (slug: string) => `/api/v1/orgs/${slug}/packages`,
  adminVerificationQueue: () => "/api/v1/admin/verification-queue",
  adminUsers: () => "/api/v1/admin/users",
  adminBanUser: (id: string) => `/api/v1/admin/users/${id}/ban`,
} as const;
