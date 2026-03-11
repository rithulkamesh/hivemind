import { clsx } from "clsx";

interface LoadingSkeletonProps {
  className?: string;
}

export function LoadingSkeleton({ className }: LoadingSkeletonProps) {
  return (
    <div
      className={clsx(
        "animate-pulse rounded bg-hm-surface",
        className
      )}
      style={{
        backgroundImage: "linear-gradient(90deg, transparent, rgba(245, 166, 35, 0.08), transparent)",
        backgroundSize: "200% 100%",
        animation: "shimmer 1.5s infinite",
      }}
    />
  );
}
