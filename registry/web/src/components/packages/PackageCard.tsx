import { Link } from "react-router-dom";
import type { Package } from "@/types";
import { Badge } from "@/components/ui/Badge";

interface PackageCardProps {
  pkg: Package;
}

export function PackageCard({ pkg }: PackageCardProps) {
  return (
    <Link
      to={`/packages/${pkg.name}`}
      className="block p-4 bg-hm-surface/40 border border-hm-border hover:border-hm-muted/50 hover:bg-hm-surface/60 transition-all"
    >
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h3 className="font-sans font-medium text-hm-text truncate">{pkg.name}</h3>
          {pkg.description && (
            <p className="mt-1 text-sm text-hm-muted line-clamp-2">{pkg.description}</p>
          )}
        </div>
        <div className="flex shrink-0 gap-1">
          {pkg.verified && <Badge variant="verified">Verified</Badge>}
          {pkg.trusted && <Badge variant="trusted">Trusted</Badge>}
        </div>
      </div>
      {typeof pkg.total_downloads === "number" && (
        <p className="mt-2 font-mono text-xs text-hm-muted">
          {pkg.total_downloads.toLocaleString()} downloads
        </p>
      )}
    </Link>
  );
}
