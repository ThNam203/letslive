import { ReceivedMessage } from "../../types/message";
import { fetchClient } from "@/utils/fetchClient";

// TODO: update erorr handling
export async function GetMessages(roomId: string): Promise<{
    messages: ReceivedMessage[];
}> {
    try {
        const data = (await fetchClient<ReceivedMessage[]>(
            `/messages?roomId=${roomId}`,
        )) as any;
        return { messages: data.data };
    } catch (error) {
        return { messages: [] };
    }
}
