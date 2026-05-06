"use client";

import { useState } from "react";
import {
    Popover,
    PopoverTrigger,
    PopoverContent,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { Twemoji } from "@/components/ui/twemoji";
import { EMOTES, EMOTE_CATEGORIES, type Emote } from "@/constant/emotes";

export default function EmotePicker({
    onSelect,
    disabled,
    searchPlaceholder,
    emptyStateText,
    getCategoryLabel,
    open,
    onOpenChange,
    searchValue,
    onSearchChange,
}: {
    onSelect: (code: string) => void;
    disabled?: boolean;
    searchPlaceholder: string;
    emptyStateText: string;
    getCategoryLabel: (category: (typeof EMOTE_CATEGORIES)[number]) => string;
    open?: boolean;
    onOpenChange?: (open: boolean) => void;
    searchValue?: string;
    onSearchChange?: (search: string) => void;
}) {
    const [internalOpen, setInternalOpen] = useState(false);
    const [internalSearch, setInternalSearch] = useState("");
    const resolvedOpen = open ?? internalOpen;
    const resolvedSearch = searchValue ?? internalSearch;

    const setOpen = (nextOpen: boolean) => {
        if (open === undefined) {
            setInternalOpen(nextOpen);
        }
        onOpenChange?.(nextOpen);
    };

    const setSearch = (nextSearch: string) => {
        if (searchValue === undefined) {
            setInternalSearch(nextSearch);
        }
        onSearchChange?.(nextSearch);
    };

    const filtered = resolvedSearch
        ? EMOTES.filter(
              (e) =>
                  e.code.includes(resolvedSearch.toLowerCase()) ||
                  e.name.toLowerCase().includes(resolvedSearch.toLowerCase()),
          )
        : EMOTES;

    const handleSelect = (emote: Emote) => {
        onSelect(`:${emote.code}:`);
        setOpen(false);
        setSearch("");
    };

    return (
        <Popover open={resolvedOpen} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    disabled={disabled}
                >
                    <Twemoji emoji="😊" className="text-lg" />
                </Button>
            </PopoverTrigger>
            <PopoverContent
                side="top"
                align="start"
                className="w-72 p-0 mr-2 bg-muted"
                onOpenAutoFocus={(e) => e.preventDefault()}
            >
                {/* Search */}
                <div className="border-b p-2">
                    <input
                        type="text"
                        placeholder={searchPlaceholder}
                        className="w-full bg-transparent text-sm outline-none"
                        value={resolvedSearch}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>

                {/* Emote grid */}
                <div className="max-h-60 overflow-y-auto p-2">
                    {resolvedSearch ? (
                        <div className="grid grid-cols-8 gap-0.5">
                            {filtered.map((emote) => (
                                <button
                                    key={emote.code}
                                    type="button"
                                    className="hover:bg-muted flex h-8 w-8 items-center justify-center rounded text-lg transition-colors"
                                    onClick={() => handleSelect(emote)}
                                    title={`:${emote.code}:`}
                                >
                                    <Twemoji
                                        emoji={emote.emoji}
                                        title={`:${emote.code}:`}
                                        ariaLabel={emote.name}
                                        className="text-lg"
                                    />
                                </button>
                            ))}
                            {filtered.length === 0 && (
                                <p className="text-muted-foreground col-span-8 py-4 text-center text-xs">
                                    {emptyStateText}
                                </p>
                            )}
                        </div>
                    ) : (
                        EMOTE_CATEGORIES.map((category) => {
                            const emotes = EMOTES.filter(
                                (e) => e.category === category,
                            );
                            return (
                                <div key={category} className="mb-2">
                                    <p className="text-muted-foreground mb-1 px-1 text-[10px] font-semibold uppercase">
                                        {getCategoryLabel(category)}
                                    </p>
                                    <div className="grid grid-cols-8 gap-0.5">
                                        {emotes.map((emote) => (
                                            <button
                                                key={emote.code}
                                                type="button"
                                                className="hover:bg-muted flex h-8 w-8 items-center justify-center rounded text-lg transition-colors"
                                                onClick={() =>
                                                    handleSelect(emote)
                                                }
                                                title={`:${emote.code}:`}
                                            >
                                                <Twemoji
                                                    emoji={emote.emoji}
                                                    title={`:${emote.code}:`}
                                                    ariaLabel={emote.name}
                                                    className="text-lg"
                                                />
                                            </button>
                                        ))}
                                    </div>
                                </div>
                            );
                        })
                    )}
                </div>
            </PopoverContent>
        </Popover>
    );
}
