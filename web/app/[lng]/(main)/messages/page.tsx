"use client";

import { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import useUser from "@/hooks/user";
import ConversationList from "./_components/conversation-list";
import NewConversationDialog from "./_components/new-conversation-dialog";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import IconClose from "@/components/icons/close";
import RequireAuth from "@/components/wrappers/RequireAuth";
import { useConversationsInfinite } from "@/hooks/queries/use-conversations";

export default function MessagesPage() {
    const params = useParams();
    const router = useRouter();
    const user = useUser((state) => state.user);
    const { data, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
        useConversationsInfinite(!!user);
    const conversations = data?.pages.flat() ?? [];
    const [showNewConversation, setShowNewConversation] = useState(false);
    const { t } = useT("messages");

    return (
        <RequireAuth>
            <div className="flex h-full w-full">
                <div className="flex h-full w-full flex-col md:w-80 md:border-r">
                    <div className="flex items-center justify-between gap-2 border-b p-4">
                        <Button
                            variant="ghost"
                            size="icon"
                            onClick={() =>
                                router.push(`/${params.lng as string}`)
                            }
                            title={t("close_section")}
                            aria-label={t("close_section")}
                            className="shrink-0"
                        >
                            <IconClose className="h-4 w-4" />
                        </Button>
                        <h1 className="min-w-0 flex-1 truncate text-lg font-semibold">
                            {t("title")}
                        </h1>
                        <Button
                            variant="outline"
                            size="sm"
                            onClick={() => setShowNewConversation(true)}
                            className="shrink-0"
                        >
                            +
                        </Button>
                    </div>

                    <ConversationList
                        conversations={conversations}
                        isLoading={isLoading}
                        hasMore={!!hasNextPage}
                        isLoadingMore={isFetchingNextPage}
                        onLoadMore={() => fetchNextPage()}
                    />

                    {showNewConversation && (
                        <NewConversationDialog
                            onClose={() => setShowNewConversation(false)}
                        />
                    )}
                </div>
                <div className="hidden flex-1 items-center justify-center md:flex">
                    <p className="text-muted-foreground">
                        {t("select_conversation")}
                    </p>
                </div>
            </div>
        </RequireAuth>
    );
}
