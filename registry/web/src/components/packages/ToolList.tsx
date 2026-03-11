interface ToolListProps {
  tools: string[];
}

export function ToolList({ tools }: ToolListProps) {
  if (!tools.length) return null;
  return (
    <div>
      <h4 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-2">
        Tools provided
      </h4>
      <ul className="flex flex-wrap gap-2">
        {tools.map((t) => (
          <li
            key={t}
            className="px-2 py-1 bg-hm-surface border border-hm-border font-mono text-sm text-hm-text-passive"
          >
            {t}
          </li>
        ))}
      </ul>
    </div>
  );
}
