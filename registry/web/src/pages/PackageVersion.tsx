import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package, PackageVersion } from "@/types";
import { PackageHeader } from "@/components/packages/PackageHeader";
import { FileList } from "@/components/packages/FileList";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";
import type { PackageFile } from "@/types";

export function PackageVersion() {
  const { name, version } = useParams<{ name: string; version: string }>();

  const { data: pkg, isLoading } = useQuery({
    queryKey: ["package", name],
    queryFn: () => api<Package>(apiRoutes.package(name!)),
    enabled: !!name,
  });

  const { data: pv } = useQuery({
    queryKey: ["package", name, "version", version],
    queryFn: () => api<PackageVersion>(apiRoutes.version(name!, version!)),
    enabled: !!name && !!version,
  });

  const files: PackageFile[] = []; // API: list files for this version

  if (isLoading || !pkg) {
    return <LoadingSkeleton className="h-64 w-full" />;
  }

  return (
    <div>
      <PackageHeader pkg={pkg} version={version ?? undefined} />
      {pv && (
        <div className="mt-6 space-y-4">
          <p className="text-hm-muted text-sm">
            Requires Python: {pv.requires_python ?? "—"} · Requires Hivemind: {pv.requires_hivemind ?? "—"}
          </p>
          {files.length > 0 && <FileList packageName={pkg.name} files={files} />}
        </div>
      )}
    </div>
  );
}
