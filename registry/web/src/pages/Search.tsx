import { useSearchParams } from "react-router-dom";
import { SearchResults } from "@/search/SearchResults";

export function Search() {
  const [params] = useSearchParams();
  const q = params.get("q") ?? "";

  return (
    <div className="max-w-4xl">
      <h1 className="font-sans text-2xl font-semibold text-hm-text mb-2">
        Search packages
      </h1>
      {q && (
        <p className="text-hm-muted text-sm mb-6">
          Results for &quot;{q}&quot;
        </p>
      )}
      <SearchResults />
    </div>
  );
}
