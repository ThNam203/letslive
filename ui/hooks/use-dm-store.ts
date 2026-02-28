import { create } from "zustand";
import { Conversation, DmMessage } from "@/types/dm";

export type DmState = {
    conversations: Conversation[];
    activeConversationId: string | null;
    messages: Record<string, DmMessage[]>;
    typingUsers: Record<string, string[]>;
    unreadCounts: Record<string, number>;
    onlineUsers: Set<string>;
    isLoading: boolean;

    setConversations: (conversations: Conversation[]) => void;
    addConversation: (conversation: Conversation) => void;
    updateConversation: (
        conversationId: string,
        update: Partial<Conversation>,
    ) => void;
    removeConversation: (conversationId: string) => void;
    setActiveConversationId: (id: string | null) => void;

    setMessages: (conversationId: string, messages: DmMessage[]) => void;
    prependMessages: (conversationId: string, messages: DmMessage[]) => void;
    addMessage: (conversationId: string, message: DmMessage) => void;
    updateMessage: (
        conversationId: string,
        messageId: string,
        update: Partial<DmMessage>,
    ) => void;
    removeMessage: (conversationId: string, messageId: string) => void;

    setTypingUser: (conversationId: string, username: string) => void;
    removeTypingUser: (conversationId: string, username: string) => void;

    setUnreadCounts: (counts: Record<string, number>) => void;
    incrementUnread: (conversationId: string) => void;
    clearUnread: (conversationId: string) => void;

    setUserOnline: (userId: string) => void;
    setUserOffline: (userId: string) => void;
    setOnlineUsers: (userIds: string[]) => void;

    setIsLoading: (isLoading: boolean) => void;
};

const useDmStore = create<DmState>((set) => ({
    conversations: [],
    activeConversationId: null,
    messages: {},
    typingUsers: {},
    unreadCounts: {},
    onlineUsers: new Set(),
    isLoading: false,

    setConversations: (conversations) => set({ conversations }),
    addConversation: (conversation) =>
        set((state) => ({
            conversations: [conversation, ...state.conversations],
        })),
    updateConversation: (conversationId, update) =>
        set((state) => ({
            conversations: state.conversations.map((c) =>
                c._id === conversationId ? { ...c, ...update } : c,
            ),
        })),
    removeConversation: (conversationId) =>
        set((state) => ({
            conversations: state.conversations.filter(
                (c) => c._id !== conversationId,
            ),
        })),
    setActiveConversationId: (id) => set({ activeConversationId: id }),

    setMessages: (conversationId, messages) =>
        set((state) => ({
            messages: { ...state.messages, [conversationId]: messages },
        })),
    prependMessages: (conversationId, messages) =>
        set((state) => ({
            messages: {
                ...state.messages,
                [conversationId]: [
                    ...messages,
                    ...(state.messages[conversationId] || []),
                ],
            },
        })),
    addMessage: (conversationId, message) =>
        set((state) => ({
            messages: {
                ...state.messages,
                [conversationId]: [
                    ...(state.messages[conversationId] || []),
                    message,
                ],
            },
        })),
    updateMessage: (conversationId, messageId, update) =>
        set((state) => ({
            messages: {
                ...state.messages,
                [conversationId]: (
                    state.messages[conversationId] || []
                ).map((m) => (m._id === messageId ? { ...m, ...update } : m)),
            },
        })),
    removeMessage: (conversationId, messageId) =>
        set((state) => ({
            messages: {
                ...state.messages,
                [conversationId]: (
                    state.messages[conversationId] || []
                ).map((m) =>
                    m._id === messageId
                        ? { ...m, isDeleted: true, text: "" }
                        : m,
                ),
            },
        })),

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

    setUnreadCounts: (counts) => set({ unreadCounts: counts }),
    incrementUnread: (conversationId) =>
        set((state) => ({
            unreadCounts: {
                ...state.unreadCounts,
                [conversationId]:
                    (state.unreadCounts[conversationId] || 0) + 1,
            },
        })),
    clearUnread: (conversationId) =>
        set((state) => {
            const newCounts = { ...state.unreadCounts };
            delete newCounts[conversationId];
            return { unreadCounts: newCounts };
        }),

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

    setIsLoading: (isLoading) => set({ isLoading }),
}));

export default useDmStore;
