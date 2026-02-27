"use client";

import Link from "next/link";
import { Conversation, ConversationType } from "@/types/dm";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";

function getConversationDisplay(
    conversation: Conversation,
    currentUserId: string,
) {
    if (conversation.type === ConversationType.DM) {
        const other = conversation.participants.find(
            (p) => p.userId !== currentUserId,
        );
        return {
            name: other?.displayName || other?.username || "Unknown",
            avatar: other?.profilePicture || null,
            initials: (other?.username || "U").charAt(0).toUpperCase(),
        };
    }

    return {
        name: conversation.name || "Group",
        avatar: conversation.avatarUrl,
        initials: (conversation.name || "G").charAt(0).toUpperCase(),
    };
}

function formatTime(dateStr: string) {
    const date = new Date(dateStr);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const dayMs = 24 * 60 * 60 * 1000;

    if (diff < dayMs) {
        return date.toLocaleTimeString([], {
            hour: "2-digit",
            minute: "2-digit",
        });
    }
    if (diff < 7 * dayMs) {
        return date.toLocaleDateString([], { weekday: "short" });
    }
    return date.toLocaleDateString([], { month: "short", day: "numeric" });
}

export default function ConversationListItem({
    conversation,
    isActive,
}: {
    conversation: Conversation;
    isActive?: boolean;
}) {
    const user = useUser((state) => state.user);
    const { unreadCounts, onlineUsers } = useDmStore();

    if (!user) return null;

    const display = getConversationDisplay(conversation, user.id);
    const unreadCount = unreadCounts[conversation._id] || 0;

    // Check online status for DM
    let isOnline = false;
    if (conversation.type === ConversationType.DM) {
        const other = conversation.participants.find(
            (p) => p.userId !== user.id,
        );
        if (other) {
            isOnline = onlineUsers.has(other.userId);
        }
    }

    return (
        <Link
            href={`./messages/${conversation._id}`}
            className={`flex items-center gap-3 px-4 py-3 transition-colors hover:bg-accent ${
                isActive ? "bg-accent" : ""
            }`}
        >
            <div className="relative">
                <Avatar className="h-10 w-10">
                    {display.avatar && <AvatarImage src={display.avatar} />}
                    <AvatarFallback>{display.initials}</AvatarFallback>
                </Avatar>
                {isOnline && (
                    <span className="absolute right-0 bottom-0 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
                )}
            </div>
            <div className="min-w-0 flex-1">
                <div className="flex items-center justify-between">
                    <span className="truncate text-sm font-medium">
                        {display.name}
                    </span>
                    {conversation.lastMessage && (
                        <span className="text-muted-foreground ml-2 text-xs whitespace-nowrap">
                            {formatTime(conversation.lastMessage.createdAt)}
                        </span>
                    )}
                </div>
                <div className="flex items-center justify-between">
                    <p className="text-muted-foreground truncate text-xs">
                        {conversation.lastMessage
                            ? conversation.lastMessage.text
                            : "No messages yet"}
                    </p>
                    {unreadCount > 0 && (
                        <span className="ml-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-blue-500 px-1.5 text-xs text-white">
                            {unreadCount > 99 ? "99+" : unreadCount}
                        </span>
                    )}
                </div>
            </div>
        </Link>
    );
}
