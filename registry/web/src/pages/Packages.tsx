import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Link, useSearchParams } from "react-router-dom";
import { api, apiRoutes } from "@/lib/api";
import type { Package } from "@/types";
import { PackageCard } from "@/components/packages/PackageCard";
import { LoadingSkeleton } from "@/components/ui/LoadingSkeleton";
import { Button } from "@/components/ui/button";

export function Packages() {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = parseInt(searchParams.get("page") ?? "1", 10);
  const [totalPages, setTotalPages] = useState(1);

  const { data, isLoading, error } = useQuery({
    queryKey: ["packages", page],
    queryFn: async () => {
      const res = await api<{ packages: Package[]; page: number }>(
        apiRoutes.packages({ page })
      );
      // Assuming API returns 20 items per page, we can infer if there's a next page
      // But ideally API should return total count.
      // For now, if we get < 20 items, we are at the end.
      return res;
    },
  });

  const handleNext = () => {
    setSearchParams({ page: String(page + 1) });
  };

  const handlePrev = () => {
    if (page > 1) setSearchParams({ page: String(page - 1) });
  };

  if (error) {
    return <div className="text-red-500">Error loading packages</div>;
  }

  return (
    <div className="max-w-4xl mx-auto py-8">
      <h1 className="font-sans text-3xl font-bold text-hm-text mb-6">All Packages</h1>
      
      {isLoading ? (
        <div className="space-y-4">
          <LoadingSkeleton className="h-24 w-full" />
          <LoadingSkeleton className="h-24 w-full" />
          <LoadingSkeleton className="h-24 w-full" />
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {(data?.packages ?? []).map((pkg) => (
            <PackageCard key={pkg.id} pkg={pkg} />
          ))}
          {data?.packages.length === 0 && (
            <p className="text-hm-muted col-span-2 text-center py-10">
              No packages found.
            </p>
          )}
        </div>
      )}

      <div className="flex justify-center mt-8 gap-4">
        <Button
          variant="outline"
          disabled={page === 1 || isLoading}
          onClick={handlePrev}
        >
          Previous
        </Button>
        <span className="flex items-center text-hm-muted">Page {page}</span>
        <Button
          variant="outline"
          disabled={!data || data.packages.length < 20 || isLoading}
          onClick={handleNext}
        >
          Next
        </Button>
      </div>
    </div>
  );
}
