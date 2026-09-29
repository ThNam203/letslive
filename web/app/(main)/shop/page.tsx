import { myGetT } from "@/lib/i18n";
import ShopItemGrid from "./_components/shop-item-grid";

export default async function ShopPage() {
    const { t } = await myGetT("shop");

    return (
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("shop.page_title")}
            </h1>
            <ShopItemGrid />
        </div>
    );
}
