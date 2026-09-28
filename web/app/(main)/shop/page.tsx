"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { CreatePurchase } from "@/lib/api/shop";
import { unwrapResponse } from "@/lib/api/api-error";
import { ShopItem } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import { useShopItems } from "@/hooks/queries/use-shop-items";
import { WALLET_BALANCE_QUERY_KEY } from "@/hooks/queries/use-wallet";
import { INVENTORY_QUERY_KEY } from "@/hooks/queries/use-inventory";

export default function ShopPage() {
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
            queryClient.invalidateQueries({ queryKey: WALLET_BALANCE_QUERY_KEY });
            queryClient.invalidateQueries({ queryKey: INVENTORY_QUERY_KEY });
        },
    });

    const handleBuy = (item: ShopItem) => {
        if (!user) return;
        buyMutation.mutate(item);
    };

    if (isLoading) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("shop:shop.page_title")}
            </h1>

            {items.length === 0 ? (
                <p className="text-muted-foreground text-center py-16">
                    {t("shop:shop.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {items.map((item) => {
                        const isBuying =
                            buyMutation.isPending &&
                            buyMutation.variables?.id === item.id;
                        return (
                            <div
                                key={item.id}
                                className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4 transition-shadow hover:shadow-md"
                            >
                                <div className="relative h-24 w-24">
                                    <Image
                                        src={item.imageUrl}
                                        alt={item.name}
                                        fill
                                        className="object-contain"
                                        unoptimized
                                    />
                                </div>
                                <p className="text-foreground text-center text-sm font-semibold">
                                    {item.name}
                                </p>
                                {item.description && (
                                    <p className="text-muted-foreground line-clamp-2 text-center text-xs">
                                        {item.description}
                                    </p>
                                )}
                                <Badge variant="secondary">
                                    {t("shop:shop.price_label", {
                                        price: item.price,
                                    })}
                                </Badge>
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
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
