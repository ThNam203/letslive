import { InfiniteData, QueryClient } from "@tanstack/react-query";
import { Conversation, DmMessage } from "@/types/dm";
import { CONVERSATIONS_QUERY_KEY } from "@/hooks/queries/use-conversations";
import { DM_UNREAD_COUNTS_QUERY_KEY } from "@/hooks/queries/use-dm-unread-counts";
import { dmMessagesQueryKey } from "@/hooks/queries/use-dm-messages";

type ConversationsData = InfiniteData<Conversation[]>;
type MessagesData = InfiniteData<DmMessage[]>;

export function prependConversation(
    queryClient: QueryClient,
    conversation: Conversation,
) {
    queryClient.setQueryData<ConversationsData>(
        CONVERSATIONS_QUERY_KEY,
        (old) => {
            if (!old) return old;
            const [first, ...rest] = old.pages;
            return { ...old, pages: [[conversation, ...(first ?? [])], ...rest] };
        },
    );
}

export function updateConversationInCache(
    queryClient: QueryClient,
    conversationId: string,
    update: Partial<Conversation>,
) {
    queryClient.setQueryData<ConversationsData>(
        CONVERSATIONS_QUERY_KEY,
        (old) =>
            old && {
                ...old,
                pages: old.pages.map((page) =>
                    page.map((c) =>
                        c._id === conversationId ? { ...c, ...update } : c,
                    ),
                ),
            },
    );
}

export function appendDmMessage(
    queryClient: QueryClient,
    conversationId: string,
    message: DmMessage,
) {
    queryClient.setQueryData<MessagesData>(
        dmMessagesQueryKey(conversationId),
        (old) => {
            if (!old) return old;
            const [first, ...rest] = old.pages;
            return { ...old, pages: [[...(first ?? []), message], ...rest] };
        },
    );
}

export function updateDmMessageInCache(
    queryClient: QueryClient,
    conversationId: string,
    messageId: string,
    update: Partial<DmMessage>,
) {
    queryClient.setQueryData<MessagesData>(
        dmMessagesQueryKey(conversationId),
        (old) =>
            old && {
                ...old,
                pages: old.pages.map((page) =>
                    page.map((m) =>
                        m._id === messageId ? { ...m, ...update } : m,
                    ),
                ),
            },
    );
}

export function incrementDmUnread(
    queryClient: QueryClient,
    conversationId: string,
) {
    queryClient.setQueryData<Record<string, number>>(
        DM_UNREAD_COUNTS_QUERY_KEY,
        (old) => ({
            ...(old ?? {}),
            [conversationId]: (old?.[conversationId] ?? 0) + 1,
        }),
    );
}

export function clearDmUnread(
    queryClient: QueryClient,
    conversationId: string,
) {
    queryClient.setQueryData<Record<string, number>>(
        DM_UNREAD_COUNTS_QUERY_KEY,
        (old) => {
            if (!old) return old;
            const next = { ...old };
            delete next[conversationId];
            return next;
        },
    );
}
