import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { motion } from "framer-motion";
import { apiRoutes, apiBase } from "@/lib/api";
import { getApiToken } from "@/lib/auth-client";
import { useMe, useSession } from "@/store/auth";
import { Sidebar } from "@/components/layout/Sidebar";
import { Button } from "@/components/ui/button";
import { TotpSetup } from "@/components/auth/TotpSetup";
import { authClient } from "@/lib/auth-client";
import { ConfirmDialog } from "@/components/ui/ConfirmDialog";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";

const schema = z.object({
  email: z.string().email().optional(),
  username: z.string().min(2),
  bio: z.string().optional(),
  website: z.string().url().optional().or(z.literal("")),
});

type FormData = z.infer<typeof schema>;

export function Settings() {
  const queryClient = useQueryClient();
  const { data: session } = useSession();
  const { data: user, isPending: mePending } = useMe();
  const sessionUser = session?.user as { email?: string; name?: string; username?: string; twoFactorEnabled?: boolean } | undefined;

  // Merge /me with session for OAuth users (GitHub may not have email in /me)
  const email = user?.email ?? sessionUser?.email ?? "";
  const username = user?.username ?? sessionUser?.username ?? (sessionUser?.name ?? "");

  const { register, handleSubmit, formState: { errors }, reset } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { email: "", username: "" },
  });
  useEffect(() => {
    if (email !== undefined || username !== undefined) {
      reset({ email: email ?? "", username: username || "" });
    }
  }, [email, username, reset]);

  const update = useMutation({
    mutationFn: async (data: FormData) => {
      const token = await getApiToken();
      const res = await fetch(apiBase + apiRoutes.me(), {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          ...(token ? { Authorization: `Bearer ${token}` } : {}),
        },
        body: JSON.stringify(data),
      });
      if (!res.ok) throw new Error(await res.text());
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["me"] }),
  });

  // ——— Has password (credential account) ———
  const [accounts, setAccounts] = useState<{ providerId: string }[] | null>(null);
  useEffect(() => {
    let cancelled = false;
    authClient.listAccounts?.().then((list) => {
      if (!cancelled && Array.isArray(list)) setAccounts(list);
      else if (!cancelled) setAccounts([]);
    }).catch(() => { if (!cancelled) setAccounts([]); });
    return () => { cancelled = true; };
  }, []);
  const hasPassword = Array.isArray(accounts) && accounts.some((a) => a.providerId === "credential" || a.providerId === "email");



  // ——— 2FA ———
  const twoFactorEnabled = !!sessionUser?.twoFactorEnabled;
  const [totpURI, setTotpURI] = useState<string | null>(null);
  const [twoFactorPassword, setTwoFactorPassword] = useState("");
  const [twoFactorError, setTwoFactorError] = useState("");
  const [enabling2FA, setEnabling2FA] = useState(false);

  const onEnable2FA = async () => {
    setTwoFactorError("");
    if (!twoFactorPassword.trim()) {
      setTwoFactorError("Enter your password to enable 2FA.");
      return;
    }
    setEnabling2FA(true);
    const { data, error } = await authClient.twoFactor.enable({ password: twoFactorPassword });
    setEnabling2FA(false);
    setTwoFactorPassword("");
    if (error) {
      setTwoFactorError(error.message ?? "Failed to enable 2FA");
      return;
    }
    if (data?.totpURI) setTotpURI(data.totpURI);
  };

  const onVerifyTOTP = async (code: string) => {
    setTwoFactorError("");
    const { error } = await authClient.twoFactor.verifyTotp({ code });
    if (error) {
      setTwoFactorError(error.message ?? "Invalid code");
      return;
    }
    setTotpURI(null);
    await authClient.getSession(); // refresh session
    queryClient.invalidateQueries({ queryKey: ["me"] });
  };

  const onDisable2FA = async () => {
    setTwoFactorError("");
    const pass = window.prompt("Enter your password to disable 2FA:");
    if (!pass) return;
    const { error } = await authClient.twoFactor.disable({ password: pass });
    if (error) setTwoFactorError(error.message ?? "Failed to disable 2FA");
    else {
      await authClient.getSession();
      queryClient.invalidateQueries({ queryKey: ["me"] });
    }
  };

  // ——— Passkeys ———
  const [passkeys, setPasskeys] = useState<{ id: string; name?: string }[]>([]);
  const [passkeysLoading, setPasskeysLoading] = useState(true);
  useEffect(() => {
    let cancelled = false;
    (async () => {
      setPasskeysLoading(true);
      const { data } = await authClient.passkey?.listUserPasskeys?.() ?? { data: null };
      if (!cancelled && Array.isArray(data)) setPasskeys(data);
      else if (!cancelled) setPasskeys([]);
      setPasskeysLoading(false);
    })();
    return () => { cancelled = true; };
  }, []);

  const onDeletePasskey = async (id: string) => {
    const { error } = await authClient.passkey?.deletePasskey?.({ id }) ?? { error: { message: "Not available" } };
    if (!error) setPasskeys((prev) => prev.filter((p) => p.id !== id));
  };

  if (mePending && !user) {
    return (
      <div className="flex gap-8">
        <Sidebar variant="dashboard" />
        <div className="flex-1 min-w-0">
          <LoadingSkeleton className="h-8 w-48 mb-6" />
          <LoadingSkeleton className="h-10 w-full max-w-md" />
          <LoadingSkeleton className="h-10 w-full max-w-md mt-4" />
        </div>
      </div>
    );
  }

  return (
    <motion.div
      className="flex gap-8"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
    >
      <Sidebar variant="dashboard" />
      <div className="flex-1 min-w-0">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Settings</h1>
        <form onSubmit={handleSubmit((d) => update.mutate(d))} className="space-y-4 max-w-md">
          <div>
            <label className="block font-mono text-xs text-hm-muted mb-1">Email</label>
            <input
              {...register("email")}
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted"
            />
            {errors.email && <p className="text-sm text-red-400">{errors.email.message}</p>}
          </div>
          <div>
            <label className="block font-mono text-xs text-hm-muted mb-1">Username</label>
            <input
              {...register("username")}
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted"
            />
            {errors.username && <p className="text-sm text-red-400">{errors.username.message}</p>}
          </div>
          <Button type="submit" disabled={update.isPending}>Save</Button>
        </form>



        {hasPassword && (
          <>
            <section className="mt-12">
              <h2 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-4">Two-factor authentication</h2>
              {twoFactorError && <p className="text-sm text-red-400 mb-2">{twoFactorError}</p>}
              {twoFactorEnabled && !totpURI && (
                <div className="space-y-2">
                  <p className="text-sm text-hm-text">2FA is enabled.</p>
                  <Button type="button" variant="outline" onClick={onDisable2FA}>Disable 2FA</Button>
                </div>
              )}
              {!twoFactorEnabled && !totpURI && (
                <div className="space-y-3 max-w-sm">
                  <p className="text-sm text-hm-muted">Add a second factor using an authenticator app.</p>
                  <input
                    type="password"
                    value={twoFactorPassword}
                    onChange={(e) => setTwoFactorPassword(e.target.value)}
                    placeholder="Your password"
                    className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted"
                  />
                  <Button type="button" onClick={onEnable2FA} disabled={enabling2FA}>
                    {enabling2FA ? "Preparing…" : "Enable 2FA"}
                  </Button>
                </div>
              )}
              {totpURI && (
                <TotpSetup
                  totpURI={totpURI}
                  onVerify={onVerifyTOTP}
                />
              )}
            </section>

            <section className="mt-8">
              <h2 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-4">Passkeys</h2>
              <p className="text-sm text-hm-muted mb-2">Use passkeys for passwordless sign-in.</p>
              <div className="flex flex-wrap gap-2 items-center mb-4">
                <Button
                  type="button"
                  variant="outline"
                  onClick={() => authClient.passkey?.addPasskey?.({ name: "Registry" }).then(() => {
                    authClient.passkey?.listUserPasskeys?.().then(({ data }) => {
                      if (Array.isArray(data)) setPasskeys(data);
                    });
                  })}
                >
                  Register passkey
                </Button>
              </div>
              {passkeysLoading ? (
                <LoadingSkeleton className="h-10 w-64" />
              ) : passkeys.length > 0 ? (
                <ul className="space-y-2 border border-hm-border rounded p-3 bg-hm-surface/50">
                  {passkeys.map((p) => (
                    <li key={p.id} className="flex items-center justify-between gap-4 py-2 border-b border-hm-border last:border-0">
                      <span className="font-mono text-sm text-hm-text">{p.name ?? p.id.slice(0, 8)}</span>
                      <ConfirmDialog
                        trigger={<Button variant="destructive" size="sm">Remove</Button>}
                        title="Remove passkey"
                        description="This passkey will no longer work for sign-in."
                        confirmLabel="Remove"
                        variant="destructive"
                        onConfirm={() => onDeletePasskey(p.id)}
                      />
                    </li>
                  ))}
                </ul>
              ) : (
                <p className="text-sm text-hm-muted">No passkeys yet. Register one above.</p>
              )}
            </section>
          </>
        )}
      </div>
    </motion.div>
  );
}
