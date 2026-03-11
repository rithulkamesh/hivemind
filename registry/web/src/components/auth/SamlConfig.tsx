import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";

const schema = z.object({
  metadata_url: z.string().url("Enter a valid metadata URL"),
});

type FormData = z.infer<typeof schema>;

interface SamlConfigProps {
  onSubmit: (data: FormData) => Promise<void>;
  onTest?: () => Promise<void>;
}

export function SamlConfig({ onSubmit, onTest }: SamlConfigProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema), defaultValues: { metadata_url: "" } });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-4 max-w-md">
      <div>
        <label className="block font-mono text-xs text-hm-muted mb-1">SAML metadata URL</label>
        <input
          {...register("metadata_url")}
          className="w-full bg-hm-surface border border-hm-border px-3 py-2 font-sans text-sm text-hm-text focus:outline-none focus:border-hm-muted"
          placeholder="https://idp.example.com/metadata"
        />
        {errors.metadata_url && (
          <p className="mt-1 text-sm text-red-400">{errors.metadata_url.message}</p>
        )}
      </div>
      <div className="flex gap-2">
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Saving…" : "Save"}
        </Button>
        {onTest && (
          <Button type="button" variant="outline" onClick={onTest}>
            Test connection
          </Button>
        )}
      </div>
    </form>
  );
}
