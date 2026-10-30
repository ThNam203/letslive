import { findHighlightRanges, foldText } from "@/utils/highlight";
import { meUser, otherUsers } from "./db";

const MAX_LIMIT = 20;

export function searchPaging(url: URL, defaultLimit: number) {
    const page = Math.max(
        0,
        parseInt(url.searchParams.get("page") ?? "0") || 0,
    );
    const rawLimit =
        parseInt(url.searchParams.get("limit") ?? "") || defaultLimit;
    const limit = Math.min(MAX_LIMIT, Math.max(1, rawLimit));
    return { page, limit };
}

export function matches(text: string | null | undefined, query: string) {
    return (
        Boolean(text) && findHighlightRanges(text as string, query).length > 0
    );
}

export function usernameOf(userId: string) {
    return [meUser, ...otherUsers].find((u) => u.id === userId)?.username;
}

// Exact (accent-insensitive) hits rank above typo-tolerant ones; within each,
// title > streamer name > description, roughly like the backend's weighting.
export function mediaMatchScore(
    item: { title: string; description: string | null; userId: string },
    query: string,
) {
    const fields = [item.title, usernameOf(item.userId), item.description];
    const phrase = foldText(query.trim());
    const exact = fields.findIndex((f) => f && foldText(f).includes(phrase));
    if (exact !== -1) return 6 - exact;
    const fuzzy = fields.findIndex((f) => matches(f, query));
    return fuzzy === -1 ? 0 : 3 - fuzzy;
}

export function pageOf<T>(items: T[], page: number, limit: number) {
    return items.slice(page * limit, page * limit + limit);
}
