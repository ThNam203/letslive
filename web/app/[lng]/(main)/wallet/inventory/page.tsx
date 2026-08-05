"use client";

import { useCallback, useEffect, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMyInventory } from "@/lib/api/gift";
import { GetShopItems } from "@/lib/api/shop";
import { ShopItem, UserInventory } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import IconLoader from "@/components/icons/loader";
import SendGiftDialog from "./send-gift-dialog";

export default function InventoryPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const [items, setItems] = useState<UserInventory[]>([]);
    const [itemsById, setItemsById] = useState<Record<string, ShopItem>>({});
    const [isLoading, setIsLoading] = useState(true);
    const [sendingItem, setSendingItem] = useState<UserInventory | null>(null);

    const fetchInventory = useCallback(async () => {
        if (!user) return;
        setIsLoading(true);
        try {
            const res = await GetMyInventory();
            if (res.success && res.data) {
                setItems(res.data);
            } else {
                toast.error(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                });
            }
        } catch (_) {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setIsLoading(false);
        }
    }, [user, t]);

    useEffect(() => {
        fetchInventory();
    }, [fetchInventory]);

    useEffect(() => {
        GetShopItems().then((res) => {
            if (res.success && res.data) {
                setItemsById(
                    Object.fromEntries(res.data.map((i) => [i.id, i])),
                );
            }
        });
    }, []);

    if (isLoading) {
        return (
            <div className="flex justify-center py-20">
                <IconLoader />
            </div>
        );
    }

    return (
        <>
            <section>
                <h2 className="text-foreground mb-4 text-xl font-semibold">
                    {t("shop:inventory.page_title")}
                </h2>

                {items.length === 0 ? (
                    <p className="text-muted-foreground py-8 text-center text-sm">
                        {t("shop:inventory.empty")}
                    </p>
                ) : (
                    <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
                        {items.map((item) => {
                            const shopItem = itemsById[item.shopItemId];
                            const name = shopItem?.name ?? t("shop:shop.unknown_item");
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
                                    <Button
                                        size="sm"
                                        variant="outline"
                                        className="w-full"
                                        onClick={() => setSendingItem(item)}
                                    >
                                        {t("shop:inventory.send_gift")}
                                    </Button>
                                </div>
                            );
                        })}
                    </div>
                )}
            </section>

            {sendingItem && (
                <SendGiftDialog
                    open={sendingItem !== null}
                    onClose={() => setSendingItem(null)}
                    shopItemId={sendingItem.shopItemId}
                    itemName={
                        itemsById[sendingItem.shopItemId]?.name ??
                        t("shop:shop.unknown_item")
                    }
                    onSent={fetchInventory}
                />
            )}
        </>
    );
}
