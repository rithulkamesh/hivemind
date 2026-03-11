import { create } from "zustand";
import { persist } from "zustand/middleware";
import { api, apiRoutes } from "@/lib/api";
import type { User, Org } from "@/types";

const BASE = import.meta.env.VITE_API_BASE_URL ?? "";

interface AuthState {
  user: User | null;
  token: string | null;
  refreshToken: string | null;
  org: Org | null;
  login: (email: string, password: string) => Promise<void>;
  loginWithGitHub: () => void;
  loginWithGoogle: () => void;
  refresh: () => Promise<void>;
  logout: () => void;
  setOrg: (org: Org | null) => void;
  fetchMe: () => Promise<void>;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: localStorage.getItem("access_token"),
      refreshToken: localStorage.getItem("refresh_token"),
      org: null,

      login: async (email, password) => {
        const res = await fetch(`${BASE}${apiRoutes.login()}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ email, password }),
        });
        if (!res.ok) throw new Error(await res.text());
        const data = (await res.json()) as {
          access_token: string;
          refresh_token: string;
        };
        localStorage.setItem("access_token", data.access_token);
        localStorage.setItem("refresh_token", data.refresh_token);
        set({ token: data.access_token, refreshToken: data.refresh_token });
        await get().fetchMe();
      },

      loginWithGitHub: () => {
        window.location.href = `${BASE}/auth/github`;
      },

      loginWithGoogle: () => {
        window.location.href = `${BASE}/auth/google`;
      },

      refresh: async () => {
        const refreshToken = get().refreshToken;
        if (!refreshToken) throw new Error("No refresh token");
        const res = await fetch(`${BASE}${apiRoutes.refresh()}`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (!res.ok) {
          set({ user: null, token: null, refreshToken: null });
          localStorage.removeItem("access_token");
          localStorage.removeItem("refresh_token");
          throw new Error("Refresh failed");
        }
        const data = (await res.json()) as { access_token: string };
        localStorage.setItem("access_token", data.access_token);
        set({ token: data.access_token });
      },

      logout: () => {
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        set({ user: null, token: null, refreshToken: null, org: null });
      },

      setOrg: (org) => set({ org }),

      fetchMe: async () => {
        const token = get().token;
        if (!token) return;
        try {
          const user = await api<User>(apiRoutes.me());
          set({ user });
        } catch {
          try {
            await get().refresh();
            const user = await api<User>(apiRoutes.me());
            set({ user });
          } catch {
            get().logout();
          }
        }
      },
    }),
    { name: "auth", partialize: (s) => ({ token: s.token, refreshToken: s.refreshToken }) }
  )
);
