"use client";

import { useState, useEffect } from "react";
import { X } from "lucide-react";

import { toast } from "react-toastify";
import { User } from "../../types/user";
import { SearchUsersByUsername } from "../../lib/api/user";
import { Input } from "../ui/input";
import { Avatar, AvatarFallback, AvatarImage } from "../ui/avatar";
import Link from "next/link";

export default function SearchBar({
    onSearch,
}: {
    onSearch?: (query: string) => void;
}) {
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<User[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [showResults, setShowResults] = useState(false);

    useEffect(() => {
        const timer = setTimeout(() => {
            if (query.trim()) {
                setIsLoading(true);
                const search = async () => {
                    const { users, fetchError } = await SearchUsersByUsername(
                        query
                    );

                    if (fetchError) {
                        toast(fetchError.message, { type: "error" });
                    } else if (users) {
                        setResults(users);
                    }

                    setIsLoading(false);
                    setShowResults(true);
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
                    placeholder="Search users..."
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    className="pr-8"
                    onFocus={() => query.trim() && setShowResults(true)}
                />
                {query && (
                    <button
                        onClick={handleClear}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                        aria-label="Clear search"
                    >
                        <X className="h-4 w-4" />
                    </button>
                )}
            </div>

            {isLoading && query && (
                <div className="absolute mt-1 w-full rounded-sm border bg-background p-4 shadow-md">
                    <div className="flex items-center justify-center">
                        <p className="text-sm text-muted-foreground">
                            Searching...
                        </p>
                    </div>
                </div>
            )}

            {showResults && results.length > 0 && !isLoading && (
                <div className="absolute mt-1 w-full rounded-sm bg-background shadow-md z-10 overflow-auto max-h-60">
                    {results.map((user) => (
                        <Link
                            key={user.id}
                            href={`/users/${user.id}`}
                            className="cursor-pointer hover:bg-gray-400 w-full flex flex-row items-center gap-3 p-2"
                        >
                                <Avatar className="h-8 w-8">
                                    <AvatarImage
                                        src={
                                            user.profilePicture
                                        }
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
                                    <p className="text-xs text-muted-foreground">
                                        {user.email}
                                    </p>
                                </div>
                        </Link>
                    ))}
                </div>
            )}

            {showResults && results.length === 0 && !isLoading && query && (
                <div className="absolute mt-1 w-full rounded-sm border bg-background p-4 shadow-md">
                    <p className="text-sm text-muted-foreground">
                        No users found
                    </p>
                </div>
            )}
        </div>
    );
}
