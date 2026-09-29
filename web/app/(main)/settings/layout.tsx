import TabbedPageLayout from "@/components/navigation/tabbed-page-layout";
import { myGetT } from "@/lib/i18n";
import SettingsLoading from "./loading";

export default async function SettingsLayout({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const { t } = await myGetT("settings");

    const tabs = [
        { name: t("navigation.profile"), href: "/settings/profile" },
        { name: t("navigation.security"), href: "/settings/security" },
        { name: t("navigation.stream"), href: "/settings/stream" },
        {
            name: t("navigation.chat_commands"),
            href: "/settings/chat-commands",
        },
        { name: t("navigation.vods"), href: "/settings/vods" },
        { name: t("navigation.upload"), href: "/settings/upload" },
    ];

    return (
        <TabbedPageLayout
            title={t("page_title")}
            tabs={tabs}
            tabClassName="w-fit px-4"
            fallback={<SettingsLoading />}
        >
            {children}
        </TabbedPageLayout>
    );
}
