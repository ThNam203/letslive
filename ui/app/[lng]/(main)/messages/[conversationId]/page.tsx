"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import useDmStore from "@/hooks/use-dm-store";
import { useDmWebSocketContext } from "@/contexts/dm-websocket-context";
import useUser from "@/hooks/user";
import {
    GetConversation,
    GetDmMessages,
    MarkConversationRead,
} from "@/lib/api/dm";
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

export default function ConversationPage() {
    const params = useParams();
    const router = useRouter();
    const conversationId = params.conversationId as string;

    const user = useUser((state) => state.user);
    const {
        conversations,
        messages,
        setMessages,
        prependMessages,
        setActiveConversationId,
        clearUnread,
        typingUsers,
    } = useDmStore();
    const { send } = useDmWebSocketContext();
    const { t } = useT("api-response");
    const { t: tMessages } = useT("messages");

    const [conversation, setConversation] = useState<Conversation | null>(null);
    const [isLoadingMessages, setIsLoadingMessages] = useState(false);
    const [hasMore, setHasMore] = useState(true);

    const currentMessages = messages[conversationId] || [];
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
            setConversation(existingConv);
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

    // Fetch initial messages
    useEffect(() => {
        if (!user || !conversationId) return;
        if (currentMessages.length > 0) return; // already loaded

        setIsLoadingMessages(true);
        GetDmMessages(conversationId)
            .then((res) => {
                if (res.data) {
                    setMessages(conversationId, res.data);
                    setHasMore(res.data.length >= 50);
                } else if (!res.success && res.key) {
                    toast.error(t(res.key));
                }
                setIsLoadingMessages(false);
            })
            .catch(() => {
                toast.error(t("fetch-error:client_fetch_error"));
                setIsLoadingMessages(false);
            });
    }, [conversationId, user, currentMessages.length, setMessages, t]);

    // Mark as read
    useEffect(() => {
        if (!user || !conversationId) return;
        clearUnread(conversationId);
        MarkConversationRead(conversationId);
    }, [conversationId, user, currentMessages.length, clearUnread]);

    // Load older messages
    const loadOlderMessages = useCallback(async () => {
        if (!hasMore || isLoadingMessages || currentMessages.length === 0)
            return;

        setIsLoadingMessages(true);
        const oldestMessageId = currentMessages[0]?._id;
        const res = await GetDmMessages(conversationId, oldestMessageId);
        if (res.data) {
            if (res.data.length === 0) {
                setHasMore(false);
            } else {
                prependMessages(conversationId, res.data);
                setHasMore(res.data.length >= 50);
            }
        }
        setIsLoadingMessages(false);
    }, [
        hasMore,
        isLoadingMessages,
        currentMessages,
        conversationId,
        prependMessages,
    ]);

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
                senderUsername: user.displayName ?? user.username,
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
            username: user.displayName ?? user.username,
        });
    }, [user, conversationId, send]);

    const handleTypingStop = useCallback(() => {
        if (!user) return;
        send({
            type: DmClientEventType.TYPING_STOP,
            conversationId,
            username: user.displayName ?? user.username,
        });
    }, [user, conversationId, send]);

    if (!user) {
        return (
            <div className="flex h-full w-full items-center justify-center">
                <p className="text-muted-foreground">
                    {tMessages("login_required")}
                </p>
            </div>
        );
    }

    return (
        <div className="flex h-full w-full">
            {/* Conversation list sidebar (hidden on mobile) */}
            <div className="hidden h-full w-80 border-r md:block">
                <div className="flex items-center gap-2 border-b p-4">
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => router.push(`/${params.lng as string}`)}
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
                    isLoading={isLoadingMessages}
                    hasMore={hasMore}
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
    );
}
