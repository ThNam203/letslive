"use client";

import { useEffect, useRef } from "react";
import IconLoader from "@/components/icons/loader";
import { SEARCH_ROWS, SearchRow } from "@/components/search/search-result-rows";
import useT from "@/hooks/use-translation";
import { useSearchInfinite } from "@/hooks/queries/use-search";
import { SearchType } from "@/lib/search";
import { flattenPages } from "@/lib/query/paginated";
import { SearchEmpty, SearchError, SearchRowsSkeleton } from "./search-states";

export default function SearchTypeTab<T extends SearchType>({
    type,
    query,
}: {
    type: T;
    query: string;
}) {
    const { t } = useT("common");
    const {
        data,
        isLoading,
        isError,
        refetch,
        hasNextPage,
        isFetchingNextPage,
        isFetchNextPageError,
        fetchNextPage,
    } = useSearchInfinite(type, query);
    const items = flattenPages(data);
    const Row: SearchRow<T> = SEARCH_ROWS[type];
    const sentinelRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        const el = sentinelRef.current;
        if (!el || typeof IntersectionObserver === "undefined") return;
        if (!hasNextPage || isFetchingNextPage || isFetchNextPageError) return;

        const observer = new IntersectionObserver(
            (entries) => {
                if (entries.some((e) => e.isIntersecting)) fetchNextPage();
            },
            { rootMargin: "300px" },
        );
        observer.observe(el);
        return () => observer.disconnect();
    }, [hasNextPage, isFetchingNextPage, isFetchNextPageError, fetchNextPage]);

    if (isLoading) return <SearchRowsSkeleton count={4} />;
    if (isError && items.length === 0) {
        return <SearchError onRetry={() => refetch()} />;
    }
    if (items.length === 0) {
        return (
            <SearchEmpty
                title={t(`common:search.no_results_type.${type}`)}
                hint={t("common:search.no_results_hint")}
            />
        );
    }

    return (
        <div>
            <div className="-mx-3">
                {items.map((item) => (
                    <Row key={item.id} item={item} query={query} />
                ))}
            </div>
            {isFetchNextPageError && (
                <SearchError onRetry={() => fetchNextPage()} />
            )}
            {hasNextPage && !isFetchNextPageError && (
                <div
                    ref={sentinelRef}
                    className="text-muted-foreground flex items-center justify-center gap-2 py-6 text-sm"
                >
                    {isFetchingNextPage && (
                        <>
                            <IconLoader className="h-4 w-4 animate-spin" />
                            {t("common:loading")}
                        </>
                    )}
                </div>
            )}
        </div>
    );
}
