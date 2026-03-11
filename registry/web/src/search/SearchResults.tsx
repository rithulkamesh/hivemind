import { useQuery } from "@tanstack/react-query";
import { useSearchParams } from "react-router-dom";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { PackageCard } from "@/components/packages/PackageCard";
import { SearchFilters } from "./SearchFilters";
import { Pagination } from "@/components/ui/Pagination";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";
import { EmptyState } from "@/components/ui/EmptyState";

export function SearchResults() {
  const [params] = useSearchParams();
  const q = params.get("q") ?? "";
  const page = Math.max(1, parseInt(params.get("page") ?? "1", 10));

  const { data, isLoading } = useQuery({
    queryKey: ["search", q, page],
    queryFn: () => api<{ results: Package[]; page: number }>(apiRoutes.search(q, page)),
    enabled: true,
  });

  if (isLoading) {
    return (
      <div className="space-y-4">
        <SearchFilters />
        <div className="space-y-2">
          {[1, 2, 3, 4].map((i) => (
            <LoadingSkeleton key={i} className="h-20 w-full" />
          ))}
        </div>
      </div>
    );
  }

  const results = data?.results ?? [];
  const totalPages = Math.max(1, Math.ceil((data?.results?.length ?? 0) / 20) || 1);

  return (
    <div className="space-y-4">
      <SearchFilters />
      {results.length === 0 ? (
        <EmptyState
          title={q ? "No packages found" : "Enter a search query"}
          description={q ? `No results for "${q}"` : "Use the search bar above."}
        />
      ) : (
        <>
          <div className="space-y-2">
            {results.map((pkg) => (
              <PackageCard key={pkg.id} pkg={pkg} />
            ))}
          </div>
          <Pagination
            basePath="/search"
            currentPage={page}
            totalPages={totalPages}
            searchParams={q ? { q } : undefined}
          />
        </>
      )}
    </div>
  );
}
