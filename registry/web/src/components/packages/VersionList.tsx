import { Link } from "react-router-dom";
import type { PackageVersion } from "@/types";
import { Badge } from "@/components/ui/Badge";

interface VersionListProps {
  packageName: string;
  versions: PackageVersion[];
}

function statusVariant(s: string | null | undefined): "pending" | "running" | "passed" | "failed" | "default" {
  if (!s) return "default";
  switch (s) {
    case "pending": return "pending";
    case "running": return "running";
    case "passed": return "passed";
    case "failed": return "failed";
    default: return "default";
  }
}

export function VersionList({ packageName, versions }: VersionListProps) {
  return (
    <div className="border border-hm-border overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-hm-border bg-hm-surface">
            <th className="text-left px-4 py-2 font-semibold text-hm-text">Version</th>
            <th className="text-left px-4 py-2 font-semibold text-hm-text">Uploaded</th>
            <th className="text-left px-4 py-2 font-semibold text-hm-text">Requires Python</th>
            <th className="text-left px-4 py-2 font-semibold text-hm-text">Status</th>
          </tr>
        </thead>
        <tbody>
          {versions.map((v) => (
            <tr key={v.id} className="border-b border-hm-border last:border-0 hover:bg-hm-surface/50">
              <td className="px-4 py-2">
                <Link
                  to={`/packages/${packageName}/v/${v.version}`}
                  className="text-hm-text hover:underline font-mono"
                >
                  {v.version}
                </Link>
                {v.yanked && <Badge variant="yanked" className="ml-2">Yanked</Badge>}
              </td>
              <td className="px-4 py-2 text-hm-muted">
                {v.uploaded_at ? new Date(v.uploaded_at).toLocaleDateString() : "—"}
              </td>
              <td className="px-4 py-2 font-mono text-hm-text-passive">
                {v.requires_python ?? "—"}
              </td>
              <td className="px-4 py-2">
                <Badge variant={statusVariant(v.verification_status)}>
                  {v.verification_status ?? "pending"}
                </Badge>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
