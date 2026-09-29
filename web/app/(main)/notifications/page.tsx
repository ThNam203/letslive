import RequireAuth from "@/components/wrappers/RequireAuth";
import { NotificationLoading } from "@/components/notification/notification-loading";
import { myGetT } from "@/lib/i18n";
import NotificationsView from "./_components/notifications-view";

export default async function NotificationsPage() {
    const { t } = await myGetT("notification");

    return (
        <RequireAuth
            fallback={
                <div className="mx-auto w-full px-4 py-6">
                    <NotificationLoading
                        message={t("loading")}
                        variant="full"
                        rows={6}
                    />
                </div>
            }
        >
            <div className="small-scrollbar h-full min-h-0 overflow-auto">
                <div className="mx-auto w-full px-4 py-6">
                    <NotificationsView
                        heading={
                            <h1 className="text-foreground text-xl font-semibold">
                                {t("title")}
                            </h1>
                        }
                    />
                </div>
            </div>
        </RequireAuth>
    );
}
