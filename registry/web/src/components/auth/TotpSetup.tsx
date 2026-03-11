import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { QRCodeSVG } from "qrcode.react";
import { Button } from "@/components/ui/button";

const schema = z.object({
  code: z.string().length(6, "Enter 6-digit code"),
});

type FormData = z.infer<typeof schema>;

interface TotpSetupProps {
  /** TOTP URI (otpauth://...) to show as QR and manual entry. */
  totpURI?: string;
  qrCodeUrl?: string;
  secret?: string;
  onVerify: (code: string) => Promise<void>;
}

export function TotpSetup({ totpURI, qrCodeUrl, secret, onVerify }: TotpSetupProps) {
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormData>({ resolver: zodResolver(schema), defaultValues: { code: "" } });

  const showQR = totpURI ?? qrCodeUrl;

  return (
    <div className="space-y-4 max-w-sm">
      {showQR && (
        <div className="flex flex-col items-center gap-2 p-4 bg-hm-surface border border-hm-border rounded">
          <p className="text-xs text-hm-muted">Scan with your authenticator app</p>
          {totpURI ? (
            <QRCodeSVG value={totpURI} size={160} level="M" className="rounded" />
          ) : (
            <img src={qrCodeUrl} alt="TOTP QR code" className="w-40 h-40" />
          )}
        </div>
      )}
      {secret && (
        <p className="font-mono text-xs text-hm-muted break-all">
          Secret: {secret}
        </p>
      )}
      <form onSubmit={handleSubmit((d) => onVerify(d.code))} className="space-y-2">
        <input
          {...register("code")}
          maxLength={6}
          placeholder="000000"
          className="w-full bg-hm-surface border border-hm-border px-3 py-2 font-mono text-lg text-hm-text text-center tracking-widest focus:outline-none focus:border-hm-muted"
        />
        {errors.code && <p className="text-sm text-red-400">{errors.code.message}</p>}
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Verifying…" : "Verify and enable"}
        </Button>
      </form>
    </div>
  );
}
