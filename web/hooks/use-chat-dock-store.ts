import { create } from "zustand";
import type { Conversation } from "@/types/dm";

// Expanded dialogs past this are shown as bubbles even on wide screens.
export const CHAT_DOCK_MAX_EXPANDED = 3;
// Oldest docked chats are dropped once this many are open.
const CHAT_DOCK_MAX_CHATS = 8;

export type DockedChat = {
    id: string;
    // what the bubble shows when the conversation is not in the cached list
    snapshot: Conversation;
    minimized: boolean;
};

type ChatDockState = {
    // ordered least to most recently opened
    chats: DockedChat[];

    openChat: (conversation: Conversation) => void;
    // opens the chat only if it is not docked yet, so a new message does not
    // undo the user's minimize
    popUpChat: (conversation: Conversation) => void;
    minimizeChat: (id: string) => void;
    closeChat: (id: string) => void;
    closeAll: () => void;
};

/**
 * The Facebook-style chat dock at the bottom right. Which chats end up
 * expanded also depends on the screen width (see ChatDock), so only the
 * user's own choices live here.
 */
const useChatDockStore = create<ChatDockState>((set, get) => ({
    chats: [],

    openChat: (conversation) =>
        set((state) => ({
            chats: [
                ...state.chats.filter((c) => c.id !== conversation._id),
                {
                    id: conversation._id,
                    snapshot: conversation,
                    minimized: false,
                },
            ].slice(-CHAT_DOCK_MAX_CHATS),
        })),
    popUpChat: (conversation) => {
        if (get().chats.some((c) => c.id === conversation._id)) return;
        get().openChat(conversation);
    },
    minimizeChat: (id) =>
        set((state) => ({
            chats: state.chats.map((c) =>
                c.id === id ? { ...c, minimized: true } : c,
            ),
        })),
    closeChat: (id) =>
        set((state) => ({ chats: state.chats.filter((c) => c.id !== id) })),
    closeAll: () => set({ chats: [] }),
}));

export default useChatDockStore;
