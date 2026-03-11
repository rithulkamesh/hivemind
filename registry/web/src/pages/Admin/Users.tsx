import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import { Button } from "@/components/ui/button";

interface UserRow {
  id: string;
  email: string;
  username: string;
  created_at: string;
  banned?: boolean;
}

export function AdminUsers() {
  const { data: users, isLoading } = useQuery({
    queryKey: ["admin", "users"],
    queryFn: () => api<UserRow[]>(apiRoutes.adminUsers()),
  });

  return (
    <div>
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Users</h1>
      {isLoading ? (
        <p className="text-hm-muted">Loading…</p>
      ) : (
        <div className="border border-hm-border overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-hm-border bg-hm-surface">
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Username</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Email</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Created</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Actions</th>
              </tr>
            </thead>
            <tbody>
              {(users ?? []).map((u) => (
                <tr key={u.id} className="border-b border-hm-border last:border-0">
                  <td className="px-4 py-2 font-mono text-hm-text">{u.username}</td>
                  <td className="px-4 py-2 text-hm-muted">{u.email}</td>
                  <td className="px-4 py-2 text-hm-muted">{new Date(u.created_at).toLocaleDateString()}</td>
                  <td className="px-4 py-2">
                    {!u.banned && (
                      <Button variant="destructive">Ban</Button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
