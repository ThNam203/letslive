"use client";

import { useRef } from "react";
import { Input } from "@/components/ui/input";
import IconClose from "@/components/icons/close";
import IconSearch from "@/components/icons/search";
import useT from "@/hooks/use-translation";
import { SEARCH_QUERY_MAX_LENGTH } from "@/constant/field-limits";
import { cn } from "@/utils/cn";

export default function SearchInput({
    value,
    onChange,
    onSubmit,
    onFocus,
    autoFocus,
    expandedPanelId,
    className,
    inputClassName,
}: {
    value: string;
    onChange: (value: string) => void;
    onSubmit?: () => void;
    onFocus?: () => void;
    autoFocus?: boolean;
    /** id of the results popup while it is open */
    expandedPanelId?: string;
    className?: string;
    inputClassName?: string;
}) {
    const { t } = useT(["common", "accessibility"]);
    const inputRef = useRef<HTMLInputElement>(null);

    const handleClear = () => {
        onChange("");
        inputRef.current?.focus();
    };

    return (
        <form
            role="search"
            className={cn("relative", className)}
            onSubmit={(e) => {
                e.preventDefault();
                onSubmit?.();
            }}
        >
            <IconSearch
                aria-hidden="true"
                className="text-muted-foreground pointer-events-none absolute top-1/2 left-2.5 h-4 w-4 -translate-y-1/2"
            />
            <Input
                ref={inputRef}
                type="search"
                enterKeyHint="search"
                placeholder={t("common:search.placeholder")}
                aria-label={t("common:search.title")}
                aria-controls={expandedPanelId}
                maxLength={SEARCH_QUERY_MAX_LENGTH}
                value={value}
                onChange={(e) => onChange(e.target.value)}
                onFocus={onFocus}
                autoFocus={autoFocus}
                className={cn(
                    "border-border pr-9 pl-8 [&::-webkit-search-cancel-button]:hidden",
                    inputClassName,
                )}
            />
            {value && (
                <button
                    type="button"
                    onClick={handleClear}
                    className="text-muted-foreground hover:text-foreground absolute top-1/2 right-1 -translate-y-1/2 rounded p-1.5"
                    aria-label={t("accessibility:clear_search")}
                >
                    <IconClose className="h-4 w-4" />
                </button>
            )}
        </form>
    );
}
