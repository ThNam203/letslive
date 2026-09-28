"use client";

import { Conversation } from "@/types/dm";
import ConversationListItem from "./conversation-list-item";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";

export default function ConversationList({
    conversations,
    isLoading,
    activeId,
    hasMore,
    isLoadingMore,
    onLoadMore,
}: {
    conversations: Conversation[];
    isLoading: boolean;
    activeId?: string;
    hasMore?: boolean;
    isLoadingMore?: boolean;
    onLoadMore?: () => void;
}) {
    const { t } = useT("messages");

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
                {t("no_conversations_yet")}
            </div>
        );
    }

    return (
        <div className="flex flex-1 flex-col overflow-hidden">
            <div className="flex-1 overflow-y-auto">
                {conversations.map((conv) => (
                    <ConversationListItem
                        key={conv._id}
                        conversation={conv}
                        isActive={conv._id === activeId}
                    />
                ))}
            </div>
            {hasMore && onLoadMore && (
                <div className="border-t p-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        className="w-full"
                        onClick={onLoadMore}
                        disabled={isLoadingMore}
                    >
                        {isLoadingMore ? (
                            <span className="size-4 animate-spin rounded-full border-2 border-current border-t-transparent" />
                        ) : (
                            t("load_more")
                        )}
                    </Button>
                </div>
            )}
        </div>
    );
}
