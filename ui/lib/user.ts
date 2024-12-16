import { FetchError } from "@/types/fetch-error";
import { User } from "@/types/user";
import { fetchClient } from "@/utils/fetchClient";

/**
 * Fetches the list of online users.
 */
export async function GetOnlineUsers(): Promise<{users: User[], fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User[]>("/user?isOnline=true");
        return { users: data };
    } catch (error) {
        return { users: [], fetchError: error as FetchError };
    }
}