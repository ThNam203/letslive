"use client";

import { useEffect, useRef, useCallback, useState } from "react";
import GLOBAL from "@/global";
import useDmStore from "./use-dm-store";
import useUser from "./user";
import {
    type DmWsClientEvent,
    type DmWsServerEvent,
    DmServerEventType,
} from "@/types/dm";
import { toast } from "@/components/utils/toast";
import useT from "./use-translation";

const MAX_RECONNECT_DELAY = 30000;
const INITIAL_RECONNECT_DELAY = 1000;
const TYPING_INDICATOR_TIMEOUT_MS = 5000;

export default function useDmWebSocket() {
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
        null,
    );
    const reconnectDelayRef = useRef(INITIAL_RECONNECT_DELAY);
    const isConnectingRef = useRef(false);
    const typingTimeoutsRef = useRef<
        Map<string, ReturnType<typeof setTimeout>>
    >(new Map());

    const [isConnected, setIsConnected] = useState(false);
    const user = useUser((state) => state.user);
    const {
        addMessage,
        updateMessage,
        removeMessage,
        setTypingUser,
        removeTypingUser,
        setUserOnline,
        setUserOffline,
        incrementUnread,
        updateConversation,
    } = useDmStore();
    const { t } = useT("api-response");

    const handleServerEvent = useCallback(
        (event: DmWsServerEvent) => {
            switch (event.type) {
                case DmServerEventType.NEW_MESSAGE:
                    addMessage(event.conversationId, event.message);
                    {
                        const activeId =
                            useDmStore.getState().activeConversationId;
                        if (event.conversationId !== activeId) {
                            incrementUnread(event.conversationId);
                        }
                    }
                    updateConversation(event.conversationId, {
                        lastMessage: {
                            _id: event.message._id,
                            senderId: event.message.senderId,
                            senderUsername: event.message.senderUsername,
                            text: event.message.text.substring(0, 100),
                            createdAt: event.message.createdAt,
                        },
                        updatedAt: event.message.createdAt,
                    });
                    break;

                case DmServerEventType.MESSAGE_EDITED:
                    updateMessage(event.conversationId, event.messageId, {
                        text: event.newText,
                        updatedAt: event.updatedAt,
                    });
                    break;

                case DmServerEventType.MESSAGE_DELETED:
                    removeMessage(event.conversationId, event.messageId);
                    break;

                case DmServerEventType.USER_TYPING:
                    setTypingUser(event.conversationId, event.username);
                    {
                        const key = `${event.conversationId}:${event.username}`;
                        const existing = typingTimeoutsRef.current.get(key);
                        if (existing) clearTimeout(existing);
                        const timeout = setTimeout(() => {
                            removeTypingUser(
                                event.conversationId,
                                event.username,
                            );
                            typingTimeoutsRef.current.delete(key);
                        }, TYPING_INDICATOR_TIMEOUT_MS);
                        typingTimeoutsRef.current.set(key, timeout);
                    }
                    break;

                case DmServerEventType.USER_STOPPED_TYPING:
                    removeTypingUser(event.conversationId, event.username);
                    {
                        const key = `${event.conversationId}:${event.username}`;
                        const existing = typingTimeoutsRef.current.get(key);
                        if (existing) {
                            clearTimeout(existing);
                            typingTimeoutsRef.current.delete(key);
                        }
                    }
                    break;

                case DmServerEventType.READ_RECEIPT:
                    break;

                case DmServerEventType.USER_ONLINE:
                    setUserOnline(event.userId);
                    break;

                case DmServerEventType.USER_OFFLINE:
                    setUserOffline(event.userId);
                    break;

                case DmServerEventType.CONVERSATION_UPDATED:
                    updateConversation(event.conversationId, event.update);
                    break;

                case DmServerEventType.SEND_FAILED:
                    toast.error(
                        t(event.key) ||
                            event.message ||
                            "Failed to send message",
                    );
                    break;
            }
        },
        [
            addMessage,
            updateMessage,
            removeMessage,
            setTypingUser,
            removeTypingUser,
            setUserOnline,
            setUserOffline,
            incrementUnread,
            updateConversation,
            t,
        ],
    );

    const connect = useCallback(() => {
        if (!user || isConnectingRef.current) return;
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN)
            return;

        isConnectingRef.current = true;
        setIsConnected(false);

        const ws = new WebSocket(GLOBAL.DM_WS_SERVER_URL);

        ws.onopen = () => {
            wsRef.current = ws;
            reconnectDelayRef.current = INITIAL_RECONNECT_DELAY;
            isConnectingRef.current = false;
            setIsConnected(true);
        };

        ws.onmessage = (event) => {
            try {
                const data: DmWsServerEvent = JSON.parse(event.data);
                handleServerEvent(data);
            } catch {
                // ignore malformed messages
            }
        };

        ws.onclose = () => {
            wsRef.current = null;
            isConnectingRef.current = false;
            setIsConnected(false);
            reconnectTimeoutRef.current = setTimeout(() => {
                reconnectDelayRef.current = Math.min(
                    reconnectDelayRef.current * 2,
                    MAX_RECONNECT_DELAY,
                );
                connect();
            }, reconnectDelayRef.current);
        };

        ws.onerror = () => {
            isConnectingRef.current = false;
        };
    }, [user, handleServerEvent]);

    const send = useCallback((event: DmWsClientEvent) => {
        if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify(event));
        }
    }, []);

    const disconnect = useCallback(() => {
        typingTimeoutsRef.current.forEach((t) => clearTimeout(t));
        typingTimeoutsRef.current.clear();
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
        setIsConnected(false);
    }, []);

    useEffect(() => {
        if (user) {
            connect();
        }
        return () => {
            disconnect();
        };
    }, [user, connect, disconnect]);

    return { send, disconnect, isConnected };
}
