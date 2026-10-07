// "1:23", "01:23" or "1:02:03" standing on their own, not inside "a1:23",
// "1:23:4" or a ratio like "16:9"
const TIMESTAMP_PATTERN =
    /(?<![\w:.])(?:(\d{1,2}):)?(\d{1,2}):([0-5]\d)(?![\w:])/g;

export type TimestampSegment =
    | { type: "text"; text: string }
    | { type: "timestamp"; text: string; seconds: number };

/**
 * Splits text into plain runs and timestamps. Timestamps past `maxSeconds`
 * stay plain text, so a comment can't link beyond the end of the video.
 */
export function splitTimestamps(
    text: string,
    maxSeconds?: number,
): TimestampSegment[] {
    const segments: TimestampSegment[] = [];
    let last = 0;

    for (const match of text.matchAll(TIMESTAMP_PATTERN)) {
        const [raw, hours, minutes, secs] = match;
        const h = hours ? Number(hours) : 0;
        const m = Number(minutes);
        if (hours && m > 59) continue;

        const seconds = h * 3600 + m * 60 + Number(secs);
        if (maxSeconds && maxSeconds > 0 && seconds > maxSeconds) continue;

        const start = match.index;
        if (start > last) {
            segments.push({ type: "text", text: text.slice(last, start) });
        }
        segments.push({ type: "timestamp", text: raw, seconds });
        last = start + raw.length;
    }

    if (last < text.length) {
        segments.push({ type: "text", text: text.slice(last) });
    }
    return segments;
}

/**
 * Parses a share link's `t` param: plain seconds ("90", "90s") or
 * "1h2m3s"-style. Returns undefined for anything else.
 */
export function parseStartTime(
    value: string | string[] | undefined,
): number | undefined {
    const raw = Array.isArray(value) ? value[0] : value;
    if (!raw) return undefined;

    const plain = raw.match(/^(\d+)s?$/);
    if (plain) return Number(plain[1]);

    const parts = raw.match(/^(?:(\d+)h)?(?:(\d+)m)?(?:(\d+)s)?$/);
    if (!parts) return undefined;
    const [, h, m, s] = parts;
    const total = Number(h ?? 0) * 3600 + Number(m ?? 0) * 60 + Number(s ?? 0);
    return total > 0 ? total : undefined;
}
