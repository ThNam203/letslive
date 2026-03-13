import { ApiResponse } from "@/types/fetch-response";
import { Notification, UnreadCountResponse } from "@/types/notification";
import { fetchClient } from "@/utils/fetchClient";

export async function GetNotifications(
    page: number = 0,
): Promise<ApiResponse<Notification[]>> {
    return fetchClient<ApiResponse<Notification[]>>(
        `/user/me/notifications?page=${page}`,
    );
}

export async function GetUnreadCount(): Promise<
    ApiResponse<UnreadCountResponse>
> {
    return fetchClient<ApiResponse<UnreadCountResponse>>(
        `/user/me/notifications/unread-count`,
    );
}

export async function MarkNotificationAsRead(
    notificationId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/user/me/notifications/${notificationId}/read`,
        { method: "PATCH" },
    );
}

export async function MarkAllNotificationsAsRead(): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(`/user/me/notifications/read-all`, {
        method: "PATCH",
    });
}

export async function DeleteNotification(
    notificationId: string,
): Promise<ApiResponse<void>> {
    return fetchClient<ApiResponse<void>>(
        `/user/me/notifications/${notificationId}`,
        { method: "DELETE" },
    );
}
