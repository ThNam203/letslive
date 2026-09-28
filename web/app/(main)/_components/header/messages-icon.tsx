"use client";

import Link from "next/link";
import useUser from "@/hooks/user";
import IconMessage from "@/components/icons/message";
import { useDmUnreadCounts } from "@/hooks/queries/use-dm-unread-counts";

export default function MessagesIcon() {
    const user = useUser((state) => state.user);
    const { data: unreadCounts = {} } = useDmUnreadCounts(!!user);

    const totalUnread = Object.values(unreadCounts).reduce(
        (sum, count) => sum + count,
        0,
    );

    return (
        <Link
            href="/messages"
            className="hover:bg-muted relative cursor-pointer rounded-md p-1.5 transition-colors"
        >
            <IconMessage className="size-5" />
            {user && totalUnread > 0 && (
                <span className="bg-destructive absolute -top-0.5 -right-0.5 flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-bold text-white">
                    {totalUnread > 99 ? "99+" : totalUnread}
                </span>
            )}
        </Link>
    );
}
