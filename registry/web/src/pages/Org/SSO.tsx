import { useParams } from "react-router-dom";
import { Sidebar } from "@/components/layout/Sidebar";
import { SamlConfig } from "@/components/auth/SamlConfig";

export function OrgSSO() {
  useParams<{ slug: string }>();

  return (
    <div className="flex gap-8">
      <Sidebar variant="org" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">SSO</h1>
        <p className="text-hm-muted mb-6">Configure SAML or OIDC for this organization.</p>
        <SamlConfig onSubmit={async () => {}} />
      </div>
    </div>
  );
}
