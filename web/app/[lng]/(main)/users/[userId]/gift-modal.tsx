"use client";

import { useEffect, useRef, useState } from "react";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { GetShopItems, CreatePurchase } from "@/lib/api/shop";
import { ShopItem } from "@/types/shop";
import { Badge } from "@/components/ui/badge";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import IconLoader from "@/components/icons/loader";

type GiftModalProps = {
    open: boolean;
    onClose: () => void;
    recipientUserId: string;
    recipientName: string;
};

export default function GiftModal({
    open,
    onClose,
    recipientUserId,
    recipientName,
}: GiftModalProps) {
    const { t } = useT(["shop", "api-response", "fetch-error"]);
    const [items, setItems] = useState<ShopItem[]>([]);
    const [isLoadingItems, setIsLoadingItems] = useState(false);
    const [sendingItemId, setSendingItemId] = useState<string | null>(null);
    const [animationUrl, setAnimationUrl] = useState<string | null>(null);
    const animationTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        return () => {
            if (animationTimerRef.current) clearTimeout(animationTimerRef.current);
        };
    }, []);

    useEffect(() => {
        if (!open) {
            setAnimationUrl(null);
            return;
        }
        const fetchItems = async () => {
            setIsLoadingItems(true);
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
                setIsLoadingItems(false);
            }
        };
        fetchItems();
    }, [open, t]);

    const dismissAnimation = () => {
        if (animationTimerRef.current) clearTimeout(animationTimerRef.current);
        setAnimationUrl(null);
        onClose();
    };

    const handleSend = async (item: ShopItem) => {
        setSendingItemId(item.id);
        try {
            const res = await CreatePurchase({
                shopItemId: item.id,
                quantity: 1,
                recipientUserId,
            });
            if (res.success && res.data) {
                toast.success(t("shop:shop.gift_sent"));
                setAnimationUrl(res.data.animationUrl);
                animationTimerRef.current = setTimeout(dismissAnimation, 3000);
            } else {
                toast.error(t(`api-response:${res.key}`), {
                    toastId: res.requestId,
                });
            }
        } catch (_) {
            toast.error(t("fetch-error:client_fetch_error"));
        } finally {
            setSendingItemId(null);
        }
    };

    return (
        <Dialog open={open} onOpenChange={(v) => !v && onClose()}>
            <DialogContent className="relative max-w-lg">
                {animationUrl && (
                    <div
                        className="absolute inset-0 z-10 flex cursor-pointer items-center justify-center rounded-lg bg-black/80"
                        onClick={dismissAnimation}
                    >
                        <Image
                            src={animationUrl}
                            alt=""
                            width={256}
                            height={256}
                            className="object-contain"
                            unoptimized
                        />
                    </div>
                )}
                <DialogHeader>
                    <DialogTitle>
                        {t("shop:shop.gift_pick_item")} — {recipientName}
                    </DialogTitle>
                </DialogHeader>

                {isLoadingItems ? (
                    <div className="flex justify-center py-8">
                        <IconLoader />
                    </div>
                ) : items.length === 0 ? (
                    <p className="text-muted-foreground py-8 text-center text-sm">
                        {t("shop:shop.gift_no_items")}
                    </p>
                ) : (
                    <div className="grid grid-cols-3 gap-3 py-2">
                        {items.map((item) => {
                            const isSending = sendingItemId === item.id;
                            return (
                                <button
                                    key={item.id}
                                    onClick={() => handleSend(item)}
                                    disabled={sendingItemId !== null}
                                    className="border-border bg-card hover:border-primary flex flex-col items-center gap-1 rounded-lg border p-3 transition-colors disabled:opacity-50"
                                >
                                    <div className="relative h-16 w-16">
                                        <Image
                                            src={item.imageUrl}
                                            alt={item.name}
                                            fill
                                            className="object-contain"
                                            unoptimized
                                        />
                                    </div>
                                    <p className="text-foreground text-center text-xs font-medium">
                                        {item.name}
                                    </p>
                                    <Badge variant="secondary" className="text-xs">
                                        {t("shop:shop.price_label", {
                                            price: item.price,
                                        })}
                                    </Badge>
                                    {isSending && (
                                        <span className="text-muted-foreground text-xs">
                                            {t("shop:shop.gift_sending")}
                                        </span>
                                    )}
                                </button>
                            );
                        })}
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}
