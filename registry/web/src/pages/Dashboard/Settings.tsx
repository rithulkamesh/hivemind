import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { User } from "@/types";
import { Sidebar } from "@/components/layout/Sidebar";
import { Button } from "@/components/ui/button";
import { TotpSetup } from "@/components/auth/TotpSetup";

const schema = z.object({
  email: z.string().email(),
  username: z.string().min(2),
});

type FormData = z.infer<typeof schema>;

export function Settings() {
  const queryClient = useQueryClient();
  const { data: user } = useQuery({
    queryKey: ["me"],
    queryFn: () => api<User>(apiRoutes.me()),
  });

  const { register, handleSubmit, formState: { errors } } = useForm<FormData>({
    resolver: zodResolver(schema),
    values: user ? { email: user.email, username: user.username } : undefined,
  });

  const update = useMutation({
    mutationFn: (data: FormData) =>
      fetch(import.meta.env.VITE_API_BASE_URL + apiRoutes.me(), {
        method: "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${localStorage.getItem("access_token")}` },
        body: JSON.stringify(data),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["me"] }),
  });

  return (
    <div className="flex gap-8">
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
        <section className="mt-12">
          <h2 className="font-mono text-xs tracking-wider uppercase text-hm-muted mb-4">Two-factor authentication</h2>
          <TotpSetup onVerify={async () => {}} />
        </section>
      </div>
    </div>
  );
}
