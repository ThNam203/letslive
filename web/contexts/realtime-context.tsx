"use client";

import {
    createContext,
    useCallback,
    useContext,
    useEffect,
    useMemo,
    useRef,
    type ReactNode,
} from "react";
import GLOBAL from "@/global";
import useUser from "@/hooks/user";
import { parseRealtimeFrame } from "@/lib/realtime/parse-frame";
import { hasAccessToken, refreshToken } from "@/utils/fetchClient";
import type { RealtimeEventFrame } from "@/types/realtime";
import {
    REALTIME_RECONNECT_INITIAL_DELAY_MS,
    REALTIME_RECONNECT_MAX_DELAY_MS,
} from "@/constant/realtime";

type EventHandler = (frame: RealtimeEventFrame) => void;
type Unsubscribe = () => void;

export type RealtimeContextValue = {
    onEvent: (type: string, handler: EventHandler) => Unsubscribe;
    onReconnect: (handler: () => void) => Unsubscribe;
};

const RealtimeContext = createContext<RealtimeContextValue | null>(null);

export function RealtimeProvider({ children }: { children: ReactNode }) {
    const userId = useUser((state) => state.user?.id ?? null);
    const eventHandlersRef = useRef(new Map<string, Set<EventHandler>>());
    const reconnectHandlersRef = useRef(new Set<() => void>());

    // userId is a dependency so login/logout reopens the socket with the new
    // cookie identity
    useEffect(() => {
        let disposed = false;
        let hasOpened = false;
        let delay = REALTIME_RECONNECT_INITIAL_DELAY_MS;
        let socket: WebSocket | null = null;
        let retryTimer: ReturnType<typeof setTimeout> | null = null;

        const connect = async () => {
            // the access cookie lives only as long as the token; without a
            // fresh one a logged-in user's socket would silently be anonymous
            if (userId !== null && !hasAccessToken()) {
                try {
                    await refreshToken();
                } catch (err: unknown) {
                    console.error("[realtime] token refresh failed:", err);
                }
                if (disposed) return;
            }

            const ws = new WebSocket(GLOBAL.REALTIME_URL);
            socket = ws;

            ws.onopen = () => {
                delay = REALTIME_RECONNECT_INITIAL_DELAY_MS;
                if (hasOpened) {
                    reconnectHandlersRef.current.forEach((handler) =>
                        handler(),
                    );
                }
                hasOpened = true;
            };

            ws.onmessage = (message: MessageEvent<unknown>) => {
                const frame = parseRealtimeFrame(message.data);
                if (frame?.op === "event") {
                    eventHandlersRef.current
                        .get(frame.type)
                        ?.forEach((handler) => handler(frame));
                } else if (frame?.op === "error") {
                    console.error("[realtime] server error:", frame.code);
                }
            };

            ws.onclose = () => {
                if (disposed || socket !== ws) return;
                retryTimer = setTimeout(() => void connect(), delay);
                delay = Math.min(delay * 2, REALTIME_RECONNECT_MAX_DELAY_MS);
            };
        };

        void connect();

        return () => {
            disposed = true;
            if (retryTimer) clearTimeout(retryTimer);
            socket?.close();
        };
    }, [userId]);

    const onEvent = useCallback(
        (type: string, handler: EventHandler): Unsubscribe => {
            const handlers = eventHandlersRef.current;
            const forType = handlers.get(type) ?? new Set<EventHandler>();
            forType.add(handler);
            handlers.set(type, forType);
            return () => {
                forType.delete(handler);
                if (forType.size === 0 && handlers.get(type) === forType) {
                    handlers.delete(type);
                }
            };
        },
        [],
    );

    const onReconnect = useCallback((handler: () => void): Unsubscribe => {
        reconnectHandlersRef.current.add(handler);
        return () => {
            reconnectHandlersRef.current.delete(handler);
        };
    }, []);

    const value = useMemo(
        () => ({ onEvent, onReconnect }),
        [onEvent, onReconnect],
    );

    return (
        <RealtimeContext.Provider value={value}>
            {children}
        </RealtimeContext.Provider>
    );
}

export function useRealtime(): RealtimeContextValue {
    const context = useContext(RealtimeContext);
    if (!context) {
        throw new Error("useRealtime must be used within RealtimeProvider");
    }
    return context;
}
