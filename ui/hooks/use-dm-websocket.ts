"use client";

import { useEffect, useRef, useCallback } from "react";
import GLOBAL from "@/global";
import useDmStore from "./use-dm-store";
import useUser from "./user";
import {
    type DmWsClientEvent,
    type DmWsServerEvent,
    DmServerEventType,
} from "@/types/dm";

const MAX_RECONNECT_DELAY = 30000;
const INITIAL_RECONNECT_DELAY = 1000;

export default function useDmWebSocket() {
    const wsRef = useRef<WebSocket | null>(null);
    const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(
        null,
    );
    const reconnectDelayRef = useRef(INITIAL_RECONNECT_DELAY);
    const isConnectingRef = useRef(false);

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
        activeConversationId,
        updateConversation,
    } = useDmStore();

    const handleServerEvent = useCallback(
        (event: DmWsServerEvent) => {
            switch (event.type) {
                case DmServerEventType.NEW_MESSAGE:
                    addMessage(event.conversationId, event.message);
                    // Increment unread if not viewing this conversation
                    if (event.conversationId !== activeConversationId) {
                        incrementUnread(event.conversationId);
                    }
                    // Update conversation's lastMessage
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
                    break;

                case DmServerEventType.USER_STOPPED_TYPING:
                    removeTypingUser(event.conversationId, event.username);
                    break;

                case DmServerEventType.READ_RECEIPT:
                    // Could be used to update read status in message bubbles
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
            activeConversationId,
            updateConversation,
        ],
    );

    const connect = useCallback(() => {
        if (!user || isConnectingRef.current) return;
        if (
            wsRef.current &&
            wsRef.current.readyState === WebSocket.OPEN
        )
            return;

        isConnectingRef.current = true;

        const ws = new WebSocket(GLOBAL.DM_WS_SERVER_URL);

        ws.onopen = () => {
            wsRef.current = ws;
            reconnectDelayRef.current = INITIAL_RECONNECT_DELAY;
            isConnectingRef.current = false;
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
            // Auto-reconnect with exponential backoff
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
        if (reconnectTimeoutRef.current) {
            clearTimeout(reconnectTimeoutRef.current);
            reconnectTimeoutRef.current = null;
        }
        if (wsRef.current) {
            wsRef.current.close();
            wsRef.current = null;
        }
    }, []);

    useEffect(() => {
        if (user) {
            connect();
        }
        return () => {
            disconnect();
        };
    }, [user, connect, disconnect]);

    return { send, disconnect, isConnected: !!wsRef.current };
}
