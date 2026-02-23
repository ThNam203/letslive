"use client";

import useT from "@/hooks/use-translation";
import { Button } from "@/components/ui/button";

type NotificationPageHeaderProps = {
    hasUnread: boolean;
    onMarkAllAsRead: () => void;
};

export function NotificationPageHeader({
    hasUnread,
    onMarkAllAsRead,
}: NotificationPageHeaderProps) {
    const { t } = useT(["notification"]);
    return (
        <div className="mb-6 flex items-center justify-between">
            <h1 className="text-xl font-semibold text-foreground">{t("title")}</h1>
            {hasUnread && (
                <Button
                    variant="ghost"
                    className="text-sm"
                    onClick={onMarkAllAsRead}
                >
                    {t("mark_all_as_read")}
                </Button>
            )}
        </div>
    );
}
