"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import { ItemGridSkeleton } from "@/components/skeletons/item-grid-skeleton";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { CreatePurchase } from "@/lib/api/shop";
import { unwrapResponse } from "@/lib/api/api-error";
import { ShopItem } from "@/types/shop";
import { Button } from "@/components/ui/button";
import ShopItemCard from "@/components/shop/shop-item-card";
import { useShopItems } from "@/hooks/queries/use-shop-items";
import { WALLET_BALANCE_QUERY_KEY } from "@/hooks/queries/use-wallet";
import { INVENTORY_QUERY_KEY } from "@/hooks/queries/use-inventory";

export default function ShopItemGrid() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const queryClient = useQueryClient();
    const { data: items = [], isLoading } = useShopItems();

    const buyMutation = useMutation({
        mutationFn: async (item: ShopItem) =>
            unwrapResponse(
                await CreatePurchase({ shopItemId: item.id, quantity: 1 }),
            ),
        onSuccess: () => {
            toast.success(t("shop:shop.purchase_success"));
            queryClient.invalidateQueries({
                queryKey: WALLET_BALANCE_QUERY_KEY,
            });
            queryClient.invalidateQueries({ queryKey: INVENTORY_QUERY_KEY });
        },
    });

    const handleBuy = (item: ShopItem) => {
        if (!user) return;
        buyMutation.mutate(item);
    };

    return (
        <>
            {isLoading ? (
                <ItemGridSkeleton
                    count={10}
                    className="grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5"
                    withAction
                />
            ) : items.length === 0 ? (
                <p className="text-muted-foreground py-16 text-center">
                    {t("shop:shop.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {items.map((item) => {
                        const isBuying =
                            buyMutation.isPending &&
                            buyMutation.variables?.id === item.id;
                        return (
                            <ShopItemCard
                                key={item.id}
                                name={item.name}
                                imageUrl={item.imageUrl}
                                description={item.description}
                                badge={t("shop:shop.price_label", {
                                    price: item.price,
                                })}
                            >
                                <Button
                                    size="sm"
                                    className="w-full"
                                    disabled={!user || buyMutation.isPending}
                                    title={
                                        !user
                                            ? t("shop:shop.login_to_buy")
                                            : undefined
                                    }
                                    onClick={() => handleBuy(item)}
                                >
                                    {isBuying
                                        ? t("shop:shop.gift_sending")
                                        : t("shop:shop.buy_button")}
                                </Button>
                            </ShopItemCard>
                        );
                    })}
                </div>
            )}
        </>
    );
}
