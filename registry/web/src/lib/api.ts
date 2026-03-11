export const apiBase = import.meta.env.VITE_API_BASE_URL ?? "";
const BASE = apiBase;

function getToken(): string | null {
  return localStorage.getItem("access_token");
}

export async function api<T>(
  path: string,
  opts?: RequestInit & { params?: Record<string, string> }
): Promise<T> {
  const url = new URL(path, BASE);
  if (opts?.params) {
    Object.entries(opts.params).forEach(([k, v]) => url.searchParams.set(k, v));
  }
  const token = getToken();
  const headers: HeadersInit = {
    "Content-Type": "application/json",
    ...opts?.headers,
  };
  if (token) (headers as Record<string, string>)["Authorization"] = `Bearer ${token}`;
  const res = await fetch(url.toString(), { ...opts, headers });
  if (!res.ok) {
    const text = await res.text();
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
  version: (name: string, version: string) =>
    `/api/v1/packages/${encodeURIComponent(name)}/${encodeURIComponent(version)}`,
  search: (q: string, page = 1) => `/api/v1/search?q=${encodeURIComponent(q)}&page=${page}`,
  stats: () => "/api/v1/stats",
  packageImages: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}/images`,
  packageDownloads: (name: string) => `/api/v1/packages/${encodeURIComponent(name)}/downloads`,
  login: () => "/auth/login",
  register: () => "/auth/register",
  refresh: () => "/auth/refresh",
  logout: () => "/auth/logout",
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
