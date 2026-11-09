"use client";

import { useMemo } from "react";
import { findHighlightRanges, splitByRanges } from "@/utils/highlight";

export default function Highlight({
    text,
    query,
    className,
}: {
    text: string;
    query: string;
    className?: string;
}) {
    const segments = useMemo(
        () => splitByRanges(text, findHighlightRanges(text, query)),
        [text, query],
    );
    return (
        <span className={className}>
            {segments.map((segment, idx) =>
                segment.isMatch ? (
                    <mark
                        key={idx}
                        className="text-foreground rounded-[2px] bg-yellow-200/80 dark:bg-yellow-500/30"
                    >
                        {segment.text}
                    </mark>
                ) : (
                    <span key={idx}>{segment.text}</span>
                ),
            )}
        </span>
    );
}
