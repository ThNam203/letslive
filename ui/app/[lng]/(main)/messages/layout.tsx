"use client";

import useDmWebSocket from "@/hooks/use-dm-websocket";

export default function MessagesLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    // Initialize the DM WebSocket connection for all messages pages
    useDmWebSocket();

    return (
        <div className="flex h-full w-full overflow-hidden">{children}</div>
    );
}
