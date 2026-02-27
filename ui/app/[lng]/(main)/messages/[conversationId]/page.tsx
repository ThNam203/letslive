"use client";

import { useEffect, useRef, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import useDmStore from "@/hooks/use-dm-store";
import useDmWebSocket from "@/hooks/use-dm-websocket";
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
import type { Conversation } from "@/types/dm";

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
    const { send } = useDmWebSocket();

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

        GetConversation(conversationId).then((res) => {
            if (res.data) {
                setConversation(res.data);
            }
        });
    }, [conversationId, user, conversations]);

    // Fetch initial messages
    useEffect(() => {
        if (!user || !conversationId) return;
        if (currentMessages.length > 0) return; // already loaded

        setIsLoadingMessages(true);
        GetDmMessages(conversationId).then((res) => {
            if (res.data) {
                setMessages(conversationId, res.data);
                setHasMore(res.data.length >= 50);
            }
            setIsLoadingMessages(false);
        });
    }, [conversationId, user, currentMessages.length, setMessages]);

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
                type: "dm:send_message",
                conversationId,
                text,
                messageType:
                    imageUrls && imageUrls.length > 0 ? "image" : "text",
                senderUsername: user.displayName ?? user.username,
                imageUrls,
            });
        },
        [user, conversationId, send],
    );

    const handleTypingStart = useCallback(() => {
        if (!user) return;
        send({
            type: "dm:typing_start",
            conversationId,
            username: user.displayName ?? user.username,
        });
    }, [user, conversationId, send]);

    const handleTypingStop = useCallback(() => {
        if (!user) return;
        send({
            type: "dm:typing_stop",
            conversationId,
            username: user.displayName ?? user.username,
        });
    }, [user, conversationId, send]);

    if (!user) {
        return (
            <div className="flex h-full w-full items-center justify-center">
                <p className="text-muted-foreground">
                    Please log in to view messages.
                </p>
            </div>
        );
    }

    return (
        <div className="flex h-full w-full">
            {/* Conversation list sidebar (hidden on mobile) */}
            <div className="hidden h-full w-80 border-r md:block">
                <div className="flex items-center border-b p-4">
                    <h1 className="text-lg font-semibold">Messages</h1>
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
                    onBack={() => router.push("./messages")}
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
