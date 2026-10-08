"use client";

import { useEffect, useMemo, useSyncExternalStore } from "react";
import { usePathname } from "next/navigation";
import { type InfiniteData, useQueryClient } from "@tanstack/react-query";
import useChatDockStore, {
    CHAT_DOCK_MAX_EXPANDED,
} from "@/hooks/use-chat-dock-store";
import useUser from "@/hooks/user";
import useMediaQuery from "@/hooks/use-media-query";
import { useRealtime } from "@/contexts/realtime-context";
import {
    CONVERSATIONS_QUERY_KEY,
    useConversationsInfinite,
} from "@/hooks/queries/use-conversations";
import { useDmUnreadCounts } from "@/hooks/queries/use-dm-unread-counts";
import { GetConversation } from "@/lib/api/dm";
import { flattenPages, type PaginatedPage } from "@/lib/query/paginated";
import { MQ_MAX_SM } from "@/constant/breakpoints";
import {
    type Conversation,
    type DmWsNewMessage,
    DmServerEventType,
} from "@/types/dm";
import { TooltipProvider } from "@/components/ui/tooltip";
import ChatDialog from "./chat-dialog";
import ChatBubble from "./chat-bubble";

// keep in sync with the dialog and bubble sizes in chat-dialog/chat-bubble
const DIALOG_WIDTH_PX = 328;
const GAP_PX = 8;
const BUBBLE_COLUMN_PX = 48;
const EDGE_GUTTER_PX = 16;

function subscribeResize(onChange: () => void) {
    window.addEventListener("resize", onChange);
    return () => window.removeEventListener("resize", onChange);
}

/** How many dialogs fit side by side next to the bubble column. */
function getExpandedCapacity() {
    const available =
        window.innerWidth - 2 * EDGE_GUTTER_PX - BUBBLE_COLUMN_PX - GAP_PX;
    const fits = Math.floor((available + GAP_PX) / (DIALOG_WIDTH_PX + GAP_PX));
    return Math.max(0, Math.min(CHAT_DOCK_MAX_EXPANDED, fits));
}

function isMessagesPage(pathname: string) {
    return pathname === "/messages" || pathname.startsWith("/messages/");
}

/**
 * Chat windows docked at the bottom right, Facebook style. The most recently
 * opened chats are expanded, as many as fit (at most three); the rest, and any
 * the user minimized, are shown as avatar bubbles. A message from someone
 * else pops their chat up unless it is already docked.
 */
export default function ChatDock() {
    const userId = useUser((state) => state.user?.id ?? null);
    const pathname = usePathname();
    const chats = useChatDockStore((state) => state.chats);
    const closeAll = useChatDockStore((state) => state.closeAll);
    const popUpChat = useChatDockStore((state) => state.popUpChat);
    const isSmallScreen = useMediaQuery(MQ_MAX_SM);
    const { onEvent } = useRealtime();
    const queryClient = useQueryClient();
    const capacity = useSyncExternalStore(
        subscribeResize,
        getExpandedCapacity,
        () => 0,
    );

    const hasChats = chats.length > 0;
    const { data: conversationsData } = useConversationsInfinite(
        !!userId && hasChats,
    );
    const { data: unreadCounts = {} } = useDmUnreadCounts(!!userId && hasChats);

    // a different account (or none) must not inherit the previous one's chats
    useEffect(() => {
        closeAll();
    }, [userId, closeAll]);

    // pop a chat up when someone messages you, Facebook style; not where the
    // dock is hidden
    const canPopUp = !!userId && !isSmallScreen && !isMessagesPage(pathname);
    useEffect(() => {
        if (!canPopUp) return;

        return onEvent(DmServerEventType.NEW_MESSAGE, (frame) => {
            const event = frame.data as Partial<DmWsNewMessage> | undefined;
            const conversationId = event?.conversationId;
            if (!conversationId || !event.message) return;
            // your own message, echoed to your other tabs
            if (event.message.senderId === userId) return;
            if (
                useChatDockStore
                    .getState()
                    .chats.some((c) => c.id === conversationId)
            ) {
                return;
            }

            const cached = flattenPages(
                queryClient.getQueryData<
                    InfiniteData<PaginatedPage<Conversation>>
                >(CONVERSATIONS_QUERY_KEY),
            ).find((c) => c._id === conversationId);
            if (cached) {
                popUpChat(cached);
                return;
            }

            // a conversation that is not loaded yet, e.g. one just started
            GetConversation(conversationId)
                .then((res) => {
                    if (res.data) popUpChat(res.data);
                })
                .catch(() => {});
        });
    }, [canPopUp, onEvent, queryClient, popUpChat, userId]);

    const cachedById = useMemo(
        () => new Map(flattenPages(conversationsData).map((c) => [c._id, c])),
        [conversationsData],
    );

    // the messages page shows the full conversation already
    if (!userId || !hasChats || isMessagesPage(pathname)) return null;

    const notMinimized = chats.filter((c) => !c.minimized);
    const expandedIds = new Set(
        notMinimized
            .slice(Math.max(0, notMinimized.length - capacity))
            .map((c) => c.id),
    );
    const expanded = chats.filter((c) => expandedIds.has(c.id));
    const bubbles = chats.filter((c) => !expandedIds.has(c.id));

    return (
        <TooltipProvider delayDuration={300}>
            <div className="pointer-events-none fixed right-0 bottom-0 z-40 hidden items-end gap-2 pr-4 sm:flex">
                {expanded.map((chat) => (
                    <ChatDialog
                        key={chat.id}
                        conversationId={chat.id}
                        snapshot={chat.snapshot}
                    />
                ))}
                {bubbles.length > 0 && (
                    <div className="pointer-events-auto flex flex-col gap-2 pb-4">
                        {bubbles.map((chat) => (
                            <ChatBubble
                                key={chat.id}
                                conversation={
                                    cachedById.get(chat.id) ?? chat.snapshot
                                }
                                unreadCount={unreadCounts[chat.id] ?? 0}
                            />
                        ))}
                    </div>
                )}
            </div>
        </TooltipProvider>
    );
}
