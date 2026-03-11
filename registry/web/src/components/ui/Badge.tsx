import { clsx } from "clsx";

type BadgeVariant = "verified" | "trusted" | "yanked" | "pending" | "passed" | "failed" | "running" | "default";

const variantClasses: Record<BadgeVariant, string> = {
  verified: "bg-amber-500/20 text-amber-400 border-amber-500/40",
  trusted: "bg-green-500/20 text-green-400 border-green-500/40",
  yanked: "bg-red-500/20 text-red-400 border-red-500/40",
  pending: "bg-hm-muted/20 text-hm-muted border-hm-border",
  passed: "bg-green-500/20 text-green-400 border-green-500/40",
  failed: "bg-red-500/20 text-red-400 border-red-500/40",
  running: "bg-amber-500/20 text-amber-400 border-amber-500/40",
  default: "bg-hm-surface text-hm-text-passive border-hm-border",
};

interface BadgeProps {
  variant?: BadgeVariant;
  children: React.ReactNode;
  className?: string;
}

export function Badge({ variant = "default", children, className }: BadgeProps) {
  return (
    <span
      className={clsx(
        "inline-flex items-center font-mono text-[10px] tracking-wider uppercase px-2 py-0.5 border",
        variantClasses[variant],
        className
      )}
    >
      {children}
    </span>
  );
}
