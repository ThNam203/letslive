"use client";

import Link from "next/link";
import { useState } from "react";
import { usePathname, useRouter } from "next/navigation";
import useUser from "@/hooks/user";
import useT from "@/hooks/use-translation";
import useMediaQuery from "@/hooks/use-media-query";
import IconMessage from "@/components/icons/message";
import { Button } from "@/components/ui/button";
import {
    Popover,
    PopoverContent,
    PopoverTrigger,
} from "@/components/ui/popover";
import { useDmUnreadCounts } from "@/hooks/queries/use-dm-unread-counts";
import { useConversationsInfinite } from "@/hooks/queries/use-conversations";
import useDmRealtime from "@/hooks/use-dm-realtime";
import useChatDockStore from "@/hooks/use-chat-dock-store";
import { flattenPages } from "@/lib/query/paginated";
import { MQ_MAX_SM } from "@/constant/breakpoints";
import type { Conversation } from "@/types/dm";
import ConversationList from "@/app/(main)/messages/_components/conversation-list";
import NewConversationDialog from "@/app/(main)/messages/_components/new-conversation-dialog";

export default function MessagesIcon() {
    const user = useUser((state) => state.user);
    const { data: unreadCounts = {} } = useDmUnreadCounts(!!user);
    useDmRealtime(!!user);

    const router = useRouter();
    const pathname = usePathname();
    const { t } = useT("messages");
    const isSmallScreen = useMediaQuery(MQ_MAX_SM);
    const openChat = useChatDockStore((state) => state.openChat);
    const [isOpen, setIsOpen] = useState(false);
    const [showNewConversation, setShowNewConversation] = useState(false);

    const { data, isLoading, isFetchingNextPage, hasNextPage, fetchNextPage } =
        useConversationsInfinite(!!user && isOpen);
    const conversations = flattenPages(data);

    const totalUnread = Object.values(unreadCounts).reduce(
        (sum, count) => sum + count,
        0,
    );

    // the dock is hidden on small screens and on the messages page itself,
    // so those go to the full conversation page instead
    const showConversation = (conversation: Conversation) => {
        setIsOpen(false);
        if (isSmallScreen || pathname.startsWith("/messages")) {
            router.push(`/messages/${conversation._id}`);
        } else {
            openChat(conversation);
        }
    };

    if (!user) {
        return (
            <Link
                href="/messages"
                className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
            >
                <IconMessage className="size-5" />
            </Link>
        );
    }

    return (
        <>
            <Popover open={isOpen} onOpenChange={setIsOpen}>
                <PopoverTrigger asChild>
                    <button
                        className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
                        aria-label={t("title")}
                    >
                        <IconMessage className="size-5" />
                        {totalUnread > 0 && (
                            <span className="bg-destructive absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white">
                                {totalUnread > 99 ? "99+" : totalUnread}
                            </span>
                        )}
                    </button>
                </PopoverTrigger>
                <PopoverContent
                    className="border-border bg-muted mr-4 flex h-[28rem] w-80 flex-col p-0"
                    align="end"
                >
                    <div className="border-border flex items-center justify-between border-b px-4 py-3">
                        <h3 className="text-foreground text-sm font-semibold">
                            {t("title")}
                        </h3>
                        <Button
                            variant="ghost"
                            size="sm"
                            className="h-7 px-2"
                            onClick={() => {
                                setIsOpen(false);
                                setShowNewConversation(true);
                            }}
                            title={t("new_conversation")}
                            aria-label={t("new_conversation")}
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
                        onSelect={showConversation}
                    />
                </PopoverContent>
            </Popover>

            {showNewConversation && (
                <NewConversationDialog
                    onClose={() => setShowNewConversation(false)}
                    onCreated={showConversation}
                />
            )}
        </>
    );
}
