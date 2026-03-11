import { useSearchParams } from "react-router-dom";

export function SearchFilters() {
  const [params, setParams] = useSearchParams();
  const verifiedOnly = params.get("verified") === "1";
  const sort = params.get("sort") ?? "relevance";

  const update = (key: string, value: string | null) => {
    const next = new URLSearchParams(params);
    if (value) next.set(key, value);
    else next.delete(key);
    setParams(next);
  };

  return (
    <div className="flex flex-wrap items-center gap-4 py-2 font-sans text-sm">
      <label className="flex items-center gap-2 text-hm-muted">
        <input
          type="checkbox"
          checked={verifiedOnly}
          onChange={(e) => update("verified", e.target.checked ? "1" : null)}
          className="rounded border-hm-border"
        />
        Verified only
      </label>
      <select
        value={sort}
        onChange={(e) => update("sort", e.target.value)}
        className="bg-hm-surface border border-hm-border px-2 py-1 text-hm-text focus:outline-none focus:border-hm-muted"
      >
        <option value="relevance">Relevance</option>
        <option value="downloads">Downloads</option>
        <option value="newest">Newest</option>
      </select>
    </div>
  );
}
