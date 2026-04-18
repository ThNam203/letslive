import { http } from "msw";
import { API_BASE, ok, notFound, noContent, created } from "../utils";
import {
    vods,
    vodComments,
    likedCommentIds,
    ME_USER_ID,
    uid,
    now,
} from "../db";
import { VOD } from "@/types/vod";
import { VODComment } from "@/types/vod-comment";
import { meUser } from "../db";

export const vodHandlers = [
    // GET /vods/author — my own VODs (all visibility)
    http.get(`${API_BASE}/vods/author`, () => {
        const mine = vods.filter((v) => v.userId === ME_USER_ID);
        return ok<VOD[]>(mine);
    }),

    // GET /vods?userId=&page=&limit=
    http.get(`${API_BASE}/vods`, ({ request }) => {
        const url = new URL(request.url);
        const userId = url.searchParams.get("userId");
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const limit = parseInt(url.searchParams.get("limit") ?? "10");

        let results = vods.filter((v) => v.visibility === "public");
        if (userId) results = results.filter((v) => v.userId === userId);

        const slice = results.slice(page * limit, page * limit + limit);
        return ok<VOD[]>(slice, {
            page,
            page_size: limit,
            total: results.length,
        });
    }),

    // GET /popular-vods
    http.get(`${API_BASE}/popular-vods`, ({ request }) => {
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const limit = parseInt(url.searchParams.get("limit") ?? "10");

        const sorted = [...vods]
            .filter((v) => v.visibility === "public" && v.status === "ready")
            .sort((a, b) => b.viewCount - a.viewCount);

        const slice = sorted.slice(page * limit, page * limit + limit);
        return ok<VOD[]>(slice, {
            page,
            page_size: limit,
            total: sorted.length,
        });
    }),

    // GET /vods/:vodId
    http.get(`${API_BASE}/vods/:vodId`, ({ params }) => {
        const { vodId } = params as { vodId: string };
        const vod = vods.find((v) => v.id === vodId);
        if (!vod) return notFound("res_err_vod_not_found", "VOD not found");
        return ok<VOD>(vod);
    }),

    // PATCH /vods/:vodId
    http.patch(`${API_BASE}/vods/:vodId`, async ({ params, request }) => {
        const { vodId } = params as { vodId: string };
        const vod = vods.find((v) => v.id === vodId);
        if (!vod) return notFound("res_err_vod_not_found", "VOD not found");
        const body = (await request.json()) as Partial<VOD>;
        Object.assign(vod, body, { updatedAt: now() });
        return noContent();
    }),

    // DELETE /vods/:vodId
    http.delete(`${API_BASE}/vods/:vodId`, ({ params }) => {
        const { vodId } = params as { vodId: string };
        const idx = vods.findIndex((v) => v.id === vodId);
        if (idx === -1)
            return notFound("res_err_vod_not_found", "VOD not found");
        vods.splice(idx, 1);
        return noContent();
    }),

    // POST /vods/upload — simulate a successful upload + instant "ready"
    http.post(`${API_BASE}/vods/upload`, async ({ request }) => {
        const formData = await request.formData();
        const newVod: VOD = {
            id: uid(),
            livestreamId: null,
            userId: ME_USER_ID,
            title: (formData.get("title") as string) ?? "Untitled",
            description: (formData.get("description") as string) ?? null,
            visibility: ((formData.get("visibility") as string) ??
                "public") as VOD["visibility"],
            thumbnailUrl: null,
            viewCount: 0,
            duration: 0,
            playbackUrl: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8",
            status: "ready",
            originalFileUrl: null,
            createdAt: now(),
            updatedAt: now(),
        };
        vods.push(newVod);
        return created<VOD>(newVod);
    }),

    // -------------------------------------------------------------------------
    // VOD Comments
    // -------------------------------------------------------------------------

    // GET /vods/:vodId/comments
    http.get(`${API_BASE}/vods/:vodId/comments`, ({ request, params }) => {
        const { vodId } = params as { vodId: string };
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const limit = parseInt(url.searchParams.get("limit") ?? "20");

        const topLevel = vodComments.filter(
            (c) => c.vodId === vodId && c.parentId === null && !c.isDeleted,
        );
        const slice = topLevel.slice(page * limit, page * limit + limit);
        return ok<VODComment[]>(slice, {
            page,
            page_size: limit,
            total: topLevel.length,
        });
    }),

    // GET /vod-comments/:commentId/replies
    http.get(
        `${API_BASE}/vod-comments/:commentId/replies`,
        ({ request, params }) => {
            const { commentId } = params as { commentId: string };
            const url = new URL(request.url);
            const page = parseInt(url.searchParams.get("page") ?? "0");
            const limit = parseInt(url.searchParams.get("limit") ?? "20");

            const replies = vodComments.filter(
                (c) => c.parentId === commentId && !c.isDeleted,
            );
            const slice = replies.slice(page * limit, page * limit + limit);
            return ok<VODComment[]>(slice, {
                page,
                page_size: limit,
                total: replies.length,
            });
        },
    ),

    // POST /vods/:vodId/comments
    http.post(
        `${API_BASE}/vods/:vodId/comments`,
        async ({ params, request }) => {
            const { vodId } = params as { vodId: string };
            const body = (await request.json()) as {
                content: string;
                parentId?: string;
            };
            const newComment: VODComment = {
                id: uid(),
                vodId,
                userId: ME_USER_ID,
                parentId: body.parentId ?? null,
                content: body.content,
                isDeleted: false,
                likeCount: 0,
                replyCount: 0,
                createdAt: now(),
                updatedAt: now(),
                user: {
                    id: ME_USER_ID,
                    username: meUser.username,
                    displayName: meUser.displayName,
                    profilePicture: meUser.profilePicture,
                },
            };
            // If it's a reply, increment parent replyCount
            if (body.parentId) {
                const parent = vodComments.find((c) => c.id === body.parentId);
                if (parent) parent.replyCount += 1;
            }
            vodComments.push(newComment);
            return created<VODComment>(newComment);
        },
    ),

    // DELETE /vod-comments/:commentId
    http.delete(`${API_BASE}/vod-comments/:commentId`, ({ params }) => {
        const { commentId } = params as { commentId: string };
        const comment = vodComments.find((c) => c.id === commentId);
        if (!comment)
            return notFound(
                "res_err_vod_comment_not_found",
                "Comment not found",
            );
        comment.isDeleted = true;
        return noContent();
    }),

    // POST /vod-comments/:commentId/like
    http.post(`${API_BASE}/vod-comments/:commentId/like`, ({ params }) => {
        const { commentId } = params as { commentId: string };
        const comment = vodComments.find((c) => c.id === commentId);
        if (comment && !likedCommentIds.has(commentId)) {
            likedCommentIds.add(commentId);
            comment.likeCount += 1;
        }
        return noContent();
    }),

    // DELETE /vod-comments/:commentId/like
    http.delete(`${API_BASE}/vod-comments/:commentId/like`, ({ params }) => {
        const { commentId } = params as { commentId: string };
        const comment = vodComments.find((c) => c.id === commentId);
        if (comment && likedCommentIds.has(commentId)) {
            likedCommentIds.delete(commentId);
            comment.likeCount = Math.max(0, comment.likeCount - 1);
        }
        return noContent();
    }),

    // POST /vod-comments/liked-ids
    http.post(`${API_BASE}/vod-comments/liked-ids`, async ({ request }) => {
        const body = (await request.json()) as { commentIds: string[] };
        const ids = (body?.commentIds ?? []).filter((id) =>
            likedCommentIds.has(id),
        );
        return ok<string[]>(ids);
    }),
];
