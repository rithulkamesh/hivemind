import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { PackageCard } from "@/components/packages/PackageCard";
import { EmptyState } from "@/components/ui/EmptyState";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

export function Dashboard() {
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
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Dashboard</h1>
        <p className="text-hm-muted mb-6">Your packages and recent activity.</p>
        <section>
          <h2 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-4">Your packages</h2>
          {isLoading ? (
            <div className="space-y-2">
              <LoadingSkeleton className="h-16 w-full" />
              <LoadingSkeleton className="h-16 w-full" />
              <LoadingSkeleton className="h-16 w-full" />
            </div>
          ) : packages.length === 0 ? (
            <EmptyState
              title="No packages yet"
              description="Create your first package to start publishing."
              action={
                <div className="flex flex-wrap gap-2 justify-center">
                  <Link to="/dashboard/packages/new" className="inline-flex items-center justify-center font-mono text-xs tracking-wider uppercase px-4 py-2 border border-hm-border text-hm-text hover:bg-hm-surface">
                    Create package
                  </Link>
                  <Link to="/dashboard/packages" className="inline-flex items-center justify-center font-mono text-xs tracking-wider uppercase px-4 py-2 text-hm-muted hover:text-hm-text">
                    Manage packages
                  </Link>
                </div>
              }
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
    </motion.div>
  );
}
