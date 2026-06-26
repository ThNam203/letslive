"use client";

import { useEffect, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetShopItems } from "@/lib/api/shop";
import { ShopItem } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function ShopPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const [items, setItems] = useState<ShopItem[]>([]);
    const [isLoading, setIsLoading] = useState(true);

    useEffect(() => {
        const fetchItems = async () => {
            setIsLoading(true);
            try {
                const res = await GetShopItems();
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
        };
        fetchItems();
    }, [t]);

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
                    {items.map((item) => (
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
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
