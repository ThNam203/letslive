"use client";

import { Conversation } from "@/types/dm";
import ConversationListItem from "./conversation-list-item";

export default function ConversationList({
    conversations,
    isLoading,
    activeId,
}: {
    conversations: Conversation[];
    isLoading: boolean;
    activeId?: string;
}) {
    if (isLoading) {
        return (
            <div className="flex flex-1 items-center justify-center p-4">
                <div className="border-primary h-6 w-6 animate-spin rounded-full border-2 border-t-transparent" />
            </div>
        );
    }

    if (conversations.length === 0) {
        return (
            <div className="text-muted-foreground flex flex-1 items-center justify-center p-4 text-sm">
                No conversations yet
            </div>
        );
    }

    return (
        <div className="flex-1 overflow-y-auto">
            {conversations.map((conv) => (
                <ConversationListItem
                    key={conv._id}
                    conversation={conv}
                    isActive={conv._id === activeId}
                />
            ))}
        </div>
    );
}
