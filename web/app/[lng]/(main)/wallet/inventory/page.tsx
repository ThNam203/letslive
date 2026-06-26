"use client";

import { useCallback, useEffect, useState } from "react";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import useUser from "@/hooks/user";
import { GetMyInventory } from "@/lib/api/gift";
import { UserInventory } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function InventoryPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const user = useUser((s) => s.user);
    const [items, setItems] = useState<UserInventory[]>([]);
    const [isLoading, setIsLoading] = useState(true);

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
                        {items.map((item) => (
                            <div
                                key={item.id}
                                className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4"
                            >
                                <p className="text-muted-foreground text-xs">
                                    {item.shopItemId}
                                </p>
                                <Badge variant="secondary">
                                    {t("shop:inventory.quantity_label", {
                                        quantity: item.quantity,
                                    })}
                                </Badge>
                            </div>
                        ))}
                    </div>
                )}
            </section>
        </>
    );
}
