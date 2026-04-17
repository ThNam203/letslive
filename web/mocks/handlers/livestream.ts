import { http } from "msw";
import { API_BASE, ok, notFound } from "../utils";
import { livestreams } from "../db";
import { Livestream } from "@/types/livestream";

export const livestreamHandlers = [
    // GET /livestreams?userId=
    http.get(`${API_BASE}/livestreams`, ({ request }) => {
        const url = new URL(request.url);
        const userId = url.searchParams.get("userId");
        const results = userId
            ? livestreams.filter((ls) => ls.userId === userId)
            : livestreams;
        return ok<Livestream[]>(results);
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
