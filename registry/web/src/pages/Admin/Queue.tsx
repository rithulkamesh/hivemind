import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import { Badge } from "@/components/ui/Badge";
import { Button } from "@/components/ui/button";

interface QueueItem {
  id: string;
  package_name: string;
  version: string;
  uploaded_at: string;
  verification_status: string;
}

export function AdminQueue() {
  const { data: list, isLoading } = useQuery({
    queryKey: ["admin", "verification-queue"],
    queryFn: () => api<QueueItem[]>(apiRoutes.adminVerificationQueue()),
  });

  return (
    <div>
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Verification queue</h1>
      {isLoading ? (
        <p className="text-hm-muted">Loading…</p>
      ) : (
        <div className="border border-hm-border overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-hm-border bg-hm-surface">
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Package</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Version</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Uploaded</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Status</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Actions</th>
              </tr>
            </thead>
            <tbody>
              {(list ?? []).map((row) => (
                <tr key={row.id} className="border-b border-hm-border last:border-0">
                  <td className="px-4 py-2 font-mono text-hm-text">{row.package_name}</td>
                  <td className="px-4 py-2 font-mono text-hm-text-passive">{row.version}</td>
                  <td className="px-4 py-2 text-hm-muted">
                    {new Date(row.uploaded_at).toLocaleString()}
                  </td>
                  <td className="px-4 py-2">
                    <Badge
                      variant={
                        row.verification_status === "passed"
                          ? "passed"
                          : row.verification_status === "failed"
                            ? "failed"
                            : "pending"
                      }
                    >
                      {row.verification_status}
                    </Badge>
                  </td>
                  <td className="px-4 py-2">
                    <Button variant="outline" className="mr-2">Approve</Button>
                    <Button variant="destructive">Reject</Button>
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
