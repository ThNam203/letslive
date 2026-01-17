import { ApiResponse } from "@/types/fetch-response";
import { LivestreamInformation, User } from "@/types/user";
import { fetchClient } from "@/utils/fetchClient";

export async function SearchUsersByUsername(
    query: string,
): Promise<ApiResponse<User[]>> {
    return fetchClient<ApiResponse<User[]>>(
        `/users/search?username=${encodeURIComponent(query)}`,
    );
}

export async function GetUserById(userId: string): Promise<ApiResponse<User>> {
    return fetchClient<ApiResponse<User>>(`/user/${userId}`);
}

export async function GetAllUsers(
    page: number = 0,
): Promise<ApiResponse<User[]>> {
    return fetchClient<ApiResponse<User[]>>(`/users?page=${page}`);
}

export async function GetMeProfile(): Promise<ApiResponse<User>> {
    return fetchClient<ApiResponse<User>>(`/user/me`);
}

export async function UpdateProfile(
    user: Partial<User>,
): Promise<ApiResponse<User>> {
    return fetchClient<ApiResponse<User>>(`/user/me`, {
        method: "PUT",
        body: JSON.stringify(user),
    });
}

export async function UpdateProfilePicture(
    file: File,
): Promise<ApiResponse<string>> {
    const formData = new FormData();
    formData.append("profile-picture", file);

    return fetchClient<ApiResponse<string>>(`/user/me/profile-picture`, {
        method: "PATCH",
        body: formData,
    });
}

export async function UpdateBackgroundPicture(
    file: File,
): Promise<ApiResponse<string>> {
    const formData = new FormData();
    formData.append("background-picture", file);

    return fetchClient<ApiResponse<string>>(`/user/me/background-picture`, {
        method: "PATCH",
        body: formData,
    });
}

export async function UpdateLivestreamInformation(
    file: File | null,
    thumbnailUrl: string | null,
    title: string,
    description: string,
): Promise<ApiResponse<LivestreamInformation>> {
    const formData = new FormData();

    // file will be used, if not then thumbnailUrl (BACKEND)
    if (file) formData.append("thumbnail", file);
    if (thumbnailUrl) formData.append("thumbnailUrl", thumbnailUrl);

    formData.append("title", title);
    formData.append("description", description);

    return fetchClient<ApiResponse<LivestreamInformation>>(
        `/user/me/livestream-information`,
        {
            method: "PATCH",
            body: formData,
        },
    );
}

export async function RequestToGenerateNewAPIKey(): Promise<
    ApiResponse<string>
> {
    return fetchClient<ApiResponse<string>>("/user/me/api-key", {
        method: "PATCH",
    });
}

export async function FollowOtherUser(
    followedId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/user/${followedId}/follow`, {
        method: "POST",
    });
}

export async function UnfollowOtherUser(
    followedId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/user/${followedId}/unfollow`, {
        method: "DELETE",
    });
}
