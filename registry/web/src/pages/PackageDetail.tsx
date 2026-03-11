import { useState } from "react";
import { useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package, PackageVersion, PackageFile } from "@/types";
import { PackageHeader } from "@/components/packages/PackageHeader";
import { VersionList } from "@/components/packages/VersionList";
import { FileList } from "@/components/packages/FileList";
import { DownloadChart } from "@/components/packages/DownloadChart";
import { ReadmeRenderer } from "@/components/packages/ReadmeRenderer";
import { ToolList } from "@/components/packages/ToolList";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

type Tab = "overview" | "versions" | "files" | "stats";

export function PackageDetail() {
  const { name } = useParams<{ name: string }>();
  const [tab, setTab] = useState<Tab>("overview");

  const { data: pkg, isLoading } = useQuery({
    queryKey: ["package", name],
    queryFn: () => api<Package>(apiRoutes.package(name!)),
    enabled: !!name,
  });

  const { data: versions } = useQuery({
    queryKey: ["package", name, "versions"],
    queryFn: async () => {
      await api<Package>(apiRoutes.package(name!));
      return api<PackageVersion[]>(`/api/v1/packages/${encodeURIComponent(name!)}/versions`).catch(() => []);
    },
    enabled: !!name && !!pkg,
  });

  const { data: downloads } = useQuery({
    queryKey: ["package", name, "downloads"],
    queryFn: () =>
      api<{ daily: { date: string; downloads: number }[] }>(apiRoutes.packageDownloads(name!)).catch(() => ({ daily: [] })),
    enabled: !!name && tab === "stats",
  });

  if (isLoading || !pkg) {
    return <LoadingSkeleton className="h-64 w-full" />;
  }

  const tabs: { id: Tab; label: string }[] = [
    { id: "overview", label: "Overview" },
    { id: "versions", label: "Versions" },
    { id: "files", label: "Files" },
    { id: "stats", label: "Stats" },
  ];

  const latestVersion = versions?.[0];
  const files: PackageFile[] = latestVersion
    ? [] // would need API: list files for version
    : [];

  const chartData = (downloads?.daily ?? []).map((d) => ({
    date: d.date,
    downloads: d.downloads,
  }));

  return (
    <div className="grid grid-cols-1 lg:grid-cols-[1fr_320px] gap-8">
      <div>
        <PackageHeader pkg={pkg} version={latestVersion?.version} />
        <div className="flex border-b border-hm-border mt-8 gap-0">
          {tabs.map((t) => (
            <button
              key={t.id}
              type="button"
              onClick={() => setTab(t.id)}
              className={`px-4 py-2 font-mono text-xs uppercase tracking-wider border-b-2 -mb-px transition-colors ${
                tab === t.id
                  ? "border-hm-text text-hm-text"
                  : "border-transparent text-hm-muted hover:text-hm-text-passive"
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <div className="mt-6">
          {tab === "overview" && (
            <div className="prose-registry max-w-none">
              {pkg.description ? (
                <ReadmeRenderer content={pkg.description} />
              ) : (
                <p className="text-hm-muted">No README for this package.</p>
              )}
            </div>
          )}
          {tab === "versions" && versions && (
            <VersionList packageName={pkg.name} versions={versions} />
          )}
          {tab === "files" && (
            files.length ? (
              <FileList packageName={pkg.name} files={files} />
            ) : (
              <p className="text-hm-muted">No files listed. Select a version to see files.</p>
            )
          )}
          {tab === "stats" && (
            <DownloadChart data={chartData} />
          )}
        </div>
      </div>
      <aside className="space-y-6">
        <div className="border border-hm-border p-4">
          <h4 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-2">Meta</h4>
          <ul className="text-sm text-hm-text-passive space-y-1">
            {pkg.license && <li>License: {pkg.license}</li>}
            {pkg.homepage && (
              <li>
                <a href={pkg.homepage} target="_blank" rel="noopener noreferrer" className="text-hm-text hover:underline">
                  Homepage
                </a>
              </li>
            )}
            {pkg.repository && (
              <li>
                <a href={pkg.repository} target="_blank" rel="noopener noreferrer" className="text-hm-text hover:underline">
                  Repository
                </a>
              </li>
            )}
            {pkg.keywords?.length ? (
              <li>Keywords: {pkg.keywords.join(", ")}</li>
            ) : null}
          </ul>
        </div>
        {latestVersion?.requires_hivemind && (
          <div className="border border-hm-border p-4">
            <h4 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-2">
              Hivemind compatibility
            </h4>
            <p className="font-mono text-sm text-hm-text">{latestVersion.requires_hivemind}</p>
          </div>
        )}
        {latestVersion && (latestVersion as PackageVersion & { tool_count?: number }).tool_count ? (
          <ToolList tools={[]} />
        ) : null}
      </aside>
    </div>
  );
}
