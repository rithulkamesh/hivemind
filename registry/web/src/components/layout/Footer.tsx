import { Link } from "react-router-dom";

export function Footer() {
  return (
    <footer className="border-t border-hm-border py-hm-lg px-hm-lg bg-hm-bg">
      <div className="max-w-6xl mx-auto flex flex-wrap items-center justify-between gap-4">
        <div className="flex flex-wrap gap-6">
          <Link to="/" className="font-mono text-[9px] tracking-widest uppercase text-hm-muted hover:text-hm-text transition-opacity">
            Registry
          </Link>
          <Link to="/search" className="font-mono text-[9px] tracking-widest uppercase text-hm-muted hover:text-hm-text transition-opacity">
            Search
          </Link>
          <Link to="/publish" className="font-mono text-[9px] tracking-widest uppercase text-hm-muted hover:text-hm-text transition-opacity">
            Publish
          </Link>
          <a
            href="https://hivemind.rithul.dev/docs"
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono text-[9px] tracking-widest uppercase text-hm-muted hover:text-hm-text transition-opacity"
          >
            Docs
          </a>
          <a
            href="https://github.com/rithulkamesh/hivemind"
            target="_blank"
            rel="noopener noreferrer"
            className="font-mono text-[9px] tracking-widest uppercase text-hm-muted hover:text-hm-text transition-opacity"
          >
            GitHub
          </a>
        </div>
        <p className="font-mono text-[9px] tracking-widest uppercase text-hm-muted">
          Hivemind Registry · registry.hivemind.rithul.dev
        </p>
      </div>
    </footer>
  );
}
