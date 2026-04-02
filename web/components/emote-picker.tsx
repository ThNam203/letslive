"use client";

import { useState } from "react";
import {
    Popover,
    PopoverTrigger,
    PopoverContent,
} from "@/components/ui/popover";
import { Button } from "@/components/ui/button";
import { EMOTES, EMOTE_CATEGORIES, type Emote } from "@/constant/emotes";

export default function EmotePicker({
    onSelect,
    disabled,
}: {
    onSelect: (code: string) => void;
    disabled?: boolean;
}) {
    const [open, setOpen] = useState(false);
    const [search, setSearch] = useState("");

    const filtered = search
        ? EMOTES.filter(
              (e) =>
                  e.code.includes(search.toLowerCase()) ||
                  e.name.toLowerCase().includes(search.toLowerCase()),
          )
        : EMOTES;

    const handleSelect = (emote: Emote) => {
        onSelect(`:${emote.code}:`);
        setOpen(false);
        setSearch("");
    };

    return (
        <Popover open={open} onOpenChange={setOpen}>
            <PopoverTrigger asChild>
                <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="h-9 w-9 shrink-0"
                    disabled={disabled}
                >
                    <span className="text-lg">😊</span>
                </Button>
            </PopoverTrigger>
            <PopoverContent
                side="top"
                align="start"
                className="w-72 p-0"
                onOpenAutoFocus={(e) => e.preventDefault()}
            >
                {/* Search */}
                <div className="border-b p-2">
                    <input
                        type="text"
                        placeholder="Search emotes..."
                        className="bg-transparent w-full text-sm outline-none"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>

                {/* Emote grid */}
                <div className="max-h-60 overflow-y-auto p-2">
                    {search ? (
                        <div className="grid grid-cols-8 gap-0.5">
                            {filtered.map((emote) => (
                                <button
                                    key={emote.code}
                                    type="button"
                                    className="hover:bg-muted flex h-8 w-8 items-center justify-center rounded text-lg transition-colors"
                                    onClick={() => handleSelect(emote)}
                                    title={`:${emote.code}:`}
                                >
                                    {emote.emoji}
                                </button>
                            ))}
                            {filtered.length === 0 && (
                                <p className="text-muted-foreground col-span-8 py-4 text-center text-xs">
                                    No emotes found
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
                                        {category}
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
                                                {emote.emoji}
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
