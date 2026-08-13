"use client";

import { useState } from "react";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { toast } from "@/components/utils/toast";
import useT from "@/hooks/use-translation";
import { SearchUsersByUsername } from "@/lib/api/user";
import { SendGift } from "@/lib/api/gift";
import { unwrapResponse } from "@/lib/api/api-error";
import { WALLET_BALANCE_QUERY_KEY } from "@/hooks/queries/use-wallet";
import { INVENTORY_QUERY_KEY } from "@/hooks/queries/use-inventory";
import { PublicUser } from "@/types/user";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
} from "@/components/ui/dialog";
import IconLoader from "@/components/icons/loader";

type SendGiftDialogProps = {
    open: boolean;
    onClose: () => void;
    shopItemId: string;
    itemName: string;
};

export default function SendGiftDialog({
    open,
    onClose,
    shopItemId,
    itemName,
}: SendGiftDialogProps) {
    const { t } = useT(["shop", "common", "api-response", "fetch-error"]);
    const queryClient = useQueryClient();
    const [query, setQuery] = useState("");
    const [results, setResults] = useState<PublicUser[]>([]);
    const [isSearching, setIsSearching] = useState(false);
    const [recipient, setRecipient] = useState<PublicUser | null>(null);
    const [message, setMessage] = useState("");

    const handleSearch = async (value: string) => {
        setQuery(value);
        if (value.trim().length < 2) {
            setResults([]);
            return;
        }

        setIsSearching(true);
        try {
            const res = await SearchUsersByUsername(value.trim());
            if (res.success && res.data) {
                setResults(res.data);
            }
        } finally {
            setIsSearching(false);
        }
    };

    const reset = () => {
        setQuery("");
        setResults([]);
        setRecipient(null);
        setMessage("");
    };

    const handleClose = () => {
        reset();
        onClose();
    };

    const sendGiftMutation = useMutation({
        mutationFn: async () =>
            unwrapResponse(
                await SendGift({
                    shopItemId,
                    recipientUserId: recipient!.id,
                    message: message.trim() || undefined,
                }),
            ),
        onSuccess: () => {
            toast.success(t("shop:inventory.gift_sent_success"));
            queryClient.invalidateQueries({ queryKey: WALLET_BALANCE_QUERY_KEY });
            queryClient.invalidateQueries({ queryKey: INVENTORY_QUERY_KEY });
            handleClose();
        },
    });

    const handleSend = () => {
        if (!recipient) return;
        sendGiftMutation.mutate();
    };

    return (
        <Dialog open={open} onOpenChange={(v) => !v && handleClose()}>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>
                        {t("shop:inventory.send_gift_title", { item: itemName })}
                    </DialogTitle>
                </DialogHeader>

                {recipient ? (
                    <div className="bg-muted flex items-center gap-3 rounded-lg p-3">
                        <Avatar className="h-8 w-8">
                            {recipient.profilePicture && (
                                <AvatarImage src={recipient.profilePicture} />
                            )}
                            <AvatarFallback>
                                {recipient.username.charAt(0).toUpperCase()}
                            </AvatarFallback>
                        </Avatar>
                        <p className="text-sm font-medium">{recipient.username}</p>
                        <button
                            onClick={() => setRecipient(null)}
                            className="text-muted-foreground ml-auto hover:text-red-500"
                        >
                            &times;
                        </button>
                    </div>
                ) : (
                    <>
                        <Input
                            placeholder={t(
                                "shop:inventory.recipient_search_placeholder",
                            )}
                            value={query}
                            onChange={(e) => handleSearch(e.target.value)}
                        />
                        <div className="max-h-48 overflow-y-auto">
                            {isSearching && (
                                <div className="flex justify-center py-2">
                                    <IconLoader />
                                </div>
                            )}
                            {!isSearching &&
                                query.trim().length >= 2 &&
                                results.length === 0 && (
                                    <p className="text-muted-foreground py-2 text-center text-sm">
                                        {t("shop:inventory.no_recipient_results")}
                                    </p>
                                )}
                            {results.map((r) => (
                                <button
                                    key={r.id}
                                    onClick={() => setRecipient(r)}
                                    className="hover:bg-accent flex w-full items-center gap-3 rounded px-3 py-2"
                                >
                                    <Avatar className="h-8 w-8">
                                        {r.profilePicture && (
                                            <AvatarImage src={r.profilePicture} />
                                        )}
                                        <AvatarFallback>
                                            {r.username.charAt(0).toUpperCase()}
                                        </AvatarFallback>
                                    </Avatar>
                                    <p className="text-sm font-medium">
                                        {r.username}
                                    </p>
                                </button>
                            ))}
                        </div>
                    </>
                )}

                <Input
                    placeholder={t("shop:shop.gift_message_placeholder")}
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                />

                <div className="flex justify-end gap-2">
                    <Button variant="outline" onClick={handleClose}>
                        {t("common:cancel")}
                    </Button>
                    <Button
                        disabled={!recipient || sendGiftMutation.isPending}
                        onClick={handleSend}
                    >
                        {sendGiftMutation.isPending
                            ? t("shop:inventory.sending_gift")
                            : t("shop:inventory.confirm_send")}
                    </Button>
                </div>
            </DialogContent>
        </Dialog>
    );
}
