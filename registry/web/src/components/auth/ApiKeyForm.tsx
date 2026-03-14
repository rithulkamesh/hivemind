import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Button } from "@/components/ui/button";

const schema = z.object({
  name: z.string().min(1, "Name required"),
  scopes: z.array(z.enum(["read", "publish", "delete", "admin"])).min(1, "Select at least one scope"),
  expires_days: z
    .union([z.string(), z.number()])
    .optional()
    .transform((v) => {
      if (v === "" || v === undefined) return undefined;
      const n = Number(v);
      if (Number.isNaN(n) || n < 0) return undefined;
      return Math.floor(n);
    }),
});

type FormData = z.output<typeof schema>;
type FormInput = z.input<typeof schema>;

const SCOPE_OPTIONS = [
  { value: "read", label: "Read" },
  { value: "publish", label: "Publish" },
  { value: "delete", label: "Delete" },
  { value: "admin", label: "Admin" },
] as const;

interface ApiKeyFormProps {
  onSubmit: (data: FormData) => Promise<{ key: string; prefix: string } | void>;
  onCancel?: () => void;
}

export function ApiKeyForm({ onSubmit, onCancel }: ApiKeyFormProps) {
  const {
    register,
    control,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormInput>({
    defaultValues: { name: "", scopes: [], expires_days: undefined },
    resolver: zodResolver(schema),
  });

  return (
    <form onSubmit={handleSubmit((data) => onSubmit(data as unknown as FormData))} className="p-4 space-y-4">
      <div>
        <label className="block font-mono text-xs text-hm-muted mb-1">Name</label>
        <input
          {...register("name")}
          className="w-full bg-hm-surface border border-hm-border px-3 py-2 font-sans text-sm text-hm-text focus:outline-none focus:border-hm-muted"
          placeholder="e.g. CI deploy"
        />
        {errors.name && <p className="mt-1 text-sm text-red-400">{errors.name.message}</p>}
      </div>
      <div>
        <label className="block font-mono text-xs text-hm-muted mb-2">Scopes</label>
        <Controller
          name="scopes"
          control={control}
          defaultValue={[]}
          render={({ field }) => (
            <div className="flex flex-wrap gap-3">
              {SCOPE_OPTIONS.map(({ value, label }) => (
                <label key={value} className="flex items-center gap-2 text-sm text-hm-text cursor-pointer">
                  <input
                    type="checkbox"
                    checked={field.value.includes(value)}
                    className="rounded border-hm-border"
                    onChange={(e) => {
                      const next = e.target.checked
                        ? [...field.value, value]
                        : field.value.filter((s) => s !== value);
                      field.onChange(next);
                    }}
                  />
                  {label}
                </label>
              ))}
            </div>
          )}
        />
        {errors.scopes && <p className="mt-1 text-sm text-red-400">{errors.scopes.message}</p>}
      </div>
      <div>
        <label className="block font-mono text-xs text-hm-muted mb-1">Expires (days, optional)</label>
        <input
          type="number"
          min={0}
          placeholder="Leave empty for no expiry"
          {...register("expires_days")}
          className="w-full bg-hm-surface border border-hm-border px-3 py-2 font-sans text-sm text-hm-text focus:outline-none focus:border-hm-muted"
        />
        {errors.expires_days && <p className="mt-1 text-sm text-red-400">{errors.expires_days.message}</p>}
      </div>
      <div className="flex gap-2">
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Creating…" : "Create key"}
        </Button>
        {onCancel && (
          <Button type="button" variant="outline" onClick={onCancel}>
            Cancel
          </Button>
        )}
      </div>
    </form>
  );
}
