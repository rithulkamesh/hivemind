import { useQuery } from "@tanstack/react-query";
import { useSession, signIn, signUp, signOut } from "@/lib/auth-client";
import { api, apiRoutes } from "@/lib/api";
import type { User } from "@/types";

export { useSession, signIn, signUp, signOut };

/** Current registry profile (/api/v1/me). Only runs when session exists. */
export function useMe() {
  const { data: session } = useSession();
  return useQuery({
    queryKey: ["me"],
    queryFn: () => api<User>(apiRoutes.me()),
    enabled: !!session?.user,
  });
}
