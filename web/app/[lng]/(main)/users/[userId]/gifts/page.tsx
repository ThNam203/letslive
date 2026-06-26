"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetUserGiftsReceived } from "@/lib/api/gift";
import { Gift } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import IconLoader from "@/components/icons/loader";

export default function UserGiftsPage() {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const params = useParams<{ userId: string }>();
    const [gifts, setGifts] = useState<Gift[]>([]);
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
                    {gifts.map((gift) => (
                        <div
                            key={gift.id}
                            className="border-border bg-card flex flex-col items-center gap-2 rounded-xl border p-4"
                        >
                            <p className="text-muted-foreground text-xs">
                                {gift.shopItemId}
                            </p>
                            <Badge variant="secondary">
                                {t("shop:gifts_received.quantity_label", {
                                    quantity: gift.quantity,
                                })}
                            </Badge>
                            {gift.message && (
                                <p className="text-muted-foreground line-clamp-2 text-center text-xs italic">
                                    "{gift.message}"
                                </p>
                            )}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
