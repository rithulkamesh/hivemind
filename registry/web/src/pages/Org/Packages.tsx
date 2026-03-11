import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { PackageCard } from "@/components/packages/PackageCard";
import { EmptyState } from "@/components/ui/EmptyState";

export function OrgPackages() {
  const { slug } = useParams<{ slug: string }>();
  const { data: packages, isLoading } = useQuery({
    queryKey: ["org", slug, "packages"],
    queryFn: () => api<Package[]>(apiRoutes.orgPackages(slug!)),
    enabled: !!slug,
  });

  return (
    <div className="flex gap-8">
      <Sidebar variant="org" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Packages</h1>
        {isLoading ? (
          <p className="text-hm-muted">Loading…</p>
        ) : (packages ?? []).length === 0 ? (
          <EmptyState title="No packages" description="This org has no packages yet." />
        ) : (
          <div className="space-y-2">
            {(packages ?? []).map((pkg) => (
              <PackageCard key={pkg.id} pkg={pkg} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
