import { FetchError } from "../../types/fetch-error";
import { LivestreamInformation, User } from "../../types/user";
import { fetchClient } from "../../utils/fetchClient";

/**
 * Fetches the list of online users.
 */
export async function GetOnlineUsers(): Promise<{
    users: User[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<User[]>("/users?liveStatus=on");
        return { users: data };
    } catch (error) {
        return { users: [], fetchError: error as FetchError };
    }
}

export async function SearchUsersByUsername(query: string): Promise<{
    users: User[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<User[]>(`/users/search?username=${encodeURIComponent(query)}`);
        return { users: data };
    } catch (error) {
        return { users: [], fetchError: error as FetchError };
    }
}

export async function GetUserById(
    userId: string
): Promise<{ user?: User; fetchError?: FetchError }> {
    try {
        const data = await fetchClient<User>("/user/" + userId);
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetAllUsers(): Promise<{
    users?: User[];
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<User[]>("/users");
        return { users: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function GetMeProfile(): Promise<{
    user?: User;
    fetchError?: FetchError;
}> {
    try {
        const data = await fetchClient<User>("/user/me");
        return { user: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function UpdateProfile(user: User): Promise<{ updatedUser?: User; fetchError?: FetchError }> {
    try {
        const data = await fetchClient<User>(`/user/me`, {
            method: "PUT",
            body: JSON.stringify(user),
        });

        return { updatedUser: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function UpdateProfilePicture(
    file: File
): Promise<{ newPath?: string; fetchError?: FetchError }> {
    try {
        const formData = new FormData();
        formData.append("profile-picture", file);

        const data = await fetchClient<string>(`/user/me/profile-picture`, {
            method: "PATCH",
            body: formData
        });

        return { newPath: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function UpdateBackgroundPicture(
    file: File
): Promise<{ newPath?: string; fetchError?: FetchError }> {
    try {
        const formData = new FormData();
        formData.append("background-picture", file);

        const data = await fetchClient<string>(`/user/me/background-picture`, {
            method: "PATCH",
            body: formData,
        });

        return { newPath: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

// if there is not file, thumbnailUrl will be used
export async function UpdateLivestreamInformation(
    file: File | null,
    thumbnailUrl: string | null,
    title: string,
    description: string
): Promise<{ updatedInfo?: LivestreamInformation, fetchError?: FetchError }> {
    try {
        const formData = new FormData();
        if (file) formData.append("thumbnail", file);
        else if (thumbnailUrl) formData.append("thumbnailUrl", thumbnailUrl);

        formData.append("title", title);
        formData.append("description", description);

        const data = await fetchClient<LivestreamInformation>(`/user/me/livestream-information`, {
            method: "PATCH",
            body: formData,
        });

        return { updatedInfo: data };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function RequestToGenerateNewAPIKey(): Promise<{ newKey?: string, fetchError?: FetchError }> {
    try {
        const newKey = await fetchClient<string>("/user/me/api-key", {
            method: "PATCH"
        });
        return { newKey };
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}

export async function FollowOtherUser(followedId: string): Promise<{ fetchError?: FetchError }> {
    try {
        await fetchClient<string>(`/user/${followedId}/follow`, {
            method: "POST"
        });

        return {};
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}


export async function UnfollowOtherUser(followedId: string): Promise<{ fetchError?: FetchError }> {
    try {
        await fetchClient<string>(`/user/${followedId}/unfollow`, {
            method: "DELETE"
        });
        
        return {};
    } catch (error) {
        return { fetchError: error as FetchError };
    }
}