"use client";

import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";

type NotificationPageHeaderProps = {
    heading: React.ReactNode;
    hasUnread: boolean;
    onMarkAllAsRead: () => void;
};

export function NotificationPageHeader({
    heading,
    hasUnread,
    onMarkAllAsRead,
}: NotificationPageHeaderProps) {
    const { t } = useT(["notification"]);
    return (
        <div className="mb-6 flex items-center justify-between">
            {heading}
            {hasUnread && (
                <Button
                    variant="ghost"
                    className="cursor-pointer text-sm"
                    onClick={onMarkAllAsRead}
                >
                    {t("mark_all_as_read")}
                </Button>
            )}
        </div>
    );
}
