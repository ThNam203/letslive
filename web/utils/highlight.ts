export type TextRange = { start: number; end: number };
export type TextSegment = { text: string; isMatch: boolean };

const COMBINING_MARKS = /\p{M}/gu;
const WORD = /[\p{L}\p{N}]+/gu;
const MIN_FUZZY_TOKEN_LENGTH = 4;

function foldChar(ch: string): string {
    return ch
        .normalize("NFD")
        .replace(COMBINING_MARKS, "")
        .replace(/đ/g, "d")
        .replace(/Đ/g, "D")
        .toLowerCase();
}

function foldWithMap(text: string): { folded: string; origin: number[] } {
    let folded = "";
    const origin: number[] = [];
    let index = 0;
    for (const ch of text) {
        const f = foldChar(ch);
        for (let i = 0; i < f.length; i++) origin.push(index);
        folded += f;
        index += ch.length;
    }
    return { folded, origin };
}

export function foldText(text: string): string {
    return foldWithMap(text).folded;
}

function maxEdits(length: number): number {
    if (length <= 5) return 1;
    if (length <= 9) return 2;
    return 3;
}

// Swapped neighbours ("jnae" for "jane") count as one typo, not two.
function editDistance(a: string, b: string, limit: number): number {
    if (Math.abs(a.length - b.length) > limit) return limit + 1;
    let before: number[] = [];
    let prev = Array.from({ length: b.length + 1 }, (_, j) => j);
    for (let i = 1; i <= a.length; i++) {
        const curr = [i];
        let rowMin = i;
        for (let j = 1; j <= b.length; j++) {
            const cost = a[i - 1] === b[j - 1] ? 0 : 1;
            let best = Math.min(
                prev[j] + 1,
                curr[j - 1] + 1,
                prev[j - 1] + cost,
            );
            if (
                i > 1 &&
                j > 1 &&
                a[i - 1] === b[j - 2] &&
                a[i - 2] === b[j - 1]
            ) {
                best = Math.min(best, before[j - 2] + 1);
            }
            curr[j] = best;
            rowMin = Math.min(rowMin, best);
        }
        if (rowMin > limit) return limit + 1;
        before = prev;
        prev = curr;
    }
    return prev[b.length];
}

function isFuzzyWordMatch(token: string, word: string): boolean {
    const limit = maxEdits(token.length);
    if (editDistance(token, word, limit) <= limit) return true;
    // "jan" should light up "jane_streams": compare against the word's prefix too
    return (
        word.length > token.length &&
        editDistance(token, word.slice(0, token.length), limit) <= limit
    );
}

function substringRanges(haystack: string, needle: string): TextRange[] {
    const ranges: TextRange[] = [];
    if (!needle) return ranges;
    let from = 0;
    while (from <= haystack.length - needle.length) {
        const at = haystack.indexOf(needle, from);
        if (at === -1) break;
        ranges.push({ start: at, end: at + needle.length });
        from = at + needle.length;
    }
    return ranges;
}

function mergeRanges(ranges: TextRange[]): TextRange[] {
    const sorted = [...ranges].sort((a, b) => a.start - b.start);
    return sorted.reduce<TextRange[]>((merged, range) => {
        const last = merged[merged.length - 1];
        if (last && range.start <= last.end) {
            return [
                ...merged.slice(0, -1),
                { start: last.start, end: Math.max(last.end, range.end) },
            ];
        }
        return [...merged, range];
    }, []);
}

/**
 * Parts of `text` that match `query`, ignoring case and Vietnamese accents.
 * Exact substrings win; a query word with no exact hit highlights the words
 * it is a likely typo of, mirroring the backend's fuzzy search.
 */
export function findHighlightRanges(text: string, query: string): TextRange[] {
    const tokens = foldText(query).split(/\s+/).filter(Boolean);
    if (!text || tokens.length === 0) return [];

    const { folded, origin } = foldWithMap(text);
    const words = [...folded.matchAll(WORD)].map((m) => ({
        word: m[0],
        start: m.index,
        end: m.index + m[0].length,
    }));

    const foldedRanges = [
        ...substringRanges(folded, tokens.join(" ")),
        ...tokens.flatMap((token) => {
            if (token.length === 1) {
                return words
                    .filter(({ word }) => word.startsWith(token))
                    .map(({ start }) => ({ start, end: start + 1 }));
            }
            const exact = substringRanges(folded, token);
            if (exact.length > 0 || token.length < MIN_FUZZY_TOKEN_LENGTH) {
                return exact;
            }
            return words
                .filter(({ word }) => isFuzzyWordMatch(token, word))
                .map(({ start, end }) => ({ start, end }));
        }),
    ];

    return mergeRanges(
        foldedRanges.map(({ start, end }) => {
            const last = origin[end - 1];
            return {
                start: origin[start],
                end:
                    last +
                    String.fromCodePoint(text.codePointAt(last) ?? 0).length,
            };
        }),
    );
}

export function splitByRanges(
    text: string,
    ranges: TextRange[],
): TextSegment[] {
    const segments: TextSegment[] = [];
    let cursor = 0;
    for (const { start, end } of ranges) {
        if (start > cursor)
            segments.push({ text: text.slice(cursor, start), isMatch: false });
        segments.push({ text: text.slice(start, end), isMatch: true });
        cursor = end;
    }
    if (cursor < text.length)
        segments.push({ text: text.slice(cursor), isMatch: false });
    return segments;
}

export function excerptAroundMatch(
    text: string,
    query: string,
    radius = 60,
): string | null {
    const [first] = findHighlightRanges(text, query);
    if (!first) return null;

    let start = Math.max(0, first.start - radius);
    let end = Math.min(text.length, first.end + radius);
    if (start > 0) {
        const space = text.indexOf(" ", start);
        start = space === -1 || space >= first.start ? start : space + 1;
    }
    if (end < text.length) {
        const space = text.lastIndexOf(" ", end);
        end = space <= first.end ? end : space;
    }
    const body = text.slice(start, end).replace(/\s+/g, " ").trim();
    return `${start > 0 ? "…" : ""}${body}${end < text.length ? "…" : ""}`;
}
