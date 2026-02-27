"use client";

import { useEffect, useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import useUser from "@/hooks/user";
import useDmStore from "@/hooks/use-dm-store";
import { GetUnreadCounts } from "@/lib/api/dm";
import IconMessage from "@/components/icons/message";

const POLL_INTERVAL_MS = 30000; // 30 seconds

export default function MessagesIcon() {
    const params = useParams();
    const lng = (params?.lng as string) ?? "en";
    const user = useUser((state) => state.user);
    const { unreadCounts, setUnreadCounts } = useDmStore();

    const totalUnread = Object.values(unreadCounts).reduce(
        (sum, count) => sum + count,
        0,
    );

    const fetchUnreadCounts = useCallback(async () => {
        if (!user) return;
        try {
            const res = await GetUnreadCounts();
            if (res.success && res.data) {
                setUnreadCounts(res.data);
            }
        } catch {
            // silently ignore
        }
    }, [user, setUnreadCounts]);

    useEffect(() => {
        if (!user) return;
        fetchUnreadCounts();
        const interval = setInterval(() => {
            if (!document.hidden) fetchUnreadCounts();
        }, POLL_INTERVAL_MS);
        return () => clearInterval(interval);
    }, [user, fetchUnreadCounts]);

    if (!user) return null;

    return (
        <Link
            href={`/${lng}/messages`}
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
        >
            <IconMessage className="size-5" />
            {totalUnread > 0 && (
                <span className="bg-destructive absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white">
                    {totalUnread > 99 ? "99+" : totalUnread}
                </span>
            )}
        </Link>
    );
}
