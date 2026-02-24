import { ApiResponse } from "@/types/fetch-response";
import { VODComment, CreateVODCommentRequest } from "@/types/vod-comment";
import { fetchClient } from "@/utils/fetchClient";

export async function GetVODComments(
    vodId: string,
    page: number = 0,
    limit: number = 10,
): Promise<ApiResponse<VODComment[]>> {
    return fetchClient<ApiResponse<VODComment[]>>(
        `/vods/${vodId}/comments?page=${page}&limit=${limit}`,
    );
}

export async function GetCommentReplies(
    commentId: string,
    page: number = 0,
    limit: number = 20,
): Promise<ApiResponse<VODComment[]>> {
    return fetchClient<ApiResponse<VODComment[]>>(
        `/vod-comments/${commentId}/replies?page=${page}&limit=${limit}`,
    );
}

export async function CreateVODComment(
    vodId: string,
    data: CreateVODCommentRequest,
): Promise<ApiResponse<VODComment>> {
    return fetchClient<ApiResponse<VODComment>>(`/vods/${vodId}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(data),
    });
}

export async function DeleteVODComment(
    commentId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/vod-comments/${commentId}`, {
        method: "DELETE",
    });
}

export async function LikeVODComment(
    commentId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/vod-comments/${commentId}/like`,
        { method: "POST" },
    );
}

export async function UnlikeVODComment(
    commentId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/vod-comments/${commentId}/like`,
        { method: "DELETE" },
    );
}

export async function GetUserLikedCommentIds(
    commentIds: string[],
): Promise<ApiResponse<string[]>> {
    return fetchClient<ApiResponse<string[]>>(
        `/vod-comments/liked-ids`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ commentIds }),
        },
    );
}
