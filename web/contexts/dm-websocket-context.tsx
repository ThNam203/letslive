"use client";

import { createContext, useContext, useMemo, type ReactNode } from "react";
import type { DmWsClientEvent } from "@/types/dm";
import useDmWebSocket from "@/hooks/use-dm-websocket";

export type DmWebSocketContextValue = {
    send: (event: DmWsClientEvent) => void;
    isConnected: boolean;
};

const DmWebSocketContext = createContext<DmWebSocketContextValue | null>(null);

export function DmWebSocketProvider({ children }: { children: ReactNode }) {
    const { send, isConnected } = useDmWebSocket();
    const value = useMemo(() => ({ send, isConnected }), [send, isConnected]);
    return (
        <DmWebSocketContext.Provider value={value}>
            {children}
        </DmWebSocketContext.Provider>
    );
}

export function useDmWebSocketContext(): DmWebSocketContextValue {
    const ctx = useContext(DmWebSocketContext);
    if (ctx == null) {
        throw new Error(
            "useDmWebSocketContext must be used within DmWebSocketProvider",
        );
    }
    return ctx;
}
