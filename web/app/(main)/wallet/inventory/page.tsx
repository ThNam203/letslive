import { myGetT } from "@/lib/i18n";
import InventoryGrid from "./_components/inventory-grid";

export default async function InventoryPage() {
    const { t } = await myGetT("shop");

    return (
        <section>
            <h2 className="text-foreground mb-4 text-xl font-semibold">
                {t("inventory.page_title")}
            </h2>
            <InventoryGrid />
        </section>
    );
}
