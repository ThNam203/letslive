"use client";

import { useEffect, useMemo, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import {
    GetConversation,
    MarkConversationRead,
    SendDmMessage,
    SendDmTyping,
} from "@/lib/api/dm";
import ConversationList from "../_components/conversation-list";
import ConversationHeader from "../_components/conversation-header";
import MessageThread from "../_components/message-thread";
import MessageInput from "../_components/message-input";
import TypingIndicator from "../_components/typing-indicator";
import { Button } from "@/components/ui/button";
import { type Conversation, DmMessageType } from "@/types/dm";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import IconClose from "@/components/icons/close";
import RequireAuth from "@/components/wrappers/RequireAuth";
import { useConversationsInfinite } from "@/hooks/queries/use-conversations";
import { useDmMessagesInfinite } from "@/hooks/queries/use-dm-messages";
import { clearDmUnread } from "@/lib/query/dm-cache";
import { flattenPages } from "@/lib/query/paginated";
import { MessagesPageSkeleton } from "../_components/messages-skeleton";

export default function ConversationPage() {
    const params = useParams();
    const router = useRouter();
    const conversationId = params.conversationId as string;

    const user = useUser((state) => state.user);
    const queryClient = useQueryClient();
    const { setActiveConversationId, typingUsers } = useDmStore();
    const { t } = useT("api-response");
    const { t: tMessages } = useT("messages");

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
    const currentMessages = useMemo(
        () => [...(messagesData?.pages ?? [])].reverse().flat(),
        [messagesData],
    );

    const [conversation, setConversation] = useState<Conversation | null>(null);
    const currentTypingUsers = typingUsers[conversationId] || [];

    // Set active conversation
    useEffect(() => {
        setActiveConversationId(conversationId);
        return () => setActiveConversationId(null);
    }, [conversationId, setActiveConversationId]);

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
    }, [conversationId, user, currentMessages.length, queryClient]);

    const loadOlderMessages = useCallback(() => {
        if (!hasNextPage) return;
        fetchNextPage();
    }, [hasNextPage, fetchNextPage]);

    const handleSendMessage = useCallback(
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

    const handleTypingStart = useCallback(() => {
        if (!user) return;

        // best effort: a lost typing hint is not worth surfacing
        SendDmTyping(conversationId, "start").catch(() => {});
    }, [user, conversationId]);

    const handleTypingStop = useCallback(() => {
        if (!user) return;

        SendDmTyping(conversationId, "stop").catch(() => {});
    }, [user, conversationId]);

    if (!user) {
        return (
            <RequireAuth fallback={<MessagesPageSkeleton withThread />}>
                {null}
            </RequireAuth>
        );
    }

    return (
        <RequireAuth fallback={<MessagesPageSkeleton withThread />}>
            <div className="flex h-full w-full">
                {/* Conversation list sidebar (hidden on mobile) */}
                <div className="hidden h-full w-80 border-r md:block">
                    <div className="flex items-center gap-2 border-b p-4">
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => router.push("/")}
                            title={tMessages("close_section")}
                            aria-label={tMessages("close_section")}
                            className="h-9 w-9 shrink-0"
                        >
                            <IconClose className="h-4 w-4" />
                        </Button>
                        <h1 className="min-w-0 flex-1 truncate text-lg font-semibold">
                            {tMessages("title")}
                        </h1>
                    </div>
                    <ConversationList
                        conversations={conversations}
                        isLoading={isLoadingConversations}
                        activeId={conversationId}
                    />
                </div>

                {/* Message thread */}
                <div className="flex h-full flex-1 flex-col">
                    <ConversationHeader
                        conversation={conversation}
                        currentUserId={user.id}
                        onBack={() => router.push("/messages")}
                        onCloseSection={() => router.push("/")}
                    />

                    <MessageThread
                        messages={currentMessages}
                        currentUserId={user.id}
                        isLoading={isLoadingMessages || isLoadingOlderMessages}
                        hasMore={!!hasNextPage}
                        onLoadMore={loadOlderMessages}
                    />

                    <TypingIndicator usernames={currentTypingUsers} />

                    <MessageInput
                        onSend={handleSendMessage}
                        onTypingStart={handleTypingStart}
                        onTypingStop={handleTypingStop}
                    />
                </div>
            </div>
        </RequireAuth>
    );
}
