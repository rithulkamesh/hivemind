import { useState, useRef, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { clsx } from "clsx";

interface SearchBarProps {
  inline?: boolean;
  onSelect?: () => void;
}

export function SearchBar({ inline = false, onSelect }: SearchBarProps) {
  const [q, setQ] = useState("");
  const [debounced, setDebounced] = useState("");
  const navigate = useNavigate();
  const wrapperRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const t = setTimeout(() => setDebounced(q), 300);
    return () => clearTimeout(t);
  }, [q]);

  const { data, isFetching } = useQuery({
    queryKey: ["search", debounced],
    queryFn: () => api<{ results: Package[]; page: number }>(apiRoutes.search(debounced, 1)),
    enabled: debounced.length >= 2,
  });

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (q.trim()) {
      navigate(`/search?q=${encodeURIComponent(q.trim())}`);
      onSelect?.();
    }
  };

  const results = data?.results ?? [];
  const showDropdown = inline && debounced.length >= 2;

  return (
    <div ref={wrapperRef} className="relative w-full">
      <form onSubmit={handleSubmit} role="search">
        <input
          type="search"
          value={q}
          onChange={(e) => setQ(e.target.value)}
          placeholder="Search packages…"
          className={clsx(
            "w-full bg-hm-surface border border-hm-border px-3 py-2 font-sans text-sm text-hm-text placeholder:text-hm-muted focus:outline-none focus:border-hm-muted",
            inline && "pr-8"
          )}
          aria-label="Search packages"
          aria-autocomplete="list"
          aria-expanded={showDropdown}
        />
      </form>
      {showDropdown && (
        <div
          className="absolute top-full left-0 right-0 mt-1 bg-hm-surface border border-hm-border shadow-lg max-h-72 overflow-y-auto z-50"
          role="listbox"
        >
          {isFetching ? (
            <div className="px-4 py-3 text-sm text-hm-muted">Searching…</div>
          ) : results.length === 0 ? (
            <div className="px-4 py-3 text-sm text-hm-muted">No packages found.</div>
          ) : (
            results.slice(0, 8).map((pkg) => (
              <a
                key={pkg.id}
                href={`/packages/${pkg.name}`}
                onClick={(e) => {
                  e.preventDefault();
                  navigate(`/packages/${pkg.name}`);
                  onSelect?.();
                }}
                className="block px-4 py-2 hover:bg-hm-border/50 text-hm-text font-sans text-sm border-b border-hm-border last:border-0"
                role="option"
              >
                {pkg.name}
              </a>
            ))
          )}
        </div>
      )}
    </div>
  );
}
