import { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useSession } from "@/store/auth";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

interface AuthGuardProps {
  children: ReactNode;
}

export function AuthGuard({ children }: AuthGuardProps) {
  const { data: session, isPending } = useSession();
  const location = useLocation();

  if (isPending) return <LoadingSkeleton />;
  if (!session?.user) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}
