import { useState, useEffect } from "react";
import { Link } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package, Stats } from "@/types";
import { PackageCard } from "@/components/packages/PackageCard";
import { CopyButton } from "@/components/ui/CopyButton";

const REGISTRY_URL = import.meta.env.VITE_REGISTRY_URL ?? "https://registry.hivemind.rithul.dev/simple/";
const PLACEHOLDER_PLUGIN = "hivemind-plugin-foo";
const DEFAULT_CMD = `pip install --index-url ${REGISTRY_URL} ${PLACEHOLDER_PLUGIN}`;

function useCountUp(end: number, duration = 1500, deps: unknown[] = []) {
  const [count, setCount] = useState(0);
  useEffect(() => {
    if (end === 0) return;
    let start = 0;
    const step = end / (duration / 16);
    const id = setInterval(() => {
      start += step;
      if (start >= end) {
        setCount(end);
        clearInterval(id);
      } else {
        setCount(Math.floor(start));
      }
    }, 16);
    return () => clearInterval(id);
  }, [end, duration, ...deps]);
  return count;
}

export function Home() {
  const { data: stats } = useQuery({
    queryKey: ["stats"],
    queryFn: () => api<Stats>(apiRoutes.stats()),
  });
  const { data: listData } = useQuery({
    queryKey: ["packages", "recent"],
    queryFn: () => api<{ packages: Package[]; page: number }>(apiRoutes.packages({ page: 1 })),
  });

  const totalPackages = stats?.total_packages ?? 0;
  const totalDownloads = Number(stats?.total_downloads ?? 0);
  const verifiedCount = 0; // API could expose this
  const packages = listData?.packages ?? [];
  const recent = packages.slice(0, 10);
  const featured = packages.filter((p) => (p as Package & { trusted?: boolean }).trusted).slice(0, 4);
  const displayFeatured = featured.length >= 4 ? featured : packages.slice(0, 4);

  const countPkg = useCountUp(totalPackages, 1200, [totalPackages]);
  const countDl = useCountUp(totalDownloads, 1200, [totalDownloads]);
  const countVerified = useCountUp(verifiedCount, 1200, [verifiedCount]);

  return (
    <div className="max-w-4xl">
      <p className="font-mono text-[11px] tracking-[0.2em] uppercase text-hm-muted mb-10">
        Hivemind Plugin Registry
      </p>
      <h1 className="font-sans text-3xl sm:text-4xl font-bold text-hm-text leading-tight tracking-tight mb-2">
        hivemind plugin registry
      </h1>
      <p className="text-hm-muted mb-6 max-w-prose">
        Discover, install, and publish hivemind plugins.
      </p>

      <div className="border-l-4 border-l-hm-amber bg-hm-code-bg border border-hm-border flex items-center justify-between gap-4 px-4 py-3 font-mono text-sm text-hm-text mb-10">
        <code className="break-all flex-1 min-w-0">
          pip install --index-url {REGISTRY_URL} {"{plugin}"}
        </code>
        <CopyButton text={DEFAULT_CMD} />
      </div>

      <div className="flex flex-wrap gap-8 font-mono text-sm text-hm-muted mb-12">
        <span>{countPkg.toLocaleString()} packages</span>
        <span>{countDl.toLocaleString()} downloads</span>
        <span>{countVerified.toLocaleString()} verified</span>
      </div>

      {displayFeatured.length > 0 && (
        <section className="mb-12">
          <h2 className="font-sans text-lg font-semibold text-hm-text mb-4">Featured packages</h2>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {displayFeatured.map((pkg) => (
              <PackageCard key={pkg.id} pkg={pkg} />
            ))}
          </div>
        </section>
      )}

      {recent.length > 0 && (
        <section>
          <h2 className="font-sans text-lg font-semibold text-hm-text mb-4">Recent packages</h2>
          <ul className="space-y-2">
            {recent.map((pkg) => (
              <li key={pkg.id}>
                <Link
                  to={`/packages/${pkg.name}`}
                  className="text-hm-text hover:underline font-sans"
                >
                  {pkg.name}
                </Link>
              </li>
            ))}
          </ul>
        </section>
      )}
    </div>
  );
}
