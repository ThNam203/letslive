import { ApiResponse } from "@/types/fetch-response";
import {
    Conversation,
    ConversationType,
    DmMessage,
    DmMessageType,
} from "@/types/dm";
import { fetchClient } from "@/utils/fetchClient";

export async function GetConversations(
    page: number = 0,
    limit: number = 20,
): Promise<ApiResponse<Conversation[]>> {
    return fetchClient<ApiResponse<Conversation[]>>(
        `/v1/conversations?page=${page}&limit=${limit}`,
    );
}

export async function GetConversation(
    conversationId: string,
): Promise<ApiResponse<Conversation>> {
    return fetchClient<ApiResponse<Conversation>>(
        `/v1/conversations/${conversationId}`,
    );
}

export async function CreateConversation(body: {
    type: ConversationType;
    participantIds: string[];
    participantUsernames?: Record<string, string>;
    participantDisplayNames?: Record<string, string>;
    participantProfilePictures?: Record<string, string>;
    creatorUsername?: string;
    creatorDisplayName?: string;
    creatorProfilePicture?: string;
    name?: string;
}): Promise<ApiResponse<Conversation>> {
    return fetchClient<ApiResponse<Conversation>>(`/v1/conversations`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
}

export async function UpdateConversation(
    conversationId: string,
    body: { name?: string; avatarUrl?: string },
): Promise<ApiResponse<Conversation>> {
    return fetchClient<ApiResponse<Conversation>>(
        `/v1/conversations/${conversationId}`,
        {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        },
    );
}

export async function LeaveConversation(
    conversationId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/v1/conversations/${conversationId}`,
        { method: "DELETE" },
    );
}

export async function AddParticipant(
    conversationId: string,
    body: {
        userId: string;
        username: string;
        displayName?: string;
        profilePicture?: string;
    },
): Promise<ApiResponse<Conversation>> {
    return fetchClient<ApiResponse<Conversation>>(
        `/v1/conversations/${conversationId}/participants`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        },
    );
}

export async function RemoveParticipant(
    conversationId: string,
    userId: string,
): Promise<ApiResponse<Conversation>> {
    return fetchClient<ApiResponse<Conversation>>(
        `/v1/conversations/${conversationId}/participants/${userId}`,
        { method: "DELETE" },
    );
}

export async function GetDmMessages(
    conversationId: string,
    before?: string,
    limit: number = 50,
): Promise<ApiResponse<DmMessage[]>> {
    let url = `/v1/conversations/${conversationId}/messages?limit=${limit}`;
    if (before) url += `&before=${before}`;
    return fetchClient<ApiResponse<DmMessage[]>>(url);
}

export async function SendDmMessage(
    conversationId: string,
    body: {
        text: string;
        type?: DmMessageType;
        senderUsername: string;
        imageUrls?: string[];
        replyTo?: string;
    },
): Promise<ApiResponse<DmMessage>> {
    return fetchClient<ApiResponse<DmMessage>>(
        `/v1/conversations/${conversationId}/messages`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(body),
        },
    );
}

export async function EditDmMessage(
    conversationId: string,
    messageId: string,
    text: string,
): Promise<ApiResponse<DmMessage>> {
    return fetchClient<ApiResponse<DmMessage>>(
        `/v1/conversations/${conversationId}/messages/${messageId}`,
        {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ text }),
        },
    );
}

export async function DeleteDmMessage(
    conversationId: string,
    messageId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/v1/conversations/${conversationId}/messages/${messageId}`,
        { method: "DELETE" },
    );
}

export async function MarkConversationRead(
    conversationId: string,
    messageId?: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/v1/conversations/${conversationId}/read`,
        {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ messageId }),
        },
    );
}

export async function GetUnreadCounts(): Promise<
    ApiResponse<Record<string, number>>
> {
    return fetchClient<ApiResponse<Record<string, number>>>(
        `/v1/conversations/unread-counts`,
    );
}
