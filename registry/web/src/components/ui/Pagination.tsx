import { Link } from "react-router-dom";

interface PaginationProps {
  basePath: string;
  currentPage: number;
  totalPages: number;
  searchParams?: Record<string, string>;
}

export function Pagination({ basePath, currentPage, totalPages, searchParams }: PaginationProps) {
  const qs = searchParams
    ? "?" + new URLSearchParams(searchParams).toString()
    : "";
  const prev = currentPage > 1 ? currentPage - 1 : null;
  const next = currentPage < totalPages ? currentPage + 1 : null;

  return (
    <nav className="flex items-center justify-center gap-2 font-mono text-xs text-hm-muted mt-hm-xl">
      {prev !== null ? (
        <Link
          to={`${basePath}?page=${prev}${qs ? "&" + qs.replace(/^\?/, "") : ""}`}
          className="px-3 py-1.5 border border-hm-border hover:border-hm-muted hover:text-hm-text transition-opacity"
        >
          ← Previous
        </Link>
      ) : (
        <span className="px-3 py-1.5 border border-hm-border opacity-50 cursor-not-allowed">
          ← Previous
        </span>
      )}
      <span className="px-3 py-1.5">
        Page {currentPage} of {totalPages || 1}
      </span>
      {next !== null ? (
        <Link
          to={`${basePath}?page=${next}${qs ? "&" + qs.replace(/^\?/, "") : ""}`}
          className="px-3 py-1.5 border border-hm-border hover:border-hm-muted hover:text-hm-text transition-opacity"
        >
          Next →
        </Link>
      ) : (
        <span className="px-3 py-1.5 border border-hm-border opacity-50 cursor-not-allowed">
          Next →
        </span>
      )}
    </nav>
  );
}
