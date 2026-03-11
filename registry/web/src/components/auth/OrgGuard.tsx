import { ReactNode } from "react";
import { Navigate, useParams } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Org } from "@/types";

interface OrgGuardProps {
  children: ReactNode;
}

export function OrgGuard({ children }: OrgGuardProps) {
  const { slug } = useParams<{ slug: string }>();
  const { data: orgs, isLoading } = useQuery({
    queryKey: ["orgs"],
    queryFn: () => api<Org[]>(apiRoutes.orgs()),
  });
  const member = orgs?.some((o) => o.name === slug);

  if (isLoading) return null;
  if (!member && slug) return <Navigate to="/dashboard" replace />;
  return <>{children}</>;
}
