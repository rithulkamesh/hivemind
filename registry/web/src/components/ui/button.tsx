import { forwardRef, type ButtonHTMLAttributes } from "react";
import { clsx } from "clsx";

type Variant = "default" | "outline" | "ghost" | "destructive";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

const variants: Record<Variant, string> = {
  default:
    "bg-hm-text text-hm-bg border border-hm-border hover:bg-hm-accent hover:text-hm-bg transition-opacity",
  outline:
    "bg-transparent border border-hm-border text-hm-text hover:bg-hm-surface transition-opacity",
  ghost: "bg-transparent text-hm-text hover:bg-hm-surface transition-opacity",
  destructive:
    "bg-red-600/20 border border-red-500/50 text-red-400 hover:bg-red-600/30 transition-opacity",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant = "default", disabled, ...props }, ref) => (
    <button
      ref={ref}
      className={clsx(
        "inline-flex items-center justify-center font-mono text-xs tracking-wider uppercase px-4 py-2 focus:outline-none focus:ring-1 focus:ring-hm-muted disabled:opacity-50 disabled:pointer-events-none",
        variants[variant],
        className
      )}
      disabled={disabled}
      {...props}
    />
  )
);
Button.displayName = "Button";
