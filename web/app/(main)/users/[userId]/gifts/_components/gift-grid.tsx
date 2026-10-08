"use client";

import { useParams } from "next/navigation";
import { ItemGridSkeleton } from "@/components/skeletons/item-grid-skeleton";
import useT from "@/hooks/use-translation";
import ShopItemCard from "@/components/shop/shop-item-card";
import { useShopItems } from "@/hooks/queries/use-shop-items";
import { useUserGiftsReceived } from "@/hooks/queries/use-user-gifts";

export default function GiftGrid() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const params = useParams<{ userId: string }>();
    const { data: shopItems = [] } = useShopItems();
    const itemsById = Object.fromEntries(shopItems.map((i) => [i.id, i]));
    const { data: gifts, isPending } = useUserGiftsReceived(params.userId);

    return (
        <>
            {isPending ? (
                <ItemGridSkeleton
                    count={10}
                    className="grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
                    imageClassName="h-16 w-16"
                />
            ) : (gifts ?? []).length === 0 ? (
                <p className="text-muted-foreground py-16 text-center">
                    {t("shop:gifts_received.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {(gifts ?? []).map((gift) => {
                        const shopItem = itemsById[gift.shopItemId];
                        const name =
                            shopItem?.name ?? t("shop:shop.unknown_item");
                        return (
                            <ShopItemCard
                                key={gift.id}
                                name={name}
                                imageUrl={shopItem?.imageUrl}
                                size="md"
                                badge={t("shop:gifts_received.quantity_label", {
                                    quantity: gift.quantity,
                                })}
                            >
                                {gift.message && (
                                    <p className="text-muted-foreground line-clamp-2 text-center text-xs italic">
                                        &quot;{gift.message}&quot;
                                    </p>
                                )}
                            </ShopItemCard>
                        );
                    })}
                </div>
            )}
        </>
    );
}
