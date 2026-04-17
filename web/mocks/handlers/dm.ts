import { http } from "msw";
import { API_BASE, ok, notFound, noContent, created, badRequest } from "../utils";
import {
    conversations,
    dmMessages,
    ME_USER_ID,
    uid,
    now,
    meUser,
    otherUsers,
} from "../db";
import {
    Conversation,
    ConversationType,
    DmMessage,
    DmMessageType,
    ParticipantRole,
} from "@/types/dm";

export const dmHandlers = [
    // GET /conversations
    http.get(`${API_BASE}/conversations`, ({ request }) => {
        const url = new URL(request.url);
        const page = parseInt(url.searchParams.get("page") ?? "0");
        const limit = parseInt(url.searchParams.get("limit") ?? "20");

        // Only return conversations the current user is in
        const mine = conversations.filter((c) =>
            c.participants.some((p) => p.userId === ME_USER_ID),
        );
        const slice = mine.slice(page * limit, page * limit + limit);
        return ok<Conversation[]>(slice, {
            page,
            page_size: limit,
            total: mine.length,
        });
    }),

    // GET /conversations/unread-counts — must come before /:id
    http.get(`${API_BASE}/conversations/unread-counts`, () => {
        const counts: Record<string, number> = {};
        conversations.forEach((conv) => {
            const me = conv.participants.find((p) => p.userId === ME_USER_ID);
            if (!me) return;
            const msgs = dmMessages[conv._id] ?? [];
            const unread = me.lastReadMessageId
                ? msgs.filter(
                      (m) =>
                          m.createdAt >
                          (msgs.find((x) => x._id === me.lastReadMessageId)
                              ?.createdAt ?? ""),
                  ).length
                : msgs.length;
            counts[conv._id] = unread;
        });
        return ok<Record<string, number>>(counts);
    }),

    // GET /conversations/:conversationId
    http.get(`${API_BASE}/conversations/:conversationId`, ({ params }) => {
        const { conversationId } = params as { conversationId: string };
        const conv = conversations.find((c) => c._id === conversationId);
        if (!conv) return notFound("res_err_conversation_not_found", "Conversation not found");
        return ok<Conversation>(conv);
    }),

    // POST /conversations
    http.post(`${API_BASE}/conversations`, async ({ request }) => {
        const body = (await request.json()) as any;
        const participantIds: string[] = body.participantIds ?? [];

        // For DMs, check no duplicate
        if (body.type === ConversationType.DM) {
            const otherId = participantIds.find((id) => id !== ME_USER_ID);
            const existing = conversations.find(
                (c) =>
                    c.type === ConversationType.DM &&
                    c.participants.some((p) => p.userId === ME_USER_ID) &&
                    c.participants.some((p) => p.userId === otherId),
            );
            if (existing) {
                return badRequest("res_err_dm_already_exists", "DM already exists");
            }
        }

        const allParticipantIds = [ME_USER_ID, ...participantIds.filter((id) => id !== ME_USER_ID)];

        const newConv: Conversation = {
            _id: `conv-${uid()}`,
            type: body.type ?? ConversationType.DM,
            name: body.name ?? null,
            avatarUrl: body.avatarUrl ?? null,
            createdBy: ME_USER_ID,
            participants: allParticipantIds.map((userId, i) => {
                const user =
                    userId === ME_USER_ID
                        ? meUser
                        : otherUsers.find((u) => u.id === userId);
                return {
                    userId,
                    username: user?.username ?? userId,
                    displayName: user?.displayName ?? null,
                    profilePicture: user?.profilePicture ?? null,
                    role: i === 0 ? ParticipantRole.OWNER : ParticipantRole.MEMBER,
                    joinedAt: now(),
                    lastReadMessageId: null,
                    isMuted: false,
                };
            }),
            lastMessage: null,
            createdAt: now(),
            updatedAt: now(),
        };
        conversations.push(newConv);
        dmMessages[newConv._id] = [];
        return created<Conversation>(newConv);
    }),

    // PUT /conversations/:conversationId
    http.put(`${API_BASE}/conversations/:conversationId`, async ({ params, request }) => {
        const { conversationId } = params as { conversationId: string };
        const conv = conversations.find((c) => c._id === conversationId);
        if (!conv) return notFound("res_err_conversation_not_found", "Conversation not found");
        const body = (await request.json()) as { name?: string; avatarUrl?: string };
        if (body.name !== undefined) conv.name = body.name;
        if (body.avatarUrl !== undefined) conv.avatarUrl = body.avatarUrl;
        conv.updatedAt = now();
        return ok<Conversation>(conv);
    }),

    // DELETE /conversations/:conversationId — leave conversation
    http.delete(`${API_BASE}/conversations/:conversationId`, ({ params }) => {
        const { conversationId } = params as { conversationId: string };
        const idx = conversations.findIndex((c) => c._id === conversationId);
        if (idx !== -1) conversations.splice(idx, 1);
        return noContent();
    }),

    // POST /conversations/:conversationId/participants
    http.post(
        `${API_BASE}/conversations/:conversationId/participants`,
        async ({ params, request }) => {
            const { conversationId } = params as { conversationId: string };
            const conv = conversations.find((c) => c._id === conversationId);
            if (!conv) return notFound("res_err_conversation_not_found", "Conversation not found");
            const body = (await request.json()) as any;
            const user = otherUsers.find((u) => u.id === body.userId);
            conv.participants.push({
                userId: body.userId,
                username: body.username ?? user?.username ?? body.userId,
                displayName: body.displayName ?? user?.displayName ?? null,
                profilePicture: body.profilePicture ?? user?.profilePicture ?? null,
                role: ParticipantRole.MEMBER,
                joinedAt: now(),
                lastReadMessageId: null,
                isMuted: false,
            });
            conv.updatedAt = now();
            return ok<Conversation>(conv);
        },
    ),

    // DELETE /conversations/:conversationId/participants/:userId
    http.delete(
        `${API_BASE}/conversations/:conversationId/participants/:userId`,
        ({ params }) => {
            const { conversationId, userId } = params as {
                conversationId: string;
                userId: string;
            };
            const conv = conversations.find((c) => c._id === conversationId);
            if (!conv) return notFound("res_err_conversation_not_found", "Conversation not found");
            conv.participants = conv.participants.filter((p) => p.userId !== userId);
            conv.updatedAt = now();
            return ok<Conversation>(conv);
        },
    ),

    // GET /conversations/:conversationId/messages
    http.get(`${API_BASE}/conversations/:conversationId/messages`, ({ params, request }) => {
        const { conversationId } = params as { conversationId: string };
        const url = new URL(request.url);
        const limit = parseInt(url.searchParams.get("limit") ?? "50");
        const before = url.searchParams.get("before");

        let msgs = dmMessages[conversationId] ?? [];
        if (before) {
            const idx = msgs.findIndex((m) => m._id === before);
            if (idx !== -1) msgs = msgs.slice(0, idx);
        }
        const slice = msgs.slice(-limit);
        return ok<DmMessage[]>(slice);
    }),

    // POST /conversations/:conversationId/messages
    http.post(
        `${API_BASE}/conversations/:conversationId/messages`,
        async ({ params, request }) => {
            const { conversationId } = params as { conversationId: string };
            const body = (await request.json()) as any;
            const newMsg: DmMessage = {
                _id: uid(),
                conversationId,
                senderId: ME_USER_ID,
                senderUsername: meUser.username,
                type: body.type ?? DmMessageType.TEXT,
                text: body.text ?? "",
                imageUrls: body.imageUrls,
                replyTo: body.replyTo,
                isDeleted: false,
                readBy: [],
                createdAt: now(),
                updatedAt: now(),
            };
            if (!dmMessages[conversationId]) dmMessages[conversationId] = [];
            dmMessages[conversationId].push(newMsg);

            // Update conversation lastMessage
            const conv = conversations.find((c) => c._id === conversationId);
            if (conv) {
                conv.lastMessage = {
                    _id: newMsg._id,
                    senderId: newMsg.senderId,
                    senderUsername: newMsg.senderUsername,
                    text: newMsg.text,
                    createdAt: newMsg.createdAt,
                };
                conv.updatedAt = now();
            }
            return created<DmMessage>(newMsg);
        },
    ),

    // PATCH /conversations/:conversationId/messages/:messageId
    http.patch(
        `${API_BASE}/conversations/:conversationId/messages/:messageId`,
        async ({ params, request }) => {
            const { conversationId, messageId } = params as {
                conversationId: string;
                messageId: string;
            };
            const msgs = dmMessages[conversationId] ?? [];
            const msg = msgs.find((m) => m._id === messageId);
            if (!msg) return notFound("res_err_dm_message_not_found", "Message not found");
            const body = (await request.json()) as { text: string };
            msg.text = body.text;
            msg.updatedAt = now();
            return ok<DmMessage>(msg);
        },
    ),

    // DELETE /conversations/:conversationId/messages/:messageId
    http.delete(
        `${API_BASE}/conversations/:conversationId/messages/:messageId`,
        ({ params }) => {
            const { conversationId, messageId } = params as {
                conversationId: string;
                messageId: string;
            };
            const msgs = dmMessages[conversationId] ?? [];
            const msg = msgs.find((m) => m._id === messageId);
            if (msg) msg.isDeleted = true;
            return noContent();
        },
    ),

    // POST /conversations/:conversationId/read
    http.post(`${API_BASE}/conversations/:conversationId/read`, ({ params }) => {
        const { conversationId } = params as { conversationId: string };
        const conv = conversations.find((c) => c._id === conversationId);
        if (conv) {
            const me = conv.participants.find((p) => p.userId === ME_USER_ID);
            const msgs = dmMessages[conversationId] ?? [];
            if (me && msgs.length > 0) {
                me.lastReadMessageId = msgs[msgs.length - 1]._id;
            }
        }
        return noContent();
    }),
];
