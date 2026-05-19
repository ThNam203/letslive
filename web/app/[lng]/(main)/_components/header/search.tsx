"use client";

import { useState, useEffect } from "react";

import Link from "next/link";
import useT from "@/hooks/use-translation";
import { PublicUser } from "@/types/user";
import { SearchUsersByUsername } from "@/lib/api/user";
import { toast } from "@/components/utils/toast";
import { Input } from "@/components/ui/input";
import IconClose from "@/components/icons/close";
import IconSearch from "@/components/icons/search";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
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

export default function SearchBar({
    onSearch,
    className,
}: {
    onSearch?: (query: string) => void;
    className?: string;
}) {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<PublicUser[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [showResults, setShowResults] = useState(false);
    const [mobileOpen, setMobileOpen] = useState(false);
    const isSmallScreen = useMediaQuery(MQ_MAX_MD);
    const { t } = useT([
        "common",
        "api-response",
        "fetch-error",
        "accessibility",
    ]);

    useEffect(() => {
        const timer = setTimeout(() => {
            if (query.trim()) {
                setIsLoading(true);
                const search = async () => {
                    await SearchUsersByUsername(query)
                        .then((res) => {
                            if (res.success) {
                                setResults(res.data ?? []);
                            } else {
                                toast(t(`api-response:${res.key}`), {
                                    toastId: res.requestId,
                                    type: "error",
                                });
                            }
                        })
                        .catch((_) => {
                            toast(t("fetch-error:client_fetch_error"), {
                                toastId: "client-fetch-error-id",
                                type: "error",
                            });
                        })
                        .finally(() => {
                            setIsLoading(false);
                            setShowResults(true);
                        });
                };

                search();
                if (onSearch) onSearch(query);
            } else {
                setResults([]);
                setShowResults(false);
            }
        }, 1000);

        return () => clearTimeout(timer);
    }, [query, onSearch, t]);

    const handleClear = () => {
        setQuery("");
        setResults([]);
        setShowResults(false);
    };

    const handleResultClick = () => {
        setMobileOpen(false);
    };

    const searchInput = (
        <div className="relative w-[300px] lg:w-[400px]">
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

            {isLoading && query && (
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
                            <Avatar className="h-8 w-8">
                                <AvatarImage
                                    src={user.profilePicture}
                                    alt={"user image"}
                                    width={32}
                                    height={32}
                                />
                                <AvatarFallback>
                                    {(user.username ?? "U")
                                        .charAt(0)
                                        .toUpperCase()}
                                </AvatarFallback>
                            </Avatar>
                            <div>
                                <p className="text-sm font-medium">
                                    {user.username}
                                </p>
                                <p className="text-muted-foreground text-xs">
                                    {user.email}
                                </p>
                            </div>
                        </Link>
                    ))}
                </div>
            )}

            {showResults && results.length === 0 && !isLoading && query && (
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
                        className="flex-1 justify-end hover:bg-background-hover flex mr-2 rounded-full"
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

    return <div className={cn("relative w-full flex flex-row justify-center", className)}>{searchInput}</div>;
}
