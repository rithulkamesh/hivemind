import { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useSession } from "@/store/auth";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

interface GuestGuardProps {
  children: ReactNode;
}

/** Renders children only when user is not logged in; otherwise redirects to dashboard or `from`. */
export function GuestGuard({ children }: GuestGuardProps) {
  const { data: session, isPending } = useSession();
  const location = useLocation();
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname;

  if (isPending) return <LoadingSkeleton />;
  if (session?.user) {
    return <Navigate to={from ?? "/dashboard"} replace />;
  }
  return <>{children}</>;
}
