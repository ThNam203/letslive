"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useQueryClient } from "@tanstack/react-query";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import useT from "@/hooks/use-translation";
import {
    GetConversation,
    MarkConversationRead,
    SendDmMessage,
    SendDmTyping,
} from "@/lib/api/dm";
import { type Conversation, DmMessageType } from "@/types/dm";
import { toast } from "@/components/utils/toast";
import { useConversationsInfinite } from "@/hooks/queries/use-conversations";
import { useDmMessagesInfinite } from "@/hooks/queries/use-dm-messages";
import { clearDmUnread } from "@/lib/query/dm-cache";
import { flattenPages } from "@/lib/query/paginated";

const NO_TYPING_USERS: string[] = [];

/**
 * Everything one open conversation needs: its details, messages, typing
 * users, and the send/typing actions. Shared by the messages page and the
 * chat dock dialogs. While mounted, the conversation counts as visible, so
 * new messages in it are not counted as unread and it is marked read.
 */
export default function useConversationChat(conversationId: string) {
    const user = useUser((state) => state.user);
    const queryClient = useQueryClient();
    const showConversation = useDmStore((state) => state.showConversation);
    const hideConversation = useDmStore((state) => state.hideConversation);
    const typingUsers = useDmStore(
        (state) => state.typingUsers[conversationId] ?? NO_TYPING_USERS,
    );
    const { t } = useT("api-response");

    const { data: conversationsData, isLoading: isLoadingConversations } =
        useConversationsInfinite(!!user);
    const conversations = useMemo(
        () => flattenPages(conversationsData),
        [conversationsData],
    );

    const {
        data: messagesData,
        isLoading: isLoadingMessages,
        isFetchingNextPage: isLoadingOlderMessages,
        hasNextPage,
        fetchNextPage,
    } = useDmMessagesInfinite(conversationId, !!user);
    const messages = useMemo(
        () => [...(messagesData?.pages ?? [])].reverse().flat(),
        [messagesData],
    );

    const [conversation, setConversation] = useState<Conversation | null>(null);

    useEffect(() => {
        showConversation(conversationId);
        return () => hideConversation(conversationId);
    }, [conversationId, showConversation, hideConversation]);

    // Fetch conversation details
    useEffect(() => {
        if (!user || !conversationId) return;

        const existingConv = conversations.find(
            (c) => c._id === conversationId,
        );
        if (existingConv) {
            queueMicrotask(() => setConversation(existingConv));
        }

        GetConversation(conversationId)
            .then((res) => {
                if (res.data) {
                    setConversation(res.data);
                } else if (!res.success && res.key) {
                    toast.error(t(res.key));
                }
            })
            .catch(() => {
                toast.error(t("fetch-error:client_fetch_error"));
            });
    }, [conversationId, user, conversations, t]);

    // Mark as read
    useEffect(() => {
        if (!user || !conversationId) return;
        clearDmUnread(queryClient, conversationId);
        MarkConversationRead(conversationId);
    }, [conversationId, user, messages.length, queryClient]);

    const loadOlderMessages = useCallback(() => {
        if (!hasNextPage) return;
        fetchNextPage();
    }, [hasNextPage, fetchNextPage]);

    const sendMessage = useCallback(
        (text: string, imageUrls?: string[]) => {
            if (!user) return;

            // the thread picks the message up from the dm:new_message push,
            // so only failures are handled here
            SendDmMessage(conversationId, {
                text,
                type:
                    imageUrls && imageUrls.length > 0
                        ? DmMessageType.IMAGE
                        : DmMessageType.TEXT,
                imageUrls,
            })
                .then((res) => {
                    if (!res.success) {
                        toast.error(
                            t(res.key) ||
                                res.message ||
                                "Failed to send message",
                        );
                    }
                })
                .catch(() => {
                    toast.error(t("fetch-error:client_fetch_error"));
                });
        },
        [user, conversationId, t],
    );

    const startTyping = useCallback(() => {
        if (!user) return;

        // best effort: a lost typing hint is not worth surfacing
        SendDmTyping(conversationId, "start").catch(() => {});
    }, [user, conversationId]);

    const stopTyping = useCallback(() => {
        if (!user) return;

        SendDmTyping(conversationId, "stop").catch(() => {});
    }, [user, conversationId]);

    return {
        user,
        conversation,
        conversations,
        isLoadingConversations,
        messages,
        isLoadingMessages: isLoadingMessages || isLoadingOlderMessages,
        hasMoreMessages: !!hasNextPage,
        loadOlderMessages,
        typingUsers,
        sendMessage,
        startTyping,
        stopTyping,
    };
}
