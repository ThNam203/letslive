"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetUserGiftsReceived } from "@/lib/api/gift";
import { GetShopItems } from "@/lib/api/shop";
import { Gift, ShopItem } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function UserGiftsPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const params = useParams<{ userId: string }>();
    const [gifts, setGifts] = useState<Gift[]>([]);
    const [itemsById, setItemsById] = useState<Record<string, ShopItem>>({});
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchGifts = async () => {
            setIsLoading(true);
            try {
                const res = await GetUserGiftsReceived(params.userId);
                if (res.success && res.data) {
                    setGifts(res.data);
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
        };
        fetchGifts();
    }, [params.userId, t]);

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
        <div className="p-6">
            <h1 className="text-foreground mb-6 text-3xl font-bold">
                {t("shop:gifts_received.page_title")}
            </h1>

            {gifts.length === 0 ? (
                <p className="text-muted-foreground py-16 text-center">
                    {t("shop:gifts_received.empty")}
                </p>
            ) : (
                <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5">
                    {gifts.map((gift) => {
                        const shopItem = itemsById[gift.shopItemId];
                        const name =
                            shopItem?.name ?? t("shop:shop.unknown_item");
                        return (
                            <div
                                key={gift.id}
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
                                    {t("shop:gifts_received.quantity_label", {
                                        quantity: gift.quantity,
                                    })}
                                </Badge>
                                {gift.message && (
                                    <p className="text-muted-foreground line-clamp-2 text-center text-xs italic">
                                        &quot;{gift.message}&quot;
                                    </p>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
