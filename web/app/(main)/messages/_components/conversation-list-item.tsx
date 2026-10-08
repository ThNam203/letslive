"use client";

import Link from "next/link";
import { Conversation, ConversationType } from "@/types/dm";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import UserAvatar from "@/components/ui/user-avatar";
import useT from "@/hooks/use-translation";
import { useDmUnreadCounts } from "@/hooks/queries/use-dm-unread-counts";
import { formatLocaleDate } from "@/utils/timeFormats";

export function getConversationDisplay(
    conversation: Conversation,
    currentUserId: string,
    t: (key: string) => string,
) {
    if (conversation.type === ConversationType.DM) {
        const other = conversation.participants.find(
            (p) => p.userId !== currentUserId,
        );
        return {
            name: other?.username || t("unknown"),
            avatar: other?.profilePicture || null,
            initials: (other?.username || "U").charAt(0).toUpperCase(),
        };
    }

    return {
        name: conversation.name || t("group"),
        avatar: conversation.avatarUrl,
        initials: (conversation.name || "G").charAt(0).toUpperCase(),
    };
}

function isSameDay(a: Date, b: Date) {
    return a.toDateString() === b.toDateString();
}

function formatTime(
    dateStr: string,
    locale: string,
    t: (key: string, options?: Record<string, string>) => string,
) {
    const date = new Date(dateStr);
    const now = new Date();
    const time = formatLocaleDate(date, locale, {
        hour: "2-digit",
        minute: "2-digit",
    });

    if (isSameDay(date, now)) {
        return time;
    }

    const yesterday = new Date(now);
    yesterday.setDate(now.getDate() - 1);
    if (isSameDay(date, yesterday)) {
        return t("yesterday_at", { time });
    }

    const dayMs = 24 * 60 * 60 * 1000;
    const diff = now.getTime() - date.getTime();
    if (diff < 7 * dayMs) {
        return formatLocaleDate(date, locale, { weekday: "short" });
    }
    return formatLocaleDate(date, locale, { month: "short", day: "numeric" });
}

export default function ConversationListItem({
    conversation,
    isActive,
    onSelect,
}: {
    conversation: Conversation;
    isActive?: boolean;
    /** Called instead of navigating to the conversation page when set. */
    onSelect?: (conversation: Conversation) => void;
}) {
    const user = useUser((state) => state.user);
    const { onlineUsers } = useDmStore();
    const { data: unreadCounts = {} } = useDmUnreadCounts(!!user);
    const { t, i18n } = useT("messages");
    const lng = i18n.resolvedLanguage ?? i18n.language;

    if (!user) return null;

    const display = getConversationDisplay(conversation, user.id, t);
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

    const className = `hover:bg-accent flex w-full items-center gap-3 px-4 py-3 text-left transition-colors ${
        isActive ? "bg-accent" : ""
    }`;
    const content = (
        <>
            <div className="relative">
                <UserAvatar
                    src={display.avatar}
                    name={display.name}
                    fallback={display.initials}
                />
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
                        <span
                            className="text-muted-foreground ml-2 text-xs whitespace-nowrap"
                            title={formatLocaleDate(
                                new Date(conversation.lastMessage.createdAt),
                                lng,
                                {
                                    year: "numeric",
                                    month: "short",
                                    day: "numeric",
                                    hour: "2-digit",
                                    minute: "2-digit",
                                },
                            )}
                        >
                            {formatTime(
                                conversation.lastMessage.createdAt,
                                lng,
                                t,
                            )}
                        </span>
                    )}
                </div>
                <div className="flex items-center justify-between">
                    <p className="text-muted-foreground truncate text-xs">
                        {conversation.lastMessage
                            ? conversation.lastMessage.text
                            : t("no_messages_yet")}
                    </p>
                    {unreadCount > 0 && (
                        <span className="ml-2 flex h-5 min-w-5 items-center justify-center rounded-full bg-blue-500 px-1.5 text-xs text-white">
                            {unreadCount > 99 ? "99+" : unreadCount}
                        </span>
                    )}
                </div>
            </div>
        </>
    );

    if (onSelect) {
        return (
            <button
                type="button"
                onClick={() => onSelect(conversation)}
                className={className}
            >
                {content}
            </button>
        );
    }

    return (
        <Link href={`/messages/${conversation._id}`} className={className}>
            {content}
        </Link>
    );
}
