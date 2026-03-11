import { useState } from "react";
import { CopyButton } from "@/components/ui/CopyButton";

const REGISTRY_URL = import.meta.env.VITE_REGISTRY_URL ?? "https://registry.hivemind.rithul.dev/simple/";

type Tab = "pip" | "uv" | "hivemind";

function command(tab: Tab, name: string, version?: string): string {
  const pkg = version ? `${name}==${version}` : name;
  switch (tab) {
    case "pip":
      return `pip install --extra-index-url=${REGISTRY_URL} ${pkg}`;
    case "uv":
      return `uv add --index ${REGISTRY_URL} ${pkg}`;
    case "hivemind":
      return `hivemind plugins install ${name}`;
    default:
      return "";
  }
}

interface InstallCommandProps {
  name: string;
  version?: string;
}

export function InstallCommand({ name, version }: InstallCommandProps) {
  const [tab, setTab] = useState<Tab>("pip");
  const cmd = command(tab, name, version);

  return (
    <div className="border border-hm-border border-l-4 border-l-hm-amber bg-hm-code-bg">
      <div className="flex border-b border-hm-border">
        {(["pip", "uv", "hivemind"] as const).map((t) => (
          <button
            key={t}
            type="button"
            onClick={() => setTab(t)}
            className={`px-4 py-2 font-mono text-xs uppercase tracking-wider transition-colors ${tab === t
                ? "bg-hm-surface text-hm-text border-b-2 border-hm-amber -mb-px"
                : "text-hm-muted hover:text-hm-text-passive"
              }`}
          >
            {t}
          </button>
        ))}
      </div>
      <div className="flex items-center justify-between gap-2 px-4 py-3 font-mono text-sm text-hm-text">
        <code className="break-all">{cmd}</code>
        <CopyButton text={cmd} className="shrink-0" />
      </div>
    </div>
  );
}
