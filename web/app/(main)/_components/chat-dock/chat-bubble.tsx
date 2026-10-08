"use client";

import { type Conversation, ConversationType } from "@/types/dm";
import useChatDockStore from "@/hooks/use-chat-dock-store";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import useT from "@/hooks/use-translation";
import UserAvatar from "@/components/ui/user-avatar";
import IconClose from "@/components/icons/close";
import {
    Tooltip,
    TooltipContent,
    TooltipTrigger,
} from "@/components/ui/tooltip";
import { getConversationDisplay } from "@/app/(main)/messages/_components/conversation-list-item";

/** A minimized chat: an avatar that reopens the dialog when clicked. */
export default function ChatBubble({
    conversation,
    unreadCount,
}: {
    conversation: Conversation;
    unreadCount: number;
}) {
    const user = useUser((state) => state.user);
    const onlineUsers = useDmStore((state) => state.onlineUsers);
    const openChat = useChatDockStore((state) => state.openChat);
    const closeChat = useChatDockStore((state) => state.closeChat);
    const { t } = useT("messages");

    if (!user) return null;

    const display = getConversationDisplay(conversation, user.id, t);
    let isOnline = false;
    if (conversation.type === ConversationType.DM) {
        const other = conversation.participants.find(
            (p) => p.userId !== user.id,
        );
        isOnline = !!other && onlineUsers.has(other.userId);
    }

    return (
        <div className="group relative">
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        onClick={() => openChat(conversation)}
                        aria-label={display.name}
                        className="relative block cursor-pointer rounded-full shadow-lg transition-transform hover:scale-105"
                    >
                        <UserAvatar
                            src={display.avatar}
                            name={display.name}
                            fallback={display.initials}
                            className="h-12 w-12"
                        />
                        {isOnline && (
                            <span className="border-background absolute right-0.5 bottom-0.5 h-3 w-3 rounded-full border-2 bg-green-500" />
                        )}
                        {unreadCount > 0 && (
                            <span className="bg-destructive absolute -top-1 -right-1 flex h-5 min-w-5 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white">
                                {unreadCount > 99 ? "99+" : unreadCount}
                            </span>
                        )}
                    </button>
                </TooltipTrigger>
                <TooltipContent side="left">{display.name}</TooltipContent>
            </Tooltip>
            <button
                type="button"
                onClick={() => closeChat(conversation._id)}
                title={t("close_chat")}
                aria-label={t("close_chat")}
                className="bg-background border-border absolute -top-1 -left-1 hidden h-5 w-5 cursor-pointer items-center justify-center rounded-full border shadow group-focus-within:flex group-hover:flex"
            >
                <IconClose className="size-3" />
            </button>
        </div>
    );
}
