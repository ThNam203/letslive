"use client";

import { useRouter } from "next/navigation";
import { type Conversation, ConversationType } from "@/types/dm";
import useConversationChat from "@/hooks/use-conversation-chat";
import useChatDockStore from "@/hooks/use-chat-dock-store";
import useDmStore from "@/hooks/use-dm-store";
import useT from "@/hooks/use-translation";
import UserAvatar from "@/components/ui/user-avatar";
import { Button } from "@/components/ui/button";
import IconClose from "@/components/icons/close";
import IconMinus from "@/components/icons/minus";
import IconFullscreen from "@/components/icons/fullscreen";
import MessageThread from "@/app/(main)/messages/_components/message-thread";
import MessageInput from "@/app/(main)/messages/_components/message-input";
import TypingIndicator from "@/app/(main)/messages/_components/typing-indicator";
import { getConversationDisplay } from "@/app/(main)/messages/_components/conversation-list-item";

/** One expanded chat window in the dock. */
export default function ChatDialog({
    conversationId,
    snapshot,
}: {
    conversationId: string;
    snapshot: Conversation;
}) {
    const router = useRouter();
    const { t } = useT("messages");
    const minimizeChat = useChatDockStore((state) => state.minimizeChat);
    const closeChat = useChatDockStore((state) => state.closeChat);
    const onlineUsers = useDmStore((state) => state.onlineUsers);

    const {
        user,
        conversation,
        messages,
        isLoadingMessages,
        hasMoreMessages,
        loadOlderMessages,
        typingUsers,
        sendMessage,
        startTyping,
        stopTyping,
    } = useConversationChat(conversationId);

    if (!user) return null;

    const shown = conversation ?? snapshot;
    const display = getConversationDisplay(shown, user.id, t);

    let statusText: string;
    let isOnline = false;
    if (shown.type === ConversationType.DM) {
        const other = shown.participants.find((p) => p.userId !== user.id);
        isOnline = !!other && onlineUsers.has(other.userId);
        statusText = isOnline ? t("online") : t("offline");
    } else {
        statusText = t("members_count", { count: shown.participants.length });
    }

    const openInMessages = () => {
        closeChat(conversationId);
        router.push(`/messages/${conversationId}`);
    };

    return (
        <div
            role="dialog"
            aria-label={display.name}
            className="bg-background border-border pointer-events-auto flex h-[455px] max-h-[calc(100dvh-4.5rem)] w-[328px] flex-col overflow-hidden rounded-t-lg border border-b-0 shadow-xl"
        >
            <div className="flex items-center gap-2 border-b px-2 py-1.5">
                <div className="relative shrink-0">
                    <UserAvatar
                        src={display.avatar}
                        name={display.name}
                        fallback={display.initials}
                        size="sm"
                    />
                    {isOnline && (
                        <span className="border-background absolute right-0 bottom-0 h-2.5 w-2.5 rounded-full border-2 bg-green-500" />
                    )}
                </div>
                <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium">
                        {display.name}
                    </p>
                    <p className="text-muted-foreground truncate text-xs">
                        {statusText}
                    </p>
                </div>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 shrink-0"
                    onClick={openInMessages}
                    title={t("open_in_messages")}
                    aria-label={t("open_in_messages")}
                >
                    <IconFullscreen className="h-4 w-4" />
                </Button>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 shrink-0"
                    onClick={() => minimizeChat(conversationId)}
                    title={t("minimize_chat")}
                    aria-label={t("minimize_chat")}
                >
                    <IconMinus className="h-4 w-4" />
                </Button>
                <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 shrink-0"
                    onClick={() => closeChat(conversationId)}
                    title={t("close_chat")}
                    aria-label={t("close_chat")}
                >
                    <IconClose className="h-4 w-4" />
                </Button>
            </div>

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
                compact
                autoFocus
            />
        </div>
    );
}
