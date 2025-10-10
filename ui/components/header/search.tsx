"use client";

import { useState, useEffect } from "react";

import { toast } from "react-toastify";
import { User } from "../../types/user";
import { SearchUsersByUsername } from "../../lib/api/user";
import { Input } from "../ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import Link from "next/link";
import IconClose from "../icons/close";
import useT from "@/hooks/use-translation";

export default function SearchBar({
    onSearch,
}: {
    onSearch?: (query: string) => void;
}) {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<User[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [showResults, setShowResults] = useState(false);
    const { t } = useT(["common", "api-response", "fetch-error"]);

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
                                type: "error" 
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
    }, [query, onSearch]);

    const handleClear = () => {
        setQuery("");
        setResults([]);
        setShowResults(false);
    };

    return (
        <div className="relative w-full max-w-sm">
            <div className="relative">
                <Input
                    type="text"
                    placeholder={t("common:search_users")}
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className="border-border pr-8"
                    onFocus={() => query.trim() && setShowResults(true)}
                />
                {query && (
                    <button
                        onClick={handleClear}
                        className="text-muted-foreground absolute right-2 top-1/2 -translate-y-1/2 hover:text-foreground"
                        aria-label="Clear search"
                    >
                        <IconClose className="h-4 w-4" />
                    </button>
                )}
            </div>

            {isLoading && query && (
                <div className="absolute mt-1 w-full rounded-sm border bg-background p-4 shadow-md">
                    <div className="flex items-center justify-center">
                        <p className="text-muted-foreground text-sm">
                            {t("common:searching")}
                        </p>
                    </div>
                </div>
            )}

            {showResults && results.length > 0 && !isLoading && (
                <div className="absolute z-10 mt-1 max-h-60 w-full overflow-auto rounded-sm bg-background shadow-md">
                    {results.map((user) => (
                        <Link
                            key={user.id}
                            href={`/users/${user.id}`}
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
                                    {user.username.charAt(0).toUpperCase()}
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
                <div className="absolute mt-1 w-full rounded-sm border bg-background p-4 shadow-md">
                    <p className="text-muted-foreground text-sm">
                        {t("common:no_users_found")}
                    </p>
                </div>
            )}
        </div>
    );
}
