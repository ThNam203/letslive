"use client";

import Image from "next/image";
import { ItemGridSkeleton } from "@/components/skeletons/item-grid-skeleton";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { Badge } from "@/components/ui/badge";
import { useShopItems } from "@/hooks/queries/use-shop-items";
import { useMyInventory } from "@/hooks/queries/use-inventory";

export default function InventoryGrid() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const { data: shopItems = [] } = useShopItems();
    const itemsById = Object.fromEntries(shopItems.map((i) => [i.id, i]));
    const { data: items = [], isLoading } = useMyInventory(!!user);

    return (
        <>
            {isLoading ? (
                <ItemGridSkeleton
                    count={8}
                    className="grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4"
                    imageClassName="h-16 w-16"
                />
            ) : items.length === 0 ? (
                <p className="text-muted-foreground py-8 text-center text-sm">
                    {t("shop:inventory.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
                    {items.map((item) => {
                        const shopItem = itemsById[item.shopItemId];
                        const name =
                            shopItem?.name ?? t("shop:shop.unknown_item");
                        return (
                            <div
                                key={item.id}
                                className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4"
                            >
                                {shopItem && (
                                    <div className="relative h-16 w-16">
                                        <Image
                                            src={shopItem.imageUrl}
                                            alt={name}
                                            fill
                                            className="object-contain"
                                            unoptimized
                                        />
                                    </div>
                                )}
                                <p className="text-foreground text-center text-sm font-medium">
                                    {name}
                                </p>
                                <Badge variant="secondary">
                                    {t("shop:inventory.quantity_label", {
                                        quantity: item.quantity,
                                    })}
                                </Badge>
                            </div>
                        );
                    })}
                </div>
            )}
        </>
    );
}
