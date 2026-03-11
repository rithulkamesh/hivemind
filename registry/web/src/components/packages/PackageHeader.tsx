import type { Package } from "@/types";
import { Badge } from "@/components/ui/Badge";
import { InstallCommand } from "./InstallCommand";

interface PackageHeaderProps {
  pkg: Package;
  version?: string;
}

export function PackageHeader({ pkg, version }: PackageHeaderProps) {
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center gap-2">
        <h1 className="font-sans text-2xl font-semibold text-hm-text">{pkg.display_name || pkg.name}</h1>
        {version && (
          <Badge variant="default">{version}</Badge>
        )}
        {pkg.verified && <Badge variant="verified">Verified</Badge>}
        {pkg.trusted && <Badge variant="trusted">Trusted</Badge>}
      </div>
      {pkg.description && (
        <p className="text-hm-text-passive max-w-prose">{pkg.description}</p>
      )}
      <InstallCommand name={pkg.name} version={version} />
    </div>
  );
}
