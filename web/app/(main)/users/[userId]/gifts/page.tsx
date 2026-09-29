import { myGetT } from "@/lib/i18n";
import GiftGrid from "./_components/gift-grid";

export default async function UserGiftsPage() {
    const { t } = await myGetT("shop");

    return (
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("gifts_received.page_title")}
            </h1>
            <GiftGrid />
        </div>
    );
}
