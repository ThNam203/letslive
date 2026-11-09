"use client";

import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import useT from "@/hooks/use-translation";
import IconSearch from "@/components/icons/search";
import SearchInput from "@/components/search/search-input";
import { SEARCH_ROWS, SearchRow } from "@/components/search/search-result-rows";
import { cn } from "@/utils/cn";
import useDebouncedValue from "@/hooks/use-debounced-value";
import useSearchUrl from "@/hooks/use-search-url";
import { useSearchPreview } from "@/hooks/queries/use-search";
import {
    RAW_INTERPOLATION,
    SEARCH_PATH,
    SEARCH_TYPES,
    SEARCH_TYPING_DEBOUNCE_MS,
    SearchType,
    searchHref,
} from "@/lib/search";

const PREVIEW_PANEL_ID = "header-search-results";

export default function SearchBar({ className }: { className?: string }) {
    const isOnSearchPage = usePathname() === SEARCH_PATH;
    const { t } = useT("common");

    return (
        <>
            {/* Phones search on the full /search page instead of a cramped popup. */}
            <div className="flex flex-1 justify-end md:hidden">
                {!isOnSearchPage && (
                    <Link
                        href={SEARCH_PATH}
                        aria-label={t("common:search.open")}
                        className="hover:bg-background-hover mr-2 rounded-full"
                    >
                        <IconSearch />
                    </Link>
                )}
            </div>
            <div
                className={cn(
                    className,
                    "hidden w-full flex-row justify-center md:flex",
                )}
            >
                {isOnSearchPage ? <SearchPageInput /> : <SearchWithPreview />}
            </div>
        </>
    );
}

function SearchPageInput() {
    const { draft, changeDraft, submit } = useSearchUrl();
    return (
        <SearchInput
            value={draft}
            onChange={changeDraft}
            onSubmit={submit}
            className="w-full max-w-[440px]"
        />
    );
}

function SearchWithPreview() {
    const router = useRouter();
    const [query, setQuery] = useState("");
    const [isOpen, setIsOpen] = useState(false);
    const debouncedQuery = useDebouncedValue(
        query.trim(),
        SEARCH_TYPING_DEBOUNCE_MS,
    );
    const containerRef = useRef<HTMLDivElement>(null);
    const isPanelShown = isOpen && Boolean(debouncedQuery);

    useEffect(() => {
        if (!isOpen) return;

        const handlePointerDown = (e: PointerEvent) => {
            if (!containerRef.current?.contains(e.target as Node)) {
                setIsOpen(false);
            }
        };
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") setIsOpen(false);
        };

        document.addEventListener("pointerdown", handlePointerDown);
        document.addEventListener("keydown", handleKeyDown);
        return () => {
            document.removeEventListener("pointerdown", handlePointerDown);
            document.removeEventListener("keydown", handleKeyDown);
        };
    }, [isOpen]);

    const handleChange = (value: string) => {
        setQuery(value);
        setIsOpen(Boolean(value.trim()));
    };

    const goToResults = () => {
        const trimmed = query.trim();
        if (!trimmed) return;
        setIsOpen(false);
        router.push(searchHref(trimmed));
    };

    const close = () => setIsOpen(false);

    return (
        <div
            ref={containerRef}
            className="relative w-full max-w-[440px]"
            onBlur={(e) => {
                // Only keyboard moves have a target; clicks are handled by
                // the pointerdown listener (Safari doesn't focus clicked links).
                const next = e.relatedTarget as Node | null;
                if (next && !e.currentTarget.contains(next)) close();
            }}
        >
            <SearchInput
                value={query}
                onChange={handleChange}
                onSubmit={goToResults}
                onFocus={() => query.trim() && setIsOpen(true)}
                expandedPanelId={isPanelShown ? PREVIEW_PANEL_ID : undefined}
            />
            {isPanelShown && (
                <SearchPreviewPanel query={debouncedQuery} onNavigate={close} />
            )}
        </div>
    );
}

function SearchPreviewPanel({
    query,
    onNavigate,
}: {
    query: string;
    onNavigate: () => void;
}) {
    const { t } = useT(["common", "fetch-error"]);
    const users = useSearchPreview("users", query);
    const lives = useSearchPreview("lives", query);
    const vods = useSearchPreview("vods", query);
    const results = [users, lives, vods];

    const isAllLoading = results.every((r) => r.isLoading);
    const isAllError = results.every((r) => r.isError);
    const isNothingFound = results.every(
        (r) => r.isSuccess && r.data.length === 0,
    );

    return (
        <div
            id={PREVIEW_PANEL_ID}
            className="bg-background border-border absolute top-full left-0 z-50 mt-1 flex max-h-[70vh] w-[min(32rem,calc(100vw-2rem))] flex-col overflow-hidden rounded-md border text-sm font-normal shadow-lg"
        >
            <div className="overflow-y-auto p-1">
                <div role="status" className="text-muted-foreground">
                    {isAllLoading && (
                        <p className="p-3">{t("common:searching")}</p>
                    )}
                    {isAllError && (
                        <p className="p-3">
                            {t("fetch-error:client_fetch_error")}
                        </p>
                    )}
                    {isNothingFound && (
                        <p className="p-3">
                            {t("common:search.no_results", {
                                query,
                                ...RAW_INTERPOLATION,
                            })}
                        </p>
                    )}
                </div>
                {SEARCH_TYPES.map((type) => (
                    <PreviewSection
                        key={type}
                        type={type}
                        query={query}
                        onNavigate={onNavigate}
                    />
                ))}
            </div>
            <Link
                href={searchHref(query)}
                onClick={onNavigate}
                className="border-border hover:bg-muted focus-visible:ring-ring text-primary flex items-center gap-2 border-t px-3 py-2.5 font-medium focus-visible:ring-2 focus-visible:outline-none focus-visible:ring-inset"
            >
                <IconSearch aria-hidden="true" className="h-4 w-4" />
                <span className="truncate">
                    {t("common:search.see_all_results", {
                        query,
                        ...RAW_INTERPOLATION,
                    })}
                </span>
            </Link>
        </div>
    );
}

function PreviewSection<T extends SearchType>({
    type,
    query,
    onNavigate,
}: {
    type: T;
    query: string;
    onNavigate: () => void;
}) {
    const { t } = useT("common");
    const { data } = useSearchPreview(type, query);
    const Row: SearchRow<T> = SEARCH_ROWS[type];
    if (!data?.length) return null;

    return (
        <div className="py-1">
            <p className="text-muted-foreground px-2 pt-1 pb-0.5 text-xs font-semibold tracking-wide uppercase">
                {t(`common:search.sections.${type}`)}
            </p>
            {data.map((item) => (
                <Row
                    key={item.id}
                    item={item}
                    query={query}
                    size="compact"
                    onNavigate={onNavigate}
                />
            ))}
        </div>
    );
}
