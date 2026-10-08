"use client";

import { useParams, useRouter } from "next/navigation";
import ConversationList from "../_components/conversation-list";
import ConversationHeader from "../_components/conversation-header";
import MessageThread from "../_components/message-thread";
import MessageInput from "../_components/message-input";
import TypingIndicator from "../_components/typing-indicator";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import IconClose from "@/components/icons/close";
import RequireAuth from "@/components/wrappers/RequireAuth";
import useConversationChat from "@/hooks/use-conversation-chat";
import { MessagesPageSkeleton } from "../_components/messages-skeleton";

export default function ConversationPage() {
    const params = useParams();
    const router = useRouter();
    const conversationId = params.conversationId as string;
    const { t: tMessages } = useT("messages");

    const {
        user,
        conversation,
        conversations,
        isLoadingConversations,
        messages,
        isLoadingMessages,
        hasMoreMessages,
        loadOlderMessages,
        typingUsers,
        sendMessage,
        startTyping,
        stopTyping,
    } = useConversationChat(conversationId);

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
                        messages={messages}
                        currentUserId={user.id}
                        isLoading={isLoadingMessages}
                        hasMore={hasMoreMessages}
                        onLoadMore={loadOlderMessages}
                    />

                    <TypingIndicator usernames={typingUsers} />

                    <MessageInput
                        onSend={sendMessage}
                        onTypingStart={startTyping}
                        onTypingStop={stopTyping}
                    />
                </div>
            </div>
        </RequireAuth>
    );
}
