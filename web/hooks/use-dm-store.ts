import { create } from "zustand";

// Conversations, messages, and unread counts are server data and live in
// the React Query cache (see hooks/queries/use-conversations.ts,
// use-dm-messages.ts, use-dm-unread-counts.ts). This store only holds
// ephemeral, WebSocket-driven UI state that has no server-fetched form.
export type DmState = {
    // conversations on screen right now (the messages page and any expanded
    // chat dock dialogs), counted so two views of one conversation can
    // mount and unmount independently
    visibleConversationIds: Record<string, number>;
    typingUsers: Record<string, string[]>;
    onlineUsers: Set<string>;

    showConversation: (id: string) => void;
    hideConversation: (id: string) => void;

    setTypingUser: (conversationId: string, username: string) => void;
    removeTypingUser: (conversationId: string, username: string) => void;

    setUserOnline: (userId: string) => void;
    setUserOffline: (userId: string) => void;
    setOnlineUsers: (userIds: string[]) => void;
};

const useDmStore = create<DmState>((set) => ({
    visibleConversationIds: {},
    typingUsers: {},
    onlineUsers: new Set(),

    showConversation: (id) =>
        set((state) => ({
            visibleConversationIds: {
                ...state.visibleConversationIds,
                [id]: (state.visibleConversationIds[id] ?? 0) + 1,
            },
        })),
    hideConversation: (id) =>
        set((state) => {
            const next = { ...state.visibleConversationIds };
            const count = (next[id] ?? 0) - 1;
            if (count > 0) next[id] = count;
            else delete next[id];
            return { visibleConversationIds: next };
        }),

    setTypingUser: (conversationId, username) =>
        set((state) => {
            const current = state.typingUsers[conversationId] || [];
            if (current.includes(username)) return state;
            return {
                typingUsers: {
                    ...state.typingUsers,
                    [conversationId]: [...current, username],
                },
            };
        }),
    removeTypingUser: (conversationId, username) =>
        set((state) => ({
            typingUsers: {
                ...state.typingUsers,
                [conversationId]: (
                    state.typingUsers[conversationId] || []
                ).filter((u) => u !== username),
            },
        })),

    setUserOnline: (userId) =>
        set((state) => {
            const next = new Set(state.onlineUsers);
            next.add(userId);
            return { onlineUsers: next };
        }),
    setUserOffline: (userId) =>
        set((state) => {
            const next = new Set(state.onlineUsers);
            next.delete(userId);
            return { onlineUsers: next };
        }),
    setOnlineUsers: (userIds) => set({ onlineUsers: new Set(userIds) }),
}));

export default useDmStore;
