import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { KeyRound } from "lucide-react";

const container = { hidden: { opacity: 0 }, show: { opacity: 1, transition: { staggerChildren: 0.06 } } };
const item = { hidden: { opacity: 0, y: 12 }, show: { opacity: 1, y: 0 } };

const schema = z.object({
  user_code: z.string().min(1, "Code required").regex(/^[A-Z]{4}-\d{4}$/, "Must be format XXXX-0000"),
});

type FormData = z.infer<typeof schema>;

export function Activate() {
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: FormData) => {
    setError("");
    setSuccess(false);

    try {
      await api("/api/v1/auth/device/approve", {
        method: "POST",
        body: JSON.stringify({ user_code: data.user_code.toUpperCase() }),
      });
      setSuccess(true);
    } catch (err: any) {
      setError(err.message || "Failed to approve device code");
    }
  };

  if (success) {
    return (
      <motion.div
        className="max-w-md mx-auto py-12 text-center"
        initial={{ opacity: 0, y: 16 }}
        animate={{ opacity: 1, y: 0 }}
      >
        <div className="w-16 h-16 bg-green-500/10 rounded-full flex items-center justify-center mx-auto mb-6">
          <KeyRound className="w-8 h-8 text-green-500" />
        </div>
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-2">Device Approved</h1>
        <p className="text-hm-muted mb-8 text-sm">You can now close this window and return to your terminal.</p>
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
        <motion.h1 variants={item} className="font-sans text-2xl font-semibold text-hm-text mb-2">Activate Device</motion.h1>
        <motion.p variants={item} className="text-hm-muted text-sm mb-6">Enter the code displayed in your terminal to authorize the CLI.</motion.p>
        
        {error && <motion.p variants={item} className="mb-4 text-sm text-red-400">{error}</motion.p>}
        
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" autoComplete="off">
          <motion.div variants={item}>
            <label className="block font-mono text-xs text-hm-muted mb-1">Device Code</label>
            <input
              {...register("user_code")}
              type="text"
              placeholder="XXXX-0000"
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 font-mono text-center text-lg uppercase tracking-widest text-hm-text focus:outline-none focus:border-hm-muted transition-colors"
            />
            {errors.user_code && <p className="mt-1 text-sm text-red-400">{errors.user_code.message}</p>}
          </motion.div>
          
          <motion.div variants={item}>
            <Button type="submit" disabled={isSubmitting} className="w-full transition-transform hover:scale-[1.01] active:scale-[0.99] mt-2">
              {isSubmitting ? "Approving…" : "Approve Device"}
            </Button>
          </motion.div>
        </form>
      </motion.div>
    </motion.div>
  );
}
