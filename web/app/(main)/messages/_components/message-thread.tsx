"use client";

import { useRef, useEffect, useCallback } from "react";
import { DmMessage } from "@/types/dm";
import MessageBubble from "./message-bubble";
import { MessageThreadSkeleton } from "./messages-skeleton";
import useT from "@/hooks/use-translation";

export default function MessageThread({
    messages,
    currentUserId,
    isLoading,
    hasMore,
    onLoadMore,
}: {
    messages: DmMessage[];
    currentUserId: string;
    isLoading: boolean;
    hasMore: boolean;
    onLoadMore: () => void;
}) {
    const { t } = useT("messages");
    const containerRef = useRef<HTMLDivElement>(null);
    const bottomRef = useRef<HTMLDivElement>(null);
    const prevMessageCountRef = useRef(0);

    // Jump to the bottom on initial load, and follow new messages after that.
    // One effect, so the count it compares against is not already updated.
    useEffect(() => {
        const prevCount = prevMessageCountRef.current;
        if (prevCount === 0 && messages.length > 0) {
            bottomRef.current?.scrollIntoView();
        } else if (messages.length - prevCount === 1) {
            bottomRef.current?.scrollIntoView({ behavior: "smooth" });
        }
        prevMessageCountRef.current = messages.length;
    }, [messages.length]);

    // Infinite scroll for loading older messages
    const handleScroll = useCallback(() => {
        const container = containerRef.current;
        if (!container || isLoading || !hasMore) return;

        if (container.scrollTop < 100) {
            onLoadMore();
        }
    }, [isLoading, hasMore, onLoadMore]);

    return (
        <div
            ref={containerRef}
            onScroll={handleScroll}
            className="flex-1 overflow-y-auto px-4 py-2"
        >
            {isLoading && messages.length === 0 && <MessageThreadSkeleton />}

            {isLoading && messages.length > 0 && (
                <div className="flex justify-center py-2">
                    <div className="border-primary h-5 w-5 animate-spin rounded-full border-2 border-t-transparent" />
                </div>
            )}

            {!hasMore && messages.length > 0 && (
                <p className="text-muted-foreground py-2 text-center text-xs">
                    {t("beginning_of_conversation")}
                </p>
            )}

            <div className="space-y-1">
                {messages.map((message, idx) => {
                    const prevMessage = idx > 0 ? messages[idx - 1] : null;
                    const showSender =
                        !prevMessage ||
                        prevMessage.senderId !== message.senderId;
                    const isOwn = message.senderId === currentUserId;

                    return (
                        <MessageBubble
                            key={message._id}
                            message={message}
                            isOwn={isOwn}
                            showSender={showSender}
                        />
                    );
                })}
            </div>

            <div ref={bottomRef} />
        </div>
    );
}
