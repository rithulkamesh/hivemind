import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { Badge } from "@/components/ui/Badge";

export function AdminPackages() {
  const { data: list, isLoading } = useQuery({
    queryKey: ["packages", "all"],
    queryFn: () => api<{ packages: Package[]; page: number }>(apiRoutes.packages({ page: 1 })),
  });
  const packages = list?.packages ?? [];

  return (
    <div>
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Packages (moderation)</h1>
      {isLoading ? (
        <p className="text-hm-muted">Loading…</p>
      ) : (
        <div className="border border-hm-border overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-hm-border bg-hm-surface">
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Package</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Downloads</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Verified</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Trusted</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Actions</th>
              </tr>
            </thead>
            <tbody>
              {packages.map((pkg) => (
                <tr key={pkg.id} className="border-b border-hm-border last:border-0">
                  <td className="px-4 py-2">
                    <Link to={`/packages/${pkg.name}`} className="font-mono text-hm-text hover:underline">
                      {pkg.name}
                    </Link>
                  </td>
                  <td className="px-4 py-2 text-hm-muted">
                    {typeof pkg.total_downloads === "number" ? pkg.total_downloads.toLocaleString() : "—"}
                  </td>
                  <td className="px-4 py-2">
                    {pkg.verified ? <Badge variant="verified">Yes</Badge> : <span className="text-hm-muted">—</span>}
                  </td>
                  <td className="px-4 py-2">
                    {pkg.trusted ? <Badge variant="trusted">Yes</Badge> : <span className="text-hm-muted">—</span>}
                  </td>
                  <td className="px-4 py-2">
                    <Link to={`/packages/${pkg.name}`} className="text-hm-muted hover:text-hm-text text-xs">
                      View
                    </Link>
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
