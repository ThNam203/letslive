"use client";

import { useEffect, useMemo, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import { useQueryClient } from "@tanstack/react-query";
import useDmStore from "@/hooks/use-dm-store";
import { useDmWebSocketContext } from "@/contexts/dm-websocket-context";
import useUser from "@/hooks/user";
import { GetConversation, MarkConversationRead } from "@/lib/api/dm";
import ConversationList from "../_components/conversation-list";
import ConversationHeader from "../_components/conversation-header";
import MessageThread from "../_components/message-thread";
import MessageInput from "../_components/message-input";
import TypingIndicator from "../_components/typing-indicator";
import { Button } from "@/components/ui/button";
import {
    type Conversation,
    DmClientEventType,
    DmMessageType,
} from "@/types/dm";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import IconClose from "@/components/icons/close";
import RequireAuth from "@/components/wrappers/RequireAuth";
import { useConversationsInfinite } from "@/hooks/queries/use-conversations";
import { useDmMessagesInfinite } from "@/hooks/queries/use-dm-messages";
import { clearDmUnread } from "@/lib/query/dm-cache";

export default function ConversationPage() {
    const params = useParams();
    const router = useRouter();
    const conversationId = params.conversationId as string;

    const user = useUser((state) => state.user);
    const queryClient = useQueryClient();
    const { setActiveConversationId, typingUsers } = useDmStore();
    const { send } = useDmWebSocketContext();
    const { t } = useT("api-response");
    const { t: tMessages } = useT("messages");

    const { data: conversationsData } = useConversationsInfinite(!!user);
    const conversations = useMemo(
        () => conversationsData?.pages.flat() ?? [],
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

            send({
                type: DmClientEventType.SEND_MESSAGE,
                conversationId,
                text,
                messageType:
                    imageUrls && imageUrls.length > 0
                        ? DmMessageType.IMAGE
                        : DmMessageType.TEXT,
                senderUsername: user.username,
                imageUrls,
            });
        },
        [user, conversationId, send],
    );

    const handleTypingStart = useCallback(() => {
        if (!user) return;

        send({
            type: DmClientEventType.TYPING_START,
            conversationId,
            username: user.username,
        });
    }, [user, conversationId, send]);

    const handleTypingStop = useCallback(() => {
        if (!user) return;

        send({
            type: DmClientEventType.TYPING_STOP,
            conversationId,
            username: user.username,
        });
    }, [user, conversationId, send]);

    if (!user) {
        return <RequireAuth>{null}</RequireAuth>;
    }

    return (
        <RequireAuth>
            <div className="flex h-full w-full">
                {/* Conversation list sidebar (hidden on mobile) */}
                <div className="hidden h-full w-80 border-r md:block">
                    <div className="flex items-center gap-2 border-b p-4">
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() =>
                                router.push(`/${params.lng as string}`)
                            }
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
                        isLoading={false}
                        activeId={conversationId}
                    />
                </div>

                {/* Message thread */}
                <div className="flex h-full flex-1 flex-col">
                    <ConversationHeader
                        conversation={conversation}
                        currentUserId={user.id}
                        onBack={() =>
                            router.push(`/${params.lng as string}/messages`)
                        }
                        onCloseSection={() =>
                            router.push(`/${params.lng as string}`)
                        }
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
