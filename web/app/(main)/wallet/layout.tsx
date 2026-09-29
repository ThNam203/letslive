import TabbedPageLayout from "@/components/navigation/tabbed-page-layout";
import { myGetT } from "@/lib/i18n";
import WalletLoading from "./loading";

export default async function WalletLayout({
    children,
}: Readonly<{ children: React.ReactNode }>) {
    const { t } = await myGetT("wallet");

    const tabs = [
        { name: t("navigation.overview"), href: "/wallet/overview" },
        { name: t("navigation.transactions"), href: "/wallet/transactions" },
        { name: t("navigation.deposit"), href: "/wallet/deposit" },
        { name: t("navigation.inventory"), href: "/wallet/inventory" },
    ];

    return (
        <TabbedPageLayout
            title={t("page_title")}
            tabs={tabs}
            tabClassName="w-28"
            fallback={<WalletLoading />}
        >
            {children}
        </TabbedPageLayout>
    );
}
