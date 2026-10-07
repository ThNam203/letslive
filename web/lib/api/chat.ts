import { ReceivedMessage } from "../../types/message";
import { ApiResponse } from "@/types/fetch-response";
import { fetchClient } from "@/utils/fetchClient";

export async function GetMessages(roomId: string): Promise<{
    messages: ReceivedMessage[];
}> {
    const res = await fetchClient<ApiResponse<ReceivedMessage[]>>(
        `/messages?roomId=${roomId}`,
    );
    return { messages: res.data ?? [] };
}

// The line reaches every viewer, the sender included, as a chat.message push
// on the room topic.
export async function SendChatMessage(
    roomId: string,
    text: string,
): Promise<ApiResponse<ReceivedMessage>> {
    return fetchClient<ApiResponse<ReceivedMessage>>(`/messages`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ roomId, text }),
    });
}
