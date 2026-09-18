import { useQuery } from "@tanstack/react-query";
import { GetMessages } from "@/lib/api/chat";
import { GetRoomChatCommands } from "@/lib/api/chat-command";
import { unwrapResponse } from "@/lib/api/api-error";

export function roomMessagesQueryKey(roomId: string) {
    return ["chat", "messages", roomId] as const;
}

export function roomChatCommandsQueryKey(roomId: string) {
    return ["chat", "commands", roomId] as const;
}

/**
 * Backlog for a room, fetched when the panel opens. Live messages arrive over
 * the WebSocket and are held separately by the panel, which drops them when it
 * unmounts — so the backlog is refetched on every mount, or reopening a room
 * would hide everything said since the first fetch.
 *
 * It is not refetched while the panel stays open: the socket is the source of
 * new messages there, and a refetch mid-session would reorder the transcript
 * under the reader.
 */
export function useRoomMessages(roomId: string | undefined) {
    return useQuery({
        queryKey: roomMessagesQueryKey(roomId ?? ""),
        queryFn: async () => (await GetMessages(roomId as string)).messages,
        enabled: Boolean(roomId),
        staleTime: Infinity,
        refetchOnMount: "always",
        refetchOnWindowFocus: false,
    });
}

export function useRoomChatCommands(roomId: string | undefined) {
    return useQuery({
        queryKey: roomChatCommandsQueryKey(roomId ?? ""),
        queryFn: async () =>
            unwrapResponse(await GetRoomChatCommands(roomId as string)),
        enabled: Boolean(roomId),
    });
}
