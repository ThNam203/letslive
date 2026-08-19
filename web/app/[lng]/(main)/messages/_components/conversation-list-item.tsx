"use client";

import Link from "next/link";
import { useParams } from "next/navigation";
import { Conversation, ConversationType } from "@/types/dm";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import useT from "@/hooks/use-translation";
import { I18N_FALLBACK_LNG } from "@/lib/i18n/settings";
import { useDmUnreadCounts } from "@/hooks/queries/use-dm-unread-counts";
import { formatLocaleDate } from "@/utils/timeFormats";

function getConversationDisplay(
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
}: {
    conversation: Conversation;
    isActive?: boolean;
}) {
    const params = useParams();
    const user = useUser((state) => state.user);
    const { onlineUsers } = useDmStore();
    const { data: unreadCounts = {} } = useDmUnreadCounts(!!user);
    const { t } = useT("messages");
    const lng = (params.lng as string) ?? I18N_FALLBACK_LNG;

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

    return (
        <Link
            href={`/${lng}/messages/${conversation._id}`}
            className={`hover:bg-accent flex items-center gap-3 px-4 py-3 transition-colors ${
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
        </Link>
    );
}
