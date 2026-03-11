import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { ApiKeyRow } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { ApiKeyForm } from "@/components/auth/ApiKeyForm";
import { Button } from "@/components/ui/button";
import { CopyButton } from "@/components/ui/CopyButton";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";

export function ApiKeys() {
  const [showForm, setShowForm] = useState(false);
  const [createdKey, setCreatedKey] = useState<{ key: string; prefix: string } | null>(null);
  const queryClient = useQueryClient();

  const { data: keys, isLoading } = useQuery({
    queryKey: ["api-keys"],
    queryFn: () => api<ApiKeyRow[]>(apiRoutes.apiKeys()),
  });

  const revoke = useMutation({
    mutationFn: (id: string) =>
      fetch(import.meta.env.VITE_API_BASE_URL + apiRoutes.revokeApiKey(id), { method: "DELETE" }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["api-keys"] }),
  });

  const handleCreate = async (data: { name: string; scopes: string[]; expires_days?: number }) => {
    const res = await fetch(import.meta.env.VITE_API_BASE_URL + apiRoutes.apiKeys(), {
      method: "POST",
      headers: { "Content-Type": "application/json", Authorization: `Bearer ${localStorage.getItem("access_token")}` },
      body: JSON.stringify(data),
    });
    if (!res.ok) throw new Error(await res.text());
    const out = (await res.json()) as { key: string; prefix: string };
    setCreatedKey(out);
    setShowForm(false);
    queryClient.invalidateQueries({ queryKey: ["api-keys"] });
    return out;
  };

  return (
    <div className="flex gap-8">
      <Sidebar variant="dashboard" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">API keys</h1>
        {createdKey && (
          <div className="mb-6 p-4 bg-amber-500/10 border border-amber-500/30">
            <p className="font-mono text-sm text-hm-text mb-2">Store this key securely. It will not be shown again.</p>
            <div className="flex items-center gap-2">
              <code className="flex-1 break-all text-amber-400">{createdKey.key}</code>
              <CopyButton text={createdKey.key} />
            </div>
            <button
              type="button"
              onClick={() => setCreatedKey(null)}
              className="mt-2 text-xs text-hm-muted hover:text-hm-text"
            >
              Dismiss
            </button>
          </div>
        )}
        {showForm ? (
          <ApiKeyForm onSubmit={handleCreate} onCancel={() => setShowForm(false)} />
        ) : (
          <Button onClick={() => setShowForm(true)}>Create key</Button>
        )}
        <div className="mt-8 border border-hm-border overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-hm-border bg-hm-surface">
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Name</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Prefix</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Scopes</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text">Last used</th>
                <th className="text-left px-4 py-2 font-semibold text-hm-text"></th>
              </tr>
            </thead>
            <tbody>
              {isLoading ? (
                <tr><td colSpan={5} className="px-4 py-4 text-hm-muted">Loading…</td></tr>
              ) : (keys ?? []).map((k) => (
                <tr key={k.id} className="border-b border-hm-border last:border-0">
                  <td className="px-4 py-2 text-hm-text">{k.name}</td>
                  <td className="px-4 py-2 font-mono text-hm-muted">{k.key_prefix}</td>
                  <td className="px-4 py-2 text-hm-muted">{k.scopes?.join(", ") ?? "—"}</td>
                  <td className="px-4 py-2 text-hm-muted">
                    {k.last_used_at ? new Date(k.last_used_at).toLocaleString() : "—"}
                  </td>
                  <td className="px-4 py-2">
                    <ConfirmDialog
                      trigger={<Button variant="destructive">Revoke</Button>}
                      title="Revoke API key"
                      description="This key will stop working immediately."
                      confirmLabel="Revoke"
                      variant="destructive"
                      onConfirm={async () => { await revoke.mutateAsync(k.id); }}
                    />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
