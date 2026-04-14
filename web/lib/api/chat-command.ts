import { ApiResponse } from "@/types/fetch-response";
import {
    ChatCommand,
    ChatCommandScope,
    MyChatCommands,
} from "@/types/chat-command";
import { fetchClient } from "@/utils/fetchClient";

export async function GetRoomChatCommands(
    roomId: string,
): Promise<ApiResponse<ChatCommand[]>> {
    return fetchClient<ApiResponse<ChatCommand[]>>(
        `/chat-commands?roomId=${encodeURIComponent(roomId)}`,
    );
}

export async function GetMyChatCommands(): Promise<
    ApiResponse<MyChatCommands>
> {
    return fetchClient<ApiResponse<MyChatCommands>>(`/chat-commands/mine`);
}

export async function CreateChatCommand(body: {
    scope: ChatCommandScope;
    name: string;
    response: string;
    description?: string;
}): Promise<ApiResponse<ChatCommand>> {
    return fetchClient<ApiResponse<ChatCommand>>(`/chat-commands`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
}

export async function UpdateChatCommand(
    id: string,
    body: {
        name?: string;
        response?: string;
        description?: string;
    },
): Promise<ApiResponse<ChatCommand>> {
    return fetchClient<ApiResponse<ChatCommand>>(`/chat-commands/${id}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });
}

export async function DeleteChatCommand(
    id: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/chat-commands/${id}`, {
        method: "DELETE",
    });
}
