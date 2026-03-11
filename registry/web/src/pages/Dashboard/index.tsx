import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { PackageCard } from "@/components/packages/PackageCard";
import { EmptyState } from "@/components/ui/EmptyState";

export function Dashboard() {
  const { data, isLoading } = useQuery({
    queryKey: ["packages", "mine"],
    queryFn: () => api<{ packages: Package[]; page: number }>(apiRoutes.packages({})),
  });
  const packages = data?.packages ?? [];

  return (
    <div className="flex gap-8">
      <Sidebar variant="dashboard" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Dashboard</h1>
        <p className="text-hm-muted mb-6">Your packages and recent activity.</p>
        <section>
          <h2 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-4">Your packages</h2>
          {isLoading ? (
            <p className="text-hm-muted">Loading…</p>
          ) : packages.length === 0 ? (
            <EmptyState
              title="No packages yet"
              description="Create your first package from the dashboard."
              action={<Link to="/dashboard/packages" className="text-hm-text underline">Manage packages</Link>}
            />
          ) : (
            <div className="space-y-2">
              {packages.slice(0, 5).map((pkg) => (
                <PackageCard key={pkg.id} pkg={pkg} />
              ))}
              <Link to="/dashboard/packages" className="text-sm text-hm-muted hover:text-hm-text">
                View all →
              </Link>
            </div>
          )}
        </section>
      </div>
    </div>
  );
}
