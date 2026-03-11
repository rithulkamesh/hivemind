import { useState } from "react";
import { Link } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import { signUp } from "@/store/auth";
import { Button } from "@/components/ui/button";

const container = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.05 } } };
const item = { hidden: { opacity: 0, y: 10 }, show: { opacity: 1, y: 0 } };

const schema = z
  .object({
    email: z.string().email(),
    username: z.string().min(2, "At least 2 characters"),
    password: z.string().min(12, "At least 12 characters").regex(/[A-Z]/, "Include an uppercase letter").regex(/\d/, "Include a number").regex(/[^A-Za-z0-9]/, "Include a special character"),
    confirm: z.string(),
  })
  .refine((d) => d.password === d.confirm, { message: "Passwords must match", path: ["confirm"] });

type FormData = z.infer<typeof schema>;

export function Register() {
  const [error, setError] = useState<string>("");
  const [registered, setRegistered] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setError("");
    const { error: err } = await signUp.email({
      email: data.email,
      password: data.password,
      name: data.username,
      username: data.username,
    });
    if (err) {
      setError(err.message ?? "Registration failed");
      return;
    }
    setRegistered(true);
  };

  if (registered) {
    return (
      <motion.div
        className="max-w-sm mx-auto py-12"
        initial={{ opacity: 0, scale: 0.98 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.3 }}
      >
        <p className="text-hm-text mb-4">
          Check your email to verify your account before logging in.
        </p>
        <Link to="/login" className="text-hm-muted hover:underline text-sm">Back to login</Link>
      </motion.div>
    );
  }

  return (
    <motion.div
      className="max-w-sm mx-auto py-12"
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, ease: [0.25, 0.46, 0.45, 0.94] }}
    >
      <motion.div variants={container} initial="hidden" animate="show">
      <motion.h1 variants={item} className="font-sans text-2xl font-semibold text-hm-text mb-6">Register</motion.h1>
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
          <label className="block font-mono text-xs text-hm-muted mb-1">Username</label>
          <input
            {...register("username")}
            autoComplete="username"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.username && <p className="mt-1 text-sm text-red-400">{errors.username.message}</p>}
        </motion.div>
        <motion.div variants={item}>
          <label className="block font-mono text-xs text-hm-muted mb-1">Password</label>
          <input
            {...register("password")}
            type="password"
            autoComplete="new-password"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          <p className="mt-1 text-xs text-hm-muted">12+ characters, uppercase, number, special character</p>
          {errors.password && <p className="mt-1 text-sm text-red-400">{errors.password.message}</p>}
        </motion.div>
        <motion.div variants={item}>
          <label className="block font-mono text-xs text-hm-muted mb-1">Confirm password</label>
          <input
            {...register("confirm")}
            type="password"
            autoComplete="new-password"
            className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
          />
          {errors.confirm && <p className="mt-1 text-sm text-red-400">{errors.confirm.message}</p>}
        </motion.div>
        <motion.div variants={item}>
        <Button type="submit" disabled={isSubmitting} className="w-full transition-transform hover:scale-[1.01] active:scale-[0.99]">
          {isSubmitting ? "Creating account…" : "Register"}
        </Button>
        </motion.div>
      </form>
      <motion.p variants={item} className="mt-6 text-sm text-hm-muted">
        Already have an account? <Link to="/login" className="text-hm-text hover:underline">Log in</Link>
      </motion.p>
      </motion.div>
    </motion.div>
  );
}
