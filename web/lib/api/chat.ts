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
