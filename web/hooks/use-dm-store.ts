import { create } from "zustand";

// Conversations, messages, and unread counts are server data and live in
// the React Query cache (see hooks/queries/use-conversations.ts,
// use-dm-messages.ts, use-dm-unread-counts.ts). This store only holds
// ephemeral, WebSocket-driven UI state that has no server-fetched form.
export type DmState = {
    activeConversationId: string | null;
    typingUsers: Record<string, string[]>;
    onlineUsers: Set<string>;

    setActiveConversationId: (id: string | null) => void;

    setTypingUser: (conversationId: string, username: string) => void;
    removeTypingUser: (conversationId: string, username: string) => void;

    setUserOnline: (userId: string) => void;
    setUserOffline: (userId: string) => void;
    setOnlineUsers: (userIds: string[]) => void;
};

const useDmStore = create<DmState>((set) => ({
    activeConversationId: null,
    typingUsers: {},
    onlineUsers: new Set(),

    setActiveConversationId: (id) => set({ activeConversationId: id }),

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
