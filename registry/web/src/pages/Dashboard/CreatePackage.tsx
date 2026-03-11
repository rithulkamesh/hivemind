import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { useQueryClient } from "@tanstack/react-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { motion } from "framer-motion";
import { apiBase, apiRoutes } from "@/lib/api";
import { getApiToken } from "@/lib/auth-client";
import { Sidebar } from "@/components/layout/Sidebar";
import { Button } from "@/components/ui/button";

const schema = z.object({
  name: z.string().min(1, "Name required").regex(/^[a-zA-Z0-9_-]+$/, "Only letters, numbers, hyphens, underscores"),
  display_name: z.string().min(1, "Display name required"),
  description: z.string().optional(),
});

type FormData = z.infer<typeof schema>;

export function CreatePackage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState("");
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormData>({
    resolver: zodResolver(schema),
    defaultValues: { name: "", display_name: "", description: "" },
  });

  const onSubmit = async (data: FormData) => {
    setError("");
    const token = await getApiToken();
    const res = await fetch(apiBase + apiRoutes.createPackage(), {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
      },
      body: JSON.stringify({
        Name: data.name,
        DisplayName: data.display_name,
        Description: data.description ?? "",
      }),
    });
    if (!res.ok) {
      const text = await res.text();
      setError(text || "Failed to create package");
      return;
    }
    
    queryClient.invalidateQueries({ queryKey: ["packages"] });
    navigate("/dashboard/packages", { replace: true });
  };

  return (
    <motion.div
      className="flex gap-8"
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.2 }}
    >
      <Sidebar variant="dashboard" />
      <div className="flex-1 min-w-0 max-w-md">
        <h1 className="font-sans text-2xl font-semibold text-hm-text mb-6">Create package</h1>
        <p className="text-sm text-hm-muted mb-6">Register a new package name. You can upload releases afterward.</p>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          {error && <p className="text-sm text-red-400">{error}</p>}
          <div>
            <label className="block font-mono text-xs text-hm-muted mb-1">Name</label>
            <input
              {...register("name")}
              placeholder="my-plugin"
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted"
            />
            {errors.name && <p className="mt-1 text-sm text-red-400">{errors.name.message}</p>}
          </div>
          <div>
            <label className="block font-mono text-xs text-hm-muted mb-1">Display name</label>
            <input
              {...register("display_name")}
              placeholder="My Plugin"
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted"
            />
            {errors.display_name && <p className="mt-1 text-sm text-red-400">{errors.display_name.message}</p>}
          </div>
          <div>
            <label className="block font-mono text-xs text-hm-muted mb-1">Description (optional)</label>
            <textarea
              {...register("description")}
              rows={3}
              className="w-full bg-hm-surface border border-hm-border px-3 py-2 text-hm-text focus:outline-none focus:border-hm-muted resize-y"
            />
          </div>
          <div className="flex gap-2">
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Creating…" : "Create package"}
            </Button>
            <Button type="button" variant="outline" onClick={() => navigate("/dashboard/packages")}>
              Cancel
            </Button>
          </div>
        </form>
      </div>
    </motion.div>
  );
}
