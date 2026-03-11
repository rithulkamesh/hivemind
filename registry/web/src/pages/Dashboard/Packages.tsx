import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { Badge } from "@/components/ui/Badge";
import { EmptyState } from "@/components/ui/EmptyState";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

export function DashboardPackages() {
  const { data, isLoading } = useQuery({
    queryKey: ["packages", "mine"],
    queryFn: () => api<{ packages: Package[]; page: number }>(apiRoutes.packages({})),
  });
  const packages = data?.packages ?? [];

  return (
    <motion.div
      className="flex gap-8"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
    >
      <Sidebar variant="dashboard" />
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-4 mb-6">
          <h1 className="font-sans text-2xl font-semibold text-hm-text">Packages</h1>
          <Link to="/dashboard/packages/new">
            <span className="inline-flex items-center justify-center font-mono text-xs tracking-wider uppercase px-4 py-2 border border-hm-border text-hm-text hover:bg-hm-surface">
              Create package
            </span>
          </Link>
        </div>
        {isLoading ? (
          <div className="space-y-2">
            <LoadingSkeleton className="h-12 w-full" />
            <LoadingSkeleton className="h-12 w-full" />
            <LoadingSkeleton className="h-12 w-full" />
          </div>
        ) : packages.length === 0 ? (
          <EmptyState
            title="No packages"
            description="Create a package to start publishing."
            action={<Link to="/dashboard/packages/new" className="inline-flex items-center justify-center font-mono text-xs tracking-wider uppercase px-4 py-2 border border-hm-border text-hm-text hover:bg-hm-surface">Create package</Link>}
          />
        ) : (
          <div className="border border-hm-border overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-hm-border bg-hm-surface">
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">Package</th>
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">Downloads</th>
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">Status</th>
                  <th className="text-left px-4 py-2 font-semibold text-hm-text">Actions</th>
                </tr>
              </thead>
              <tbody>
                {packages.map((pkg) => (
                  <tr key={pkg.id} className="border-b border-hm-border last:border-0">
                    <td className="px-4 py-2">
                      <Link to={`/packages/${pkg.name}`} className="text-hm-text hover:underline font-mono">
                        {pkg.name}
                      </Link>
                    </td>
                    <td className="px-4 py-2 text-hm-muted">
                      {typeof pkg.total_downloads === "number" ? pkg.total_downloads.toLocaleString() : "—"}
                    </td>
                    <td className="px-4 py-2">
                      <Badge variant="default">—</Badge>
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
    </motion.div>
  );
}
