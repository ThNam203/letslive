"use client";

import { useState, useEffect, useRef } from "react";

import Link from "next/link";
import useT from "@/hooks/use-translation";
import { Input } from "@/components/ui/input";
import IconClose from "@/components/icons/close";
import IconSearch from "@/components/icons/search";
import UserAvatar from "@/components/ui/user-avatar";
import { cn } from "@/utils/cn";
import { MQ_MAX_MD } from "@/constant/breakpoints";
import { SEARCH_QUERY_MAX_LENGTH } from "@/constant/field-limits";
import useMediaQuery from "@/hooks/use-media-query";
import {
    Dialog,
    DialogContent,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog";
import { useUserSearch } from "@/hooks/queries/use-user-search";

export default function SearchBar({
    onSearch,
    className,
}: {
    onSearch?: (query: string) => void;
    className?: string;
}) {
    const [query, setQuery] = useState("");
    const [debouncedQuery, setDebouncedQuery] = useState("");
    const [showResults, setShowResults] = useState(false);
    const [mobileOpen, setMobileOpen] = useState(false);
    const isSmallScreen = useMediaQuery(MQ_MAX_MD);
    const containerRef = useRef<HTMLDivElement>(null);
    const { t } = useT([
        "common",
        "api-response",
        "fetch-error",
        "accessibility",
    ]);

    const {
        data: results = [],
        isFetching: isLoading,
        isError,
    } = useUserSearch(debouncedQuery);

    // The input stays responsive while the query the server sees only moves
    // once typing pauses.
    useEffect(() => {
        const timer = setTimeout(() => {
            const trimmed = query.trim();
            setDebouncedQuery(trimmed);
            setShowResults(Boolean(trimmed));
            if (trimmed && onSearch) onSearch(query);
        }, 1000);

        return () => clearTimeout(timer);
    }, [query, onSearch]);

    // Close the results when the user clicks outside the search or hits Escape;
    // focusing the input again reopens them.
    useEffect(() => {
        if (!showResults) return;

        const handlePointerDown = (e: PointerEvent) => {
            if (!containerRef.current?.contains(e.target as Node)) {
                setShowResults(false);
            }
        };
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.key === "Escape") setShowResults(false);
        };

        document.addEventListener("pointerdown", handlePointerDown);
        document.addEventListener("keydown", handleKeyDown);
        return () => {
            document.removeEventListener("pointerdown", handlePointerDown);
            document.removeEventListener("keydown", handleKeyDown);
        };
    }, [showResults]);

    const handleClear = () => {
        setQuery("");
        setDebouncedQuery("");
        setShowResults(false);
    };

    const handleResultClick = () => {
        setShowResults(false);
        setMobileOpen(false);
    };

    const searchInput = (
        <div ref={containerRef} className="relative w-[300px] lg:w-[400px]">
            <div className="relative">
                <Input
                    type="text"
                    placeholder={t("common:search_users")}
                    maxLength={SEARCH_QUERY_MAX_LENGTH}
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className="border-border pr-8"
                    onFocus={() => query.trim() && setShowResults(true)}
                    autoFocus={isSmallScreen}
                />
                {query && (
                    <button
                        onClick={handleClear}
                        className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 -translate-y-1/2"
                        aria-label={t("accessibility:clear_search")}
                    >
                        <IconClose className="h-4 w-4" />
                    </button>
                )}
            </div>

            {showResults && isLoading && query && (
                <div className="bg-background absolute mt-1 w-full rounded-sm border p-4 shadow-md">
                    <div className="flex items-center justify-center">
                        <p className="text-muted-foreground text-sm">
                            {t("common:searching")}
                        </p>
                    </div>
                </div>
            )}

            {showResults && results.length > 0 && !isLoading && (
                <div className="bg-background absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-sm border shadow-md">
                    {results.map((user) => (
                        <Link
                            key={user.id}
                            href={`/users/${user.id}`}
                            onClick={handleResultClick}
                            className="flex w-full cursor-pointer flex-row items-center gap-3 p-2 hover:bg-gray-400"
                        >
                            <UserAvatar
                                size="sm"
                                src={user.profilePicture}
                                name={user.username}
                                alt={t("accessibility:user_avatar")}
                            />
                            <p className="text-sm font-medium">
                                {user.username}
                            </p>
                        </Link>
                    ))}
                </div>
            )}

            {/* a failed search is not an empty one: saying "no users found"
                when the request never completed is a lie about the data */}
            {showResults && isError && !isLoading && query && (
                <div className="bg-background absolute mt-1 w-full rounded-sm border p-4 shadow-md">
                    <p className="text-muted-foreground text-sm">
                        {t("fetch-error:client_fetch_error")}
                    </p>
                </div>
            )}

            {showResults &&
                !isError &&
                results.length === 0 &&
                !isLoading &&
                query && (
                    <div className="bg-background absolute mt-1 w-full rounded-sm border p-4 shadow-md">
                        <p className="text-muted-foreground text-sm">
                            {t("common:no_users_found")}
                        </p>
                    </div>
                )}
        </div>
    );

    if (isSmallScreen) {
        return (
            <Dialog open={mobileOpen} onOpenChange={setMobileOpen}>
                <DialogTrigger asChild>
                    <button
                        type="button"
                        aria-label={t("common:search_users")}
                        className="hover:bg-background-hover mr-2 flex flex-1 justify-end rounded-full"
                    >
                        <IconSearch />
                    </button>
                </DialogTrigger>
                <DialogContent
                    showCloseButton={false}
                    className="top-20 left-1/2 w-[92vw] max-w-md translate-y-0 p-4"
                >
                    <DialogTitle className="sr-only">
                        {t("common:search_users")}
                    </DialogTitle>
                    {searchInput}
                </DialogContent>
            </Dialog>
        );
    }

    return (
        <div
            className={cn(
                "relative flex w-full flex-row justify-center",
                className,
            )}
        >
            {searchInput}
        </div>
    );
}
