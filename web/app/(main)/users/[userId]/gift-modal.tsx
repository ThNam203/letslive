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
    const queryClient = useQueryClient();
    const { data: items = [], isLoading: isLoadingItems } = useShopItems({
        enabled: open,
    });
    const [animationUrl, setAnimationUrl] = useState<string | null>(null);
    const animationTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        return () => {
            if (animationTimerRef.current) clearTimeout(animationTimerRef.current);
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
            queryClient.invalidateQueries({ queryKey: WALLET_BALANCE_QUERY_KEY });
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
                            const isSending =
                                sendGiftMutation.isPending &&
                                sendGiftMutation.variables?.id === item.id;
                            return (
                                <button
                                    key={item.id}
                                    onClick={() => handleSend(item)}
                                    disabled={sendGiftMutation.isPending}
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
