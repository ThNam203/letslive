"use client";

import { Button } from "@/components/ui/button";
import IconRefresh from "@/components/icons/refresh";
import IconSearch from "@/components/icons/search";
import { SearchRowSkeleton } from "@/components/search/search-result-rows";
import useT from "@/hooks/use-translation";

export function SearchRowsSkeleton({ count = 3 }: { count?: number }) {
    return (
        <div aria-busy="true">
            {Array.from({ length: count }, (_, i) => (
                <SearchRowSkeleton key={i} />
            ))}
        </div>
    );
}

export function SearchError({ onRetry }: { onRetry: () => void }) {
    const { t } = useT("common");
    return (
        <div
            role="alert"
            className="flex flex-col items-center gap-3 py-8 text-center"
        >
            <p className="text-sm">{t("common:search.error")}</p>
            <Button variant="outline" size="sm" onClick={onRetry}>
                <IconRefresh />
                {t("common:search.retry")}
            </Button>
        </div>
    );
}

export function SearchEmpty({ title, hint }: { title: string; hint?: string }) {
    return (
        <div className="flex flex-col items-center px-6 py-16 text-center">
            <div className="bg-muted mb-4 rounded-full p-4">
                <IconSearch className="text-muted-foreground h-8 w-8" />
            </div>
            <p className="text-lg font-semibold">{title}</p>
            {hint && (
                <p className="text-muted-foreground mt-1 max-w-sm text-sm">
                    {hint}
                </p>
            )}
        </div>
    );
}
