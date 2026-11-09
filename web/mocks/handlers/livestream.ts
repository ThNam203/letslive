import { http } from "msw";
import { API_BASE, ok, notFound } from "../utils";
import { livestreams } from "../db";
import { Livestream } from "@/types/livestream";
import { mediaMatchScore, pageOf, searchPaging } from "../search";

export const livestreamHandlers = [
    // GET /livestreams?userId= — current livestream or null; the backend returns one object, not a list
    http.get(`${API_BASE}/livestreams`, ({ request }) => {
        const url = new URL(request.url);
        const userId = url.searchParams.get("userId");
        const current =
            livestreams.find(
                (ls) => ls.userId === userId && ls.endedAt === null,
            ) ?? null;
        return ok<Livestream | null>(current);
    }),

    // GET /livestreams/search?q=&page=&limit= — only streams that are live now
    http.get(`${API_BASE}/livestreams/search`, ({ request }) => {
        const url = new URL(request.url);
        const query = url.searchParams.get("q")?.trim() ?? "";
        const { page, limit } = searchPaging(url, 20);
        const results = livestreams
            .filter((ls) => ls.endedAt === null && ls.visibility === "public")
            .map((ls) => ({ ls, score: mediaMatchScore(ls, query) }))
            .filter(({ score }) => score > 0)
            .sort(
                (a, b) => b.score - a.score || b.ls.viewCount - a.ls.viewCount,
            )
            .map(({ ls }) => ls);
        return ok<Livestream[]>(pageOf(results, page, limit), {
            page,
            page_size: limit,
        });
    }),

    // GET /popular-livestreams?page=&limit=
    http.get(`${API_BASE}/popular-livestreams`, ({ request }) => {
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const limit = parseInt(url.searchParams.get("limit") ?? "10");
        const active = livestreams.filter((ls) => ls.endedAt === null);
        const slice = active.slice(page * limit, page * limit + limit);
        return ok<Livestream[]>(slice, {
            page,
            page_size: limit,
            total: active.length,
        });
    }),
];
