"use client";

import { useEffect } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useRealtime } from "@/contexts/realtime-context";
import { REALTIME_EVENT } from "@/constant/realtime";
import { NOTIFICATIONS_ROOT_QUERY_KEY } from "@/hooks/queries/use-notifications";

export default function useNotificationRealtime(enabled: boolean) {
    const { onEvent, onReconnect } = useRealtime();
    const queryClient = useQueryClient();

    useEffect(() => {
        if (!enabled) return;

        const refresh = () => {
            queryClient.invalidateQueries({
                queryKey: NOTIFICATIONS_ROOT_QUERY_KEY,
            });
        };
        const offEvent = onEvent(REALTIME_EVENT.NOTIFICATION_CREATED, refresh);
        const offReconnect = onReconnect(refresh);

        return () => {
            offEvent();
            offReconnect();
        };
    }, [enabled, onEvent, onReconnect, queryClient]);
}
