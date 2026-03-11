import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import { Sidebar } from "@/components/layout/Sidebar";

interface Member {
  user_id: string;
  email: string;
  username: string;
  role: string;
}

export function OrgMembers() {
  const { slug } = useParams<{ slug: string }>();
  const { data: members, isLoading } = useQuery({
    queryKey: ["org", slug, "members"],
    queryFn: () => api<Member[]>(apiRoutes.orgMembers(slug!)),
    enabled: !!slug,
  });

  return (
    <div className="flex gap-8">
      <Sidebar variant="org" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Members</h1>
        {isLoading ? (
          <p className="text-hm-muted">Loading…</p>
        ) : (
          <div className="border border-hm-border overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-hm-border bg-hm-surface">
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">User</th>
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">Role</th>
                </tr>
              </thead>
              <tbody>
                {(members ?? []).map((m) => (
                  <tr key={m.user_id} className="border-b border-hm-border last:border-0">
                    <td className="px-4 py-2 text-hm-text">{m.username} ({m.email})</td>
                    <td className="px-4 py-2 text-hm-muted">{m.role}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
