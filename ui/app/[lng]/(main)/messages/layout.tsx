"use client";

import { DmWebSocketProvider } from "@/contexts/dm-websocket-context";

export default function MessagesLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    return (
        <DmWebSocketProvider>
            <div className="flex h-full w-full overflow-hidden">{children}</div>
        </DmWebSocketProvider>
    );
}
