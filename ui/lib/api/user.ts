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

export async function GetUserById(userId: string): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>("/user/" + userId);
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetAllUsers(): Promise<{users?: User[], fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User[]>("/users");
        return { users: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetMeProfile(): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>("/user/me");
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function UpdateProfile({id, username, bio}:{id: string, username: string, bio: string}): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>(`/user/${id}`, {
            method: "PUT",
            body: JSON.stringify({ username, bio }),
        });

        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}