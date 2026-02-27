"use client";

import { Conversation } from "@/types/dm";
import useDmStore from "@/hooks/use-dm-store";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import IconClose from "@/components/icons/close";

export default function ConversationHeader({
    conversation,
    currentUserId,
    onBack,
}: {
    conversation: Conversation | null;
    currentUserId: string;
    onBack: () => void;
}) {
    const { onlineUsers } = useDmStore();

    if (!conversation) {
        return <div className="border-b p-4" />;
    }

    let name: string;
    let avatar: string | null;
    let initials: string;
    let isOnline = false;
    let memberCount: number | null = null;

    if (conversation.type === "dm") {
        const other = conversation.participants.find(
            (p) => p.userId !== currentUserId,
        );
        name = other?.displayName || other?.username || "Unknown";
        avatar = other?.profilePicture || null;
        initials = (other?.username || "U").charAt(0).toUpperCase();
        if (other) {
            isOnline = onlineUsers.has(other.userId);
        }
    } else {
        name = conversation.name || "Group";
        avatar = conversation.avatarUrl;
        initials = (conversation.name || "G").charAt(0).toUpperCase();
        memberCount = conversation.participants.length;
    }

    return (
        <div className="flex items-center gap-3 border-b px-4 py-3">
            <Button
                variant="ghost"
                size="icon"
                onClick={onBack}
                className="md:hidden"
            >
                <IconClose className="h-4 w-4" />
            </Button>

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
                <p className="text-muted-foreground text-xs">
                    {conversation.type === "dm"
                        ? isOnline
                            ? "Online"
                            : "Offline"
                        : `${memberCount} members`}
                </p>
            </div>
        </div>
    );
}
