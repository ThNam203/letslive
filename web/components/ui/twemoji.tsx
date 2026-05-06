"use client";

import { useEffect, useRef } from "react";
import twemoji from "twemoji";

type TwemojiProps = {
    emoji: string;
    className?: string;
    title?: string;
    ariaLabel?: string;
};

export function Twemoji({ emoji, className, title, ariaLabel }: TwemojiProps) {
    const containerRef = useRef<HTMLSpanElement>(null);

    useEffect(() => {
        if (!containerRef.current) {
            return;
        }

        containerRef.current.textContent = emoji;
        twemoji.parse(containerRef.current, {
            folder: "svg",
            ext: ".svg",
            className: "twemoji",
        });
    }, [emoji]);

    return (
        <span
            ref={containerRef}
            role="img"
            aria-label={ariaLabel}
            title={title}
            className={`inline-block leading-none [&>img]:m-0 [&>img]:inline-block [&>img]:h-[1em] [&>img]:w-[1em] [&>img]:align-[-0.1em] ${className ?? ""}`}
        >
            {emoji}
        </span>
    );
}
