"use client";

import { useEffect, useState, useCallback } from "react";
import { useParams, useRouter } from "next/navigation";
import useDmStore from "@/hooks/use-dm-store";
import useUser from "@/hooks/user";
import { GetConversations, GetUnreadCounts } from "@/lib/api/dm";
import ConversationList from "./_components/conversation-list";
import NewConversationDialog from "./_components/new-conversation-dialog";
import { Button } from "@/components/ui/button";
import useT from "@/hooks/use-translation";
import { toast } from "@/components/utils/toast";
import IconClose from "@/components/icons/close";

const CONVERSATIONS_PAGE_SIZE = 20;

export default function MessagesPage() {
    const params = useParams();
    const router = useRouter();
    const user = useUser((state) => state.user);
    const {
        conversations,
        setConversations,
        appendConversations,
        setUnreadCounts,
        setIsLoading,
        isLoading,
    } = useDmStore();
    const [showNewConversation, setShowNewConversation] = useState(false);
    const [page, setPage] = useState(0);
    const [hasMore, setHasMore] = useState(true);
    const [isLoadingMore, setIsLoadingMore] = useState(false);
    const { t } = useT("messages");

    useEffect(() => {
        if (!user) return;

        const fetchData = async () => {
            setIsLoading(true);
            setPage(0);
            setHasMore(true);
            try {
                const [convRes, unreadRes] = await Promise.all([
                    GetConversations(0, CONVERSATIONS_PAGE_SIZE),
                    GetUnreadCounts(),
                ]);

                if (convRes.data) {
                    setConversations(convRes.data);
                    const total = convRes.meta?.total ?? 0;
                    setHasMore(convRes.data.length < total);
                }
                if (unreadRes.data) {
                    setUnreadCounts(unreadRes.data);
                }
            } finally {
                setIsLoading(false);
            }
        };

        fetchData();
    }, [user, setConversations, setUnreadCounts, setIsLoading]);

    const loadMore = useCallback(async () => {
        if (!user || !hasMore || isLoadingMore) return;
        const nextPage = page + 1;
        setIsLoadingMore(true);
        try {
            const convRes = await GetConversations(
                nextPage,
                CONVERSATIONS_PAGE_SIZE,
            );
            if (convRes.data && convRes.data.length > 0) {
                appendConversations(convRes.data);
                setPage(nextPage);
                const total = convRes.meta?.total ?? 0;
                const newLength = conversations.length + convRes.data.length;
                setHasMore(
                    total > 0
                        ? newLength < total
                        : convRes.data.length >= CONVERSATIONS_PAGE_SIZE,
                );
            } else {
                setHasMore(false);
            }
        } catch {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setIsLoadingMore(false);
        }
    }, [
        user,
        hasMore,
        isLoadingMore,
        page,
        conversations.length,
        appendConversations,
        t,
    ]);

    if (!user) {
        return (
            <div className="flex h-full w-full items-center justify-center">
                <p className="text-muted-foreground">{t("login_required")}</p>
            </div>
        );
    }

    return (
        <div className="flex h-full w-full">
            <div className="flex h-full w-full flex-col md:w-80 md:border-r">
                <div className="flex items-center justify-between gap-2 border-b p-4">
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => router.push(`/${params.lng as string}`)}
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
                    hasMore={hasMore}
                    isLoadingMore={isLoadingMore}
                    onLoadMore={loadMore}
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
    );
}
