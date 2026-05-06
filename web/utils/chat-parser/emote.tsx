import React from "react";
import { Twemoji } from "@/components/ui/twemoji";
import { EMOTE_MAP } from "@/constant/emotes";

const EMOTE_REGEX = /:([a-z0-9_]+):/g;

/**
 * Parse message text and replace :shortcodes: with rendered emotes.
 * Returns an array of React nodes (text spans + emote spans).
 *
 * If the entire message is a single emote, it renders larger (sticker-style).
 */
export function parseEmotes(
    text: string | undefined | null,
    options?: {
        emoteClassName?: string;
    },
): React.ReactNode[] {
    if (text == null) {
        return [];
    }

    const parts: React.ReactNode[] = [];
    let lastIndex = 0;
    let match: RegExpExecArray | null;

    // Reset regex state
    EMOTE_REGEX.lastIndex = 0;

    while ((match = EMOTE_REGEX.exec(text)) !== null) {
        const code = match[1];
        const emote = EMOTE_MAP.get(code);

        if (emote) {
            // Add preceding text
            if (match.index > lastIndex) {
                parts.push(text.slice(lastIndex, match.index));
            }

            parts.push(
                <Twemoji
                    key={match.index}
                    className={options?.emoteClassName}
                    ariaLabel={emote.name}
                    title={`:${emote.code}:`}
                    emoji={emote.emoji}
                />,
            );

            lastIndex = EMOTE_REGEX.lastIndex;
        }
    }

    // Add remaining text
    if (lastIndex < text.length) {
        parts.push(text.slice(lastIndex));
    }

    return parts.length > 0 ? parts : [text];
}
