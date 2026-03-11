import { useState } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { apiRoutes, apiBase } from "@/lib/api";
import { Button } from "@/components/ui/button";

const schema = z
  .object({
    email: z.string().email(),
    username: z.string().min(2, "At least 2 characters"),
    password: z.string().min(8, "At least 8 characters"),
    confirm: z.string(),
  })
  .refine((d) => d.password === d.confirm, { message: "Passwords must match", path: ["confirm"] });

type FormData = z.infer<typeof schema>;

export function Register() {
  const navigate = useNavigate();
  const [error, setError] = useState<string>("");

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setError("");
    try {
      const r = await fetch(apiBase + apiRoutes.register(), {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          email: data.email,
          username: data.username,
          password: data.password,
        }),
      });
      if (!r.ok) {
        const body = await r.text();
        throw new Error(body || "Registration failed");
      }
      navigate("/login", { replace: true });
    } catch (e) {
      setError(e instanceof Error ? e.message : "Registration failed");
    }
  };

  return (
    <div className="max-w-sm mx-auto py-12 animate-in">
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Register</h1>
      {error && <p className="mb-4 text-sm text-red-400">{error}</p>}
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" autoComplete="off">
        <div>
          <label className="block font-mono text-xs text-hm-muted mb-1">Email</label>
          <input
            {...register("email")}
            type="email"
            autoComplete="email"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.email && <p className="mt-1 text-sm text-red-400">{errors.email.message}</p>}
        </div>
        <div>
          <label className="block font-mono text-xs text-hm-muted mb-1">Username</label>
          <input
            {...register("username")}
            autoComplete="username"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.username && <p className="mt-1 text-sm text-red-400">{errors.username.message}</p>}
        </div>
        <div>
          <label className="block font-mono text-xs text-hm-muted mb-1">Password</label>
          <input
            {...register("password")}
            type="password"
            autoComplete="new-password"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.password && <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>}
        </div>
        <div>
          <label className="block font-mono text-xs text-hm-muted mb-1">Confirm password</label>
          <input
            {...register("confirm")}
            type="password"
            autoComplete="new-password"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.confirm && <p className="mt-1 text-sm text-red-400">{errors.confirm.message}</p>}
        </div>
        <Button type="submit" disabled={isSubmitting} className="w-full transition-transform hover:scale-[1.01] active:scale-[0.99]">
          {isSubmitting ? "Creating account…" : "Register"}
        </Button>
      </form>
      <p className="mt-6 text-sm text-hm-muted">
        Already have an account? <Link to="/login" className="text-hm-text hover:underline">Log in</Link>
      </p>
    </div>
  );
}
