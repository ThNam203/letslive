import { http } from "msw";
import { API_BASE, ok, notFound, noContent } from "../utils";
import { meUser, otherUsers, uid, now, notifications, likedCommentIds } from "../db";
import { MeUser, PublicUser } from "@/types/user";
import { Notification, UnreadCountResponse } from "@/types/notification";

// Combined list for look-ups
const getAllUsers = (): (PublicUser | MeUser)[] => [meUser, ...otherUsers];

export const userHandlers = [
    // GET /user/me
    http.get(`${API_BASE}/user/me`, () => {
        return ok<MeUser>(meUser);
    }),

    // PUT /user/me
    http.put(`${API_BASE}/user/me`, async ({ request }) => {
        const body = (await request.json()) as Partial<MeUser>;
        Object.assign(meUser, body);
        return ok<MeUser>(meUser);
    }),

    // PATCH /user/me/profile-picture
    http.patch(`${API_BASE}/user/me/profile-picture`, async () => {
        const fakeUrl = `https://api.dicebear.com/9.x/avataaars/svg?seed=${Date.now()}`;
        meUser.profilePicture = fakeUrl;
        return ok<string>(fakeUrl);
    }),

    // PATCH /user/me/background-picture
    http.patch(`${API_BASE}/user/me/background-picture`, async () => {
        const fakeUrl = `https://images.unsplash.com/photo-1506905925346-21bda4d32df4?w=1200&q=80&t=${Date.now()}`;
        meUser.backgroundPicture = fakeUrl;
        return ok<string>(fakeUrl);
    }),

    // PATCH /user/me/livestream-information
    http.patch(`${API_BASE}/user/me/livestream-information`, async ({ request }) => {
        const formData = await request.formData();
        const title = formData.get("title") as string | null;
        const description = formData.get("description") as string | null;
        const thumbnailUrl = formData.get("thumbnailUrl") as string | null;
        if (title !== null) meUser.livestreamInformation.title = title;
        if (description !== null) meUser.livestreamInformation.description = description;
        if (thumbnailUrl !== null) meUser.livestreamInformation.thumbnailUrl = thumbnailUrl;
        return ok(meUser.livestreamInformation);
    }),

    // PATCH /user/me/api-key
    http.patch(`${API_BASE}/user/me/api-key`, () => {
        const newKey = `mock-stream-key-${uid()}`;
        meUser.streamAPIKey = newKey;
        return ok<string>(newKey);
    }),

    // GET /user/:userId
    http.get(`${API_BASE}/user/:userId`, ({ params }) => {
        const { userId } = params as { userId: string };
        if (userId === meUser.id || userId === "me") {
            return ok<MeUser>(meUser);
        }
        const user = otherUsers.find((u) => u.id === userId);
        if (!user) return notFound("res_err_user_not_found", "User not found");
        return ok<PublicUser>(user);
    }),

    // POST /user/:followedId/follow
    http.post(`${API_BASE}/user/:followedId/follow`, ({ params }) => {
        const { followedId } = params as { followedId: string };
        const user = otherUsers.find((u) => u.id === followedId);
        if (user) {
            user.isFollowing = true;
            user.followerCount += 1;
        }
        return noContent();
    }),

    // DELETE /user/:followedId/unfollow
    http.delete(`${API_BASE}/user/:followedId/unfollow`, ({ params }) => {
        const { followedId } = params as { followedId: string };
        const user = otherUsers.find((u) => u.id === followedId);
        if (user) {
            user.isFollowing = false;
            user.followerCount = Math.max(0, user.followerCount - 1);
        }
        return noContent();
    }),

    // GET /users/search?username=
    http.get(`${API_BASE}/users/search`, ({ request }) => {
        const url = new URL(request.url);
        const query = url.searchParams.get("username")?.toLowerCase() ?? "";
        const results = getAllUsers().filter(
            (u) =>
                u.username.toLowerCase().includes(query) ||
                u.displayName?.toLowerCase().includes(query),
        ) as PublicUser[];
        return ok<PublicUser[]>(results);
    }),

    // GET /users/recommendations
    http.get(`${API_BASE}/users/recommendations`, () => {
        return ok<PublicUser[]>(otherUsers, { page: 0, page_size: 10, total: otherUsers.length });
    }),

    // GET /user/me/following
    http.get(`${API_BASE}/user/me/following`, () => {
        const following = otherUsers.filter((u) => u.isFollowing);
        return ok<PublicUser[]>(following);
    }),

    // -------------------------------------------------------------------------
    // Notifications
    // -------------------------------------------------------------------------

    // GET /user/me/notifications
    http.get(`${API_BASE}/user/me/notifications`, () => {
        return ok<Notification[]>(
            [...notifications].sort(
                (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
            ),
        );
    }),

    // GET /user/me/notifications/unread-count
    http.get(`${API_BASE}/user/me/notifications/unread-count`, () => {
        const count = notifications.filter((n) => !n.isRead).length;
        return ok<UnreadCountResponse>({ count });
    }),

    // PATCH /user/me/notifications/:id/read
    http.patch(`${API_BASE}/user/me/notifications/:notificationId/read`, ({ params }) => {
        const { notificationId } = params as { notificationId: string };
        const notif = notifications.find((n) => n.id === notificationId);
        if (notif) notif.isRead = true;
        return noContent();
    }),

    // PATCH /user/me/notifications/read-all
    http.patch(`${API_BASE}/user/me/notifications/read-all`, () => {
        notifications.forEach((n) => (n.isRead = true));
        return noContent();
    }),

    // DELETE /user/me/notifications/:id
    http.delete(`${API_BASE}/user/me/notifications/:notificationId`, ({ params }) => {
        const { notificationId } = params as { notificationId: string };
        const idx = notifications.findIndex((n) => n.id === notificationId);
        if (idx !== -1) notifications.splice(idx, 1);
        return noContent();
    }),
];
