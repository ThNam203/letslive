"use client";

import { Button } from "@/components/ui/button";
import IconChevronRight from "@/components/icons/chevron-right";
import { SEARCH_ROWS, SearchRow } from "@/components/search/search-result-rows";
import useT from "@/hooks/use-translation";
import {
    SEARCH_PREVIEW_LIMIT,
    useSearchPreview,
} from "@/hooks/queries/use-search";
import { RAW_INTERPOLATION, SEARCH_TYPES, SearchType } from "@/lib/search";
import { SearchEmpty, SearchError, SearchRowsSkeleton } from "./search-states";

function PreviewSection<T extends SearchType>({
    type,
    query,
    onSeeAll,
}: {
    type: T;
    query: string;
    onSeeAll: () => void;
}) {
    const { t } = useT("common");
    const { data, isLoading, isError, refetch } = useSearchPreview(type, query);
    const Row: SearchRow<T> = SEARCH_ROWS[type];

    if (!isLoading && !isError && (data?.length ?? 0) === 0) return null;

    const headingId = `search-section-${type}`;
    return (
        <section aria-labelledby={headingId}>
            <div className="mb-1 flex items-center justify-between">
                <h2 id={headingId} className="text-lg font-semibold">
                    {t(`common:search.sections.${type}`)}
                </h2>
                {data?.length === SEARCH_PREVIEW_LIMIT && (
                    <Button variant="ghost" size="sm" onClick={onSeeAll}>
                        {t("common:search.see_all")}
                        <IconChevronRight />
                    </Button>
                )}
            </div>
            {isLoading ? (
                <SearchRowsSkeleton count={2} />
            ) : isError ? (
                <SearchError onRetry={() => refetch()} />
            ) : (
                <div className="-mx-3">
                    {data?.map((item) => (
                        <Row key={item.id} item={item} query={query} />
                    ))}
                </div>
            )}
        </section>
    );
}

export default function SearchAllTab({
    query,
    onSeeAll,
}: {
    query: string;
    onSeeAll: (type: SearchType) => void;
}) {
    const { t } = useT("common");
    const users = useSearchPreview("users", query);
    const lives = useSearchPreview("lives", query);
    const vods = useSearchPreview("vods", query);
    const results = [users, lives, vods];
    const nothingFound = results.every(
        (r) => r.isSuccess && r.data.length === 0,
    );

    if (nothingFound) {
        return (
            <SearchEmpty
                title={t("common:search.no_results", {
                    query,
                    ...RAW_INTERPOLATION,
                })}
                hint={t("common:search.no_results_hint")}
            />
        );
    }

    return (
        <div className="space-y-8">
            {SEARCH_TYPES.map((type) => (
                <PreviewSection
                    key={type}
                    type={type}
                    query={query}
                    onSeeAll={() => onSeeAll(type)}
                />
            ))}
        </div>
    );
}
