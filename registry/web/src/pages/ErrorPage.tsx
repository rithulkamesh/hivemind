import { useSearchParams, Link } from "react-router-dom";

export function ErrorPage() {
  const [params] = useSearchParams();
  const code = params.get("code") ?? "unknown_error";

  return (
    <div className="min-h-screen flex items-center justify-center p-4 bg-hm-bg bg-[linear-gradient(rgba(255,255,255,.02)_1px,transparent_1px),linear-gradient(90deg,rgba(255,255,255,.02)_1px,transparent_1px)] bg-[size:24px_24px]">
      <div className="relative w-full max-w-md border border-hm-border rounded-lg bg-hm-surface/80 p-8 shadow-xl">
        {/* Corner brackets */}
        <div className="absolute top-4 left-4 w-6 h-6 border-l-2 border-t-2 border-hm-muted rounded-tl" />
        <div className="absolute bottom-4 right-4 w-6 h-6 border-r-2 border-b-2 border-hm-muted rounded-br" />

        <div className="flex flex-col items-center text-center space-y-4">
          <span className="inline-block px-3 py-1 text-sm font-bold tracking-wider text-white border-2 border-amber-500/80 rounded">
            ERROR
          </span>
          <h1 className="font-sans text-xl font-semibold text-hm-text">
            Something went wrong
          </h1>
          <p className="text-hm-muted text-sm">
            CODE:{" "}
            <code className="px-2 py-0.5 bg-hm-bg border border-hm-border rounded font-mono text-hm-text text-xs">
              {code}
            </code>
          </p>
          <p className="text-hm-text-passive text-sm max-w-sm">
            We encountered an unexpected error. Please try again or return to the
            home page. If you&apos;re a developer, you can find more information
            about the error{" "}
            <a
              href="https://www.better-auth.com/docs"
              target="_blank"
              rel="noopener noreferrer"
              className="text-amber-400 hover:text-amber-300 underline underline-offset-2"
            >
              here
            </a>
            .
          </p>
          <div className="flex gap-3 pt-2">
            <Link
              to="/"
              className="px-4 py-2 rounded border border-hm-border text-hm-text hover:bg-hm-border/50 transition-colors text-sm font-medium"
            >
              Go Home
            </Link>
            <a
              href="https://github.com/rithulkamesh/hivemind/issues"
              target="_blank"
              rel="noopener noreferrer"
              className="px-4 py-2 rounded bg-hm-border text-hm-text hover:bg-hm-muted transition-colors text-sm font-medium"
            >
              Ask AI / Report
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}
