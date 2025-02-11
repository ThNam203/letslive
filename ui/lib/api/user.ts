import { FetchError } from "@/types/fetch-error";
import { FetchOptions } from "@/types/fetch-options";
import { User } from "@/types/user";
import { fetchClient } from "@/utils/fetchClient";

/**
 * Fetches the list of online users.
 */
export async function GetOnlineUsers(headers: Record<string, string> = {}): Promise<{users: User[], fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User[]>("/user?isOnline=true", {headers: headers});
        return { users: data };
    } catch (error) {
        return { users: [], fetchError: error as FetchError };
    }
}

export async function GetUserById(userId: string, headers: Record<string, string> = {}): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>("/user/" + userId, {headers: headers});
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetAllUsers(headers: Record<string, string> = {}): Promise<{users?: User[], fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User[]>("/users", {headers: headers});
        return { users: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetMeProfile(headers: Record<string, string> = {}): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>("/user/me", {headers: headers});
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function UpdateProfile({id, username, bio}:{id: string, username: string, bio: string}, headers: Record<string, string> = {}): Promise<{user?: User, fetchError?: FetchError}> {
    try {
        const data = await fetchClient<User>(`/user/${id}`, {
            method: "PUT",
            body: JSON.stringify({ username, bio }),
            headers: headers
        }, );

        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}