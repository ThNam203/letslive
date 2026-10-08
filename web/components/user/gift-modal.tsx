"use client";

import { useEffect, useRef, useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import Image from "next/image";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { CreatePurchase } from "@/lib/api/shop";
import { unwrapResponse } from "@/lib/api/api-error";
import { ShopItem } from "@/types/shop";
import { useShopItems } from "@/hooks/queries/use-shop-items";
import { WALLET_BALANCE_QUERY_KEY } from "@/hooks/queries/use-wallet";
import ShopItemCard from "@/components/shop/shop-item-card";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import { ItemGridSkeleton } from "@/components/skeletons/item-grid-skeleton";

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
    const queryClient = useQueryClient();
    const { data: items = [], isLoading: isLoadingItems } = useShopItems({
        enabled: open,
    });
    const [animationUrl, setAnimationUrl] = useState<string | null>(null);
    const animationTimerRef = useRef<ReturnType<typeof setTimeout> | null>(
        null,
    );

    useEffect(() => {
        return () => {
            if (animationTimerRef.current)
                clearTimeout(animationTimerRef.current);
        };
    }, []);

    const dismissAnimation = () => {
        if (animationTimerRef.current) clearTimeout(animationTimerRef.current);
        setAnimationUrl(null);
        onClose();
    };

    const sendGiftMutation = useMutation({
        mutationFn: async (item: ShopItem) =>
            unwrapResponse(
                await CreatePurchase({
                    shopItemId: item.id,
                    quantity: 1,
                    recipientUserId,
                }),
            ),
        onSuccess: (data) => {
            toast.success(t("shop:shop.gift_sent"));
            queryClient.invalidateQueries({
                queryKey: WALLET_BALANCE_QUERY_KEY,
            });
            if (data?.animationUrl) {
                setAnimationUrl(data.animationUrl);
                animationTimerRef.current = setTimeout(dismissAnimation, 3000);
            }
        },
    });

    const handleSend = (item: ShopItem) => {
        sendGiftMutation.mutate(item);
    };

    return (
        <Dialog
            open={open}
            onOpenChange={(v) => {
                if (v) return;
                setAnimationUrl(null);
                onClose();
            }}
        >
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
                    <ItemGridSkeleton
                        count={6}
                        className="grid-cols-3 gap-3 py-2"
                        imageClassName="h-16 w-16"
                    />
                ) : items.length === 0 ? (
                    <p className="text-muted-foreground py-8 text-center text-sm">
                        {t("shop:shop.gift_no_items")}
                    </p>
                ) : (
                    <div className="grid grid-cols-3 gap-3 py-2">
                        {items.map((item) => {
                            const isSending =
                                sendGiftMutation.isPending &&
                                sendGiftMutation.variables?.id === item.id;
                            return (
                                <ShopItemCard
                                    key={item.id}
                                    name={item.name}
                                    imageUrl={item.imageUrl}
                                    badge={t("shop:shop.price_label", {
                                        price: item.price,
                                    })}
                                    size="sm"
                                    onClick={() => handleSend(item)}
                                    disabled={sendGiftMutation.isPending}
                                >
                                    {isSending && (
                                        <span className="text-muted-foreground text-xs">
                                            {t("shop:shop.gift_sending")}
                                        </span>
                                    )}
                                </ShopItemCard>
                            );
                        })}
                    </div>
                )}
            </DialogContent>
        </Dialog>
    );
}
