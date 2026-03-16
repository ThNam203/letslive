"use client";

import { Conversation, ConversationType } from "@/types/dm";
import useDmStore from "@/hooks/use-dm-store";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import IconClose from "@/components/icons/close";
import useT from "@/hooks/use-translation";

export default function ConversationHeader({
    conversation,
    currentUserId,
    onBack,
    onCloseSection,
}: {
    conversation: Conversation | null;
    currentUserId: string;
    onBack: () => void;
    onCloseSection?: () => void;
}) {
    const { onlineUsers } = useDmStore();
    const { t } = useT("messages");

    if (!conversation) {
        return <div className="border-b p-4" />;
    }

    let name: string;
    let avatar: string | null;
    let initials: string;
    let isOnline = false;
    let memberCount: number | null = null;

    if (conversation.type === ConversationType.DM) {
        const other = conversation.participants.find(
            (p) => p.userId !== currentUserId,
        );
        name = other?.displayName || other?.username || t("unknown");
        avatar = other?.profilePicture || null;
        initials = (other?.username || "U").charAt(0).toUpperCase();
        if (other) {
            isOnline = onlineUsers.has(other.userId);
        }
    } else {
        name = conversation.name || t("group");
        avatar = conversation.avatarUrl;
        initials = (conversation.name || "G").charAt(0).toUpperCase();
        memberCount = conversation.participants.length;
    }
    const statusText =
        conversation.type === ConversationType.DM
            ? isOnline
                ? t("online")
                : t("offline")
            : t("members_count", { count: memberCount ?? 0 });

    return (
        <div className="flex items-center gap-3 border-b px-4 py-3">
            <Button
                variant="ghost"
                size="icon"
                onClick={onBack}
                className="md:hidden"
                aria-label={t("back_to_list")}
            >
                <IconClose className="h-4 w-4" />
            </Button>

            {onCloseSection && (
                <Button
                    variant="ghost"
                    size="icon"
                    onClick={onCloseSection}
                    className="hidden md:flex"
                    title={t("close_section")}
                    aria-label={t("close_section")}
                >
                    <IconClose className="h-4 w-4" />
                </Button>
            )}

            <div className="relative">
                <Avatar className="h-9 w-9">
                    {avatar && <AvatarImage src={avatar} />}
                    <AvatarFallback>{initials}</AvatarFallback>
                </Avatar>
                {isOnline && (
                    <span className="absolute right-0 bottom-0 h-2.5 w-2.5 rounded-full border-2 border-white bg-green-500" />
                )}
            </div>

            <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{name}</p>
                <p className="text-muted-foreground text-xs">{statusText}</p>
            </div>
        </div>
    );
}
