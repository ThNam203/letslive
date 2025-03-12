import { FetchError } from "../../types/fetch-error";
import { ReceivedMessage } from "../../types/message";
import { fetchClient } from "../../utils/fetchClient";


export async function GetMessages(roomId: string): Promise<{
    messages: ReceivedMessage[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<ReceivedMessage[]>(`/messages?roomId=${roomId}`);
        return { messages: data };
    } catch (error) {
        return { messages: [], fetchError: error as FetchError };
    }
}