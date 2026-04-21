import { ReceivedMessage } from "../../types/message";
import { fetchClient } from "@/utils/fetchClient";

export async function GetMessages(roomId: string): Promise<{
    messages: ReceivedMessage[];
}> {
    const data = (await fetchClient<ReceivedMessage[]>(
        `/messages?roomId=${roomId}`,
    )) as any;
    return { messages: data.data ?? [] };
}
