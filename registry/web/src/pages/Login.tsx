import { useState } from "react";
import { Link, useNavigate, useLocation } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import { Github, KeyRound } from "lucide-react";
import { signIn } from "@/store/auth";
import { authClient } from "@/lib/auth-client";
import { Button } from "@/components/ui/button";

const container = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.06 } } };
const item = { hidden: { opacity: 0, y: 12 }, show: { opacity: 1, y: 0 } };

function GoogleIcon({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden>
      <path
        fill="currentColor"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"
      />
      <path
        fill="currentColor"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"
      />
      <path
        fill="currentColor"
        d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"
      />
      <path
        fill="currentColor"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"
      />
    </svg>
  );
}

const schema = z.object({
  email: z.string().email(),
  password: z.string().min(1, "Password required"),
});

type FormData = z.infer<typeof schema>;

export function Login() {
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname ?? "/dashboard";
  const [error, setError] = useState("");

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setError("");
    const { error: err } = await signIn.email({ email: data.email, password: data.password });
    if (err) {
      setError(err.message ?? "Login failed");
      return;
    }
    navigate(from, { replace: true });
  };

  const onGitHub = async () => {
    setError("");
    await signIn.social({ provider: "github", callbackURL: from });
  };

  const onGoogle = async () => {
    setError("");
    await signIn.social({ provider: "google", callbackURL: from });
  };

  const onPasskey = async () => {
    setError("");
    const { error: err } = await authClient.signIn.passkey({
      fetchOptions: {
        onSuccess: () => navigate(from, { replace: true }),
        onError: (ctx) => setError(ctx.error?.message ?? "Passkey sign-in failed"),
      },
    });
    if (err) setError(err.message ?? "Passkey sign-in failed");
  };

  return (
    <motion.div
      className="max-w-sm mx-auto py-12"
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: [0.25, 0.46, 0.45, 0.94] }}
    >
      <motion.div variants={container} initial="hidden" animate="show">
        <motion.h1 variants={item} className="font-sans text-2xl font-semibold text-hm-text mb-6">Log in</motion.h1>
        {error && <motion.p variants={item} className="mb-4 text-sm text-red-400">{error}</motion.p>}
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" autoComplete="off">
        <motion.div variants={item}>
          <label className="block font-mono text-xs text-hm-muted mb-1">Email</label>
          <input
            {...register("email")}
            type="email"
            autoComplete="email"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.email && <p className="mt-1 text-sm text-red-400">{errors.email.message}</p>}
        </motion.div>
        <motion.div variants={item}>
          <label className="block font-mono text-xs text-hm-muted mb-1">Password</label>
          <input
            {...register("password")}
            type="password"
            autoComplete="current-password"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.password && <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>}
        </motion.div>
        <motion.div variants={item}>
        <Button type="submit" disabled={isSubmitting} className="w-full transition-transform hover:scale-[1.01] active:scale-[0.99]">
          {isSubmitting ? "Logging in…" : "Log in"}
        </Button>
        </motion.div>
      </form>
      <div className="mt-6 flex flex-col gap-2">
        <motion.div variants={item}>
        <Button
          variant="outline"
          onClick={onGitHub}
          className="w-full gap-2 transition-transform hover:scale-[1.01] active:scale-[0.99]"
        >
          <Github className="size-5 shrink-0" aria-hidden />
          Log in with GitHub
        </Button>
        <Button
          variant="outline"
          onClick={onGoogle}
          className="w-full gap-2 transition-transform hover:scale-[1.01] active:scale-[0.99]"
        >
          <GoogleIcon className="size-5 shrink-0" />
          Log in with Google
        </Button>
        <Button
          variant="outline"
          onClick={onPasskey}
          className="w-full gap-2 transition-transform hover:scale-[1.01] active:scale-[0.99]"
        >
          <KeyRound className="size-5 shrink-0" aria-hidden />
          Sign in with passkey
        </Button>
        </motion.div>
      </div>
      <motion.p variants={item} className="mt-6 text-sm text-hm-muted">
        Don&apos;t have an account? <Link to="/register" className="text-hm-text hover:underline">Register</Link>
      </motion.p>
      </motion.div>
    </motion.div>
  );
}
