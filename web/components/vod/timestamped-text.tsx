"use client";

import { Fragment } from "react";
import { useVODPlayback } from "@/contexts/vod-playback-context";
import { splitTimestamps } from "@/utils/vod-timestamps";
import useT from "@/hooks/use-translation";

/** Renders text with "1:23"-style timestamps as buttons that seek the VOD. */
export default function TimestampedText({ text }: { text: string }) {
    const { t } = useT("common");
    const playback = useVODPlayback();
    if (!playback) return <>{text}</>;

    return (
        <>
            {splitTimestamps(text, playback.duration).map((segment, i) =>
                segment.type === "text" ? (
                    <Fragment key={i}>{segment.text}</Fragment>
                ) : (
                    <button
                        key={i}
                        type="button"
                        onClick={() => playback.seekTo(segment.seconds)}
                        aria-label={t("common:vod.seek_to", {
                            time: segment.text,
                        })}
                        className="text-primary font-medium hover:underline"
                    >
                        {segment.text}
                    </button>
                ),
            )}
        </>
    );
}
