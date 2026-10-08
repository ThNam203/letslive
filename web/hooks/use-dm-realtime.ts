"use client";

import { useEffect } from "react";
import { type InfiniteData, useQueryClient } from "@tanstack/react-query";
import { useRealtime } from "@/contexts/realtime-context";
import useDmStore from "./use-dm-store";
import useUser from "./user";
import {
    type Conversation,
    type DmWsServerEvent,
    DmServerEventType,
} from "@/types/dm";
import {
    appendDmMessage,
    incrementDmUnread,
    updateConversationInCache,
    updateDmMessageInCache,
} from "@/lib/query/dm-cache";
import { CONVERSATIONS_QUERY_KEY } from "@/hooks/queries/use-conversations";
import { DM_UNREAD_COUNTS_QUERY_KEY } from "@/hooks/queries/use-dm-unread-counts";
import { dmMessagesQueryKey } from "@/hooks/queries/use-dm-messages";
import type { PaginatedPage } from "@/lib/query/paginated";

const TYPING_INDICATOR_TIMEOUT_MS = 5000;

const DM_EVENT_TYPES = Object.values(DmServerEventType);

function isRecord(value: unknown): value is Record<string, unknown> {
    return typeof value === "object" && value !== null;
}

/**
 * Applies DM pushes from the realtime socket to the query cache and the DM
 * store. Mounted once in the header so unread counts and presence stay live
 * on every page, not only under /messages or in the chat dock.
 */
export default function useDmRealtime(enabled: boolean) {
    const { onEvent, onReconnect } = useRealtime();
    const queryClient = useQueryClient();
    const userId = useUser((state) => state.user?.id ?? null);

    useEffect(() => {
        if (!enabled) return;

        const typingTimeouts = new Map<string, ReturnType<typeof setTimeout>>();

        const isConversationCached = (conversationId: string) =>
            queryClient
                .getQueryData<InfiniteData<PaginatedPage<Conversation>>>(
                    CONVERSATIONS_QUERY_KEY,
                )
                ?.pages.some((page) =>
                    page.items.some((c) => c._id === conversationId),
                ) ?? false;

        const handleEvent = (event: DmWsServerEvent) => {
            const store = useDmStore.getState();

            switch (event.type) {
                case DmServerEventType.NEW_MESSAGE: {
                    appendDmMessage(
                        queryClient,
                        event.conversationId,
                        event.message,
                    );
                    // your own message (echoed to your other tabs) is never
                    // unread; the server count excludes it too
                    if (
                        !store.visibleConversationIds[event.conversationId] &&
                        event.message.senderId !== userId
                    ) {
                        incrementDmUnread(queryClient, event.conversationId);
                    }
                    if (!isConversationCached(event.conversationId)) {
                        // a conversation someone just started with you
                        queryClient.invalidateQueries({
                            queryKey: CONVERSATIONS_QUERY_KEY,
                        });
                        break;
                    }
                    updateConversationInCache(
                        queryClient,
                        event.conversationId,
                        {
                            lastMessage: {
                                _id: event.message._id,
                                senderId: event.message.senderId,
                                senderUsername: event.message.senderUsername,
                                text: event.message.text.substring(0, 100),
                                createdAt: event.message.createdAt,
                            },
                            updatedAt: event.message.createdAt,
                        },
                    );
                    break;
                }

                case DmServerEventType.MESSAGE_EDITED:
                    updateDmMessageInCache(
                        queryClient,
                        event.conversationId,
                        event.messageId,
                        { text: event.newText, updatedAt: event.updatedAt },
                    );
                    break;

                case DmServerEventType.MESSAGE_DELETED:
                    updateDmMessageInCache(
                        queryClient,
                        event.conversationId,
                        event.messageId,
                        { isDeleted: true, text: "" },
                    );
                    break;

                case DmServerEventType.USER_TYPING: {
                    store.setTypingUser(event.conversationId, event.username);
                    const key = `${event.conversationId}:${event.username}`;
                    const existing = typingTimeouts.get(key);
                    if (existing) clearTimeout(existing);
                    typingTimeouts.set(
                        key,
                        setTimeout(() => {
                            useDmStore
                                .getState()
                                .removeTypingUser(
                                    event.conversationId,
                                    event.username,
                                );
                            typingTimeouts.delete(key);
                        }, TYPING_INDICATOR_TIMEOUT_MS),
                    );
                    break;
                }

                case DmServerEventType.USER_STOPPED_TYPING: {
                    store.removeTypingUser(
                        event.conversationId,
                        event.username,
                    );
                    const key = `${event.conversationId}:${event.username}`;
                    const existing = typingTimeouts.get(key);
                    if (existing) {
                        clearTimeout(existing);
                        typingTimeouts.delete(key);
                    }
                    break;
                }

                case DmServerEventType.READ_RECEIPT:
                    break;

                case DmServerEventType.USER_ONLINE:
                    store.setUserOnline(event.userId);
                    break;

                case DmServerEventType.USER_OFFLINE:
                    store.setUserOffline(event.userId);
                    break;

                case DmServerEventType.CONVERSATION_UPDATED:
                    updateConversationInCache(
                        queryClient,
                        event.conversationId,
                        event.update,
                    );
                    break;
            }
        };

        const offEvents = DM_EVENT_TYPES.map((type) =>
            onEvent(type, (frame) => {
                if (!isRecord(frame.data)) return;
                handleEvent({ ...frame.data, type } as DmWsServerEvent);
            }),
        );

        // pushes sent while the socket was down are lost; refetch what they
        // would have changed
        const offReconnect = onReconnect(() => {
            queryClient.invalidateQueries({
                queryKey: DM_UNREAD_COUNTS_QUERY_KEY,
            });
            queryClient.invalidateQueries({
                queryKey: CONVERSATIONS_QUERY_KEY,
            });
            const visibleIds = Object.keys(
                useDmStore.getState().visibleConversationIds,
            );
            for (const id of visibleIds) {
                queryClient.invalidateQueries({
                    queryKey: dmMessagesQueryKey(id),
                });
            }
        });

        return () => {
            offEvents.forEach((off) => off());
            offReconnect();
            typingTimeouts.forEach((timeout) => clearTimeout(timeout));
            typingTimeouts.clear();
        };
    }, [enabled, onEvent, onReconnect, queryClient, userId]);
}
