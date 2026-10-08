"use client";

import { Conversation } from "@/types/dm";
import ConversationListItem from "./conversation-list-item";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import IconLoader from "@/components/icons/loader";
import { ConversationListSkeleton } from "./messages-skeleton";

export default function ConversationList({
    conversations,
    isLoading,
    activeId,
    hasMore,
    isLoadingMore,
    onLoadMore,
    onSelect,
}: {
    conversations: Conversation[];
    isLoading: boolean;
    activeId?: string;
    hasMore?: boolean;
    isLoadingMore?: boolean;
    onLoadMore?: () => void;
    onSelect?: (conversation: Conversation) => void;
}) {
    const { t } = useT("messages");

    if (isLoading) {
        return <ConversationListSkeleton />;
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
                        onSelect={onSelect}
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
                            <IconLoader className="size-4" />
                        ) : (
                            t("load_more")
                        )}
                    </Button>
                </div>
            )}
        </div>
    );
}
